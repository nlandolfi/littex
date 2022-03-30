package lit

import (
	"fmt"

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
	default:
		panic(fmt.Sprintf("unknown node type: %d", t))
	}
}

type NodeClass string

const (
	ErrorClass       = "gba3-error"
	FragmentClass    = "gba3-fragment"
	ParagraphClass   = "gba3-paragraph"
	FootnoteClass    = "gba3-footnote"
	DisplayMathClass = "gba3-displaymath"
	RunClass         = "gba3-run"
	TextClass        = "gba3-text"
	TokenClass       = "gba3-token"
	ListClass        = "gba3-list"
	ListItemClass    = "gba3-listitem"
)

type Attribute html.Attribute

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

func classOf(n *html.Node) string {
	for _, a := range n.Attr {
		if a.Key == "class" {
			return a.Val
		}
	}

	return ""
}
