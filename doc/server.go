package main

import (
	"flag"
	"log"
	"net/http"
)

var (
	staticDir = flag.String("static-dir", "/static", "directory of static files")
	address   = flag.String("address", ":8080", "address to listen on")
)

func main() {
	flag.Parse()

	m := http.NewServeMux()
	m.Handle("/", http.FileServer(http.Dir(*staticDir)))
	log.Fatal(http.ListenAndServe(*address, m))
}
