package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var in = flag.String("in", "text.gba", "in file")
var mode = flag.String("m", "json", "mode")

// var singlelineComments = regexp.MustCompile("/\/\/[^\r\n]*/")
// var multilineComments = regexp.MustCompile("/\\*[^*]*\\*+(?:[^/*][^*]*\\*+)*/")

func GBAReplacements(s string) string {
	s = strings.Replace(s, "¶ ⦊", "<div class='¶'>", -1)
	s = strings.Replace(s, "† ⦊", "<div class='†'>", -1)
	s = strings.Replace(s, "◇ ⦊", "<div class='◇'>", -1)
	s = strings.Replace(s, "‖", "<div class='‖'>", -1)
	s = strings.Replace(s, "⦉", "</div>", -1)
	return s
}

const FragmentNode = "fragment"
const ParagraphNode = "¶"
const FootnoteNode = "†"
const DisplayMathNode = "◇"
const RunNode = "‖"
const TextNode = "text"
const TokenNode = "?"

type Node struct {
	Type  string
	Data  string
	Attr  []html.Attribute
	Token *Token

	Parent                   *Node `json:"-"`
	FirstChild, LastChild    *Node
	PrevSibling, NextSibling *Node
}

func class(n *html.Node) string {
	for _, a := range n.Attr {
		if a.Key == "class" {
			return a.Val
		}
	}

	return ""
}

const maxWidth int = 70 - 2

func WriteGBA(w io.Writer, n *Node, prefix, indent string) {
	switch n.Type {
	case FragmentNode:
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			WriteGBA(w, c, prefix, indent)
		}
	case ParagraphNode:
		w.Write([]byte(prefix + "¶ ⦊\n"))
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			WriteGBA(w, c, prefix+indent, indent)
		}
		w.Write([]byte(prefix + "⦉\n"))
		if n.NextSibling != nil {
			w.Write([]byte("\n"))
		}
	case FootnoteNode:
		w.Write([]byte(prefix + "† ⦊\n"))
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			WriteGBA(w, c, prefix+indent, indent)
		}
		w.Write([]byte(prefix + "⦉\n"))
	case DisplayMathNode:
		w.Write([]byte(prefix + "† ⦊\n"))
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			WriteGBA(w, c, prefix+indent, indent)
		}
		w.Write([]byte(prefix + "⦉\n"))
	case RunNode:
		out := prefix + "‖ "

		w.Write([]byte(out))

		offset := utf8.RuneCountInString(out)

		relOffset := 0

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			switch c.Type {
			case TokenNode:
				runes := utf8.RuneCountInString(c.Token.Value)
				if relOffset != 0 && runes+relOffset+offset > maxWidth {
					w.Write([]byte("\n" + prefix + indent))
					relOffset = 0
				}
				w.Write([]byte(c.Token.Value))
				relOffset += runes
			default:
				if c.PrevSibling.Type == TokenNode {
					w.Write([]byte("\n"))
				}
				WriteGBA(w, n, prefix+indent, indent)
				relOffset = 0
			}
		}

		// will need to do overflow check
		w.Write([]byte(" ⦉\n"))
		if n.NextSibling != nil && n.NextSibling.Type == RunNode {
			w.Write([]byte("\n"))
		}
	case TextNode:
		lines := strings.Split(n.Data, "\n")
		for i, l := range lines {
			lines[i] = prefix + l
		}
		out := strings.Join(lines, "\n")
		w.Write([]byte(out + "\n"))
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
	ts, err = LexText(in.Data)
	if err != nil {
		return
	}

	//	var prev *Node
	for _, t := range ts {
		tn := &Node{Type: TokenNode, Token: t}
		tokens = append(tokens, tn)
	}
	return
}

/*
		tn := &Node{Type: TokenNode, Token: t}

		child.Parent = parent
		child.PrevSibling = prev

		if prev == nil { // dont check in.LastChild, cause we skip some nodes
			n.FirstChild = child
		} else {
			prev.NextSibling = child
		}

		prev = child

		tn.PrevSibling = prev
		if prev != nil {
			prev.NextSibling = tn
		}
		if i == 0 {
			parent.FirstChild = tn
		}
		if i == len(tokens)-1 {
			parent.LastChild = tn
		}
		prev = tn
	}
*/

func unmarshalHTML(in *html.Node, parent *Node) (*Node, error) {
	var n Node

	switch in.Type {
	case html.TextNode:
		/*
			if strings.TrimSpace(in.Data) == "" {
				return nil, nil
			}
			n.Type = TextNode
			n.Data = in.Data
			return &n, nil
		*/
	case html.ElementNode:
		switch in.DataAtom {
		case atom.Div:
			switch c := class(in); {
			case c == "¶":
				n.Type = ParagraphNode
			case c == "‖":
				n.Type = RunNode
			case c == "◇":
				n.Type = DisplayMathNode
			case c == "†":
				n.Type = FootnoteNode
			case c == "fragment":
				n.Type = FragmentNode
			}

			var prev *Node
			for c := in.FirstChild; c != nil; c = c.NextSibling {
				if c.Type == html.TextNode {
					ts, err := unmarshalText(c)
					if err != nil {
						return nil, err
					}

					for _, child := range ts {
						child.Parent = &n
						child.PrevSibling = prev

						if prev == nil { // dont check in.LastChild, cause we skip some nodes
							n.FirstChild = child
						} else {
							prev.NextSibling = child
						}

						prev = child
					}
				} else {
					child, err := unmarshalHTML(c, &n)
					if err != nil {
						return nil, err
					}
					if child == nil { // for example an empty text node
						continue
					}

					child.Parent = &n
					child.PrevSibling = prev

					if prev == nil { // dont check in.LastChild, cause we skip some nodes
						n.FirstChild = child
					} else {
						prev.NextSibling = child
					}

					prev = child
				}
			}
			n.LastChild = prev
		default:
			return nil, fmt.Errorf("unsupported ElementNode DataAtom: %s", in.DataAtom)
		}
	default:
		return nil, fmt.Errorf("unsupported node type: %s", in.Type)
	}

	return &n, nil
}

func Parse3(s string) (*Node, error) {
	s = GBAReplacements(s)
	//fmt.Fprint(os.Stdout, s)

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
	for i, n := range ns {
		if i == 0 {
			fragment.FirstChild = n
		}
		if i == len(ns)-1 {
			fragment.LastChild = n
		}
		n.Parent = &fragment
	}
	//	html.Render(os.Stdout, &fragment)
	nGBA, err := UnmarshalHTML(&fragment)
	return nGBA, err
}

func main() {
	flag.Parse()
	bs, err := os.ReadFile(*in)
	if err != nil {
		log.Fatal(err)
	}

	n, err := Parse3(string(bs))
	if err != nil {
		log.Fatal(err)
	}

	switch *mode {
	case "json":
		bs, err = json.MarshalIndent(n, "", "  ")
		if err != nil {
			log.Fatal(err)
		}

		fmt.Fprint(os.Stdout, string(bs))
	case "tex":
		panic("d")
	case "gba":
		WriteGBA(os.Stdout, n, "", "  ")
	}
}

const OpenOpaqueRune = '❲'
const CloseOpaqueRune = '❳'

/*
const OpenMathOpaqueRune = '⧼'
const CloseMathOpaqueRune = '⧽'
*/

const WordToken = "word"
const PunctuationToken = "punctuation"
const OpaqueToken = "opaque"
const SymbolToken = "symbol"

type Token struct {
	Type                 string
	Value                string
	Implicit             bool
	StartLine, StartChar int
	width                int
}

func (t *Token) String() string {
	return fmt.Sprintf("%s(%q)%d:%d", t.Type, t.Value, t.StartLine, t.StartChar)
}

func LexText(s string) (tokens []*Token, err error) {
	s = strings.TrimSpace(s)
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = strings.TrimSpace(l)
	}
	s = strings.Join(lines, " ")

	var opaque bool

	for _, r := range s {
		if opaque {
			if r == CloseOpaqueRune {
				opaque = false
				continue
			}
			tokens[len(tokens)-1].Value += string(r)
			continue
		}

		switch {
		case r == OpenOpaqueRune:
			tokens = append(tokens, &Token{
				Type:  OpaqueToken,
				Value: "",
				//				StartLine: line,
				//				StartChar: char,
			})
			opaque = true
		case r == ' ' && len(tokens) > 0: // unicode.IsSpace(r):
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
					//StartLine: line,
					//StartChar: char,
				})
			}
		case unicode.IsSymbol(r):
			tokens = append(tokens, &Token{
				Type:  SymbolToken,
				Value: string(r),
				//StartLine: line,
				//StartChar: char,
			})
		case unicode.IsPunct(r):
			tokens = append(tokens, &Token{
				Type:  PunctuationToken,
				Value: string(r),
				//StartLine: line,
				//StartChar: char,
			})
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			if len(tokens) == 0 || tokens[len(tokens)-1].Type != WordToken { // start a new word
				tokens = append(tokens, &Token{
					Type:  WordToken,
					Value: string(r),
					//	StartLine: line,
					//	StartChar: char,
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

type TokenList []*Token

func (ts TokenList) String() string {
	var s strings.Builder

	for _, t := range ts {
		s.WriteString(t.Value)
	}

	return s.String()
}

func ToGBAFormat(indent, max int, tokens []*Token) string {
	return TokenList(tokens).String()
}
