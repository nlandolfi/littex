package main

import (
	"log"
	"os"

	"github.com/greatbooksadventure/gba"
)

func main() {
	bs, err := os.ReadFile("halmos.gba")
	if err != nil {
		log.Fatal(err)
	}

	f, _ := gba.ParseSource1(bs)

	/*
		bs, err = json.MarshalIndent(f, "", "  ")
		if err != nil {
			log.Fatal(err)
		}

		fmt.Fprint(os.Stdout, string(bs))
	*/

	f.WriteTo(os.Stdout)
}
