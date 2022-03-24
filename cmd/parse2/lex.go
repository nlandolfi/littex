package main

import (
	"flag"
	"fmt"
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

	for _, t := range ts {
		fmt.Println(t.String())
	}
}
