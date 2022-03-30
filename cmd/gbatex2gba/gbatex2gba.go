package main

import (
	"flag"
	"log"
	"os"

	"github.com/greatbooksadventure/lit"
)

var in = flag.String("in", "text.gba", "in file")

func main() {
	flag.Parse()
	bs, err := os.ReadFile(*in)
	if err != nil {
		log.Fatal(err)
	}
	s := string(bs)

	n, err := lit.ParseTex(s)
	if err != nil {
		log.Fatal(err)
	}

	lit.WriteLit(os.Stdout, n, "", "  ")
}
