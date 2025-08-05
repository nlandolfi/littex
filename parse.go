package lit

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"regexp"
	"strings"
	"unicode/utf8"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"gopkg.in/yaml.v3"
)

// ParseHTML parses HTML content into a LitTex Node tree. It creates a fragment
// node as the root and processes all HTML elements into their corresponding
// LitTex node types.
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

// Must is a helper that wraps a call to a function returning (*Node, error)
// and panics if the error is non-nil. It's intended for use in variable
// initializations such as:
//
//	var document = lit.Must(lit.ParseLit(...))
//
// This is similar to template.Must in the standard library.
func Must(n *Node, err error) *Node {
	if err != nil {
		panic(err)
	}
	return n
}

// ParseLit parses LitTex markup into a Node tree. It first converts LitTex
// syntax to an intermediate HTML representation and then parses that HTML
// into the final Node tree structure.
func ParseLit(s string) (*Node, error) {
	s = litReplace(s)
	return ParseHTML(s)
}

func litReplace(s string) string {
	s = " " + s // to ensure a first character match,
	// for the picrow etc escapes

	s = strings.Replace(s, "\\<", "&lt;", -1)
	s = strings.Replace(s, "\\>", "&gt;", -1)

	// runs
	re := regexp.MustCompile(`[^\\]‖`)
	s = re.ReplaceAllString(s, "<div data-littype='"+RunClass+"'>")
	s = strings.Replace(s, "\\‖", "‖", -1)

	// pilcrow
	re = regexp.MustCompile(`([^\\])¶⦊`)
	s = re.ReplaceAllString(s, `$1¶ ⦊`)
	re = regexp.MustCompile(`([^\\])¶ ⦊`)
	s = re.ReplaceAllString(s, "$1<div data-littype='"+ParagraphClass+"'>")
	s = strings.Replace(s, "\\¶", "¶", -1)

	// footnote
	re = regexp.MustCompile(`([^\\])†⦊`)
	s = re.ReplaceAllString(s, `$1† ⦊`)
	re = regexp.MustCompile(`([^\\])† ⦊`)
	s = re.ReplaceAllString(s, "$1<div data-littype='"+FootnoteClass+"'>")
	s = strings.Replace(s, "\\†", "†", -1)

	// display math
	re = regexp.MustCompile(`([^\\])◇⦊`)
	s = re.ReplaceAllString(s, `$1◇ ⦊`)
	re = regexp.MustCompile(`([^\\])◇ ⦊`)
	s = re.ReplaceAllString(s, "$1<div data-littype='"+DisplayMathClass+"'>")
	s = strings.Replace(s, "\\◇", "◇", -1)

	// unordered lists
	re = regexp.MustCompile(`([^\\])⁝⦊`)
	s = re.ReplaceAllString(s, `$1⁝ ⦊`)
	re = regexp.MustCompile(`([^\\])⁝ ⦊`)
	s = re.ReplaceAllString(s, "$1<div data-littype='"+ListClass+"' data-litlisttype='unordered'>")
	s = strings.Replace(s, "\\⁝", "⁝", -1)

	// ordered lists
	re = regexp.MustCompile(`([^\\])𝍫⦊`)
	s = re.ReplaceAllString(s, `$1𝍫 ⦊`)
	re = regexp.MustCompile(`([^\\])𝍫 ⦊`)
	s = re.ReplaceAllString(s, "$1<div data-littype='"+ListClass+"' data-litlisttype='ordered'>")
	s = strings.Replace(s, "\\𝍫", "𝍫", -1)

	// list items
	re = regexp.MustCompile(`([^\\])‣`)
	s = re.ReplaceAllString(s, "$1<div data-littype='"+ListItemClass+"'>")
	s = strings.Replace(s, "\\‣", "‣", -1)

	// sections
	// first, replace repeats
	re = regexp.MustCompile(`[^\\]§+`)
	s = re.ReplaceAllStringFunc(s, func(og string) string {
		//		log.Printf("in: %q", og)
		// drop the first non \ match
		index := strings.Index(og, "§")
		in := og[index:len(og)]
		//log.Printf("slice: %q", og[0:index])
		//log.Print(utf8.RuneCountInString(in))
		out := fmt.Sprintf("%s§%d", og[0:index], utf8.RuneCountInString(in))
		//		log.Printf("out: %q", out)
		return out
	})
	// numbered
	re = regexp.MustCompile(`#§([[:digit:]]+)`)
	s = re.ReplaceAllString(s, "<div data-littype='"+SectionClass+"' data-litsectionlevel='$1' data-litsectionnumbered='true'>")
	// unnumbered
	re = regexp.MustCompile(`([^\\#])§([[:digit:]]+)`)
	s = re.ReplaceAllString(s, "$1<div data-littype='"+SectionClass+"' data-litsectionlevel='$2' data-litsectionnumbered='false'>")
	// section symbol
	s = strings.Replace(s, "\\§", "§", -1)

	// closes
	// the naive single match doesn't work, misses some of them
	// so need this more complicated thing
	re = regexp.MustCompile(`([^\\])⦉+`)
	s = re.ReplaceAllStringFunc(s, func(og string) string {
		// drop the first non \ match
		index := strings.Index(og, "⦉")
		in := og[index:len(og)]
		out := og[0:index]
		for i := 0; i < utf8.RuneCountInString(in); i++ {
			out += "</div>"
		}
		return out
	})
	// all to get the escape functionality
	s = strings.Replace(s, "\\⦉", "⦉", -1)

	// Update: Unfortunately the below doesn't work
	// because it will write out the replacements, instead
	// of the compact form...more to be done here.
	// -1 hack, this should be improved to only make
	// the replacement when in math mode, and to think
	// through edge cases, but I think the gains in
	// readability for now outweigh the fragileness of
	// this solution
	s = strings.Replace(s, "⁻¹", "^{-1}", -1)
	// the same goes for the below
	s = strings.Replace(s, "¹", "^{1}", -1)
	s = strings.Replace(s, "²", "^{2}", -1)
	s = strings.Replace(s, "₁", "_{1}", -1)
	s = strings.Replace(s, "₂", "_{2}", -1)
	s = strings.Replace(s, "ᵢ", "_{i}", -1)
	s = strings.Replace(s, "ⱼ", "_{j}", -1)
	s = strings.Replace(s, "ₖ", "_{k}", -1)
	s = strings.Replace(s, "ₘ", "_{m}", -1)
	s = strings.Replace(s, "ₙ", "_{n}", -1)

	//	s = strings.Replace(s, "⦉", "</div>", -1)

	//re = regexp.MustCompile(`\[(.+?)\]\((.+?)\)`)
	//s = re.ReplaceAllString(s, `<a href='$2'> ‖ $1 ⦉</a>`)
	return s
}

// ParseTex converts LaTeX content into a LitTex Node tree. It processes LaTeX
// commands and environments, transforming them into their corresponding LitTex
// syntax before parsing the result into a Node tree.
func ParseTex(s string) (*Node, error) {
	for _, c := range commentsR.FindAllString(s, -1) {
		log.Printf("dropping comment: %q", c)
	}

	for _, r := range order {
		replace := res[r]
		s = r.ReplaceAllString(s, replace)
	}
	s = strings.Replace(s, "\\item", " ‣", -1)
	s = strings.Replace(s, "\\begin{itemize}", " ⁝ ⦊", -1)
	s = strings.Replace(s, "\\begin{enumerate}", " 𝍫 ⦊", -1)
	s = strings.Replace(s, "\\end{itemize}", "⦉", -1)
	s = strings.Replace(s, "\\end{enumerate}", "⦉", -1)
	s = strings.Replace(s, "\\[\n", "◇ ⦊ ‖ ", -1)
	s = strings.Replace(s, "\n\\]", " ⦉", -1)
	s = strings.Replace(s, "---", "—", -1)
	s = strings.Replace(s, "``", "\"\"", -1)
	s = strings.Replace(s, "''", "\"\"", -1)
	s = strings.Replace(s, "`", "'", -1) // MUST BE AFTER DOUBLE
	// s = strings.Replace(s, "'", "'", -1)
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
		fmt.Fprintf(w, " ¶ ⦊")
		ls := strings.Split(p, "\n")
		for _, l := range ls {
			fmt.Fprintf(w, " ‖ ")
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
var propositionWithText = regexp.MustCompile(`\\begin{proposition}\[([\w| ]*)\]`)
var proposition = regexp.MustCompile(`\\begin{proposition}`)
var propositionEnd = regexp.MustCompile(`\\end{proposition}`)
var proof = regexp.MustCompile(`\\begin{proof}`)
var proofEnd = regexp.MustCompile(`\\end{proof}`)
var ssection = regexp.MustCompile(`\\ssection{(\w*)}`)
var section = regexp.MustCompile(`\\section{(\w*)}`)
var ssubsection = regexp.MustCompile(`\\ssubsection{(\w*)}`)
var subsection = regexp.MustCompile(`\\subsection{(\w*)}`)

// useful: https://gist.github.com/claybridges/8f9d51a1dc365f2e64fa
var res = map[*regexp.Regexp]string{
	propositionWithText: " <statement type='proposition' text='$1'>",
	proposition:         " <statement type='proposition'>",
	propositionEnd:      " </statement>",
	proof:               " <proof>",
	proofEnd:            " </proof>",
	ssection:            " § $1 ⦉",
	section:             " #§ $1 ⦉",
	subsection:          " #§§ $1 ⦉",
	ssubsection:         " §§ $1 ⦉",
	textitR:             "‹$1›",
	textbfR:             "«$1»",
	footnoteR:           " † ⦊ ‖ $1 ⦉⦉",
	textscR:             "⸤$1⸥",
	tR:                  "❬$1❭",
	cR:                  "⁅$1⁆",
	dblqR:               "\"$1\"",
	sglqR:               "'$1'",
	sayR:                "\"$1\"",
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

// UnmarshalHTML converts an HTML node tree into a LitTex Node tree.
// It recursively processes the HTML structure and transforms it into
// the corresponding LitTex node types based on element attributes and content.
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

		if strings.HasPrefix(in.Data, "yaml") {
			d := strings.TrimPrefix(in.Data, "yaml")
			n.Type = YAMLNode
			n.Data = d
			n.IsComment = true
			if strings.TrimSpace(d) != "" {
				n.YAML = make(map[interface{}]interface{})
				if err := yaml.Unmarshal([]byte(d), &n.YAML); err != nil {
					// TODO: do something else?
					log.Fatalf("yaml.Unmarshal: %v", err)
				}
			}
			return &n, nil
		}
		if strings.HasPrefix(in.Data, "json") {
			d := strings.TrimPrefix(in.Data, "json")
			n.Type = JSONNode
			n.Data = d
			n.IsComment = true
			if strings.TrimSpace(d) != "" {
				n.JSON = make(map[string]interface{})
				if err := json.Unmarshal([]byte(d), &n.JSON); err != nil {
					// TODO: do something else?
					log.Fatalf("json.Unmarshal: %v", err)
				}
			}
			return &n, nil
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
			case "json":
				n.Type = JSONNode
				n.Attr = copyAttr(in.Attr)

				c := in.FirstChild
				if c == nil {
					return &n, nil
				}
				if c.Type != html.TextNode {
					return &n, nil
				}
				if strings.TrimSpace(c.Data) == "" {
					return &n, nil
				}
				n.JSON = make(map[string]interface{})
				if err := json.Unmarshal([]byte(c.Data), &n.JSON); err != nil {
					// TODO: do something else?
					log.Fatalf("json.Unmarshal: %v", err)
				}
				n.Data = c.Data // TODO: remove?? - NCL 1/25/23
			case "yaml":
				n.Type = YAMLNode
				n.Attr = copyAttr(in.Attr)

				c := in.FirstChild
				if c == nil {
					return &n, nil
				}
				if c.Type != html.TextNode {
					return &n, nil
				}
				if strings.TrimSpace(c.Data) == "" {
					return &n, nil
				}
				n.YAML = make(map[interface{}]interface{})
				if err := yaml.Unmarshal([]byte(c.Data), &n.YAML); err != nil {
					// TODO: do something else?
					log.Fatalf("yaml.Unmarshal: %v", err)
				}
				n.Data = c.Data // TODO: remove?? - NCL 1/25/23
			default:
				n.Type = OpaqueNode
				n.Attr = copyAttr(in.Attr)
				n.Data = in.Data
			}
		default:
			n.Type = OpaqueNode
			n.Attr = copyAttr(in.Attr)
			n.DataAtom = in.DataAtom
		}

		for c := in.FirstChild; c != nil; c = c.NextSibling {
			switch c.Type {
			case html.TextNode:
				if strings.TrimSpace(c.Data) == "" {
					continue
				}

				r := &n
				if r.Type != RunNode && r.Type != ListItemNode && r.Type != SectionNode {
					r = &Node{}
					r.Type = RunNode
					n.AppendChild(r)
				}

				ts, err := unmarshalHTMLText(c)
				if err != nil {
					return nil, err
				}

				for _, child := range ts {
					r.AppendChild(child)
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
		return nil, fmt.Errorf("unsupported node type: %d", in.Type)
	}
	return &n, nil
}

// ParseCSV parses CSV content into a LitTex Node tree. Each row is represented
// as a list node, and each field in the row becomes a list item with appropriate
// token nodes for the content.
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
