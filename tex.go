package gba

import (
	"strings"
	"unicode/utf8"
)

func Tex(t *Token) string {
	switch t.Type {
	case WordToken:
		return t.Value
	case PunctuationToken:
		switch r, _ := utf8.DecodeRuneInString(t.Value); r {
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
		case '↦':
			return "{\\indent}"
		case '↤':
			return "{\\noindent}"
		}
	case SymbolToken:
		r, _ := utf8.DecodeRuneInString(t.Value)
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
	'∅': "\\varnothing",
	'∪': "\\cup",
	'∩': "\\cap",
	'×': "\\times",
	'𝒞': "\\mathcal{C}",
	'∕': "/",
	'∏': "\\prod",
	'∑': "\\sum",
	'≡': "\\equiv",
	'≪': "\\ll",
	'≫': "\\gg",
	'≦': "\\leqq",
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
	'𝗥': "\\mathbfsf{R}",
	'𝗤': "\\mathbfsf{Q}",
	'𝗡': "\\mathbfsf{N}",
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
}
