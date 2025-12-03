package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"

	"github.com/nlandolfi/lit"
)

var inmode = flag.String("i", "", "input format {lit|tex|html|csv}, auto-detected from extension")
var in = flag.String("in", "", "input file or directory (required)")
var outmode = flag.String("o", "", "output format {debug|lit|tex|html|slides|tmpl}")
var out = flag.String("out", "", "output file or directory; if dir input, -out dir is required")
var tmpl = flag.String("tmpl", "text.tmpl", "template file for -o tmpl mode")
var v = flag.Bool("v", false, "print version and exit")
var verbose = flag.Bool("verbose", false, "print file processing info")

// Set using link flags; e.g., -X main.Version=...
var (
	Version   string // e.g. 0.1.0
	GitSHA    string
	BuildDate string
	GoVersion = runtime.Version()
)

func main() {
	flag.Parse()

	if *v {
		fmt.Printf("lit version %s (%s)\n  SHA %s \n  Built at %s\n", Version, GoVersion, GitSHA, BuildDate)
		return
	}

	if *in == "" {
		fmt.Println("usage: lit -in <file|dir> [-out <file|dir>] [-o format]")
		fmt.Println()
		fmt.Println("single file:")
		fmt.Println("  lit -in foo.lit -o html          # output to stdout")
		fmt.Println("  lit -in foo.lit -out foo.html    # output to file")
		fmt.Println()
		fmt.Println("directory (batch):")
		fmt.Println("  lit -in src/ -out dist/ -o html  # compile all .lit files")
		fmt.Println()
		fmt.Println("in directory mode, each .lit file can specify a template via yaml frontmatter:")
		fmt.Println("  <!--yaml")
		fmt.Println("  template: page.tmpl")
		fmt.Println("  title: My Page")
		fmt.Println("  -->")
		fmt.Println()
		fmt.Println("falls back to default.tmpl in input dir, or raw output if no template found.")
		fmt.Println()
		flag.PrintDefaults()
		return
	}

	// check if input is a directory
	fi, err := os.Stat(*in)
	if err != nil {
		log.Fatalf("stat %q: %v", *in, err)
	}
	if fi.IsDir() {
		if *out == "" {
			log.Fatalf("directory mode requires -out <dir>")
		}
		processDirectory(*in, *out)
		return
	}

	if *inmode == "" {
		switch path.Ext(*in) {
		case ".lit":
			*inmode = "lit"
		case ".tex":
			*inmode = "tex"
		case ".html":
			*inmode = "html"
		case ".csv":
			*inmode = "csv"
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
	case "csv":
		n, err = lit.ParseCSV(string(bs))
	default:
		log.Fatalf("unknown input type: %q", *inmode)
	}
	if err != nil {
		log.Fatalf("parsing: %v", err)
	}

	if *outmode == "" && *out != "" {
		switch path.Ext(*out) {
		case ".lit":
			*outmode = "lit"
		case ".tex":
			*outmode = "tex"
		case ".html":
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

	var opts = lit.DefaultWriteOpts
	switch *outmode {
	case "debug":
		lit.WriteDebug(w, n, opts)
	case "", "lit":
		if err := lit.WriteLit(w, n, opts); err != nil {
			log.Fatal(err)
		}
	case "tex":
		lit.WriteTex(w, n, opts)
	case "html":
		lit.WriteHTMLInBody(w, n, opts)
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

func processDirectory(inDir, outDir string) {
	// determine output extension based on -o flag
	outExt := ".html"
	switch *outmode {
	case "tex":
		outExt = ".tex"
	case "lit":
		outExt = ".lit"
	}

	// check for default template
	defaultTmplPath := filepath.Join(inDir, "default.tmpl")
	var defaultTmpl string
	if bs, err := os.ReadFile(defaultTmplPath); err == nil {
		defaultTmpl = string(bs)
	}

	err := filepath.WalkDir(inDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// skip hidden files and directories
		if strings.HasPrefix(d.Name(), ".") || strings.HasPrefix(d.Name(), "_") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// skip template files and non-.lit files
		if d.IsDir() || filepath.Ext(p) == ".tmpl" || filepath.Ext(p) != ".lit" {
			return nil
		}

		// compute relative path and output path
		rel, err := filepath.Rel(inDir, p)
		if err != nil {
			return err
		}
		outPath := filepath.Join(outDir, strings.TrimSuffix(rel, ".lit")+outExt)

		// ensure output directory exists
		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			return err
		}

		// read and parse input file
		bs, err := os.ReadFile(p)
		if err != nil {
			return fmt.Errorf("reading %s: %w", p, err)
		}

		n, err := lit.ParseLit(string(bs))
		if err != nil {
			return fmt.Errorf("parsing %s: %w", p, err)
		}

		// check for yaml frontmatter to find template
		var tmplContent string
		if n.FirstChild != nil && n.FirstChild.Type == lit.YAMLNode {
			if t, ok := n.FirstChild.YAML["template"].(string); ok {
				tmplPath := filepath.Join(inDir, t)
				if bs, err := os.ReadFile(tmplPath); err == nil {
					tmplContent = string(bs)
				} else {
					return fmt.Errorf("reading template %s: %w", tmplPath, err)
				}
			}
		}
		if tmplContent == "" {
			tmplContent = defaultTmpl
		}

		// create output file
		f, err := os.Create(outPath)
		if err != nil {
			return fmt.Errorf("creating %s: %w", outPath, err)
		}
		defer f.Close()

		// write output
		if tmplContent != "" {
			// use template, pass node directly like single-file mode
			t, err := template.New("").Funcs(templateFuncs).Parse(tmplContent)
			if err != nil {
				return fmt.Errorf("parsing template for %s: %w", p, err)
			}
			if err := t.Execute(f, n); err != nil {
				return fmt.Errorf("executing template for %s: %w", p, err)
			}
		} else {
			// no template, output raw
			switch *outmode {
			case "tex":
				lit.WriteTex(f, n, lit.DefaultWriteOpts)
			case "lit":
				lit.WriteLit(f, n, lit.DefaultWriteOpts)
			default:
				lit.WriteHTMLInBody(f, n, lit.DefaultWriteOpts)
			}
		}

		if *verbose {
			log.Printf("%s -> %s", p, outPath)
		}
		return nil
	})

	if err != nil {
		log.Fatalf("processing directory: %v", err)
	}
}

var templateFuncs = template.FuncMap{
	"tex": func(n *lit.Node) string {
		var b bytes.Buffer
		lit.WriteTex(&b, n, &lit.WriteOpts{Prefix: "    ", Indent: ""})
		return b.String()
	},
	"texpi": func(n *lit.Node, pr, in string) string {
		var b bytes.Buffer
		lit.WriteTex(&b, n, &lit.WriteOpts{Prefix: pr, Indent: in})
		return b.String()
	},
	"lit": func(n *lit.Node) string {
		var b bytes.Buffer
		lit.WriteLit(&b, n, &lit.WriteOpts{Prefix: "", Indent: "  "})
		return b.String()
	},
	"html": func(n *lit.Node) string {
		var b bytes.Buffer
		lit.WriteHTML(&b, n, lit.DefaultWriteOpts)
		return b.String()
	},
	"body": func(n *lit.Node) string {
		var b bytes.Buffer
		// skip yaml frontmatter, render rest
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == lit.YAMLNode {
				continue
			}
			lit.WriteHTML(&b, c, lit.DefaultWriteOpts)
		}
		return b.String()
	},
}

func execute(w io.Writer, t string, n *lit.Node) {
	tmpl, err := template.New("").Funcs(templateFuncs).Parse(t)
	if err != nil {
		log.Fatalf("template parsing: %s", err)
	}
	if err = tmpl.Execute(w, n); err != nil {
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
     AND the number of items is correct, this works,
     otherwise it breaks
		 */}}
    \titleslide
    { {{ $tslide.FirstTokenString }} }
    {{- range $tslide.FirstListNode.Kids -}}
      { {{  .FirstTokenString }} }
    {{- end }}
  {{ end }}

  {{- range $slide := slice . 1 -}}
{{ if $slide.IsListItem }}
\slide{ {{ $slide.FirstTokenString }} }{
  {{ range $slide.KidsExcludingTokens }}

  {{- texpi . "  " "  " -}}

{{- end }}
}
{{ else }}
  {{- texpi $slide "  " "  " -}}
{{ end }}
  {{ end }}
{{ end }}
\end{document}
`
