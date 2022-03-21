package gba

import (
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

// this might be a mistake if we ever want to use these symbols!
const FootnoteRune = '†'
const RunRune = '‖'
const ParagraphRune = '¶'

func (f *Fragment) TexWriteTo(w io.Writer) (n int, err error) {
	return f.texWrite(w, "")
}

func (f *Fragment) WriteTo(w io.Writer) (n int, err error) {
	return f.write(w, "")
}

func (f *Fragment) texWrite(w io.Writer, prefix string) (n int, err error) {
	for i, p := range f.Paragraphs {
		if i != 0 {
			fmt.Fprint(w, "\n\n")
		}
		var nn int
		nn, err = p.texWrite(w, prefix+" ")
		n += nn
		if err != nil {
			return
		}
	}
	return
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

func (p *Paragraph) texWrite(w io.Writer, prefix string) (n int, er error) {
	//	nn, err := fmt.Fprint(w, "¶ {\n")
	//	n += nn
	//	if err != nil {
	//		return
	//	}
	for _, r := range p.Runs {
		//	if i != 0 {
		//			nn, err := fmt.Fprint(w, "\n")
		//		n += nn
		//		if err != nil {
		//			return
		//		}
		//	}
		nn, err := r.texWrite(w, prefix) // for new lines
		n += nn
		if err != nil {
			return
		}
	}
	//	nn, err = fmt.Fprint(w, "}")
	//	n += nn
	//	if err != nil {
	//		return
	//	}
	return
}

func (p *Paragraph) write(w io.Writer, prefix string) (n int, err error) {
	var nn int
	nn, err = fmt.Fprint(w, "¶ {\n")
	n += nn
	if err != nil {
		return
	}
	for i, l := range p.Runs {
		if i != 0 {
			nn, err = fmt.Fprint(w, "\n")
			n += nn
			if err != nil {
				return
			}
		}
		nn, err = l.write(w, prefix) // for new lines
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

func (r *Run) Pieces(insertions []int, prefix string) (outs []string) {
	var pieces [][]*Token
	var last int
	for _, i := range insertions {
		pieces = append(pieces, r.Tokens[last:i])
	}
	pieces = append(pieces, r.Tokens[last:])
	var nonempties [][]*Token
	for _, p := range pieces {
		if len(p) > 0 {
			nonempties = append(nonempties, p)
		}
	}
	for i, p := range nonempties {
		var spaced string
		for _, t := range p {
			spaced += t.Data
		}
		splits := strings.Split(spaced, " ")
		var o string = prefix
		for _, s := range splits {
			if len(o+s) > 70 {
				outs = append(outs, o)
				o = prefix
			}
			o += s
		}

		if m := r.Maths[insertions[i]]; m != nil {
			outs = append(outs, prefix+"◇ ")
			outs = append(outs, m.Token.Data)
		}
	}
	return
}

func (l *Run) texWrite(w io.Writer, prefix string) (n int, err error) {
	var charsOnLine = 0
	var nn int
	nn, err = fmt.Fprint(w, prefix+"")
	charsOnLine += nn
	n += nn
	if err != nil {
		return
	}
	//	log.Printf("%+v", l.Notes)
	var prev *Token
	for i, t := range l.Tokens {
		if note := l.Notes[i]; note != nil { // assume can be index 0
			//log.Printf("writing")
			note.texWrite(w, prefix+"")
			nn, err = fmt.Fprint(w, prefix+"  ")
			n += nn
			if err != nil {
				return
			}
			charsOnLine += nn
		}
		if math := l.Maths[i]; math != nil { // assume can be index 0
			//log.Printf("writing")
			math.texWrite(w, prefix+"")
			nn, err = fmt.Fprint(w, prefix+"  ")
			n += nn
			if err != nil {
				return
			}
			charsOnLine += nn
		}

		willWrite := t.TexTokenString(prev)
		prev = t
		nn, err = fmt.Fprint(w, willWrite)
		n += nn
		if err != nil {
			return
		}
		charsOnLine += nn
	}

	if note := l.Notes[len(l.Tokens)]; note != nil {
		nn, err = note.texWrite(w, prefix+"  ")
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
	for i, t := range l.Tokens {
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

	if note := l.Notes[len(l.Tokens)]; note != nil {
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

func (note *Note) texWrite(w io.Writer, prefix string) (n int, err error) {
	//log.Printf("%+v", n.Runs)
	var nn int
	nn, err = fmt.Fprint(w, "\n"+prefix+"\\footnote{\n")
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
		nn, err = l.texWrite(w, prefix+"  ")
		n += nn
		if err != nil {
			return
		}
	}
	nn, err = fmt.Fprint(w, prefix+"}")
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

func (t *Token) TexTokenString(prev *Token) string {
	switch t.Type {
	case ErrorToken:
		panic("error token")
		//	case WordToken:
		//		if prev == nil || prev.Type == WordToken || prev.Type == PunctuationToken {
		//			return " " + t.Data
		//		}
		//	case PunctuationToken, StyleToken, GlueToken, OpaqueToken:
		//		return t.Data
	case OpaqueToken:
		return t.Data[1 : len(t.Data)-1]
	case PunctuationToken:
		switch r, _ := utf8.DecodeRuneInString(t.Data); r {
		case '‹':
			return "\\textit{"
		case '›':
			return "}"
		case '«':
			return "\\textbf{"
		case '»':
			return "}"
		}
	}

	return t.Data
	//	panic("not reached")
}

func (t *Token) TokenString(prev *Token) string {
	switch t.Type {
	case ErrorToken:
		panic("error token")
	}

	return t.Data
	//	panic("not reached")
}
