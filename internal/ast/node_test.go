package ast

import (
	"testing"

	"github.com/nlandolfi/lit/internal/lexer"
)

func TestNodeTypeString(t *testing.T) {
	tests := []struct {
		nt       NodeType
		expected string
	}{
		{ErrorNode, "error"},
		{FragmentNode, "fragment"},
		{ParagraphNode, "paragraph"},
		{TokenNode, "token"},
		{ListNode, "list"},
	}

	for _, tt := range tests {
		if got := tt.nt.String(); got != tt.expected {
			t.Errorf("NodeType(%d).String() = %q, expected %q", tt.nt, got, tt.expected)
		}
	}
}

func TestNodeAppendChild(t *testing.T) {
	parent := &Node{
		Type: FragmentNode,
	}

	child1 := &Node{
		Type:  TokenNode,
		Token: &lexer.Token{Type: lexer.WordToken, Value: "hello"},
	}

	child2 := &Node{
		Type:  TokenNode,
		Token: &lexer.Token{Type: lexer.WordToken, Value: "world"},
	}

	parent.AppendChild(child1)
	parent.AppendChild(child2)

	if parent.FirstChild != child1 {
		t.Error("FirstChild should be child1")
	}
	if parent.LastChild != child2 {
		t.Error("LastChild should be child2")
	}
	if child1.Parent != parent {
		t.Error("child1.Parent should be parent")
	}
	if child2.Parent != parent {
		t.Error("child2.Parent should be parent")
	}
	if child1.NextSibling != child2 {
		t.Error("child1.NextSibling should be child2")
	}
	if child2.PrevSibling != child1 {
		t.Error("child2.PrevSibling should be child1")
	}
}

func TestNodeKids(t *testing.T) {
	parent := &Node{
		Type: FragmentNode,
	}

	child1 := &Node{Type: TokenNode}
	child2 := &Node{Type: TokenNode}
	child3 := &Node{Type: TokenNode}

	parent.AppendChild(child1)
	parent.AppendChild(child2)
	parent.AppendChild(child3)

	kids := parent.Kids()
	if len(kids) != 3 {
		t.Errorf("expected 3 kids, got %d", len(kids))
	}
	if kids[0] != child1 || kids[1] != child2 || kids[2] != child3 {
		t.Error("kids order is incorrect")
	}
}

func TestSetAttr(t *testing.T) {
	n := &Node{
		Type: FragmentNode,
	}

	// Test adding new attribute
	n.SetAttr("key1", "value1")
	if len(n.Attr) != 1 {
		t.Errorf("expected 1 attribute, got %d", len(n.Attr))
	}
	if n.Attr[0].Key != "key1" || n.Attr[0].Val != "value1" {
		t.Error("attribute not set correctly")
	}

	// Test replacing existing attribute
	n.SetAttr("key1", "value2")
	if len(n.Attr) != 1 {
		t.Errorf("expected 1 attribute, got %d", len(n.Attr))
	}
	if n.Attr[0].Key != "key1" || n.Attr[0].Val != "value2" {
		t.Error("attribute not updated correctly")
	}

	// Test adding second attribute
	n.SetAttr("key2", "value3")
	if len(n.Attr) != 2 {
		t.Errorf("expected 2 attributes, got %d", len(n.Attr))
	}
}
