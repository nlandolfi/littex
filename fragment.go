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
	Tokenized []*Token
	Notes     map[int]*Note
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
	l.Tokenized = append(l.Tokenized, t)
}

func (l *Run) AddNote(n *Note) {
	if l.Notes == nil {
		l.Notes = make(map[int]*Note)
	}
	l.Notes[n.Index] = n
}

func (l *Run) CurrentNote() *Note {
	return l.Notes[len(l.Tokenized)]
}

func (n *Note) AddRun(r *Run) {
	n.Runs = append(n.Runs, r)
}

func (n *Note) LastRun() *Run {
	return n.Runs[len(n.Runs)-1]
}

type TokenType uint

const (
	Error TokenType = iota
	Word
	Punctuation
	Style
	Glue
)

func (t *TokenType) String() string {
	switch *t {
	case Error:
		return "ERROR"
	case Punctuation:
		return "PUNCTUATION"
	case Word:
		return "WORD"
	case Style:
		return "STYLE"
	case Glue:
		return "GLUE"
	}
	panic("not reached")
}

func (t TokenType) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%q", t.String())), nil
}

func (t *TokenType) UnmarshalJSON(bs []byte) error {
	switch string(bs) {
	case `"ERROR"`:
		*t = Error
	case `"PUNCTUATION"`:
		*t = Punctuation
	case `"WORD"`:
		*t = Word
		return nil
	case `"STYLE"`:
		*t = Style
		return nil
	case `"GLUE"`:
		*t = Glue
		return nil
	}
	return fmt.Errorf("unknown token type %q", bs)
}

type Token struct {
	Type TokenType
	Data string
}

func (t *Token) Add(r rune) {
	t.Data += string(r)
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
	}
	panic("not reached")
}
