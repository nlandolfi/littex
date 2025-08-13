package writer

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/nlandolfi/lit/internal/ast"
	"github.com/nlandolfi/lit/internal/lexer"
)

func TestWriteOpts(t *testing.T) {
	opts := &WriteOpts{Prefix: ">> ", Indent: "  ", InMath: false}

	// Test InMath
	mathOpts := InMath(opts)
	if !mathOpts.InMath {
		t.Error("InMath should set InMath to true")
	}
	if mathOpts.Prefix != ">> " {
		t.Error("InMath should preserve prefix")
	}
	if mathOpts.Indent != "  " {
		t.Error("InMath should preserve indent")
	}

	// Original should be unchanged
	if opts.InMath {
		t.Error("Original opts should not be modified")
	}

	// Test Indented
	indentedOpts := Indented(opts)
	if indentedOpts.Prefix != ">>   " {
		t.Errorf("Indented should add indent to prefix, expected '>>   ', got '%s'", indentedOpts.Prefix)
	}
	if indentedOpts.Indent != "  " {
		t.Error("Indented should preserve indent")
	}
	if indentedOpts.InMath != false {
		t.Error("Indented should preserve InMath")
	}

	// Test NoPrefix
	noPrefixOpts := NoPrefix(opts)
	if noPrefixOpts.Prefix != "" {
		t.Error("NoPrefix should clear prefix")
	}
	if noPrefixOpts.Indent != "  " {
		t.Error("NoPrefix should preserve indent")
	}
	if noPrefixOpts.InMath != false {
		t.Error("NoPrefix should preserve InMath")
	}
}

func TestWriteDebug(t *testing.T) {
	node := &ast.Node{
		Type: ast.ParagraphNode,
		Data: "test-data",
	}
	token := &lexer.Token{Type: lexer.WordToken, Value: "hello"}
	tokenNode := &ast.Node{
		Type:  ast.TokenNode,
		Token: token,
	}
	node.AppendChild(tokenNode)

	var buf bytes.Buffer
	opts := &WriteOpts{Prefix: "  ", Indent: "  "}
	WriteDebug(&buf, node, opts)

	output := buf.String()
	if !strings.Contains(output, "paragraph") {
		t.Error("Debug output should contain node type")
	}
	if !strings.Contains(output, "test-data") {
		t.Error("Debug output should contain node data")
	}
	if !strings.Contains(output, "token") {
		t.Error("Debug output should contain child token info")
	}
}

func TestVal(t *testing.T) {
	tests := []struct {
		name     string
		token    *lexer.Token
		inMath   bool
		expected string
	}{
		{
			name:     "word token",
			token:    &lexer.Token{Type: lexer.WordToken, Value: "hello"},
			inMath:   false,
			expected: "hello",
		},
		{
			name:     "punctuation token",
			token:    &lexer.Token{Type: lexer.PunctuationToken, Value: "!"},
			inMath:   false,
			expected: "!",
		},
		{
			name:     "space symbol token",
			token:    &lexer.Token{Type: lexer.SymbolToken, Value: "␣"},
			inMath:   false,
			expected: " ",
		},
		{
			name:     "other symbol token",
			token:    &lexer.Token{Type: lexer.SymbolToken, Value: "→"},
			inMath:   false,
			expected: "→",
		},
		{
			name:     "opaque token",
			token:    &lexer.Token{Type: lexer.OpaqueToken, Value: "test"},
			inMath:   false,
			expected: "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Val(tt.token, tt.inMath)
			if result != tt.expected {
				t.Errorf("Val() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestHTMLVal(t *testing.T) {
	tests := []struct {
		name     string
		token    *lexer.Token
		inMath   bool
		expected string
	}{
		{
			name:     "word token",
			token:    &lexer.Token{Type: lexer.WordToken, Value: "hello"},
			inMath:   false,
			expected: "hello",
		},
		{
			name:     "less than",
			token:    &lexer.Token{Type: lexer.PunctuationToken, Value: "<"},
			inMath:   false,
			expected: "&lt;",
		},
		{
			name:     "greater than",
			token:    &lexer.Token{Type: lexer.PunctuationToken, Value: ">"},
			inMath:   false,
			expected: "&gt;",
		},
		{
			name:     "ampersand",
			token:    &lexer.Token{Type: lexer.PunctuationToken, Value: "&"},
			inMath:   false,
			expected: "&amp;",
		},
		{
			name:     "italic start",
			token:    &lexer.Token{Type: lexer.PunctuationToken, Value: "‹"},
			inMath:   false,
			expected: "<i>",
		},
		{
			name:     "italic end",
			token:    &lexer.Token{Type: lexer.PunctuationToken, Value: "›"},
			inMath:   false,
			expected: "</i>",
		},
		{
			name:     "space symbol",
			token:    &lexer.Token{Type: lexer.SymbolToken, Value: "␣"},
			inMath:   false,
			expected: " ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HTMLVal(tt.token, tt.inMath)
			if result != tt.expected {
				t.Errorf("HTMLVal() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestIsSpace(t *testing.T) {
	spaceToken := &lexer.Token{Type: lexer.SymbolToken, Value: "␣"}
	wordToken := &lexer.Token{Type: lexer.WordToken, Value: "hello"}
	nonSpaceSymbol := &lexer.Token{Type: lexer.SymbolToken, Value: "→"}

	if !IsSpace(spaceToken) {
		t.Error("IsSpace should return true for space symbol")
	}
	if IsSpace(wordToken) {
		t.Error("IsSpace should return false for word token")
	}
	if IsSpace(nonSpaceSymbol) {
		t.Error("IsSpace should return false for non-space symbol")
	}
}

func TestTokenBlockStartingAt(t *testing.T) {
	// Create a sequence of token nodes followed by a non-token
	token1 := &ast.Node{Type: ast.TokenNode, Token: &lexer.Token{Type: lexer.WordToken, Value: "hello"}}
	token2 := &ast.Node{Type: ast.TokenNode, Token: &lexer.Token{Type: lexer.SymbolToken, Value: "␣"}}
	token3 := &ast.Node{Type: ast.TokenNode, Token: &lexer.Token{Type: lexer.WordToken, Value: "world"}}
	paragraph := &ast.Node{Type: ast.ParagraphNode}

	// Set up sibling relationships
	token1.NextSibling = token2
	token2.NextSibling = token3
	token3.NextSibling = paragraph

	block, last := TokenBlockStartingAt(token1)

	if len(block) != 3 {
		t.Errorf("Expected 3 tokens, got %d", len(block))
	}
	if last != token3 {
		t.Error("Last should be token3")
	}
	if block[0].Value != "hello" || block[1].Value != "␣" || block[2].Value != "world" {
		t.Error("Token block contents don't match expected values")
	}
}

func TestWriteLines(t *testing.T) {
	lines := []string{"line1", "line2", "line3"}
	var buf bytes.Buffer

	// Test with prefix on first line
	buf.Reset()
	WriteLines(&buf, lines, ">> ", true)
	expected := ">> line1\nline2\nline3"
	if buf.String() != expected {
		t.Errorf("WriteLines with prefixFirst=true: expected %q, got %q", expected, buf.String())
	}

	// Test without prefix on first line
	buf.Reset()
	WriteLines(&buf, lines, ">> ", false)
	expected = "line1\n>> line2\n>> line3"
	if buf.String() != expected {
		t.Errorf("WriteLines with prefixFirst=false: expected %q, got %q", expected, buf.String())
	}

	// Test empty lines
	buf.Reset()
	WriteLines(&buf, []string{}, ">> ", true)
	if buf.String() != "" {
		t.Error("WriteLines with empty lines should produce empty output")
	}

	// Test single line
	buf.Reset()
	WriteLines(&buf, []string{"single"}, ">> ", true)
	if buf.String() != ">> single" {
		t.Errorf("WriteLines with single line: expected '>> single', got %q", buf.String())
	}
}

func TestWriteKids(t *testing.T) {
	parent := &ast.Node{Type: ast.FragmentNode}
	child1 := &ast.Node{Type: ast.TokenNode, Token: &lexer.Token{Type: lexer.WordToken, Value: "hello"}}
	child2 := &ast.Node{Type: ast.TokenNode, Token: &lexer.Token{Type: lexer.WordToken, Value: "world"}}

	parent.AppendChild(child1)
	parent.AppendChild(child2)

	var buf bytes.Buffer
	writerFunc := func(w io.Writer, n *ast.Node, opts *WriteOpts) {
		if n.Type == ast.TokenNode && n.Token != nil {
			w.Write([]byte(n.Token.Value + " "))
		}
	}

	WriteKids(&buf, parent, DefaultWriteOpts, writerFunc)

	if buf.String() != "hello world " {
		t.Errorf("WriteKids output: expected 'hello world ', got %q", buf.String())
	}
}

func TestWriteKidsWithError(t *testing.T) {
	parent := &ast.Node{Type: ast.FragmentNode}
	child1 := &ast.Node{Type: ast.TokenNode, Token: &lexer.Token{Type: lexer.WordToken, Value: "hello"}}
	child2 := &ast.Node{Type: ast.TokenNode, Token: &lexer.Token{Type: lexer.WordToken, Value: "world"}}

	parent.AppendChild(child1)
	parent.AppendChild(child2)

	var buf bytes.Buffer
	writerFunc := func(w io.Writer, n *ast.Node, opts *WriteOpts) error {
		if n.Type == ast.TokenNode && n.Token != nil {
			w.Write([]byte(n.Token.Value + " "))
		}
		return nil
	}

	err := WriteKidsWithError(&buf, parent, DefaultWriteOpts, writerFunc)
	if err != nil {
		t.Errorf("WriteKidsWithError should not return error: %v", err)
	}

	if buf.String() != "hello world " {
		t.Errorf("WriteKidsWithError output: expected 'hello world ', got %q", buf.String())
	}

	// Test error propagation
	errorFunc := func(w io.Writer, n *ast.Node, opts *WriteOpts) error {
		return fmt.Errorf("test error")
	}

	err = WriteKidsWithError(&buf, parent, DefaultWriteOpts, errorFunc)
	if err == nil {
		t.Error("WriteKidsWithError should propagate errors")
	}
	if err.Error() != "test error" {
		t.Errorf("WriteKidsWithError should propagate correct error message, got: %v", err)
	}
}
