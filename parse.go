package gba

import (
	"fmt"
	"unicode"
)

// list stuff
// ½ {
// }
// ⁑d
// ︰
// ⋮ {
// }

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

	var lastToken *Token

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
					f.LastParagraph().LastRun().AddNote(&Note{Index: len(f.LastParagraph().LastRun().Tokens)})
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
				lastToken = &Token{Type: OpaqueToken, Data: string(r)}
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
				if lastToken != nil && lastToken.Type == GlueToken {
					continue // ignore repeated glues
				}
				lastToken = &Token{Type: GlueToken, Data: string(r)}
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
				lastToken = &Token{Type: OpaqueToken, Data: string(r)}
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
				lastToken = &Token{Type: StyleToken, Data: string(r)}
				switch state {
				case StateInL:
					f.LastParagraph().LastRun().AddToken(lastToken)
				case StateInLF:
					f.LastParagraph().LastRun().CurrentNote().LastRun().AddToken(lastToken)
				}
			case unicode.IsPunct(r):
				lastToken = &Token{Type: PunctuationToken, Data: string(r)}
				switch state {
				case StateInL:
					f.LastParagraph().LastRun().AddToken(lastToken)
				case StateInLF:
					f.LastParagraph().LastRun().CurrentNote().LastRun().AddToken(lastToken)
				}
			default:
				switch {
				case lastToken == nil || (lastToken != nil && lastToken.Type != WordToken):
					lastToken = &Token{Type: WordToken, Data: string(r)}
					switch state {
					case StateInL:
						f.LastParagraph().LastRun().AddToken(lastToken)
					case StateInLF:
						f.LastParagraph().LastRun().CurrentNote().LastRun().AddToken(lastToken)
					}
				default:
					lastToken.Add(r)
				}
			}
		}
	}

	return &f, nil

	//f.Write(os.Stdout)

	//	log.Printf("%#v", f)
}
