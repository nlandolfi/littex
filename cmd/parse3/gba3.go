package main

import (
	"flag"
	"log"
	"os"

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
	case "tmpl":
		// load template
		// load in some helper functions
		// execute it
	}
}
