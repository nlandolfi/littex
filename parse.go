package gba

import (
	"log"
	"unicode"
)

func ParseSource1(bs []byte) (*Fragment, error) {
	var state State
	var f Fragment

	var lastToken *Token

	for _, r := range string(bs) {
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
				log.Fatalf("unexpected rune %q in state %s:", r, state)
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
					f.LastParagraph().LastRun().AddNote(&Note{Index: len(f.LastParagraph().LastRun().Tokenized)})
				}
				continue
			default:
				log.Fatalf("unexpected rune %q in state %s:", r, state)
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
					lastToken = nil
				case StateInF:
					state = StateInLF
					f.LastParagraph().LastRun().CurrentNote().AddRun(&Run{})
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
				log.Fatalf("unexpected rune %q in state %s:", r, state)
			}
		case StateInL, StateInLF:
			switch {
			case unicode.IsSpace(r):
				lastToken = &Token{Type: Glue, Data: string(r)}
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
			case r == '‖':
				switch state {
				case StateInL:
					f.LastParagraph().AddRun(&Run{})
					continue
				case StateInLF:
					f.LastParagraph().LastRun().CurrentNote().AddRun(&Run{})
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
				lastToken = &Token{Type: Style, Data: string(r)}
				switch state {
				case StateInL:
					f.LastParagraph().LastRun().AddToken(lastToken)
				case StateInLF:
					f.LastParagraph().LastRun().CurrentNote().LastRun().AddToken(lastToken)
				}
			case unicode.IsPunct(r):
				lastToken = &Token{Type: Punctuation, Data: string(r)}
				switch state {
				case StateInL:
					f.LastParagraph().LastRun().AddToken(lastToken)
				case StateInLF:
					f.LastParagraph().LastRun().CurrentNote().LastRun().AddToken(lastToken)
				}
			default:
				switch {
				case lastToken == nil || (lastToken != nil && lastToken.Type != Word):
					lastToken = &Token{Type: Word, Data: string(r)}
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
