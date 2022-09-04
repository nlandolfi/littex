package lit

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"regexp"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func ParseHTML(s string) (*Node, error) {
	var fragment html.Node = html.Node{
		Type:     html.ElementNode,
		DataAtom: atom.Div,
		Data:     "div",
		Attr: []html.Attribute{
			html.Attribute{
				Key: "data-littype", Val: "fragment",
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

func ParseLit(s string) (*Node, error) {
	/*
		var b bytes.Buffer

		var prev rune
		var left bool = truek
		for _, r := range s {
			if r == '$' && prev != '\\' {
				if left {
					b.WriteRune('⟅')
					left = false
				} else {
					b.WriteRune('⟆')
					left = true
				}
			} else {
				b.WriteRune(r)
			}
			prev = r
		}
	*/

	s = litReplace(s)
	return ParseHTML(s)
}

func litReplace(s string) string {
	s = strings.Replace(s, "¶⦊", "¶ ⦊", -1)
	s = strings.Replace(s, "†⦊", "† ⦊", -1)
	s = strings.Replace(s, "◇⦊", "◇ ⦊", -1)
	s = strings.Replace(s, "⁝⦊", "⁝ ⦊", -1)
	s = strings.Replace(s, "𝍫⦊", "𝍫 ⦊", -1)
	s = strings.Replace(s, "§⦊", "§ ⦊", -1)
	s = strings.Replace(s, "¶ ⦊", "<div data-littype='"+ParagraphClass+"'>", -1)
	s = strings.Replace(s, "† ⦊", "<div data-littype='"+FootnoteClass+"'>", -1)
	s = strings.Replace(s, "◇ ⦊", "<div data-littype='"+DisplayMathClass+"'>", -1)
	s = strings.Replace(s, "‖", "<div data-littype='"+RunClass+"'>", -1)
	s = strings.Replace(s, "⁝ ⦊", "<div data-littype='"+ListClass+"' data-litlisttype='unordered'>", -1)
	s = strings.Replace(s, "𝍫 ⦊", "<div data-littype='"+ListClass+"' data-litlisttype='ordered'>", -1)
	s = strings.Replace(s, "‣", "<div data-littype='"+ListItemClass+"'>", -1)
	s = strings.Replace(s, "#§§§", "<div data-littype='"+SectionClass+"' data-litsectionlevel='3' data-litsectionnumbered='true'>", -1)
	s = strings.Replace(s, "#§§", "<div data-littype='"+SectionClass+"' data-litsectionlevel='2' data-litsectionnumbered='true'>", -1)
	s = strings.Replace(s, "#§", "<div data-littype='"+SectionClass+"' data-litsectionlevel='1' data-litsectionnumbered='true'>", -1)
	s = strings.Replace(s, "§§§", "<div data-littype='"+SectionClass+"' data-litsectionlevel='3' data-litsectionnumbered='false'>", -1)
	s = strings.Replace(s, "§§", "<div data-littype='"+SectionClass+"' data-litsectionlevel='2' data-litsectionnumbered='false'>", -1)
	s = strings.Replace(s, "§", "<div data-littype='"+SectionClass+"' data-litsectionlevel='1' data-litsectionnumbered='false'>", -1)
	s = strings.Replace(s, "⦉", "</div>", -1)
	//	s = strings.Replace(s, "\\S", "§", -1)
	re := regexp.MustCompile(`\[(.+?)\]\((.+?)\)`)
	s = re.ReplaceAllString(s, `<a href='$2'> ‖ $1 ⦉</a>`)
	return s
}

func ParseTex(s string) (*Node, error) {
	for _, c := range commentsR.FindAllString(s, -1) {
		log.Printf("dropping comment: %q", c)
	}

	for _, r := range order {
		replace := res[r]
		s = r.ReplaceAllString(s, replace)
	}
	s = strings.Replace(s, "\\item", "‣", -1)
	s = strings.Replace(s, "\\begin{itemize}", "⁝ ⦊", -1)
	s = strings.Replace(s, "\\begin{enumerate}", "𝍫 ⦊", -1)
	s = strings.Replace(s, "\\end{itemize}", "⦉", -1)
	s = strings.Replace(s, "\\end{enumerate}", "⦉", -1)
	s = strings.Replace(s, "\\[", "◇ ⦊ ‖ ", -1)
	s = strings.Replace(s, "\\]", "⦉", -1)
	s = strings.Replace(s, "---", "—", -1)
	s = strings.Replace(s, "``", "“", -1)
	s = strings.Replace(s, "''", "”", -1)
	s = strings.Replace(s, "`", "‘", -1) // MUST BE AFTER DOUBLE
	// s = strings.Replace(s, "'", "’", -1)
	s = strings.Replace(s, "\\&", "&", -1)
	s = strings.Replace(s, "\\\\", "᜶", -1)
	s = strings.Replace(s, "\\indent", "↦", -1)
	s = strings.Replace(s, "\\noindent", "↤", -1)

	for from, to := range LatexMathReplacements {
		s = strings.Replace(s, to, string(from), -1)
	}

	// TODO better comments handling
	s = strings.Replace(s, "\\%", "%", -1)

	var b bytes.Buffer
	w := &b
	ps := strings.Split(s, "\n\n")
	for _, p := range ps {
		fmt.Fprintf(w, "¶ ⦊")
		ls := strings.Split(p, "\n")
		for _, l := range ls {
			fmt.Fprintf(w, "‖ ")
			if len(l) > 0 && l[0] == '%' { // comments
				fmt.Fprintf(w, "❲%s❳", l)
			} else {
				fmt.Fprint(w, l)
			}
			fmt.Fprintf(w, "⦉")
		}
		fmt.Fprintf(w, "⦉")
	}

	return ParseLit(b.String())
}

var textitR = regexp.MustCompile(`\\textit{((.|\n)*?)}`)
var textbfR = regexp.MustCompile(`\\textbf{((.|\n)*?)}`)
var textscR = regexp.MustCompile(`\\textsc{((.|\n)*?)}`)
var footnoteR = regexp.MustCompile(`\\footnote{((.|\n)*?)}`)
var tR = regexp.MustCompile(`\\t{((.|\n)*?)}`)
var cR = regexp.MustCompile(`\\c{((.|\n)*?)}`)
var dblqR = regexp.MustCompile("``((.|\n)*)?''")
var sglqR = regexp.MustCompile("`((.|\n)*)?'")
var sayR = regexp.MustCompile(`\\say{((.|\n)*)?}`)
var commentsR = regexp.MustCompile(`%(.*?)\n`)
var propositionWithText = regexp.MustCompile(`\\begin{proposition}\[(\w*)\]`)
var proposition = regexp.MustCompile(`\\begin{proposition}`)
var propositionEnd = regexp.MustCompile(`\\end{proposition}`)
var ssection = regexp.MustCompile(`\\ssection{(\w*)}`)
var section = regexp.MustCompile(`\\section{(\w*)}`)
var ssubsection = regexp.MustCompile(`\\ssubsection{(\w*)}`)
var subsection = regexp.MustCompile(`\\subsection{(\w*)}`)

// useful: https://gist.github.com/claybridges/8f9d51a1dc365f2e64fa
var res = map[*regexp.Regexp]string{
	propositionWithText: "<statement type='proposition' text='$1'>",
	proposition:         "<statement type='proposition'>",
	propositionEnd:      "</statement>",
	ssection:            "§ $1 ⦉",
	section:             "#§ $1 ⦉",
	subsection:          "#§§ $1 ⦉",
	ssubsection:         "§§ $1 ⦉",
	textitR:             "‹$1›",
	textbfR:             "«$1»",
	footnoteR:           "† ⦊ ‖ $1 ⦉⦉",
	textscR:             "⸤$1⸥",
	tR:                  "❬$1❭",
	cR:                  "⁅$1⁆",
	dblqR:               "“$1”",
	sglqR:               "‘$1’",
	sayR:                "“$1”",
}

var order = []*regexp.Regexp{
	ssection,
	section,
	ssubsection,
	subsection,
	textitR,
	textbfR,
	footnoteR,
	textscR,
	tR,
	cR,
	dblqR,
	sglqR,
	sayR,
	propositionWithText,
	proposition,
	propositionEnd,
}

// func MarshalHTML(n *Node) *html.Node

func UnmarshalHTML(in *html.Node) (*Node, error) {
	return unmarshalHTML(in, nil)
}

func unmarshalHTMLText(in *html.Node) (tokens []*Node, err error) {
	if in.Type != html.TextNode {
		panic("lit.unmarshalHTMLText called on non-text node")
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
		if strings.TrimSpace(in.Data) == "" {
			return nil, nil
		}

		n.Type = CommentNode
		n.Data = in.Data
		return &n, nil
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
		case atom.Center:
			n.Type = CenterAlignNode
		case atom.Img:
			n.Type = ImageNode
			n.setAttr("src", getAttr(in.Attr, "src"))
			n.setAttr("width", getAttr(in.Attr, "width"))
		case atom.A:
			n.Type = LinkNode
			n.setAttr("href", getAttr(in.Attr, "href"))
		case atom.Table:
			n.Type = TableNode
			n.Attr = copyAttr(in.Attr)
		case atom.Thead:
			n.Type = TableHeadNode
			n.Attr = copyAttr(in.Attr)
		case atom.Tbody:
			n.Type = TableBodyNode
			n.Attr = copyAttr(in.Attr)
		case atom.Tr:
			n.Type = TableRowNode
			n.Attr = copyAttr(in.Attr)
		case atom.Th:
			n.Type = THNode
			n.Attr = copyAttr(in.Attr)
		case atom.Td:
			n.Type = TDNode
			n.Attr = copyAttr(in.Attr)
		case atom.Code:
			n.Type = CodeNode
			n.Attr = copyAttr(in.Attr)
		case atom.Div:
			switch c := littypeOf(in); {
			case c == ParagraphClass:
				n.Type = ParagraphNode
			case c == RunClass:
				n.Type = RunNode
			case c == DisplayMathClass:
				n.Type = DisplayMathNode
			case c == FootnoteClass:
				n.Type = FootnoteNode
			case c == ListClass:
				n.Type = ListNode
				n.setAttr("list-type", litlisttypeOf(in.Attr))
			case c == ListItemClass:
				n.Type = ListItemNode
			case c == FragmentClass:
				n.Type = FragmentNode
			case c == SectionClass:
				n.Type = SectionNode
				n.setAttr("section-level", litsectionlevelOf(in.Attr))
				n.setAttr("section-numbered", litsectionnumbered(in.Attr))
			case c == "":
				n.Type = DivNode
				n.Attr = copyAttr(in.Attr)
			default:
				panic(fmt.Sprintf("unrecognized littype: %q", c))
			}
		case 0:
			switch in.Data {
			case "tex":
				n.Type = TexOnlyNode
			case "right":
				n.Type = RightAlignNode
			case "center":
				n.Type = CenterAlignNode
			case "equation":
				n.Type = EquationNode
				n.setAttr("id", getAttr(in.Attr, "id"))
			case "subequations":
				n.Type = SubequationsNode
			case "statement":
				n.Type = StatementNode
				n.setAttr("id", getAttr(in.Attr, "id"))
				n.setAttr("type", getAttr(in.Attr, "type"))
				n.setAttr("text", getAttr(in.Attr, "text"))
			case "proof":
				n.Type = ProofNode
			case "quote":
				n.Type = QuoteNode
			default:
				log.Printf("%+v", in)
				return nil, fmt.Errorf("unsupported ElementNode DataAtom: %s", in.DataAtom)
			}
		default:
			return nil, fmt.Errorf("unsupported ElementNode DataAtom: %s", in.DataAtom)
		}

		if n.Type == CodeNode {
			if len(n.Kids()) > 1 {
				log.Fatal("code node can only have one child")
			}
		} else {

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
		}
	default:
		return nil, fmt.Errorf("unsupported node type: %d", in.Type)
	}

	return &n, nil
}

// super simple
func ParseCSV(s string) (*Node, error) {
	fragment := Node{Type: FragmentNode}

	r := csv.NewReader(strings.NewReader(s))
	for {
		list := &Node{Type: ListNode}
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		for _, field := range record {
			li := &Node{Type: ListItemNode}
			var ts []*Token
			ts, err = Lex(field)
			if err != nil {
				return nil, err
			}

			for _, t := range ts {
				tn := &Node{Type: TokenNode, Token: t}
				li.AppendChild(tn)
			}
			list.AppendChild(li)
		}
		fragment.AppendChild(list)
	}

	return &fragment, nil
}
