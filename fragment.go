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
	Tokens []*Token1
	Notes  map[int]*Note
	Maths  map[int]*Math
}

type Math struct {
	Index int
	Token *Token1 // The opaque token corresponding
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

func (l *Run) AddToken(t *Token1) {
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

type Token1Type uint

// change these for better export
const (
	ErrorToken1 Token1Type = iota
	WordToken1
	PunctuationToken1
	StyleToken1
	GlueToken1
	OpaqueToken1
)

func (t *Token1Type) String() string {
	switch *t {
	case ErrorToken1:
		return "ERROR"
	case PunctuationToken1:
		return "PUNCTUATION"
	case WordToken1:
		return "WORD"
	case StyleToken1:
		return "STYLE"
	case GlueToken1:
		return "GLUE"
	case OpaqueToken1:
		return "Opaque"
	}
	panic("TokentType.String not reached")
}

func (t Token1Type) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%q", t.String())), nil
}

func (t *Token1Type) UnmarshalJSON(bs []byte) error {
	switch string(bs) {
	case `"ERROR"`:
		*t = ErrorToken1
		return nil
	case `"PUNCTUATION"`:
		*t = PunctuationToken1
		return nil
	case `"WORD"`:
		*t = WordToken1
		return nil
	case `"STYLE"`:
		*t = StyleToken1
		return nil
	case `"GLUE"`:
		*t = GlueToken1
		return nil
	case `"OPAQUE"`:
		*t = OpaqueToken1
		return nil
	}
	return fmt.Errorf("unknown token type %q", bs)
}

type Token1 struct {
	Type  Token1Type
	Data  string
	Index int
}

func (t *Token1) Add(r rune) {
	t.Data += string(r)
}
