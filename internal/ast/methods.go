package ast

import (
	"golang.org/x/net/html"
)

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

// RemoveChild removes a node c that is a child of n. Afterwards, c will have
// no parent and no siblings.
//
// It will panic if c's parent is not n.
func (n *Node) RemoveChild(c *Node) {
	if c.Parent != n {
		panic("html: RemoveChild called for a non-child Node")
	}
	if n.FirstChild == c {
		n.FirstChild = c.NextSibling
	}
	if c.NextSibling != nil {
		c.NextSibling.PrevSibling = c.PrevSibling
	}
	if n.LastChild == c {
		n.LastChild = c.PrevSibling
	}
	if c.PrevSibling != nil {
		c.PrevSibling.NextSibling = c.NextSibling
	}
	c.Parent = nil
	c.PrevSibling = nil
	c.NextSibling = nil
}

func copyAttr(as []Attribute) []Attribute {
	out := make([]Attribute, len(as))
	copy(out, as)
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

func LittypeOf(n *html.Node) string {
	for _, a := range n.Attr {
		if a.Key == "data-littype" {
			return a.Val
		}
	}

	return ""
}

func LitlisttypeOf(a []html.Attribute) string {
	for _, at := range a {
		if at.Key == "data-litlisttype" {
			return at.Val
		}
	}

	return "unordered"
}

func LitsectionlevelOf(a []html.Attribute) string {
	for _, at := range a {
		if at.Key == "data-litsectionlevel" {
			return at.Val
		}
	}

	return "1"
}

func LitsectionnumberedOf(a []html.Attribute) string {
	for _, at := range a {
		if at.Key == "data-litsectionnumbered" {
			return at.Val
		}
	}

	return "false"
}

func LitimgsrcOf(a []html.Attribute) string {
	for _, at := range a {
		if at.Key == "src" {
			return at.Val
		}
	}

	return ""
}

func (n *Node) SetAttr(k, v string) {
	for i := range n.Attr {
		if n.Attr[i].Key == k {
			n.Attr[i].Val = v
			return
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

func (n *Node) GetAttr(key string) string {
	for _, attr := range n.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}
