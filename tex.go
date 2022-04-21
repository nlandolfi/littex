package lit

import (
	"strings"
	"unicode/utf8"
)

func Tex(t *Token) string {
	switch t.Type {
	case WordToken:
		if utf8.RuneCountInString(t.Value) == 1 {
			r, _ := utf8.DecodeRuneInString(t.Value)
			if replacement, ok := LatexMathReplacements[r]; ok {
				return replacement
			}
		}
		return t.Value
	case PunctuationToken:
		switch r, _ := utf8.DecodeRuneInString(t.Value); r {
		case '·':
			if t.Implicit {
				return " "
			}
			return "·"
		case '&':
			return "\\&"
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
		case '❬':
			return "\\t{"
		case '❭':
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
		case '“': //left
			return "\\say{"
		case '”': //right
			return "}"
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
		}
	case SymbolToken:
		r, _ := utf8.DecodeRuneInString(t.Value)
		switch r {
		case '᜶':
			return "\\\\"
		case '↦':
			return "\\indent"
		case '↤':
			return "\\noindent"
		}
		if replacement, ok := LatexMathReplacements[r]; ok {
			return replacement
		}
		return t.Value
	case OpaqueToken:
		x := t.Value
		for r, to := range LatexMathReplacements {
			x = strings.Replace(x, string(r), to, -1)
		}
		return x
	}

	if utf8.RuneCountInString(t.Value) == 1 {
		r, _ := utf8.DecodeRuneInString(t.Value)
		if replacement, ok := LatexMathReplacements[r]; ok {
			return replacement
		}
	}
	return t.Value
}

// TODO: clean up
var LatexMathReplacements = map[rune]string{
	'→': "\\to",
	'↦': "\\mapsto",
	'≠': "\\neq",
	'∈': "\\in",
	'∉': "\\not\\in",
	'⊃': "\\supset",
	'⊇': "\\supseteq",
	'⊂': "\\subset",
	'⊆': "\\subseteq",
	'⊊': "\\subsetneq",
	'∅': "\\varnothing",
	'∪': "\\cup",
	'∩': "\\cap",
	'×': "\\times",
	'𝒞': "\\mathcal{C}",
	'𝒰': "\\mathcal{U}",
	'𝒱': "\\mathcal{V}",
	'★': "\\star",
	'𝒢': "\\mathcal{G}",
	'ℋ': "\\mathcal{H}",
	'𝒩': "\\mathcal{N}",
	'ℓ': "\\ell",
	'∕': "/",
	'∏': "\\prod",
	'∑': "\\sum",
	'≈': "\\approx",
	'≡': "\\equiv",
	'≪': "\\ll",
	'≫': "\\gg",
	'≦': "\\leqq",
	'≥': "\\geq",
	'≤': "\\leq",
	'≺': "\\prec",
	'≻': "\\succ",
	'≼': "\\preceq",
	'≽': "\\succeq",
	'∫': "\\int",
	'∀': "\\forall",
	'∃': "\\exists",
	'∄': "\\not\\exists",
	'∞': "\\infty",
	'∝': "\\propto",
	'∘': "\\ocirc",
	'⋮': "\\vdots",
	'⋯': "\\cdots",
	'⋱': "\\ddots",
	'∼': "\\sim",
	'√': "\\sqrt",
	'±': "\\pm",
	'𝗥': "\\R",
	'𝗤': "\\Q",
	'𝗡': "\\N",
	'𝗭': "\\Z",
	'∇': "\\nabla",
	'∂': "\\partial",
	'α': "\\alpha",
	'β': "\\beta",
	'ψ': "\\psi",
	'δ': "\\delta",
	'ε': "\\varepsilon",
	'ϵ': "\\epsilon",
	'φ': "\\phi",
	'γ': "\\gamma",
	'η': "\\eta",
	'ι': "\\iota",
	'ξ': "\\xi",
	'κ': "\\kappa",
	'λ': "\\lambda",
	'μ': "\\mu",
	'ν': "\\nu",
	'ο': "\\omicron",
	'π': "\\pi",
	'ρ': "\\rho",
	'σ': "\\sigma",
	'τ': "\\tau",
	'θ': "\\theta",
	'ω': "\\omega",
	//	'ς':
	'χ': "\\chi",
	'υ': "\\upsilon",
	'ζ': "\\zeta",
	'⇒': "\\implies",
	'Ξ': "\\Xi",
	'½': "\\nicefrac{1}{2}", // does not work?
	'∖': "\\setminus",       // doesnot work?
	'…': "\\dots",
}
