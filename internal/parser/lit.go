package parser

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/nlandolfi/lit/internal/ast"
)

// Class constants for lit types
const (
	RunClass         = "run"
	ParagraphClass   = "paragraph"
	FootnoteClass    = "footnote"
	DisplayMathClass = "displaymath"
	ListClass        = "list"
	ListItemClass    = "listitem"
	SectionClass     = "section"
)

// ParseLit parses a Lit format string and returns the corresponding AST
func ParseLit(s string) (*ast.Node, error) {
	s = litReplace(s)
	return ParseHTML(s)
}

func litReplace(s string) string {
	s = " " + s // to ensure a first character match,
	// for the picrow etc escapes

	s = strings.Replace(s, "\\<", "&lt;", -1)
	s = strings.Replace(s, "\\>", "&gt;", -1)

	// runs
	re := regexp.MustCompile(`[^\\]‖`)
	s = re.ReplaceAllString(s, "<div data-littype='"+RunClass+"'>")
	s = strings.Replace(s, "\\‖", "‖", -1)

	// pilcrow
	re = regexp.MustCompile(`([^\\])¶⧊`)
	s = re.ReplaceAllString(s, `$1¶ ⧊`)
	re = regexp.MustCompile(`([^\\])¶ ⧊`)
	s = re.ReplaceAllString(s, "$1<div data-littype='"+ParagraphClass+"'>")
	s = strings.Replace(s, "\\¶", "¶", -1)

	// footnote
	re = regexp.MustCompile(`([^\\])†⧊`)
	s = re.ReplaceAllString(s, `$1† ⧊`)
	re = regexp.MustCompile(`([^\\])† ⧊`)
	s = re.ReplaceAllString(s, "$1<div data-littype='"+FootnoteClass+"'>")
	s = strings.Replace(s, "\\†", "†", -1)

	// display math
	re = regexp.MustCompile(`([^\\])◇⧊`)
	s = re.ReplaceAllString(s, `$1◇ ⧊`)
	re = regexp.MustCompile(`([^\\])◇ ⧊`)
	s = re.ReplaceAllString(s, "$1<div data-littype='"+DisplayMathClass+"'>")
	s = strings.Replace(s, "\\◇", "◇", -1)

	// unordered lists
	re = regexp.MustCompile(`([^\\])⁝⧊`)
	s = re.ReplaceAllString(s, `$1⁝ ⧊`)
	re = regexp.MustCompile(`([^\\])⁝ ⧊`)
	s = re.ReplaceAllString(s, "$1<div data-littype='"+ListClass+"' data-litlisttype='unordered'>")
	s = strings.Replace(s, "\\⁝", "⁝", -1)

	// ordered lists
	re = regexp.MustCompile(`([^\\])𝝫⧊`)
	s = re.ReplaceAllString(s, `$1𝝫 ⧊`)
	re = regexp.MustCompile(`([^\\])𝝫 ⧊`)
	s = re.ReplaceAllString(s, "$1<div data-littype='"+ListClass+"' data-litlisttype='ordered'>")
	s = strings.Replace(s, "\\𝝫", "𝝫", -1)

	// list items
	re = regexp.MustCompile(`([^\\])‣`)
	s = re.ReplaceAllString(s, "$1<div data-littype='"+ListItemClass+"'>")
	s = strings.Replace(s, "\\‣", "‣", -1)

	// sections
	// first, replace repeats
	re = regexp.MustCompile(`[^\\]§+`)
	s = re.ReplaceAllStringFunc(s, func(og string) string {
		// drop the first non \\ match
		index := strings.Index(og, "§")
		in := og[index:]
		out := fmt.Sprintf("%s§%d", og[0:index], utf8.RuneCountInString(in))
		return out
	})
	// numbered
	re = regexp.MustCompile(`#§([[:digit:]]+)`)
	s = re.ReplaceAllString(s, "<div data-littype='"+SectionClass+"' data-litsectionlevel='$1' data-litsectionnumbered='true'>")
	// unnumbered
	re = regexp.MustCompile(`([^\\#])§([[:digit:]]+)`)
	s = re.ReplaceAllString(s, "$1<div data-littype='"+SectionClass+"' data-litsectionlevel='$2' data-litsectionnumbered='false'>")
	// section symbol
	s = strings.Replace(s, "\\§", "§", -1)

	// closes
	// the naive single match doesn't work, misses some of them
	// so need this more complicated thing
	re = regexp.MustCompile(`([^\\])⧉+`)
	s = re.ReplaceAllStringFunc(s, func(og string) string {
		// drop the first non \\ match
		index := strings.Index(og, "⧉")
		in := og[index:]
		out := og[0:index]
		for i := 0; i < utf8.RuneCountInString(in); i++ {
			out += "</div>"
		}
		return out
	})
	// all to get the escape functionality
	s = strings.Replace(s, "\\⧉", "⧉", -1)

	// Update: Unfortunately the below doesn't work
	// because it will write out the replacements, instead
	// of the compact form...more to be done here.
	// -1 hack, this should be improved to only make
	// the replacement when in math mode, and to think
	// through edge cases, but I think the gains in
	// readability for now outweigh the fragileness of
	// this solution
	s = strings.Replace(s, "⁻¹", "^{-1}", -1)
	// the same goes for the below
	s = strings.Replace(s, "¹", "^{1}", -1)
	s = strings.Replace(s, "²", "^{2}", -1)
	s = strings.Replace(s, "₁", "_{1}", -1)
	s = strings.Replace(s, "₂", "_{2}", -1)
	s = strings.Replace(s, "ᵢ", "_{i}", -1)
	s = strings.Replace(s, "ⱼ", "_{j}", -1)
	s = strings.Replace(s, "ₖ", "_{k}", -1)
	s = strings.Replace(s, "ₘ", "_{m}", -1)
	s = strings.Replace(s, "ₙ", "_{n}", -1)

	return s
}
