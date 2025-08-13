package lit

import (
	"bytes"
	"strings"
	"testing"

	"github.com/nlandolfi/lit/internal/ast"
	"github.com/nlandolfi/lit/internal/lexer"
	"github.com/nlandolfi/lit/internal/writer"
)

func TestWriteLit(t *testing.T) {
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
			name: "run node",
			node: &ast.Node{
				Type: ast.RunNode,
				FirstChild: &ast.Node{
					Type:  ast.TokenNode,
					Token: &lexer.Token{Type: lexer.WordToken, Value: "hello"},
				},
			},
			expected: "‖hello⦉",
		},
		{
			name: "error node",
			node: &ast.Node{
				Type: ast.ErrorNode,
				Data: "test error",
			},
			expected: "ERROR: test error",
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
			err := WriteLit(&buf, tt.node, nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			result := buf.String()
			if !strings.Contains(result, tt.expected) {
				t.Errorf("expected result to contain %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestWriteLitWithOpts(t *testing.T) {
	node := &ast.Node{
		Type:  ast.TokenNode,
		Token: &lexer.Token{Type: lexer.WordToken, Value: "test"},
	}

	opts := &writer.WriteOpts{
		Prefix: ">> ",
		Indent: "  ",
	}

	var buf bytes.Buffer
	err := WriteLit(&buf, node, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := buf.String()
	if result != "test" {
		t.Errorf("expected %q, got %q", "test", result)
	}
}
