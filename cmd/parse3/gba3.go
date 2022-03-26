package main

import (
	"bytes"
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
var mode = flag.String("m", "gba", "mode")

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

const maxWidth int = 74

func val(t *Token) string {
	out := t.Value
	if t.Implicit && isSpace(t) {
		out = " "
	}
	return out
}

func isSpace(t *Token) bool {
	return t.Type == PunctuationToken && t.Value == "·"
}

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
		//log.Printf("offset; %d", offset)

		//		lineBuffer = ""
		//relOffset := 0

		var afterFirstLine bool
		for c := n.FirstChild; c != nil; afterFirstLine = true {
			switch c.Type {
			case TokenNode:

				var block []*Token = []*Token{c.Token}

				// in case its a token, go a find all tokens to next non-token
				var cc *Node
				for cc = c.NextSibling; cc != nil && cc.Type == TokenNode; cc = cc.NextSibling {
					block = append(block, cc.Token)
				}

				var pieces = []string{""}
				var spaces []*Token
				for _, t := range block {
					if isSpace(t) {
						spaces = append(spaces, t)
						pieces = append(pieces, "")
					} else {
						pieces[len(pieces)-1] = pieces[len(pieces)-1] + val(t)
					}
				}

				//	log.Print("PIECES")
				//	for _, p := range pieces {
				//		log.Printf("%q", p)
				//	}

				allowedWidth := maxWidth - offset
				//log.Printf("allowed width: %d", allowedWidth)
				var lines []string = []string{""}
				var curRuneCount = 0
				var onePieceOnLine bool
				for i, p := range pieces {
					var lastPiece = (len(pieces)-1 == i)

					c := utf8.RuneCountInString(p) + 1 // for the space
					if lastPiece {
						c -= 1 // except the last piece
					}

					if curRuneCount+c > allowedWidth {
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
							//							log.Printf("%q adding space?", nl)
							nl += val(spaces[i])
						}
						lines[len(lines)-1] = nl
						curRuneCount += c + 1
						onePieceOnLine = false
					}
				}

				//log.Print("LINES")
				//for _, p := range lines {
				//		log.Printf("%q", p)
				//	}

				for i, line := range lines {
					lastLine := (i == len(lines)-1)
					if afterFirstLine {
						w.Write([]byte(prefix + indent + line))
					} else {
						w.Write([]byte(line))
					}
					if !lastLine {
						w.Write([]byte("\n"))
					}
					afterFirstLine = true
				}

				c = cc
			default:
				if c.PrevSibling.Type == TokenNode {
					w.Write([]byte("\n"))
				}
				WriteGBA(w, n, prefix+indent, indent)
				c = c.NextSibling
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

	log.Print(n)

	switch *mode {
	/* doesn't work cause node points have cycles
	case "json":
		bs, err = json.MarshalIndent(n, "", "  ")
		if err != nil {
			log.Fatal(err)
		}

		fmt.Fprint(os.Stdout, string(bs))
	*/
	case "tex":
		panic("d")
	case "debug":
		WriteDebug(os.Stdout, n, "", "  ")
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

func WriteDebug(w io.Writer, n *Node, prefix, indent string) {
	fmt.Fprintf(w, "%s", n.Type)
	if n.Type == TokenNode {
		fmt.Fprintf(w, ":%v", n.Token)
	}
	fmt.Fprintf(w, "\n")
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		WriteDebug(w, c, prefix+indent, indent)
	}
}
