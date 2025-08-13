package ast

import (
	"fmt"

	"github.com/nlandolfi/lit/internal/lexer"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// Node is similar to *html.Node.
//
// The main difference is that we lex HTML text nodes
// into TokenNodes.
type Node struct {
	Type      NodeType // The Type of Node, see NodeType.
	DataAtom  atom.Atom
	Data      string
	Attr      []Attribute                 // The attributes, as in html.Node. Attribute is a type alias of html.Attribute
	Token     *lexer.Token                // The token value if Type==TokenNode; see Token.
	JSON      map[string]interface{}      // The JSON if Type == JSONNode
	YAML      map[interface{}]interface{} // The YAML if Type == YAMLNode
	IsComment bool                        // For JSON and YAML nodes, only if they are comment form

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
	DrawNode
	JSONNode
	YAMLNode
	ImageNode
	StatementNode
	ProofNode
	LinkNode
	DivNode
	CodeNode
	PreNode
	OpaqueNode
)

func (t NodeType) String() string {
	switch t {
	case ErrorNode:
		return "error"
	case FragmentNode:
		return "fragment"
	case ParagraphNode:
		return "paragraph"
	case FootnoteNode:
		return "footnote"
	case DisplayMathNode:
		return "displaymath"
	case RunNode:
		return "run"
	case TextNode:
		return "text"
	case TokenNode:
		return "token"
	case ListNode:
		return "list"
	case ListItemNode:
		return "list-item"
	case SectionNode:
		return "section"
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
	case DrawNode:
		return "draw"
	case JSONNode:
		return "json"
	case YAMLNode:
		return "yaml"
	case ImageNode:
		return "image"
	case StatementNode:
		return "statement"
	case ProofNode:
		return "proof"
	case LinkNode:
		return "link"
	case DivNode:
		return "div"
	case CodeNode:
		return "code"
	case PreNode:
		return "pre"
	case OpaqueNode:
		return "opaque"
	default:
		panic(fmt.Sprintf("unknown node type: %d", t))
	}
}

type NodeClass string

const (
	InlineClass  NodeClass = "inline"
	BlockClass   NodeClass = "block"
	UnknownClass NodeClass = "unknown"
) // Kids returns the children of the node.
func (n *Node) Kids() (ks []*Node) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		ks = append(ks, c)
	}
	return
}
