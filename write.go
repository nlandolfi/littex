package gba

import (
	"fmt"
	"io"
)

// this might be a mistake if we ever want to use these symbols!
const FootnoteRune = '†'
const RunRune = '‖'
const ParagraphRune = '¶'

func (f *Fragment) WriteTo(w io.Writer) (n int, err error) {
	return f.write(w, "")
}

func (f *Fragment) write(w io.Writer, prefix string) (n int, err error) {
	for i, p := range f.Paragraphs {
		if i != 0 {
			fmt.Fprint(w, "\n\n")
		}
		var nn int
		nn, err = p.write(w, prefix+" ")
		n += nn
		if err != nil {
			return
		}
	}
	return
}

func (p *Paragraph) write(w io.Writer, prefix string) (n int, er error) {
	nn, err := fmt.Fprint(w, "¶ {\n")
	n += nn
	if err != nil {
		return
	}
	for i, l := range p.Runs {
		if i != 0 {
			nn, err := fmt.Fprint(w, "\n")
			n += nn
			if err != nil {
				return
			}
		}
		nn, err := l.write(w, prefix) // for new lines
		n += nn
		if err != nil {
			return
		}
	}
	nn, err = fmt.Fprint(w, "}")
	n += nn
	if err != nil {
		return
	}
	return
}
func (l *Run) write(w io.Writer, prefix string) (n int, err error) {
	var charsOnLine = 0
	var nn int
	nn, err = fmt.Fprint(w, prefix+"‖")
	charsOnLine += nn
	n += nn
	if err != nil {
		return
	}
	//	log.Printf("%+v", l.Notes)
	var prev *Token
	for i, t := range l.Tokenized {
		if note := l.Notes[i]; note != nil { // assume can be index 0
			//log.Printf("writing")
			note.write(w, prefix+"")
			nn, err = fmt.Fprint(w, prefix+"  ")
			n += nn
			if err != nil {
				return
			}
			charsOnLine += nn
		}

		willWrite := t.TokenString(prev)
		prev = t
		nn, err = fmt.Fprint(w, willWrite)
		n += nn
		if err != nil {
			return
		}
		charsOnLine += nn
	}

	if note := l.Notes[len(l.Tokenized)]; note != nil {
		nn, err = note.write(w, prefix+"  ")
		n += nn
		if err != nil {
			return
		}
	}
	nn, err = fmt.Fprint(w, "\n")
	n += nn
	if err != nil {
		return
	}
	return
}

func (note *Note) write(w io.Writer, prefix string) (n int, err error) {
	//log.Printf("%+v", n.Runs)
	var nn int
	nn, err = fmt.Fprint(w, "\n"+prefix+" † {\n")
	n += nn
	if err != nil {
		return
	}
	for i, l := range note.Runs {
		if i != 0 {
			nn, err = fmt.Fprint(w, "\n")
			n += nn
			if err != nil {
				return
			}
		}
		nn, err = l.write(w, prefix+"  ")
		n += nn
		if err != nil {
			return
		}
	}
	nn, err = fmt.Fprint(w, prefix+"}\n")
	n += nn
	if err != nil {
		return
	}
	return
}

func (t *Token) TokenString(prev *Token) string {
	switch t.Type {
	case Error:
		panic("error token")
	case Word:
		if prev == nil || prev.Type == Word || prev.Type == Punctuation {
			return " " + t.Data
		}
	case Punctuation, Style:
		return t.Data
	}
	panic("not reached")
}
