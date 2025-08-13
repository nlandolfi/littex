package parser

import (
	"fmt"
	"strings"
	"testing"

	"github.com/nlandolfi/lit/internal/ast"
)

func TestParseLit(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected ast.NodeType
	}{
		{
			name:     "simple text",
			input:    "hello world",
			expected: ast.FragmentNode,
		},
		{
			name:     "paragraph",
			input:    "¶⧊hello world⧉",
			expected: ast.FragmentNode,
		},
		{
			name:     "run",
			input:    "‖hello world⧉",
			expected: ast.FragmentNode,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, err := ParseLit(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if node.Type != tt.expected {
				t.Errorf("expected node type %v, got %v", tt.expected, node.Type)
			}
		})
	}
}

func TestLitReplace(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{
			name:     "paragraph marker",
			input:    "¶⧊hello⧉",
			contains: "data-littype='paragraph'",
		},
		{
			name:     "run marker",
			input:    "‖hello⧉",
			contains: "data-littype='run'",
		},
		{
			name:     "escaped less than",
			input:    "\\<test>",
			contains: "&lt;test>",
		},
		{
			name:     "escaped greater than",
			input:    "\\>test<",
			contains: "&gt;test<",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := litReplace(tt.input)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("expected result to contain %q, got %q", tt.contains, result)
			}
		})
	}
}

func TestMust(t *testing.T) {
	node := &ast.Node{Type: ast.FragmentNode}
	result := Must(node, nil)
	if result != node {
		t.Error("Must should return the node when no error")
	}

	// Test panic on error
	defer func() {
		if r := recover(); r == nil {
			t.Error("Must should panic on error")
		}
	}()
	Must(nil, fmt.Errorf("test error"))
}
