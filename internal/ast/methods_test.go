package ast

import (
	"testing"

	"github.com/nlandolfi/lit/internal/lexer"
)

func TestInsertBefore(t *testing.T) {
	parent := &Node{Type: FragmentNode}
	oldChild := &Node{Type: TokenNode}
	newChild := &Node{Type: TokenNode}

	parent.AppendChild(oldChild)
	parent.InsertBefore(newChild, oldChild)

	if parent.FirstChild != newChild {
		t.Error("newChild should be first child")
	}
	if newChild.NextSibling != oldChild {
		t.Error("newChild.NextSibling should be oldChild")
	}
	if oldChild.PrevSibling != newChild {
		t.Error("oldChild.PrevSibling should be newChild")
	}
}

func TestInsertBeforeNil(t *testing.T) {
	parent := &Node{Type: FragmentNode}
	existingChild := &Node{Type: TokenNode}
	newChild := &Node{Type: TokenNode}

	parent.AppendChild(existingChild)
	parent.InsertBefore(newChild, nil) // Should append at end

	if parent.LastChild != newChild {
		t.Error("newChild should be last child when inserting before nil")
	}
}

func TestRemoveChild(t *testing.T) {
	parent := &Node{Type: FragmentNode}
	child1 := &Node{Type: TokenNode}
	child2 := &Node{Type: TokenNode}
	child3 := &Node{Type: TokenNode}

	parent.AppendChild(child1)
	parent.AppendChild(child2)
	parent.AppendChild(child3)

	// Remove middle child
	parent.RemoveChild(child2)

	if child1.NextSibling != child3 {
		t.Error("child1.NextSibling should be child3 after removing child2")
	}
	if child3.PrevSibling != child1 {
		t.Error("child3.PrevSibling should be child1 after removing child2")
	}
	if child2.Parent != nil || child2.NextSibling != nil || child2.PrevSibling != nil {
		t.Error("child2 should have no relationships after removal")
	}
}

func TestRemoveFirstChild(t *testing.T) {
	parent := &Node{Type: FragmentNode}
	child1 := &Node{Type: TokenNode}
	child2 := &Node{Type: TokenNode}

	parent.AppendChild(child1)
	parent.AppendChild(child2)

	parent.RemoveChild(child1)

	if parent.FirstChild != child2 {
		t.Error("child2 should be first child after removing child1")
	}
	if child2.PrevSibling != nil {
		t.Error("child2.PrevSibling should be nil after becoming first child")
	}
}

func TestRemoveLastChild(t *testing.T) {
	parent := &Node{Type: FragmentNode}
	child1 := &Node{Type: TokenNode}
	child2 := &Node{Type: TokenNode}

	parent.AppendChild(child1)
	parent.AppendChild(child2)

	parent.RemoveChild(child2)

	if parent.LastChild != child1 {
		t.Error("child1 should be last child after removing child2")
	}
	if child1.NextSibling != nil {
		t.Error("child1.NextSibling should be nil after becoming last child")
	}
}

func TestSectionNumbered(t *testing.T) {
	n := &Node{Type: SectionNode}
	n.SetAttr("section-numbered", "true")

	if !n.SectionNumbered() {
		t.Error("SectionNumbered should return true when attribute is 'true'")
	}

	n.SetAttr("section-numbered", "false")
	if n.SectionNumbered() {
		t.Error("SectionNumbered should return false when attribute is 'false'")
	}
}

func TestSectionLevel(t *testing.T) {
	n := &Node{Type: SectionNode}
	n.SetAttr("section-level", "3")

	if level := n.SectionLevel(); level != "3" {
		t.Errorf("SectionLevel should return '3', got '%s'", level)
	}
}

func TestKidsExcludingTokens(t *testing.T) {
	parent := &Node{Type: FragmentNode}
	tokenChild := &Node{Type: TokenNode, Token: &lexer.Token{Type: lexer.WordToken, Value: "test"}}
	paragraphChild := &Node{Type: ParagraphNode}
	listChild := &Node{Type: ListNode}

	parent.AppendChild(tokenChild)
	parent.AppendChild(paragraphChild)
	parent.AppendChild(listChild)

	kids := parent.KidsExcludingTokens()
	if len(kids) != 2 {
		t.Errorf("Expected 2 non-token children, got %d", len(kids))
	}
	if kids[0] != paragraphChild || kids[1] != listChild {
		t.Error("KidsExcludingTokens should only return non-token children")
	}
}

func TestIsListItem(t *testing.T) {
	listItem := &Node{Type: ListItemNode}
	paragraph := &Node{Type: ParagraphNode}

	if !listItem.IsListItem() {
		t.Error("ListItemNode should return true for IsListItem")
	}
	if paragraph.IsListItem() {
		t.Error("ParagraphNode should return false for IsListItem")
	}
}

func TestGetAttr(t *testing.T) {
	n := &Node{Type: FragmentNode}
	n.SetAttr("test-key", "test-value")
	n.SetAttr("another-key", "another-value")

	if value := n.GetAttr("test-key"); value != "test-value" {
		t.Errorf("GetAttr should return 'test-value', got '%s'", value)
	}
	if value := n.GetAttr("nonexistent"); value != "" {
		t.Errorf("GetAttr should return empty string for nonexistent key, got '%s'", value)
	}
}

func TestUtilityFunctions(t *testing.T) {
	// Test LitlisttypeOf
	attrs := []Attribute{{Key: "data-litlisttype", Val: "ordered"}}
	if listType := LitlisttypeOf(attrs); listType != "ordered" {
		t.Errorf("LitlisttypeOf should return 'ordered', got '%s'", listType)
	}

	// Test default case
	emptyAttrs := []Attribute{}
	if listType := LitlisttypeOf(emptyAttrs); listType != "unordered" {
		t.Errorf("LitlisttypeOf should return 'unordered' by default, got '%s'", listType)
	}

	// Test LitsectionlevelOf
	sectionAttrs := []Attribute{{Key: "data-litsectionlevel", Val: "2"}}
	if level := LitsectionlevelOf(sectionAttrs); level != "2" {
		t.Errorf("LitsectionlevelOf should return '2', got '%s'", level)
	}

	// Test default case
	if level := LitsectionlevelOf(emptyAttrs); level != "1" {
		t.Errorf("LitsectionlevelOf should return '1' by default, got '%s'", level)
	}
}
