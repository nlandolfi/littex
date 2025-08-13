package html

import (
	"bytes"
	"strings"
	"testing"

	"github.com/nlandolfi/lit/internal/ast"
	"github.com/nlandolfi/lit/internal/lexer"
)

func TestWriteHTML(t *testing.T) {
	tests := []struct {
		name     string
		node     *ast.Node
		expected string
	}{
		{
			name: "simple token",
			node: &ast.Node{
				Type:  ast.TokenNode,
				Token: &lexer.Token{Type: lexer.WordToken, Value: "hello"},
			},
			expected: "hello",
		},
		{
			name: "paragraph",
			node: &ast.Node{
				Type: ast.ParagraphNode,
				FirstChild: &ast.Node{
					Type:  ast.TokenNode,
					Token: &lexer.Token{Type: lexer.WordToken, Value: "hello"},
				},
			},
			expected: "<p>hello</p>",
		},
		{
			name: "error node",
			node: &ast.Node{
				Type: ast.ErrorNode,
				Data: "test error",
			},
			expected: "<div class='error'>ERROR: test error</div>",
		},
		{
			name: "run node",
			node: &ast.Node{
				Type: ast.RunNode,
				FirstChild: &ast.Node{
					Type:  ast.TokenNode,
					Token: &lexer.Token{Type: lexer.WordToken, Value: "hello"},
				},
			},
			expected: "<span class='run'>hello</span>",
		},
		{
			name: "display math",
			node: &ast.Node{
				Type: ast.DisplayMathNode,
				FirstChild: &ast.Node{
					Type:  ast.TokenNode,
					Token: &lexer.Token{Type: lexer.WordToken, Value: "x+y"},
				},
			},
			expected: "<div class='displaymath'>$$x+y$$</div>",
		},
		{
			name: "comment node",
			node: &ast.Node{
				Type: ast.CommentNode,
				Data: "test comment",
			},
			expected: "<!-- test comment -->",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up child relationships properly
			if tt.node.FirstChild != nil {
				tt.node.FirstChild.Parent = tt.node
				tt.node.LastChild = tt.node.FirstChild
			}

			var buf bytes.Buffer
			err := WriteHTML(&buf, tt.node, nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			result := buf.String()
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestWriteHTMLList(t *testing.T) {
	// Create unordered list
	ulist := &ast.Node{Type: ast.ListNode}
	ulist.SetAttr("list-type", "unordered")
	listItem := &ast.Node{Type: ast.ListItemNode}
	token := &ast.Node{Type: ast.TokenNode, Token: &lexer.Token{Type: lexer.WordToken, Value: "item"}}

	ulist.AppendChild(listItem)
	listItem.AppendChild(token)

	var buf bytes.Buffer
	err := WriteHTML(&buf, ulist, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := buf.String()
	expected := "<ul><li>item</li></ul>"
	if result != expected {
		t.Errorf("unordered list: expected %q, got %q", expected, result)
	}

	// Test ordered list
	olist := &ast.Node{Type: ast.ListNode}
	olist.SetAttr("list-type", "ordered")
	listItem2 := &ast.Node{Type: ast.ListItemNode}
	token2 := &ast.Node{Type: ast.TokenNode, Token: &lexer.Token{Type: lexer.WordToken, Value: "item"}}

	olist.AppendChild(listItem2)
	listItem2.AppendChild(token2)

	buf.Reset()
	err = WriteHTML(&buf, olist, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result = buf.String()
	expected = "<ol><li>item</li></ol>"
	if result != expected {
		t.Errorf("ordered list: expected %q, got %q", expected, result)
	}
}

func TestWriteHTMLSection(t *testing.T) {
	section := &ast.Node{Type: ast.SectionNode}
	section.SetAttr("section-level", "2")
	token := &ast.Node{Type: ast.TokenNode, Token: &lexer.Token{Type: lexer.WordToken, Value: "Title"}}
	section.AppendChild(token)

	var buf bytes.Buffer
	err := WriteHTML(&buf, section, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := buf.String()
	if !strings.Contains(result, "<h2") {
		t.Error("Section should generate h2 tag")
	}
	if !strings.Contains(result, "Title") {
		t.Error("Section should contain title")
	}
	if !strings.Contains(result, "</h2>") {
		t.Error("Section should close h2 tag")
	}
}

func TestWriteHTMLInBody(t *testing.T) {
	node := &ast.Node{
		Type: ast.ParagraphNode,
		FirstChild: &ast.Node{
			Type:  ast.TokenNode,
			Token: &lexer.Token{Type: lexer.WordToken, Value: "hello"},
		},
	}
	node.FirstChild.Parent = node
	node.LastChild = node.FirstChild

	var buf bytes.Buffer
	WriteHTMLInBody(&buf, node, nil)

	result := buf.String()
	expected := "<p>hello</p>"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestWriteHTMLEscaping(t *testing.T) {
	// Test HTML escaping of special characters
	node := &ast.Node{
		Type:  ast.TokenNode,
		Token: &lexer.Token{Type: lexer.PunctuationToken, Value: "<"},
	}

	var buf bytes.Buffer
	err := WriteHTML(&buf, node, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := buf.String()
	if result != "&lt;" {
		t.Errorf("expected '&lt;', got %q", result)
	}
}
