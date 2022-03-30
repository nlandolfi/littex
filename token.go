package lit

import (
	"fmt"
	"unicode"
)

type TokenType int

const (
	ErrorToken TokenType = iota
	WordToken
	PunctuationToken
	SymbolToken
	OpaqueToken
)

func (t TokenType) String() string {
	switch t {
	case ErrorToken:
		return "error"
	case WordToken:
		return "word"
	case PunctuationToken:
		return "punctuation"
	case SymbolToken:
		return "symbol"
	case OpaqueToken:
		return "opaque"
	default:
		panic(fmt.Sprintf("unknown token type: %d", t))
	}
}

type Token struct {
	Type     TokenType
	Value    string
	Implicit bool
}

func (t *Token) String() string {
	// return fmt.Sprintf("%s(%q)%d:%d", t.Type, t.Value, t.StartLine, t.StartChar)
	return fmt.Sprintf("%s(%q)", t.Type, val(t))
}

func Lex(s string) (tokens []*Token, err error) {
	var opaque bool

	for i, r := range s {
		if opaque {
			if r == OpaqueCloseRune {
				opaque = false
				continue
			}
			tokens[len(tokens)-1].Value += string(r)
			continue
		}

		switch {
		case r == OpaqueOpenRune:
			tokens = append(tokens, &Token{
				Type:  OpaqueToken,
				Value: "",
			})
			opaque = true
		case r == ' ':
			if len(tokens) == 0 {
				continue
			}
			if i == len(s)-1 {
				continue // last one can't be an implicit space
			}
			last := tokens[len(tokens)-1]
			if last.Type == WordToken ||
				last.Type == SymbolToken ||
				last.Type == OpaqueToken ||
				(last.Type == PunctuationToken && last.Value != "·") {
				// convert it to a space
				tokens = append(tokens, &Token{
					Type:     PunctuationToken,
					Value:    "·",
					Implicit: true,
				})
			}
		case r == '\n' || r == '\r' || r == '\t':
			continue
		case unicode.IsSymbol(r):
			tokens = append(tokens, &Token{
				Type:  SymbolToken,
				Value: string(r),
			})
		case unicode.IsPunct(r):
			// TODO should we detect that if the previous token was a word
			// then add the hyphen so that clear-cut lexes as a word?
			tokens = append(tokens, &Token{
				Type:  PunctuationToken,
				Value: string(r),
			})
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			if len(tokens) == 0 || tokens[len(tokens)-1].Type != WordToken { // start a new word
				tokens = append(tokens, &Token{
					Type:  WordToken,
					Value: string(r),
				})
				continue
			}
			// continue that word
			tokens[len(tokens)-1].Value += string(r)
		default:
			err = fmt.Errorf("unrecognized rune %q", r)
			return
		}
	}

	return
}
