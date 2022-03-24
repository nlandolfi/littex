package main

import (
	"encoding/json"
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
	b := gba.Parse2(ts)
	bs, err = json.MarshalIndent(b, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Fprint(os.Stdout, string(bs))
}
