package main

import (
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	bs, err := os.ReadFile("slides2.gba")
	if err != nil {
		log.Fatal(err)
	}

	s := string(bs)

	s = strings.Replace(s, "\n\n", "</div>", -1)
	s = strings.Replace(s, "‣", "<div>", -1)
	s = strings.Replace(s, "⁝", "<div>", -1)

	fmt.Fprint(os.Stdout, s)
}
