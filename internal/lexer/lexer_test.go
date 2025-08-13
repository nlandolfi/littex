package lexer

import (
	"testing"
)

func TestLex(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []Token
	}{
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:  "single word",
			input: "hello",
			expected: []Token{
				{Type: WordToken, Value: "hello", Implicit: false},
			},
		},
		{
			name:  "word with space",
			input: "hello world",
			expected: []Token{
				{Type: WordToken, Value: "hello", Implicit: false},
				{Type: SymbolToken, Value: "␣", Implicit: true},
				{Type: WordToken, Value: "world", Implicit: false},
			},
		},
		{
			name:  "punctuation",
			input: "hello,",
			expected: []Token{
				{Type: WordToken, Value: "hello", Implicit: false},
				{Type: PunctuationToken, Value: ",", Implicit: false},
			},
		},
		{
			name:  "symbol",
			input: "hello→world",
			expected: []Token{
				{Type: WordToken, Value: "hello", Implicit: false},
				{Type: SymbolToken, Value: "→", Implicit: false},
				{Type: WordToken, Value: "world", Implicit: false},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens, err := Lex(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(tokens) != len(tt.expected) {
				t.Fatalf("expected %d tokens, got %d", len(tt.expected), len(tokens))
			}

			for i, expected := range tt.expected {
				token := tokens[i]
				if token.Type != expected.Type {
					t.Errorf("token %d: expected type %v, got %v", i, expected.Type, token.Type)
				}
				if token.Value != expected.Value {
					t.Errorf("token %d: expected value %q, got %q", i, expected.Value, token.Value)
				}
				if token.Implicit != expected.Implicit {
					t.Errorf("token %d: expected implicit %v, got %v", i, expected.Implicit, token.Implicit)
				}
			}
		})
	}
}

func TestTokenTypeString(t *testing.T) {
	tests := []struct {
		tt       TokenType
		expected string
	}{
		{ErrorToken, "error"},
		{WordToken, "word"},
		{PunctuationToken, "punctuation"},
		{SymbolToken, "symbol"},
		{OpaqueToken, "opaque"},
	}

	for _, tt := range tests {
		if got := tt.tt.String(); got != tt.expected {
			t.Errorf("TokenType(%d).String() = %q, expected %q", tt.tt, got, tt.expected)
		}
	}
}
