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

	f, _ := gba.ParseSource2(bs)
	switch *mode {
	case "json":
		bs, err = json.MarshalIndent(f, "", "  ")
		if err != nil {
			log.Fatal(err)
		}

		fmt.Fprint(os.Stdout, string(bs))
	case "tex":
		f.TexWriteTo(os.Stdout)
	case "gba":
		f.WriteTo(os.Stdout)
	}

}
