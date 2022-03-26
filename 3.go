package gba

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// var singlelineComments = regexp.MustCompile("/\/\/[^\r\n]*/")
// var multilineComments = regexp.MustCompile("/\\*[^*]*\\*+(?:[^/*][^*]*\\*+)*/")

func GBAReplacements(s string) string {
	s = strings.Replace(s, "¶ ⦊", "<div class='¶'>", -1)
	s = strings.Replace(s, "† ⦊", "<div class='†'>", -1)
	s = strings.Replace(s, "◇ ⦊", "<div class='◇'>", -1)
	s = strings.Replace(s, "‖", "<div class='‖'>", -1)
	s = strings.Replace(s, "⁝ ⦊", "<div class='⁝'>", -1)
	s = strings.Replace(s, "‣", "<div class='‣'>", -1)
	s = strings.Replace(s, "⦉", "</div>", -1)
	return s
}

const FragmentNode = "fragment"
const ParagraphNode = "¶"
const FootnoteNode = "†"
const DisplayMathNode = "◇"
const RunNode = "‖"
const TextNode = "text"
const TokenNode = "token"
const ListNode = "⁝"
const ListItemNode = "‣"

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
	if t.Type == OpaqueToken {
		return string(OpaqueOpenRune) + t.Value + string(OpaqueCloseRune)
	}
	out := t.Value
	if t.Implicit && isSpace(t) {
		out = " "
	}
	return out
}

func texval(t *Token) string {
	switch t.Type {
	case WordToken:
		return t.Value
	case PunctuationToken:
		switch r, _ := utf8.DecodeRuneInString(t.Value); r {
		case '‹':
			return "\\textit{"
		case '›':
			return "}"
		case '«':
			return "\textbf{"
		case '»':
			return "}"
		case '❬':
			return "\\t{"
		case '❭':
			return "}"
		case '❮':
			return "\\textbf{"
		case '❯':
			return "}"
		case '⧼':
			return "\\t{"
		case '⧽':
			return "}"
		case '“': //left
			return "\\say{"
		case '”': //right
			return "}"
		case '–': // en dash
			return "--"
		case '—': // em dash
			return "---"
		case '‘': // left
			return "`"
		case '’': // right
			return "'"
		case '᜶':
			return "\\\\"
		case '↦':
			return "{\\indent}"
		case '↤':
			return "{\\noindent}"
		}
	case SymbolToken:
		r, _ := utf8.DecodeRuneInString(t.Value)
		if replacement, ok := latexMathReplacements[r]; ok {
			return replacement
		}
		return t.Value
	case OpaqueToken:
		x := t.Value[1 : len(t.Value)-1]
		for r, to := range latexMathReplacements {
			x = strings.Replace(x, string(r), to, -1)
		}
		return x
	}

	return t.Value
}

func isSpace(t *Token) bool {
	return t.Type == PunctuationToken && t.Value == "·"
}

func lineBlocks(ts []*Token, v func(*Token) string, width int) []string {
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

	//	log.Print("PIECES")
	//	for _, p := range pieces {
	//		log.Printf("%q", p)
	//	}

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
				//							log.Printf("%q adding space?", nl)
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

func WriteGBA(w io.Writer, n *Node, prefix, indent string) {
	switch n.Type {
	case FragmentNode:
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			WriteGBA(w, c, prefix, indent)
		}
	case ParagraphNode, ListNode:
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

				var block []*Token = []*Token{c.Token}

				// in case its a token, go a find all tokens to next non-token
				var cc *Node
				for cc = c.NextSibling; cc != nil && cc.Type == TokenNode; cc = cc.NextSibling {
					block = append(block, cc.Token)
				}

				allowedWidth := maxWidth - offset
				lines := lineBlocks(block, val, allowedWidth)
				if len(lines) > 0 {
					writeLines(w, lines, prefix+indent, afterFirstLine)
					afterFirstLine = true
				}

				c = cc
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

		if n.LastChild.Type == TokenNode {
			w.Write([]byte(" "))
		}
		// will need to do overflow check
		w.Write([]byte("⦉"))
	case TextNode:
		lines := strings.Split(n.Data, "\n")
		for i, l := range lines {
			lines[i] = prefix + l
		}
		out := strings.Join(lines, "\n")
		w.Write([]byte(out + "\n"))
	}
}

func WriteTex(w io.Writer, n *Node, prefix, indent string) {
	switch n.Type {
	case FragmentNode:
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			WriteGBA(w, c, prefix, indent)
		}
	case ParagraphNode, ListNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}
		if n.PrevSibling != nil && (n.PrevSibling.Type == ParagraphNode || n.PrevSibling.Type == ListNode) {
			w.Write([]byte("\n"))
		}
		if n.Type == ParagraphNode {
			w.Write([]byte(prefix + "\n"))
		} else {
			w.Write([]byte(prefix + "\\begin{itemize}\n"))
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			WriteGBA(w, c, prefix+indent, indent)
		}

		if n.Type == ParagraphNode {
			w.Write([]byte(prefix + "\n"))
		} else {
			w.Write([]byte(prefix + "\\end{itemize}\n"))
		}

		// always puts this on the next node
		//		if n.NextSibling != nil {
		//			w.Write([]byte("\n"))
		//		}
	case FootnoteNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}
		w.Write([]byte(prefix + "\\footnote{\n"))
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			WriteGBA(w, c, prefix+indent, indent)
		}
		w.Write([]byte("\n" + prefix + "}"))
	case DisplayMathNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}
		w.Write([]byte(prefix + "\\[\n"))
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			WriteGBA(w, c, prefix+indent, indent)
		}
		w.Write([]byte("\n" + prefix + "\\]"))
	case RunNode, ListItemNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}

		if n.PrevSibling != nil && (n.PrevSibling.Type == RunNode || n.PrevSibling.Type == ListItemNode) {
			w.Write([]byte("\n"))
		}

		var out string
		if n.Type == RunNode {
			out = prefix + ""
		} else {
			out = prefix + "\\item "
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

				var block []*Token = []*Token{c.Token}

				// in case its a token, go a find all tokens to next non-token
				var cc *Node
				for cc = c.NextSibling; cc != nil && cc.Type == TokenNode; cc = cc.NextSibling {
					block = append(block, cc.Token)
				}

				allowedWidth := maxWidth - offset
				lines := lineBlocks(block, texval, allowedWidth)
				if len(lines) > 0 {
					writeLines(w, lines, prefix+indent, afterFirstLine)
					afterFirstLine = true
				}

				c = cc
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

		if n.LastChild.Type == TokenNode {
			w.Write([]byte(" "))
		}
		// will need to do overflow check
		w.Write([]byte("⦉"))
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

	for _, t := range ts {
		tn := &Node{Type: TokenNode, Token: t}
		tokens = append(tokens, tn)
	}
	return
}

func unmarshalHTML(in *html.Node, parent *Node) (*Node, error) {
	var n Node

	switch in.Type {
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
			switch c := class(in); {
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
	for _, n := range ns {
		/*
			if n.Type == html.ElementNode {
				log.Printf("%s:%s", n.DataAtom, class(n))
			} else if n.Type == html.TextNode {
				log.Printf("text")
			}
		*/
		fragment.AppendChild(n)
	}
	for c := fragment.FirstChild; c != nil; c = c.NextSibling {
		/*
			if c.Type == html.ElementNode {
				log.Printf("%s:%s", c.DataAtom, class(c))
			} else if c.Type == html.TextNode {
				log.Printf("text")
			}
		*/
	}
	//	html.Render(os.Stdout, &fragment)
	nGBA, err := UnmarshalHTML(&fragment)
	return nGBA, err
}

const OpaqueOpenRune = '❲'
const OpaqueCloseRune = '❳'

/*
const OpenMathOpaqueRune = '⧼'
const CloseMathOpaqueRune = '⧽'
*/

const WordToken = "word"
const PunctuationToken = "punctuation"
const SymbolToken = "symbol"
const OpaqueToken = "opaque"

type Token struct {
	Type                 string
	Value                string
	Implicit             bool
	StartLine, StartChar int
	width                int
}

func (t *Token) String() string {
	// return fmt.Sprintf("%s(%q)%d:%d", t.Type, t.Value, t.StartLine, t.StartChar)
	return fmt.Sprintf("%s(%q)", t.Type, val(t))
}

func LexText(s string) (tokens []*Token, err error) {
	//	s = strings.TrimSpace(s)
	//	lines := strings.Split(s, "\n")
	//	for i, l := range lines {
	//		lines[i] = strings.TrimSpace(l)
	//	}
	//	s = strings.Join(lines, " ")

	var opaque bool

	//	log.Printf("lexing %q", s)
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
				//				StartLine: line,
				//				StartChar: char,
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
					//StartLine: line,
					//StartChar: char,
				})
			}
		case r == '\n' || r == '\r' || r == '\t':
			continue
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
	fmt.Fprintf(w, prefix+"%s", n.Type)
	if n.Type == TokenNode {
		fmt.Fprintf(w, ":%v", n.Token)
	}
	fmt.Fprintf(w, "\n")
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		WriteDebug(w, c, prefix+indent, indent)
	}
}

// InsertBefore inserts newChild as a child of n, immediately before oldChild
// in the sequence of n's children. oldChild may be nil, in which case newChild
// is appended to the end of n's children.
//
// It will panic if newChild already has a parent or siblings.
func (n *Node) InsertBefore(newChild, oldChild *Node) {
	if newChild.Parent != nil || newChild.PrevSibling != nil || newChild.NextSibling != nil {
		panic("html: InsertBefore called for an attached child Node")
	}
	var prev, next *Node
	if oldChild != nil {
		prev, next = oldChild.PrevSibling, oldChild
	} else {
		prev = n.LastChild
	}
	if prev != nil {
		prev.NextSibling = newChild
	} else {
		n.FirstChild = newChild
	}
	if next != nil {
		next.PrevSibling = newChild
	} else {
		n.LastChild = newChild
	}
	newChild.Parent = n
	newChild.PrevSibling = prev
	newChild.NextSibling = next
}

// AppendChild adds a node c as a child of n.
//
// It will panic if c already has a parent or siblings.
func (n *Node) AppendChild(c *Node) {
	if c.Parent != nil || c.PrevSibling != nil || c.NextSibling != nil {
		panic("html: AppendChild called for an attached child Node")
	}
	last := n.LastChild
	if last != nil {
		last.NextSibling = c
	} else {
		n.FirstChild = c
	}
	n.LastChild = c
	c.Parent = n
	c.PrevSibling = last
}
