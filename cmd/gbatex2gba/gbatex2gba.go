package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/greatbooksadventure/gba"
)

var in = flag.String("in", "text.gba", "in file")

var textitR = regexp.MustCompile(`\\textit{((.|\n)*?)}`)
var textbfR = regexp.MustCompile(`\\textbf{((.|\n)*?)}`)
var footnoteR = regexp.MustCompile(`\\footnote{((.|\n)*?)}`)
var tR = regexp.MustCompile(`\\t{((.|\n)*?)}`)
var cR = regexp.MustCompile(`\\c{((.|\n)*?)}`)
var dblqR = regexp.MustCompile("``((.|\n)*)?''")
var sglqR = regexp.MustCompile("`((.|\n)*)?'")
var sayR = regexp.MustCompile(`\\say{((.|\n)*)?}`)
var commentsR = regexp.MustCompile(`%(.*?)\n`)

var res = map[*regexp.Regexp]string{
	textitR:   "‹$1›",
	textbfR:   "«$1»",
	footnoteR: "† ⦊ ‖ $1 ⦉⦉",
	tR:        "❬$1❭",
	cR:        "⁅$1⁆",
	dblqR:     "“$1”",
	sglqR:     "‘$1’",
	sayR:      "“$1”",
}

var order = []*regexp.Regexp{
	textitR,
	textbfR,
	footnoteR,
	tR,
	cR,
	dblqR,
	sglqR,
	sayR,
}

func main() {
	flag.Parse()
	bs, err := os.ReadFile(*in)
	if err != nil {
		log.Fatal(err)
	}
	s := string(bs)

	for _, c := range commentsR.FindAllString(s, -1) {
		log.Printf("dropping comment: %q", c)
	}

	for _, r := range order {
		replace := res[r]
		s = r.ReplaceAllString(s, replace)
	}
	s = strings.Replace(s, "\\item", "‣", -1)
	s = strings.Replace(s, "\\begin{itemize}", "⁝ ⦊", -1)
	s = strings.Replace(s, "\\end{itemize}", "⦉", -1)
	s = strings.Replace(s, "---", "—", -1)
	s = strings.Replace(s, "``", "“", -1)
	s = strings.Replace(s, "''", "”", -1)
	s = strings.Replace(s, "`", "‘", -1) // MUST BE AFTER DOUBLE
	s = strings.Replace(s, "'", "’", -1)
	s = strings.Replace(s, "\\&", "&", -1)
	s = strings.Replace(s, "\\\\", "᜶", -1)
	s = strings.Replace(s, "\\indent", "↦", -1)
	s = strings.Replace(s, "\\noindent", "↤", -1)

	for from, to := range gba.LatexMathReplacements {
		s = strings.Replace(s, to, string(from), -1)
	}

	// TODO better comments handling
	s = strings.Replace(s, "\\%", "%", -1)

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
