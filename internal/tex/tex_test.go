package tex

import (
	"testing"

	"github.com/nlandolfi/lit/internal/lexer"
)

func TestToTeX(t *testing.T) {
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
			name:     "word token with math replacement",
			token:    &lexer.Token{Type: lexer.WordToken, Value: "α"},
			inMath:   true,
			expected: "\\alpha ",
		},
		{
			name:     "ampersand not in math",
			token:    &lexer.Token{Type: lexer.PunctuationToken, Value: "&"},
			inMath:   false,
			expected: "\\&",
		},
		{
			name:     "ampersand in math",
			token:    &lexer.Token{Type: lexer.PunctuationToken, Value: "&"},
			inMath:   true,
			expected: "&",
		},
		{
			name:     "percent sign",
			token:    &lexer.Token{Type: lexer.PunctuationToken, Value: "%"},
			inMath:   false,
			expected: "\\%",
		},
		{
			name:     "underscore not in math",
			token:    &lexer.Token{Type: lexer.PunctuationToken, Value: "_"},
			inMath:   false,
			expected: "\\_",
		},
		{
			name:     "underscore in math",
			token:    &lexer.Token{Type: lexer.PunctuationToken, Value: "_"},
			inMath:   true,
			expected: "_",
		},
		{
			name:     "symbol token space",
			token:    &lexer.Token{Type: lexer.SymbolToken, Value: "␣"},
			inMath:   false,
			expected: " ",
		},
		{
			name:     "symbol token arrow",
			token:    &lexer.Token{Type: lexer.SymbolToken, Value: "→"},
			inMath:   false,
			expected: "\\to",
		},
		{
			name:     "opaque token",
			token:    &lexer.Token{Type: lexer.OpaqueToken, Value: "α + β"},
			inMath:   false,
			expected: "\\alpha + \\beta",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToTeX(tt.token, tt.inMath)
			if result != tt.expected {
				t.Errorf("ToTeX() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestLatexMathReplacements(t *testing.T) {
	// Test a few key replacements
	tests := []struct {
		rune     rune
		expected string
	}{
		{'→', "\\to"},
		{'∈', "\\in"},
		{'∀', "\\forall"},
		{'α', "\\alpha"},
		{'β', "\\beta"},
		{'∞', "\\infty"},
	}

	for _, tt := range tests {
		if replacement, exists := LatexMathReplacements[tt.rune]; !exists {
			t.Errorf("Expected replacement for rune %q to exist", tt.rune)
		} else if replacement != tt.expected {
			t.Errorf("Expected replacement for rune %q to be %q, got %q", tt.rune, tt.expected, replacement)
		}
	}
}
