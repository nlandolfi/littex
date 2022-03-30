package gba

import (
	"fmt"
	"log"
	"strings"
	"unicode"
)

//‣

// list stuff
// ½ {
// }
// ⁑d
// ︰
// ⋮ {
// }
var recognizedPunctuation = map[rune]bool{
	// standard
	',': true,
	'.': true,
	';': true,
	':': true,
	'-': true,
	'(': true,
	')': true,
	'[': true,
	']': true,
	'?': true,
	'!': true,

	'‹': true,
	'›': true,
	'❮': true,
	'❯': true,

	'«': true,
	'»': true,
	'⧼': true, // terms
	'⧽': true,
	'“': true, //left
	'”': true, //right
	'–': true, // en dash
	'—': true, // em dash
	'‘': true, // left
	'’': true, // right
	'᜶': true,

	'↦': true,
	'↤': true,
	'·': true,
}

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

type State int

const (
	StateFresh State = iota
	StateOpenP
	StateInP
	StateInL
	StateOpenF
	StateInF
	StateInLF
	StateInOpaque
	StateInOpaqueF
	StateOpenMath
	StateInMath
)

func (s State) String() string {
	switch s {
	case StateFresh:
		return "fresh"
	case StateOpenP:
		return "openp"
	case StateInP:
		return "inp"
	case StateInL:
		return "inl"
	case StateOpenF:
		return "openf"
	case StateInF:
		return "inf"
	case StateInLF:
		return "inlf"
	case StateInOpaque:
		return "opaque"
	case StateInOpaqueF:
		return "opaquef"
	case StateOpenMath:
		return "openmath"
	case StateInMath:
		return "math"
	}
	panic("not reached")
}

func ParseSource1(bs []byte) (*Fragment, error) {
	var line, char int = 1, 0

	var state State
	var f Fragment

	var lastToken *Token1

	var opaqueLevel int

	for _, r := range string(bs) {
		if r == '\n' {
			char = 0
			line += 1
		} else {
			char += 1
		}

		//		log.Printf("%q", r)
		//		log.Printf("%s", string(r))
		//		log.Printf("%#v", f)
		switch state {
		case StateFresh:
			if unicode.IsSpace(r) {
				continue
			}
			switch r {
			case '¶':
				state = StateOpenP
				continue
			default:
				panic(fmt.Sprintf("unexpected rune %q in state %s at %d:%d", r, state, line, char))
			}
		case StateOpenP, StateOpenF:
			if unicode.IsSpace(r) {
				continue
			}
			switch r {
			case '{':
				switch state {
				case StateOpenP:
					state = StateInP
					f.AddParagraph(&Paragraph{})
				case StateOpenF:
					state = StateInF
					f.LastParagraph().LastRun().AddNote(
						&Note{Index: len(f.LastParagraph().LastRun().Tokens)},
					)
				}
				continue
			default:
				panic(fmt.Sprintf("unexpected rune %q in state %s at %d:%d", r, state, line, char))
			}
		case StateInP, StateInF:
			if unicode.IsSpace(r) {
				continue
			}

			switch r {
			case '‖':
				switch state {
				case StateInP:
					state = StateInL
					f.LastParagraph().AddRun(&Run{})
					//					lastToken = &Token{Type: GlueToken, Data: " "}
					//					f.LastParagraph().LastRun().AddToken(lastToken)
				case StateInF:
					state = StateInLF
					f.LastParagraph().LastRun().CurrentNote().AddRun(&Run{})
					//		lastToken = &Token{Type: GlueToken, Data: string(r)}
					//			f.LastParagraph().LastRun().CurrentNote().LastRun().AddToken(lastToken)
				}
				continue
			case '}':
				switch state {
				case StateInP:
					state = StateFresh
				case StateInF:
					state = StateInP
				}
			default:
				panic(fmt.Sprintf("unexpected rune %q in state %s at %d:%d", r, state, line, char))
			}
		case StateOpenMath:
			switch {
			case r == '{':
				state = StateInMath
				lastToken = &Token1{Type: OpaqueToken1, Data: string(r)}
				state = StateInMath
				f.LastParagraph().LastRun().AddMath(&Math{Index: len(f.LastParagraph().LastRun().Tokens), Token: lastToken})
				continue
			}
		case StateInOpaque, StateInOpaqueF, StateInMath:
			lastToken.Add(r)
			switch {
			case r == '{':
				opaqueLevel += 1
			case r == '}':
				if opaqueLevel == 0 {
					switch state {
					case StateInOpaque:
						state = StateInL
					case StateInOpaqueF:
						state = StateInLF
					case StateInMath:
						state = StateInL
					}
				} else {
					opaqueLevel -= 1
				}
			}
		case StateInL, StateInLF:
			switch {
			case r == '\t' || r == '\r' || r == '\n':
				continue
			case r == ' ':
				if lastToken != nil && lastToken.Type == GlueToken1 {
					continue // ignore repeated glues
				}
				lastToken = &Token1{Type: GlueToken1, Data: string(r)}
				switch state {
				case StateInL:
					f.LastParagraph().LastRun().AddToken(lastToken)
				case StateInLF:
					f.LastParagraph().LastRun().CurrentNote().LastRun().AddToken(lastToken)
				}
				continue
			case r == '†':
				switch state {
				case StateInL:
					state = StateOpenF
					continue
				case StateInLF:
					panic("no double footnotes")
				}
				continue
			case r == '◇':
				switch state {
				case StateInL:
					state = StateOpenMath
					continue
				case StateInLF:
					panic("no display math in footnotes, sorry")
				}
			case r == '‖':
				switch state {
				case StateInL:
					f.LastParagraph().AddRun(&Run{})
					continue
				case StateInLF:
					f.LastParagraph().LastRun().CurrentNote().AddRun(&Run{})
					continue
				}
			case r == '{':
				lastToken = &Token1{Type: OpaqueToken1, Data: string(r)}
				switch state {
				case StateInL:
					f.LastParagraph().LastRun().AddToken(lastToken)
					state = StateInOpaque
					continue
				case StateInLF:
					f.LastParagraph().LastRun().CurrentNote().LastRun().AddToken(lastToken)
					state = StateInOpaqueF
					continue
				}
			case r == '}':
				switch state {
				case StateInL:
					state = StateFresh
					continue
				case StateInLF:
					state = StateInL
					continue
				}
			case r == '_' || r == '*' || r == '$':
				lastToken = &Token1{Type: StyleToken1, Data: string(r)}
				switch state {
				case StateInL:
					f.LastParagraph().LastRun().AddToken(lastToken)
				case StateInLF:
					f.LastParagraph().LastRun().CurrentNote().LastRun().AddToken(lastToken)
				}
			case unicode.IsPunct(r) || r == '↦' || r == '↤':
				if !recognizedPunctuation[r] {
					panic(fmt.Sprintf("unrecognized punctuation %q in state %s at %d:%d", r, state, line, char))
				}
				lastToken = &Token1{Type: PunctuationToken1, Data: string(r)}
				switch state {
				case StateInL:
					f.LastParagraph().LastRun().AddToken(lastToken)
				case StateInLF:
					f.LastParagraph().LastRun().CurrentNote().LastRun().AddToken(lastToken)
				}
			default:
				switch {
				case lastToken == nil || (lastToken != nil && lastToken.Type != WordToken1):
					lastToken = &Token1{Type: WordToken1, Data: string(r)}
					switch state {
					case StateInL:
						f.LastParagraph().LastRun().AddToken(lastToken)
					case StateInLF:
						f.LastParagraph().LastRun().CurrentNote().LastRun().AddToken(lastToken)
					}
				default:
					if !(unicode.IsLetter(r) || unicode.IsNumber(r)) {
						panic(fmt.Sprintf("unrecognized unicode %q in state %s at %d:%d", r, state, line, char))
					}
					lastToken.Add(r)
				}
			}
		}
	}

	return &f, nil

	//f.Write(os.Stdout)

	//	log.Printf("%#v", f)
}

const NewRunGroupRune = '¶' // paragraphs
const NewRunRune = '‖'      // runs
const NewMathRune = '◇'     // math
const NewFootnoteRune = '†' // footnote
const NewListRune = '⁝'     // lists
const NewListItemRune = '‣' // list items
const OpenRune = '{'        // open
const CloseRune = '}'       // close
const OpenOpaqueRune = '❲'
const CloseOpaqueRune = '❳'

/*
const OpenMathOpaqueRune = '⧼'
const CloseMathOpaqueRune = '⧽'
*/

const BlockTToken = "block"
const WordTToken = "word"
const PunctuationTToken = "punctuation"
const OpenTToken = "open"
const CloseTToken = "close"
const OpaqueTToken = "opaque"
const SymbolTToken = "symbol"

type TToken struct {
	Type                 string
	Value                string
	StartLine, StartChar int
	width                int
}

func (t *TToken) String() string {
	return fmt.Sprintf("%s(%q)%d:%d", t.Type, t.Value, t.StartLine, t.StartChar)
}

func Lex2(s string) []*TToken {
	// destroys nice errors with char and lines
	/*
		lines := strings.Split(s, "\n")
		for i, l := range lines {
			lines[i] = strings.TrimSpace(l)
		}
		s = strings.Join(lines
	*/

	var line, char int = 1, 0
	var ts []*TToken = []*TToken{&TToken{Type: "start", Value: ""}}
	var opaque bool

	for _, r := range s {
		if r == '\n' {
			char = 0
			line += 1
		} else {
			char += 1
		}

		if opaque {
			if r == CloseOpaqueRune {
				opaque = false
				continue
			}
			ts[len(ts)-1].Value += string(r)
			continue
		}

		switch {
		case r == NewRunGroupRune || r == NewListRune || r == NewRunRune || r == NewListItemRune || r == NewFootnoteRune || r == NewMathRune:
			ts = append(ts, &TToken{
				Type:      BlockTToken,
				Value:     string(r),
				StartLine: line,
				StartChar: char,
			})
		case r == OpenRune:
			ts = append(ts, &TToken{
				Type: OpenTToken, Value: "{",
				StartLine: line,
				StartChar: char,
			})
			// lex ◇ as opaque...
			if ts[len(ts)-1].Value == "◇" {
				ts = append(ts, &TToken{
					Type:      OpaqueTToken,
					Value:     "",
					StartLine: line,
					StartChar: char,
				})
				opaque = true
			}
		case r == CloseRune:
			ts = append(ts, &TToken{
				Type:      CloseTToken,
				Value:     "}",
				StartLine: line,
				StartChar: char,
			})
		case r == OpenOpaqueRune:
			ts = append(ts, &TToken{
				Type:      OpaqueTToken,
				Value:     "",
				StartLine: line,
				StartChar: char,
			})
			opaque = true
			/*
				case r == OpenMathOpaqueRune:
					ts = append(ts, TToken{typ: OpaqueTToken, Value: "$"})
					opaque = true
				case r == CloseMathOpaqueRune:
					ts[len(ts)-1].Value += "$"
					opaque = false
			*/
		case unicode.IsSpace(r):
			if r != ' ' {
				continue
			}

			last := ts[len(ts)-1]

			// ok this is some magic nonsense
			if last.Type == WordTToken ||
				last.Type == SymbolTToken ||
				last.Type == OpaqueTToken ||
				(last.Type == PunctuationTToken && last.Value != "·") {
				// convert it to a space
				ts = append(ts, &TToken{
					Type:      PunctuationTToken,
					Value:     "·",
					StartLine: line,
					StartChar: char,
				})
			}
		case unicode.IsSymbol(r):
			ts = append(ts, &TToken{
				Type:      SymbolTToken,
				Value:     string(r),
				StartLine: line,
				StartChar: char,
			})
		case unicode.IsPunct(r): // revert to IsPunctuation?
			ts = append(ts, &TToken{
				Type:      PunctuationTToken,
				Value:     string(r),
				StartLine: line,
				StartChar: char,
			})
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			last := ts[len(ts)-1]
			if last.Type != WordTToken { // start a new word
				ts = append(ts, &TToken{
					Type:      WordTToken,
					Value:     string(r),
					StartLine: line,
					StartChar: char,
				})
				continue
			}
			// continue that word
			last.Value += string(r)
		default:
			panic(fmt.Sprintf("unrecognized rune %q at line %d char %d", r, line, char))
		}
	}

	return ts
}

type Block struct {
	Type  string
	Token *TToken
	Kids  []*Block
}

const (
	ContainerBlock = "container"
	AtomicBlock    = "atomic"
)

func prints(s []*Block) string {
	var ss []string
	for _, b := range s {
		ss = append(ss, fmt.Sprintf("%s-%s", b.Type, b.Token))
	}
	return strings.Join(ss, " > ")
}

func Parse2(ts []*TToken) *Block {
	var stack = []*Block{
		&Block{
			Type: ContainerBlock, Token: &TToken{Type: BlockTToken, Value: "ROOT"},
		},
	}

	var pendingBlock *Block

	for _, t := range ts {
		if len(stack) == 0 {
			panic("stack is empty!")
		}
		cur := stack[len(stack)-1]
		log.Println(prints(stack))
		log.Printf("recv %s in %s token %s", t, cur.Type, cur.Token.String())

		if pendingBlock != nil {
			if t.Type != OpenTToken {
				panic("expected open")
			}
			cur.Kids = append(cur.Kids, pendingBlock)
			stack = append(stack, pendingBlock)
			pendingBlock = nil
			continue
		}

		switch t.Type {
		case BlockTToken:
			switch t.Value {
			case "‖", "‣": // special ones
				if cur.Token.Value == "‖" || cur.Token.Value == "‣" {
					stack = stack[:len(stack)-1]
				}
				p := stack[len(stack)-1]
				b := &Block{Type: ContainerBlock, Token: t}
				p.Kids = append(p.Kids, b)
				stack = append(stack, b)
			default:
				pendingBlock = &Block{Type: ContainerBlock, Token: t}
			}
		case CloseTToken:
			if v := stack[len(stack)-1].Token.Value; v == "‖" || v == "‣" {
				if len(stack) > 2 {
					stack = stack[:len(stack)-2] // close it and the previous
					log.Print("double close!")
				} else {
					stack = stack[:len(stack)-1] // close it
				}
				continue
			}

			log.Print("normal close")
			stack = stack[:len(stack)-1] // close it
		case WordTToken, PunctuationTToken, OpaqueTToken, SymbolTToken:
			cur.Kids = append(cur.Kids, &Block{Type: AtomicBlock, Token: t})
		}
	}

	// we have hit the end of the file,
	// so count it as a close
	if v := stack[len(stack)-1].Token.Value; v == "‖" || v == "‣" {
		stack = stack[:len(stack)-1]
	}

	if len(stack) != 1 {
		log.Printf("unclosed %s", stack[len(stack)-1].Token)
		panic(fmt.Sprintf("stack has %d elements, want %d", len(stack), 1))
	}

	return stack[0]
}
