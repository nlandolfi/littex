package tex

import (
	"strings"
	"unicode/utf8"

	"github.com/nlandolfi/lit/internal/lexer"
)

// ToTeX converts a token to its LaTeX representation
func ToTeX(t *lexer.Token, inMath bool) string {
	switch t.Type {
	case lexer.WordToken:
		out := ""

		for _, r := range t.Value {
			// TODO avoid dictionary lookup if not in math
			if replacement, ok := LatexMathReplacements[r]; inMath && ok {
				out += replacement + " " // I think we need the space here.
			} else {
				out += string(r)
			}
		}

		return out
	case lexer.PunctuationToken:
		switch r, _ := utf8.DecodeRuneInString(t.Value); r {
		case '&':
			if !inMath {
				return "\\&"
			} else {
				return "&"
			}
		case '＆':
			return "&"
		case '%':
			return "\\%"
		case '‹':
			return "\\textit{"
		case '›':
			return "}"
		case '«':
			return "\\textbf{"
		case '»':
			return "}"
		case '❰':
			return "\\t{"
		case '❱':
			return "}"
		case '⁅':
			return "\\c{"
		case '⁆':
			return "}"
		case '❮':
			return "\\textbf{"
		case '❯':
			return "}"
		case '⧼':
			return "\\t{"
		case '⧽':
			return "}"
		case '“': // left
			return "``"
		case '”': // right
			return "''"
		case '–': // en dash
			return "--"
		case '—': // em dash
			return "---"
		case '‘': // left
			return "`"
		case '’': // right
			return "'"
		case '᜶':
			return "\\\\"
		case '⸤':
			return "\\textsc{"
		case '⸥':
			return "}"
		case '⅛':
			return "$\\nicefrace{1}{8}$"
		case '½':
			return "$\\nicefrace{1}{2}$"
		case '¼':
			return "$\\nicefrace{1}{4}$"
		case '_':
			if !inMath {
				return "\\_"
			} else {
				return "_"
			}
		}
	case lexer.SymbolToken:
		r, _ := utf8.DecodeRuneInString(t.Value)
		switch r {
		case '᜶':
			return "\\\\"
		case '↦':
			if inMath {
				return "\\mapsto"
			} else {
				return "\\indent"
			}
		case '↤':
			return "{\\noindent}"
		case '␣':
			return " "
		}
		if replacement, ok := LatexMathReplacements[r]; ok {
			return replacement
		}
		return t.Value
	case lexer.OpaqueToken:
		x := t.Value
		for r, to := range LatexMathReplacements {
			if r == '|' { // don't replace to mid; reason: table headers
				continue
			}
			x = strings.Replace(x, string(r), to, -1)
		}
		return x
	}

	if utf8.RuneCountInString(t.Value) == 1 {
		r, _ := utf8.DecodeRuneInString(t.Value)
		if replacement, ok := LatexMathReplacements[r]; ok {
			return replacement + " "
		}
	}
	return t.Value
}
