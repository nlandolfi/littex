package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"text/template"

	"github.com/greatbooksadventure/lit"
)

var inmode = flag.String("i", "", "the type of the input file")
var in = flag.String("in", "", "in file, required")
var outmode = flag.String("o", "", "the type of the output file")
var out = flag.String("out", "", "out file, if unset writes to stdout")
var tmpl = flag.String("tmpl", "text.tmpl", "in case -o tmpl, the template file to execute")

func main() {
	flag.Parse()

	if *in == "" {
		fmt.Printf("lit -in <filename>\n")
		return
	}

	if *inmode == "" {
		switch path.Ext(*in) {
		case "lit":
			*inmode = "lit"
		case "tex":
			*inmode = "tex"
		case "html":
			*inmode = "html"
		default:
			*inmode = "lit"
		}
	}

	bs, err := os.ReadFile(*in)
	if err != nil {
		log.Fatalf("reading: %v", err)
	}

	var n *lit.Node
	switch *inmode {
	case "html":
		n, err = lit.ParseHTML(string(bs))
	case "tex":
		n, err = lit.ParseTex(string(bs))
	case "lit":
		n, err = lit.ParseLit(string(bs))
	default:
		log.Fatalf("unknown input type: %q", *inmode)
	}
	if err != nil {
		log.Fatalf("parsing: %v", err)
	}

	if *outmode == "" && *out != "" {
		switch path.Ext(*out) {
		case "lit":
			*outmode = "lit"
		case "tex":
			*outmode = "tex"
		case "html":
			*outmode = "html"
		default:
			*outmode = "lit"
		}
	}

	var w = os.Stdout
	if *out != "" {
		var f *os.File
		f, err = os.Create(*out)
		if err != nil {
			log.Fatalf("creating out file %q: %v", *out, err)
		}
		w = f
		defer f.Close()
	}

	switch *outmode {
	case "debug":
		lit.WriteDebug(w, n, "", "  ")
	case "lit":
		lit.WriteLit(w, n, "", "  ")
	case "tex":
		lit.WriteTex(w, n, "", "  ")
	case "slides":
		execute(w, slidesTemplate, n)
	case "tmpl":
		bs, err := os.ReadFile(*tmpl)
		if err != nil {
			log.Fatalf("reading template file: %v", err)
		}
		execute(w, string(bs), n)
	default:
		log.Fatalf("unknown output type: %q", *outmode)
	}
}

func execute(w io.Writer, t string, n *lit.Node) {
	// Create a template, add the function map, and parse the text.
	tmpl, err := template.New("").Funcs(
		template.FuncMap{
			"tex": func(n *lit.Node) string {
				var b bytes.Buffer
				lit.WriteTex(&b, n, "", "  ")
				return b.String()
			},
			"texpi": func(n *lit.Node, pr, in string) string {
				var b bytes.Buffer
				lit.WriteTex(&b, n, pr, in)
				return b.String()
			},
			"lit": func(n *lit.Node) string {
				var b bytes.Buffer
				lit.WriteLit(&b, n, "", "  ")
				return b.String()
			},
		},
	).Parse(t)
	if err != nil {
		log.Fatalf("template parsing: %s", err)
	}

	// Run the template to verify the output.
	err = tmpl.Execute(w, n)
	if err != nil {
		log.Fatalf("template execution: %s", err)
	}
}

const slidesTemplate = `
\documentclass[9pt]{extarticle}
\input{macros.tex}
\begin{document}
{{ with $slides := .Kids }}
  {{- with $tslide := index . 0 }}
	{{/*
     if we assume the last node of this slide is the list
     AND the number of items is ocrrect, this works,
     otherwise it breaks
		 */}}
    \titleslide
    { {{ $tslide.FirstTokenString }} }
    {{- range $tslide.FirstListNode.Kids -}}
      { {{  .FirstTokenString }} }
    {{- end }}
  {{ end }}

  {{- range $slide := slice . 1 -}}
\slidei{ {{ $slide.FirstTokenString }} }{
{{ range $slide.FirstListNode.Kids -}}

  {{- texpi . "  " "  " -}}

{{- end }}
}
  {{ end }}
{{ end }}
\end{document}
`
