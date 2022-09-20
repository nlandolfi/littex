package lit

import (
	"strings"
	"unicode/utf8"
)

func Tex(t *Token, inMath bool) string {
	switch t.Type {
	case WordToken:
		var out = ""

		for _, r := range t.Value {
			// TODO avoid dictionary lookup if not in math
			if replacement, ok := LatexMathReplacements[r]; inMath && ok {
				out += replacement + " " // I think we need the space here.
			} else {
				out += string(r)
			}
		}

		return out
	case PunctuationToken:
		switch r, _ := utf8.DecodeRuneInString(t.Value); r {
		case '&':
			return "\\&"
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
	case SymbolToken:
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
			return "\\noindent"
		case '␣':
			return " "
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
			return replacement + " "
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
	'⋃': "\\bigcup",
	'∩': "\\cap ",
	'⋂': "\\bigcap",
	'×': "\\times ",
	'★': "\\star",
	'𝒜': "\\mathcal{A}",
	'ℬ': "\\mathcal{B}",
	'𝒞': "\\mathcal{C}",
	'𝒟': "\\mathcal{D}",
	'ℰ': "\\mathcal{E}",
	'ℱ': "\\mathcal{F}",
	'𝒢': "\\mathcal{G}",
	'ℋ': "\\mathcal{H}",
	'ℐ': "\\mathcal{I}",
	'𝒥': "\\mathcal{J}",
	'𝒦': "\\mathcal{K}",
	'ℒ': "\\mathcal{L}",
	'ℳ': "\\mathcal{M}",
	'𝒩': "\\mathcal{N}",
	'𝒪': "\\mathcal{O}",
	'𝒫': "\\mathcal{P}",
	'𝒬': "\\mathcal{Q}",
	'ℛ': "\\mathcal{R}",
	'𝒮': "\\mathcal{S}",
	'𝒯': "\\mathcal{T}",
	'𝒰': "\\mathcal{U}",
	'𝒱': "\\mathcal{V}",
	'𝒳': "\\mathcal{X}",
	'ℓ': "\\ell",
	//	'∕': "/", causes confusion with </div>
	'∏': "\\prod",
	'∑': "\\sum",
	'≈': "\\approx",
	'≡': "\\equiv",
	'≪': "\\ll",
	'≫': "\\gg",
	'≦': "\\leqq",
	'≧': "\\geqq",
	'≥': "\\geq",
	'≤': "\\leq",
	'≺': "\\prec",
	'≻': "\\succ",
	'≼': "\\preceq",
	'≽': "\\succeq",
	'∫': "\\int ",
	'∀': "\\forall",
	'∃': "\\exists ",
	'∄': "\\not\\exists",
	'∞': "\\infty",
	'∝': "\\propto",
	'∘': "\\circ",
	'⋮': "\\vdots",
	'⋯': "\\cdots",
	'⋱': "\\ddots",
	'·': "\\cdot",
	'∼': "\\sim",
	'√': "\\sqrt",
	'±': "\\pm",
	'∓': "\\mp",
	'𝗥': "\\R",
	'𝗤': "\\Q",
	'𝗡': "\\N",
	'𝗭': "\\Z",
	'𝗖': "\\C",
	'𝗙': "\\F",
	'𝗘': "\\E",
	'𝗣': "\\mathbfsf{P}",
	'𝗦': "\\mathbfsf{S}",
	'𝐑': "\\mathbf{R}",
	'𝐒': "\\mathbf{S}",
	'𝐄': "\\mathbf{E}",
	'𝟏': "\\mathbf{1}",
	'𝟎': "\\mathbf{0}",
	'∇': "\\nabla ",
	'∂': "\\partial ",
	'α': "\\alpha",
	'β': "\\beta",
	'ψ': "\\psi",
	'δ': "\\delta",
	'ε': "\\varepsilon",
	'ϵ': "\\epsilon",
	'φ': "\\phi",
	'Φ': "\\Phi",
	'γ': "\\gamma",
	'η': "\\eta",
	'ι': "\\iota",
	'ξ': "\\xi",
	'κ': "\\kappa",
	'λ': "\\lambda",
	'Λ': "\\Lambda ",
	'μ': "\\mu",
	'ν': "\\nu",
	'ο': "\\omicron",
	'π': "\\pi",
	'ρ': "\\rho",
	'σ': "\\sigma",
	'Σ': "\\Sigma",
	'⇒': "\\Rightarrow",
	'⇐': "\\Leftarrow",
	'τ': "\\tau",
	'θ': "\\theta",
	'Θ': "\\Theta",
	'ω': "\\omega",
	'Ω': "\\Omega",
	//	'ς':
	'χ': "\\chi",
	'υ': "\\upsilon",
	'ζ': "\\zeta",
	'⟹': "\\implies",
	'Ξ': "\\Xi",
	//	'½': "\\nicefrac{1}{2}", // does not work?
	'½': "1/2",
	'¼': "1/4",
	'⅙': "1/6",
	'⅓': "1/3",
	'∖': "\\setminus", // doesnot work?
	'…': "\\dots",
	'∔': "\\dotplus",
	'|': "\\mid ",
	'⟂': "\\perp ",
	'Π': "\\Pi",
	'¬': "\\neg",
	'∨': "\\lor",
	'∧': "\\land",
	'Ψ': "\\Psi",
	'Γ': "\\Gamma",
	'Δ': "\\Delta",
	'⊔': "\\sqcup",
	'ℑ': "\\Im",
	'ℜ': "\\Re",
	'∠': "\\angle",
	'⊤': "\\top ",
	'⊥': "\\perp",
}
