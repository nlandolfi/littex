package tex

import (
	"bytes"
	"strings"
	"testing"

	"github.com/nlandolfi/lit/internal/ast"
	"github.com/nlandolfi/lit/internal/lexer"
)

func TestWriteTex(t *testing.T) {
	tests := []struct {
		name     string
		node     *ast.Node
		contains string
	}{
		{
			name: "simple token",
			node: &ast.Node{
				Type:  ast.TokenNode,
				Token: &lexer.Token{Type: lexer.WordToken, Value: "hello"},
			},
			contains: "hello",
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
			contains: "hello",
		},
		{
			name: "error node",
			node: &ast.Node{
				Type: ast.ErrorNode,
				Data: "test error",
			},
			contains: "% ERROR: test error",
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
			contains: "\\begin{displaymath}",
		},
		{
			name: "comment",
			node: &ast.Node{
				Type: ast.CommentNode,
				Data: "test comment",
			},
			contains: "% test comment",
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
			WriteTex(&buf, tt.node, nil)

			result := buf.String()
			if !strings.Contains(result, tt.contains) {
				t.Errorf("expected result to contain %q, got %q", tt.contains, result)
			}
		})
	}
}

func TestWriteTexList(t *testing.T) {
	// Create unordered list
	ulist := &ast.Node{Type: ast.ListNode}
	ulist.SetAttr("list-type", "unordered")
	listItem := &ast.Node{Type: ast.ListItemNode}
	token := &ast.Node{Type: ast.TokenNode, Token: &lexer.Token{Type: lexer.WordToken, Value: "item"}}

	ulist.AppendChild(listItem)
	listItem.AppendChild(token)

	var buf bytes.Buffer
	WriteTex(&buf, ulist, nil)

	result := buf.String()
	if !strings.Contains(result, "\\begin{itemize}") {
		t.Error("unordered list should generate itemize environment")
	}
	if !strings.Contains(result, "\\item") {
		t.Error("list item should generate \\item command")
	}
	if !strings.Contains(result, "\\end{itemize}") {
		t.Error("unordered list should close itemize environment")
	}

	// Test ordered list
	olist := &ast.Node{Type: ast.ListNode}
	olist.SetAttr("list-type", "ordered")
	listItem2 := &ast.Node{Type: ast.ListItemNode}
	token2 := &ast.Node{Type: ast.TokenNode, Token: &lexer.Token{Type: lexer.WordToken, Value: "item"}}

	olist.AppendChild(listItem2)
	listItem2.AppendChild(token2)

	buf.Reset()
	WriteTex(&buf, olist, nil)

	result = buf.String()
	if !strings.Contains(result, "\\begin{enumerate}") {
		t.Error("ordered list should generate enumerate environment")
	}
	if !strings.Contains(result, "\\end{enumerate}") {
		t.Error("ordered list should close enumerate environment")
	}
}

func TestWriteTexSection(t *testing.T) {
	section := &ast.Node{Type: ast.SectionNode}
	section.SetAttr("section-level", "1")
	section.SetAttr("section-numbered", "true")
	token := &ast.Node{Type: ast.TokenNode, Token: &lexer.Token{Type: lexer.WordToken, Value: "Title"}}
	section.AppendChild(token)

	var buf bytes.Buffer
	WriteTex(&buf, section, nil)

	result := buf.String()
	if !strings.Contains(result, "\\section{") {
		t.Error("Level 1 numbered section should generate \\section command")
	}
	if !strings.Contains(result, "Title") {
		t.Error("Section should contain title")
	}

	// Test unnumbered section
	section.SetAttr("section-numbered", "false")
	buf.Reset()
	WriteTex(&buf, section, nil)

	result = buf.String()
	if !strings.Contains(result, "\\section*{") {
		t.Error("Unnumbered section should generate \\section* command")
	}
}

func TestWriteTexMathToken(t *testing.T) {
	// Test token with math symbol
	node := &ast.Node{
		Type:  ast.TokenNode,
		Token: &lexer.Token{Type: lexer.SymbolToken, Value: "→"}, // right arrow
	}

	var buf bytes.Buffer
	WriteTex(&buf, node, nil)

	result := buf.String()
	if !strings.Contains(result, "\\to") {
		t.Errorf("Arrow symbol should be converted to \\to, got %q", result)
	}
}
