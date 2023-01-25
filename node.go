package lit

import (
	"fmt"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// Node is similar to *html.Node.
//
// The main difference is that we lex HTML text nodes
// into TokenNodes.
type Node struct {
	Type     NodeType // The Type of Node, see NodeType.
	DataAtom atom.Atom
	Data     string
	Attr     []Attribute            // The attributes, as in html.Node. Attribute is a type alias of html.Attribute
	Token    *Token                 // The token value if Type==TokenNode; see Token.
	JSON     map[string]interface{} // The JSON if Type == JSONNode

	Parent                   *Node `json:"-"`
	FirstChild, LastChild    *Node `json:"-"`
	PrevSibling, NextSibling *Node `json:"-"`
}

type Attribute = html.Attribute

type NodeType int

const (
	ErrorNode NodeType = iota
	FragmentNode
	ParagraphNode
	FootnoteNode
	DisplayMathNode
	RunNode
	TextNode
	TokenNode
	ListNode
	ListItemNode
	SectionNode
	CommentNode
	TexOnlyNode
	CenterAlignNode
	RightAlignNode
	EquationNode
	SubequationsNode
	ImageNode
	StatementNode
	ProofNode
	LinkNode
	TableNode
	TableHeadNode
	TableBodyNode
	TableRowNode
	THNode
	TDNode
	QuoteNode
	DivNode
	CodeNode
	PreNode
	JSONNode
	OpaqueNode // Any other node type, for extending to lit to arbitrarty HTML
)

func (t NodeType) String() string {
	switch t {
	case ErrorNode:
		return "error"
	case FragmentNode:
		return "fragment"
	case ParagraphNode:
		return "¶"
	case FootnoteNode:
		return "†"
	case DisplayMathNode:
		return "◇"
	case RunNode:
		return "‖"
	case TextNode:
		return "text"
	case TokenNode:
		return "token"
	case ListNode:
		return "⁝"
	case ListItemNode:
		return "‣"
	case SectionNode:
		return "§"
	case CommentNode:
		return "comment"
	case TexOnlyNode:
		return "tex"
	case CenterAlignNode:
		return "center"
	case RightAlignNode:
		return "right"
	case EquationNode:
		return "equation"
	case ImageNode:
		return "image"
	case StatementNode:
		return "statement"
	case ProofNode:
		return "statement"
	case LinkNode:
		return "link"
	case DivNode:
		return "div"
	case CodeNode:
		return "code"
	case PreNode:
		return "pre"
	case JSONNode:
		return "json"
	case OpaqueNode:
		return "opaque" // do we need this? or the above? - NCL 1/25/23
	default:
		panic(fmt.Sprintf("unknown node type: %d", t))
	}
}

type NodeClass string

const (
	ErrorClass       = "error"
	FragmentClass    = "fragment"
	ParagraphClass   = "paragraph"
	FootnoteClass    = "footnote"
	DisplayMathClass = "displaymath"
	RunClass         = "run"
	TextClass        = "text"
	TokenClass       = "token"
	ListClass        = "list"
	ListItemClass    = "listitem"
	SectionClass     = "section"
)

func (n *Node) Kids() (ks []*Node) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		ks = append(ks, c)
	}
	return
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

func copyAttr(as []Attribute) []Attribute {
	var out []Attribute = make([]Attribute, len(as))
	for i, a := range as {
		out[i] = a
	}
	return out
}

func getAttr(as []Attribute, k string) string {
	for _, a := range as {
		if a.Key == k {
			return a.Val
		}
	}

	return ""
}

func littypeOf(n *html.Node) string {
	for _, a := range n.Attr {
		if a.Key == "data-littype" {
			return a.Val
		}
	}

	return ""
}

func litlisttypeOf(a []html.Attribute) string {
	for _, at := range a {
		if at.Key == "data-litlisttype" {
			return at.Val
		}
	}

	return "unordered"
}

func litsectionlevelOf(a []html.Attribute) string {
	for _, at := range a {
		if at.Key == "data-litsectionlevel" {
			return at.Val
		}
	}

	return "1"
}

func litsectionnumbered(a []html.Attribute) string {
	for _, at := range a {
		if at.Key == "data-litsectionnumbered" {
			return at.Val
		}
	}

	return "false"
}

func litimgsrc(a []html.Attribute) string {
	for _, at := range a {
		if at.Key == "src" {
			return at.Val
		}
	}

	return ""
}

func (n *Node) setAttr(k, v string) {
	for _, a := range n.Attr {
		if a.Key == k {
			a.Val = v
		}
	}
	n.Attr = append(n.Attr, Attribute{Key: k, Val: v})
}

func (n *Node) SectionNumbered() bool {
	return getAttr(n.Attr, "section-numbered") == "true"
}

func (n *Node) SectionLevel() string {
	return getAttr(n.Attr, "section-level")
}

// Convenient for templates (esp. slides) {{{

func (n *Node) FirstTokenString() string {
	if n.Type != ListItemNode {
		panic("FirstTokenSTring only for list items")
	}
	if n.FirstChild == nil {
		return ""
	}
	block, _ := tokenBlockStartingAt(n.FirstChild)
	lines := lineBlocks(block, Tex, new(WriteOpts), true, maxWidth)
	if len(lines) > 1 {
		return strings.Join(lines, "\n")
	}
	return lines[0]
}

func (n *Node) FirstListNode() *Node {
	// kids of the first ⁝ node
	if n.Type != ListItemNode {
		panic("FirstListNode only for list items")
	}
	c := n.FirstChild
	for c != nil && c.Type != ListNode {
		c = c.NextSibling
	}
	if c == nil {
		// quick fix for "‣ Slide title ⦉" fillers
		li := &Node{Type: ListItemNode}
		l := &Node{Type: ListNode}
		l.AppendChild(li)
		return l
	}
	return c
}

func (n *Node) KidsExcludingTokens() (ks []*Node) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == TokenNode {
			continue
		}
		ks = append(ks, c)
	}
	return
}

func (n *Node) IsListItem() bool {
	return n.Type == ListItemNode
}

func (n *Node) headerTokenString() string {
	if n.Type != TokenNode {
		panic("TokenString only for tokens")
	}
	block, _ := tokenBlockStartingAt(n)
	lines := lineBlocks(block, Tex, new(WriteOpts), true, maxWidth)
	out := lines[0]
	if len(lines) > 1 {
		out = strings.Join(lines, " ")
	}
	return strings.Replace(out, " ", "_", -1)
}

// }}}
