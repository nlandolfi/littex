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
var mode = flag.String("m", "json", "mode")

func main() {
	flag.Parse()
	bs, err := os.ReadFile(*in)
	if err != nil {
		log.Fatal(err)
	}

	s := string(bs)
	ts := gba.Lex2(s)
	b := gba.Parse2(ts)

	switch *mode {
	case "json":
		bs, err = json.MarshalIndent(b, "", "  ")
		if err != nil {
			log.Fatal(err)
		}

		fmt.Fprint(os.Stdout, string(bs))
	case "tex":
		b.TexWriteTo(os.Stdout, "")
	case "gba":
		b.WriteTo(os.Stdout, "")
	}

}
