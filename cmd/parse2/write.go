package main

import (
	"flag"
	"log"
	"os"

	"github.com/greatbooksadventure/gba"
)

var in = flag.String("in", "text.gba", "in file")

func main() {
	flag.Parse()
	bs, err := os.ReadFile(*in)
	if err != nil {
		log.Fatal(err)
	}

	s := string(bs)
	ts := gba.Lex2(s)
	b := gba.Parse2(ts)
	b.WriteTo(os.Stdout, "")
}
