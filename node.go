package lit

import (
	"fmt"
	"strings"

	"golang.org/x/net/html"
)

// Node is a GBA node, similar to *html.Node, except that
// we lex the html text nodes into Tokens.
type Node struct {
	Type  NodeType
	Data  string
	Attr  []Attribute
	Token *Token

	Parent                   *Node `json:"-"`
	FirstChild, LastChild    *Node
	PrevSibling, NextSibling *Node
}

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

type Attribute = html.Attribute

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

func (n *Node) setAttr(k, v string) {
	for _, a := range n.Attr {
		if a.Key == k {
			a.Val = v
		}
	}
	n.Attr = append(n.Attr, Attribute{Key: k, Val: v})
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
	lines := lineBlocks(block, Tex, maxWidth)
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

// }}}
