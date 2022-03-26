package main

import (
	"bytes"
	"flag"
	"log"
	"os"
	"text/template"

	"github.com/greatbooksadventure/gba"
)

var in = flag.String("in", "text.gba", "in file")
var mode = flag.String("m", "gba", "mode")
var tmpl = flag.String("tmpl", "text.tmpl", "template")

func main() {
	flag.Parse()
	bs, err := os.ReadFile(*in)
	if err != nil {
		log.Fatal(err)
	}

	n, err := gba.Parse3(string(bs))
	if err != nil {
		log.Fatal(err)
	}

	//log.Print(n)

	switch *mode {
	case "debug":
		gba.WriteDebug(os.Stdout, n, "", "  ")
	case "gba":
		gba.WriteGBA(os.Stdout, n, "", "  ")
	case "tex":
		gba.WriteTex(os.Stdout, n, "", "  ")
	case "slides":
		execute(slidesTemplate, n)
	case "tmpl":
		bs, err := os.ReadFile(*tmpl)
		if err != nil {
			log.Fatal(err)
		}
		execute(string(bs), n)

	}
}

func execute(t string, n *gba.Node) {
	// Create a template, add the function map, and parse the text.
	tmpl, err := template.New("").Funcs(
		template.FuncMap{
			"tex": func(n *gba.Node) string {
				var b bytes.Buffer
				gba.WriteTex(&b, n, "", "  ")
				return b.String()
			},
			"texpi": func(n *gba.Node, pr, in string) string {
				var b bytes.Buffer
				gba.WriteTex(&b, n, pr, in)
				return b.String()
			},
			"gba": func(n *gba.Node) string {
				var b bytes.Buffer
				gba.WriteGBA(&b, n, "", "  ")
				return b.String()
			},
		},
	).Parse(t)
	if err != nil {
		log.Fatalf("parsing: %s", err)
	}

	// Run the template to verify the output.
	err = tmpl.Execute(os.Stdout, n)
	if err != nil {
		log.Fatalf("execution: %s", err)
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
