package lit

import (
	"fmt"
	"io"
	"log"
	"strings"
	"unicode/utf8"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

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

func UnmarshalHTML(in *html.Node) (*Node, error) {
	return unmarshalHTML(in, nil)
}

func unmarshalHTMLText(in *html.Node) (tokens []*Node, err error) {
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

			for c := in.FirstChild; c != nil; c = c.NextSibling {
				switch c.Type {
				case html.TextNode:
					ts, err := unmarshalHTMLText(c)
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

const OpaqueOpenRune = '❲'
const OpaqueCloseRune = '❳'

/*
const OpenMathOpaqueRune = '⧼'
const CloseMathOpaqueRune = '⧽'
*/

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
