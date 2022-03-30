package gba

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"strings"
	"unicode/utf8"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// GBAReplacements turns a .gba file into parseable HTML form.
func GBAReplacements(s string) string {
	s = strings.Replace(s, "¶⦊", "¶ ⦊", -1)
	s = strings.Replace(s, "†⦊", "† ⦊", -1)
	s = strings.Replace(s, "◇⦊", "◇ ⦊", -1)
	s = strings.Replace(s, "⁝⦊", "⁝ ⦊", -1)
	s = strings.Replace(s, "¶ ⦊", "<div class='¶'>", -1)
	s = strings.Replace(s, "† ⦊", "<div class='†'>", -1)
	s = strings.Replace(s, "◇ ⦊", "<div class='◇'>", -1)
	s = strings.Replace(s, "‖", "<div class='‖'>", -1)
	s = strings.Replace(s, "⁝ ⦊", "<div class='⁝'>", -1)
	s = strings.Replace(s, "‣", "<div class='‣'>", -1)
	s = strings.Replace(s, "⦉", "</div>", -1)
	return s
}

const maxWidth int = 74

func val(t *Token) string {
	if t.Type == OpaqueToken {
		return string(OpaqueOpenRune) + t.Value + string(OpaqueCloseRune)
	}
	out := t.Value
	if t.Implicit && isSpace(t) {
		out = " "
	}
	return out
}

func isSpace(t *Token) bool {
	return t.Type == PunctuationToken && t.Value == "·"
}

type tokenStringer func(*Token) string

func lineBlocks(ts []*Token, v tokenStringer, width int) []string {
	var pieces = []string{""}
	var spaces []*Token
	for _, t := range ts {
		if isSpace(t) {
			spaces = append(spaces, t)
			pieces = append(pieces, "")
		} else {
			pieces[len(pieces)-1] = pieces[len(pieces)-1] + v(t)
		}
	}

	var lines []string = []string{""}
	var curRuneCount = 0
	var onePieceOnLine bool
	for i, p := range pieces {
		var lastPiece = (len(pieces)-1 == i)

		c := utf8.RuneCountInString(p) + 1 // for the space
		if lastPiece {
			c -= 1 // except the last piece
		}

		if curRuneCount+c > width {
			lines = append(lines, p)
			curRuneCount = c
			onePieceOnLine = true
		} else {
			nl := lines[len(lines)-1]
			if onePieceOnLine {
				nl += val(spaces[i-1])
			}
			nl += p
			if !lastPiece {
				nl += val(spaces[i])
			}
			lines[len(lines)-1] = nl
			curRuneCount += c + 1
			onePieceOnLine = false
		}
	}

	var o []string = make([]string, 0, len(lines))
	for _, line := range lines {
		// these are sort of quick fix hacks
		// to avoid end of line spaces and
		// empty lines, but really the code above
		// should ultimately be fixed
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		o = append(o, line)
	}

	return o
}

func writeLines(w io.Writer, lines []string, prefix string, prefixFirst bool) {
	for i, line := range lines {
		lastLine := (i == len(lines)-1)
		if i == 0 && !prefixFirst {
			w.Write([]byte(line))
		} else {
			w.Write([]byte(prefix + line))
		}

		if !lastLine {
			w.Write([]byte("\n"))
		}
	}
}

func tokenBlockStartingAt(c *Node) (block []*Token, last *Node) {
	block = append(block, c.Token)
	last = c

	var cc *Node
	for cc = c.NextSibling; cc != nil && cc.Type == TokenNode; cc = cc.NextSibling {
		block = append(block, cc.Token)
		last = cc
	}
	return block, last
}

func WriteGBA(w io.Writer, n *Node, prefix, indent string) {
	switch n.Type {
	case FragmentNode:
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			WriteGBA(w, c, prefix, indent)
		}
	case ParagraphNode, ListNode:
		if n.FirstChild == nil {
			log.Printf("skipping empty paragraph or list")
			return // just skip!
		}
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}
		if n.PrevSibling != nil && (n.PrevSibling.Type == ParagraphNode || n.PrevSibling.Type == ListNode) {
			w.Write([]byte("\n"))
		}
		if n.Type == ParagraphNode {
			w.Write([]byte(prefix + "¶ ⦊\n"))
		} else {
			w.Write([]byte(prefix + "⁝ ⦊\n"))
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			WriteGBA(w, c, prefix+indent, indent)
		}
		w.Write([]byte("\n" + prefix + "⦉"))

		// always puts this on the next node
		//		if n.NextSibling != nil {
		//			w.Write([]byte("\n"))
		//		}
	case FootnoteNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}
		w.Write([]byte(prefix + "† ⦊\n"))
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			WriteGBA(w, c, prefix+indent, indent)
		}
		w.Write([]byte("\n" + prefix + "⦉"))
	case DisplayMathNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}
		w.Write([]byte(prefix + "◇ ⦊\n"))
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			WriteGBA(w, c, prefix+indent, indent)
		}
		w.Write([]byte("\n" + prefix + "⦉"))
	case RunNode, ListItemNode:
		if n.FirstChild == nil {
			log.Printf("skipping empty run")
			return // just skip!
		}
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}

		if n.PrevSibling != nil && (n.PrevSibling.Type == RunNode || n.PrevSibling.Type == ListItemNode) {
			w.Write([]byte("\n"))
		}

		var out string
		if n.Type == RunNode {
			out = prefix + "‖ "
		} else {
			out = prefix + "‣ "
		}

		w.Write([]byte(out))

		offset := utf8.RuneCountInString(out)
		//log.Printf("offset; %d", offset)

		//		lineBuffer = ""
		//relOffset := 0

		var afterFirstLine bool
		// Looping over the children.
		for c := n.FirstChild; c != nil; afterFirstLine = true {
			//log.Printf("RUN NODE: %s", c.Type)
			switch c.Type {
			case TokenNode:
				if c.PrevSibling != nil {
					w.Write([]byte("\n"))
				}

				// in case its a token, go a find all tokens to next non-token
				block, lastTokenNode := tokenBlockStartingAt(c)
				allowedWidth := maxWidth - offset
				lines := lineBlocks(block, val, allowedWidth)
				if len(lines) > 0 {
					writeLines(w, lines, prefix+indent, afterFirstLine)
					afterFirstLine = true
				}

				c = lastTokenNode.NextSibling
			default:
				//	log.Print("DEFAU⦊")
				// no need, WriteGBA should do this new line
				//				if c.PrevSibling.Type == TokenNode {
				//					w.Write([]byte("\n"))
				//				}
				WriteGBA(w, c, prefix+indent, indent)
				c = c.NextSibling
			}
		}

		if n.LastChild != nil && n.LastChild.Type == TokenNode {
			w.Write([]byte(" "))
		}
		// will need to do overflow check
		w.Write([]byte("⦉"))
	case TextNode:
		log.Printf("text nodes should not appear...")
		lines := strings.Split(n.Data, "\n")
		for i, l := range lines {
			lines[i] = prefix + l
		}
		out := strings.Join(lines, "\n")
		w.Write([]byte(out + "\n"))
	default:
		log.Printf("prev: %v; cur: %v; next: %v", n.PrevSibling, n, n.NextSibling)
		panic(fmt.Sprintf("unhandled node type: %s", n.Type))
	}
}

func writeKids(
	w io.Writer, n *Node, in, pr string,
	write func(w io.Writer, n *Node, pr, in string),
) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		write(w, c, pr, in)
	}
}

func WriteTex(w io.Writer, n *Node, prefix, indent string) {
	switch n.Type {
	case FragmentNode:
		writeKids(w, n, prefix, indent, WriteTex)
	case ParagraphNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}
		writeKids(w, n, prefix+indent, indent, WriteTex)
		w.Write([]byte("\n"))
	case ListNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}
		w.Write([]byte(prefix + "\\begin{itemize}\n"))
		writeKids(w, n, prefix+indent, indent, WriteTex)
		w.Write([]byte("\n" + prefix + "\\end{itemize}"))
	case FootnoteNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}
		// the first little bit here removes the space.
		// it means there is no way to have a space in between
		// text and a footnote, but I don't think there's a way
		// to specify that within .gba files.
		// Could in the future check if the previous child here was a token
		// and a non implicit space token
		w.Write([]byte(prefix + "\\ifhmode\\unskip\\fi\\footnote{\n"))
		writeKids(w, n, prefix+indent, indent, WriteTex)
		w.Write([]byte("\n" + prefix + "}"))
	case DisplayMathNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}
		w.Write([]byte(prefix + "\\[\n"))
		writeKids(w, n, prefix+indent, indent, WriteTex)
		w.Write([]byte("\n" + prefix + "\\]"))
	case RunNode, ListItemNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}

		var offset int
		if n.Type == ListItemNode {
			out := prefix + "\\item "
			w.Write([]byte(out))
			offset = utf8.RuneCountInString(out)
		}

		var afterFirstLine bool
		// Looping over the children.
		for c := n.FirstChild; c != nil; afterFirstLine = true {
			//log.Printf("RUN NODE: %s", c.Type)
			switch c.Type {
			case TokenNode:
				if c.PrevSibling != nil && c.PrevSibling.Type != RunNode {
					w.Write([]byte("\n"))
				}

				// in case its a token, go a find all tokens to next non-token
				block, lastTokenNode := tokenBlockStartingAt(c)
				allowedWidth := maxWidth - offset
				lines := lineBlocks(block, Tex, allowedWidth)
				if len(lines) > 0 {
					writeLines(w, lines, prefix+indent, afterFirstLine)
					afterFirstLine = true
				}

				c = lastTokenNode.NextSibling
			default:
				WriteTex(w, c, prefix+indent, indent)
				c = c.NextSibling
			}
		}
	default:
		panic(fmt.Sprintf("unhandled node type: %s", n.Type))
	}
}

func UnmarshalHTML(in *html.Node) (*Node, error) {
	return unmarshalHTML(in, nil)
}

func unmarshalText(in *html.Node) (tokens []*Node, err error) {
	if in.Type != html.TextNode {
		panic("die")
	}

	var ts []*Token
	ts, err = Lex(in.Data)
	if err != nil {
		return
	}

	for _, t := range ts {
		tn := &Node{Type: TokenNode, Token: t}
		tokens = append(tokens, tn)
	}
	return
}

func unmarshalHTML(in *html.Node, parent *Node) (*Node, error) {
	var n Node

	switch in.Type {
	case html.CommentNode:
		return nil, nil // for now
	case html.TextNode:
		log.Printf("warning, unexpected text node")
		if strings.TrimSpace(in.Data) == "" {
			return nil, nil
		}
		n.Type = TextNode
		n.Data = in.Data
		return &n, nil
	case html.ElementNode:
		switch in.DataAtom {
		case atom.Div:
			switch c := classOf(in); {
			case c == "¶":
				n.Type = ParagraphNode
			case c == "‖":
				n.Type = RunNode
			case c == "◇":
				n.Type = DisplayMathNode
			case c == "†":
				n.Type = FootnoteNode
			case c == "⁝":
				n.Type = ListNode
			case c == "‣":
				n.Type = ListItemNode
			case c == "fragment":
				n.Type = FragmentNode
			}
			//log.Printf("CLASS: %s", class(in))

			for c := in.FirstChild; c != nil; c = c.NextSibling {
				switch c.Type {
				case html.TextNode:
					ts, err := unmarshalText(c)
					if err != nil {
						return nil, err
					}

					for _, child := range ts {
						n.AppendChild(child)
					}
				default:
					child, err := unmarshalHTML(c, &n)
					if err != nil {
						return nil, err
					}
					if child != nil {
						n.AppendChild(child)
					}
				}
			}
		default:
			return nil, fmt.Errorf("unsupported ElementNode DataAtom: %s", in.DataAtom)
		}
	default:
		return nil, fmt.Errorf("unsupported node type: %d", in.Type)
	}

	return &n, nil
}

func ParseHTML(s string) (*Node, error) {
	var fragment html.Node = html.Node{
		Type:     html.ElementNode,
		DataAtom: atom.Div,
		Data:     "div",
		Attr: []html.Attribute{
			html.Attribute{
				Key: "class", Val: "fragment",
			},
		},
	}

	ns, err := html.ParseFragment(bytes.NewBufferString(s), &fragment)
	if err != nil {
		return nil, err
	}
	for _, n := range ns {
		fragment.AppendChild(n)
	}
	nGBA, err := UnmarshalHTML(&fragment)
	return nGBA, err

}

func ParseGBA(s string) (*Node, error) {
	s = GBAReplacements(s)
	return ParseHTML(s)
}

const OpaqueOpenRune = '❲'
const OpaqueCloseRune = '❳'

/*
const OpenMathOpaqueRune = '⧼'
const CloseMathOpaqueRune = '⧽'
*/

func WriteDebug(w io.Writer, n *Node, prefix, indent string) {
	fmt.Fprintf(w, prefix+"%s", n.Type)
	if n.Type == TokenNode {
		fmt.Fprintf(w, ":%v", n.Token)
	}
	fmt.Fprintf(w, "\n")
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		WriteDebug(w, c, prefix+indent, indent)
	}
}

// Slides
func (n *Node) FirstTokenString() string {
	if n.Type != ListItemNode {
		panic("SlideTitle only for list items")
	}
	if n.FirstChild == nil {
		return ""
	}
	block, _ := tokenBlockStartingAt(n.FirstChild)
	lines := lineBlocks(block, Tex, maxWidth)
	if len(lines) > 1 {
		panic("SlideTitle multi-line")
	}
	return lines[0]
}

func (n *Node) FirstListNode() *Node {
	// kids of the first ⁝ node
	if n.Type != ListItemNode {
		panic("SlideItems only for list items")
	}
	c := n.FirstChild
	for c != nil && c.Type != ListNode {
		c = c.NextSibling
	}
	return c
}
