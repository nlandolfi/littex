package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

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
	// only heuristics
	// each of the first several are incomplete, in fact
	s = strings.Replace(s, "\\textit{", "‹", -1)
	s = strings.Replace(s, "\\textbf{", "«", -1)
	s = strings.Replace(s, "\\%", "%", -1)
	s = strings.Replace(s, "\\footnote", "†", -1)
	s = strings.Replace(s, "\\t{", "❬", -1)
	s = strings.Replace(s, "\\c{", "⁅", -1)
	s = strings.Replace(s, "\\item", "‣", -1)
	s = strings.Replace(s, "\\begin{itemize}", "⁝ ⦊", -1)
	s = strings.Replace(s, "\\end{itemize}", "⦉", -1)
	s = strings.Replace(s, "---", "—", -1)
	s = strings.Replace(s, "``", "“", -1)
	s = strings.Replace(s, "''", "”", -1)
	s = strings.Replace(s, "`", "‘", -1) // MUST BE AFTER DOUBLE
	s = strings.Replace(s, "'", "’", -1)

	var b bytes.Buffer
	w := &b
	ps := strings.Split(s, "\n\n")
	for _, p := range ps {
		fmt.Fprintf(w, "¶ ⦊")
		ls := strings.Split(p, "\n")
		for _, l := range ls {
			fmt.Fprintf(w, "‖ ")
			fmt.Fprint(w, l)
			fmt.Fprintf(w, "⦉")
		}
		fmt.Fprintf(w, "⦉")
	}

	n, err := gba.Parse3(b.String())
	if err != nil {
		log.Fatal(err)
	}

	gba.WriteGBA(os.Stdout, n, "", "  ")
}
