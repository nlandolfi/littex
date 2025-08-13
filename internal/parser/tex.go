package parser

import (
	"bytes"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/nlandolfi/lit/internal/ast"
	"github.com/nlandolfi/lit/internal/tex"
)

// ParseTex parses LaTeX format and converts it to the AST
func ParseTex(s string) (*ast.Node, error) {
	for _, c := range commentsR.FindAllString(s, -1) {
		log.Printf("dropping comment: %q", c)
	}

	for _, r := range order {
		replace := res[r]
		s = r.ReplaceAllString(s, replace)
	}
	s = strings.Replace(s, "\\item", " ‣", -1)
	s = strings.Replace(s, "\\begin{itemize}", " ⁝ ⦊", -1)
	s = strings.Replace(s, "\\begin{enumerate}", " 𝝫 ⦊", -1)
	s = strings.Replace(s, "\\end{itemize}", "⧉", -1)
	s = strings.Replace(s, "\\end{enumerate}", "⧉", -1)
	s = strings.Replace(s, "\\[\n", "◇ ⦊ ‖ ", -1)
	s = strings.Replace(s, "\n\\]", " ⧉", -1)
	s = strings.Replace(s, "---", "—", -1)
	s = strings.Replace(s, "``", "“", -1)
	s = strings.Replace(s, "''", "”", -1)
	s = strings.Replace(s, "`", "‘", -1) // MUST BE AFTER DOUBLE
	s = strings.Replace(s, "\\&", "&", -1)
	s = strings.Replace(s, "\\\\", "᜶", -1)
	s = strings.Replace(s, "\\indent", "↦", -1)
	s = strings.Replace(s, "\\noindent", "↤", -1)

	for from, to := range tex.LatexMathReplacements {
		s = strings.Replace(s, to, string(from), -1)
	}

	// TODO better comments handling
	s = strings.Replace(s, "\\%", "%", -1)

	var b bytes.Buffer
	w := &b
	ps := strings.Split(s, "\n\n")
	for _, p := range ps {
		fmt.Fprintf(w, " ¶ ⦊")
		ls := strings.Split(p, "\n")
		for _, l := range ls {
			fmt.Fprintf(w, " ‖ ")
			if len(l) > 0 && l[0] == '%' { // comments
				fmt.Fprintf(w, "❲%s❳", l)
			} else {
				fmt.Fprint(w, l)
			}
			fmt.Fprintf(w, "⧉")
		}
		fmt.Fprintf(w, "⧉")
	}

	return ParseLit(b.String())
}

var (
	textitR             = regexp.MustCompile(`\\textit{([\s\S]*?)}`)
	textbfR             = regexp.MustCompile(`\\textbf{([\s\S]*?)}`)
	textscR             = regexp.MustCompile(`\\textsc{([\s\S]*?)}`)
	footnoteR           = regexp.MustCompile(`\\footnote{([\s\S]*?)}`)
	tR                  = regexp.MustCompile(`\\t{([\s\S]*?)}`)
	cR                  = regexp.MustCompile(`\\c{([\s\S]*?)}`)
	dblqR               = regexp.MustCompile("``([\\s\\S]*?)''")
	sglqR               = regexp.MustCompile("`([\\s\\S]*?)'")
	sayR                = regexp.MustCompile(`\\say{([\s\S]*?)}`)
	commentsR           = regexp.MustCompile(`%(.*?)\n`)
	propositionWithText = regexp.MustCompile(`\\begin{proposition}\[([\w| ]*)\]`)
	proposition         = regexp.MustCompile(`\\begin{proposition}`)
	propositionEnd      = regexp.MustCompile(`\\end{proposition}`)
	proof               = regexp.MustCompile(`\\begin{proof}`)
	proofEnd            = regexp.MustCompile(`\\end{proof}`)
	ssection            = regexp.MustCompile(`\\ssection{(\w*)}`)
	section             = regexp.MustCompile(`\\section{(\w*)}`)
	ssubsection         = regexp.MustCompile(`\\ssubsection{(\w*)}`)
	subsection          = regexp.MustCompile(`\\subsection{(\w*)}`)
)

// useful: https://gist.github.com/claybridges/8f9d51a1dc365f2e64fa
var res = map[*regexp.Regexp]string{
	propositionWithText: " <statement type='proposition' text='$1'>",
	proposition:         " <statement type='proposition'>",
	propositionEnd:      " </statement>",
	proof:               " <proof>",
	proofEnd:            " </proof>",
	ssection:            " § $1 ⧉",
	section:             " #§ $1 ⧉",
	subsection:          " #§§ $1 ⧉",
	ssubsection:         " §§ $1 ⧉",
	textitR:             "‹$1›",
	textbfR:             "«$1»",
	footnoteR:           " † ⦊ ‖ $1 ⧉⧉",
	textscR:             "⸤$1⸥",
	tR:                  "❰$1❱",
	cR:                  "⁅$1⁆",
	dblqR:               "“$1”",
	sglqR:               "‘$1’",
	sayR:                "“$1”",
}

var order = []*regexp.Regexp{
	ssection,
	section,
	ssubsection,
	subsection,
	textitR,
	textbfR,
	footnoteR,
	textscR,
	tR,
	cR,
	dblqR,
	sglqR,
	sayR,
	propositionWithText,
	proposition,
	propositionEnd,
}
