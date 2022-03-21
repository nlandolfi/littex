package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/greatbooksadventure/gba"
)

func main() {
	bs, err := os.ReadFile("text.gba")
	if err != nil {
		log.Fatal(err)
	}

	f, _ := gba.ParseSource1(bs)

	bs, err = json.MarshalIndent(f, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Fprint(os.Stdout, string(bs))
}
