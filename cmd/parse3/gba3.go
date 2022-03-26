package main

import (
	"flag"
	"log"
	"os"

	"github.com/greatbooksadventure/gba"
)

var in = flag.String("in", "text.gba", "in file")
var mode = flag.String("m", "gba", "mode")

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
	/* doesn't work cause node points have cycles
	case "json":
		bs, err = json.MarshalIndent(n, "", "  ")
		if err != nil {
			log.Fatal(err)
		}

		fmt.Fprint(os.Stdout, string(bs))
	*/
	case "tex":
		panic("d")
	case "debug":
		gba.WriteDebug(os.Stdout, n, "", "  ")
	case "gba":
		gba.WriteGBA(os.Stdout, n, "", "  ")
	}
}
