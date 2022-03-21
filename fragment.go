package gba

import (
	"fmt"
)

type Fragment struct {
	Paragraphs []*Paragraph
}

type Paragraph struct {
	Runs []*Run
}

type Run struct {
	// change this to Tokens
	Tokens []*Token
	Notes  map[int]*Note
	Maths  map[int]*Math
}

type Math struct {
	Index int
	Token *Token // The opaque token corresponding
}

type Note struct {
	Index int
	Runs  []*Run
}

func (f *Fragment) AddParagraph(p *Paragraph) {
	f.Paragraphs = append(f.Paragraphs, p)
}

func (f *Fragment) LastParagraph() *Paragraph {
	return f.Paragraphs[len(f.Paragraphs)-1]
}

func (f *Fragment) LastRun() *Run {
	return f.LastParagraph().LastRun()
}

func (p *Paragraph) AddRun(l *Run) {
	p.Runs = append(p.Runs, l)
}

func (p *Paragraph) LastRun() *Run {
	return p.Runs[len(p.Runs)-1]
}

func (l *Run) AddToken(t *Token) {
	t.Index = len(l.Tokens)
	l.Tokens = append(l.Tokens, t)
}

func (l *Run) AddMath(n *Math) {
	if l.Maths == nil {
		l.Maths = make(map[int]*Math)
	}
	l.Maths[n.Index] = n
}

func (l *Run) AddNote(n *Note) {
	if l.Notes == nil {
		l.Notes = make(map[int]*Note)
	}
	l.Notes[n.Index] = n
}

func (l *Run) CurrentMath() *Math {
	return l.Maths[len(l.Tokens)]
}

func (l *Run) CurrentNote() *Note {
	return l.Notes[len(l.Tokens)]
}

func (n *Note) AddRun(r *Run) {
	n.Runs = append(n.Runs, r)
}

func (n *Note) LastRun() *Run {
	return n.Runs[len(n.Runs)-1]
}

type TokenType uint

// change these for better export
const (
	ErrorToken TokenType = iota
	WordToken
	PunctuationToken
	StyleToken
	GlueToken
	OpaqueToken
)

func (t *TokenType) String() string {
	switch *t {
	case ErrorToken:
		return "ERROR"
	case PunctuationToken:
		return "PUNCTUATION"
	case WordToken:
		return "WORD"
	case StyleToken:
		return "STYLE"
	case GlueToken:
		return "GLUE"
	case OpaqueToken:
		return "Opaque"
	}
	panic("TokentType.String not reached")
}

func (t TokenType) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%q", t.String())), nil
}

func (t *TokenType) UnmarshalJSON(bs []byte) error {
	switch string(bs) {
	case `"ERROR"`:
		*t = ErrorToken
		return nil
	case `"PUNCTUATION"`:
		*t = PunctuationToken
		return nil
	case `"WORD"`:
		*t = WordToken
		return nil
	case `"STYLE"`:
		*t = StyleToken
		return nil
	case `"GLUE"`:
		*t = GlueToken
		return nil
	case `"OPAQUE"`:
		*t = OpaqueToken
		return nil
	}
	return fmt.Errorf("unknown token type %q", bs)
}

type Token struct {
	Type  TokenType
	Data  string
	Index int
}

func (t *Token) Add(r rune) {
	t.Data += string(r)
}
