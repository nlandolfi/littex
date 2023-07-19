package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/nlandolfi/lit"
)

var (
	staticDir = flag.String("static-dir", "/static", "directory of static files")
	address   = flag.String("address", ":8080", "address to listen on")
)

func main() {
	flag.Parse()

	m := http.NewServeMux()
	m.Handle("/", http.FileServer(http.Dir(*staticDir)))
	m.HandleFunc("/lit", litHandler)
	log.Fatal(http.ListenAndServe(*address, m))
}

type LitRequest struct {
	InMode  string
	In      string
	OutMode string
}

func litHandler(w http.ResponseWriter, r *http.Request) {
	var lr LitRequest
	// Decode the JSON object from the request body
	err := json.NewDecoder(r.Body).Decode(&lr)
	if err != nil {
		// Handle the error, e.g., return an HTTP 400 Bad Request status
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var n *lit.Node
	switch lr.InMode {
	case "html":
		n, err = lit.ParseHTML(lr.In)
	case "tex":
		n, err = lit.ParseTex(lr.In)
	case "lit":
		n, err = lit.ParseLit(lr.In)
	case "csv":
		n, err = lit.ParseCSV(lr.In)
	default:
		http.Error(w, fmt.Sprintf("unknown input type: %q", lr.InMode), http.StatusBadRequest)
	}
	if err != nil {
		http.Error(w, fmt.Sprintf("parsing: %v", err), http.StatusInternalServerError)
	}

	var opts = lit.DefaultWriteOpts
	switch lr.OutMode {
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
	default:
		http.Error(w, fmt.Sprintf("unknown output type: %q", lr.OutMode), http.StatusBadRequest)
	}
}
