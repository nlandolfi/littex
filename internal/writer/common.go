package writer

import (
	"io"
	"strings"
	"unicode/utf8"

	"github.com/nlandolfi/lit/internal/ast"
	"github.com/nlandolfi/lit/internal/lexer"
	"github.com/nlandolfi/lit/internal/tex"
)

type WriteOpts struct {
	Prefix, Indent string
	InMath         bool
}

var DefaultWriteOpts = &WriteOpts{
	Prefix: "",
	Indent: "  ",
}

func InMath(o *WriteOpts) *WriteOpts {
	var out WriteOpts = *o
	out.InMath = true
	return &out
}

func Indented(o *WriteOpts) *WriteOpts {
	return &WriteOpts{
		Prefix: o.Prefix + o.Indent,
		Indent: o.Indent,
		InMath: o.InMath,
	}
}

func NoPrefix(o *WriteOpts) *WriteOpts {
	return &WriteOpts{
		Indent: o.Indent,
		InMath: o.InMath,
	}
}

// WriteDebug prints the node tree in a pretty format.
func WriteDebug(w io.Writer, n *ast.Node, opts *WriteOpts) {
	if opts == nil {
		opts = DefaultWriteOpts
	}

	io.WriteString(w, opts.Prefix)
	io.WriteString(w, n.Type.String())
	if n.Token != nil {
		io.WriteString(w, ": ")
		io.WriteString(w, n.Token.String())
	}
	if n.Data != "" {
		io.WriteString(w, " [")
		io.WriteString(w, n.Data)
		io.WriteString(w, "]")
	}
	io.WriteString(w, "\n")

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		WriteDebug(w, c, Indented(opts))
	}
}

// writeKids writes all children of a node using the given writer function
func WriteKids(
	w io.Writer,
	n *ast.Node,
	opts *WriteOpts,
	writer func(io.Writer, *ast.Node, *WriteOpts),
) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		writer(w, c, opts)
	}
}

// WriteKidsWithError writes all children of a node using the given writer function that returns an error
func WriteKidsWithError(
	w io.Writer,
	n *ast.Node,
	opts *WriteOpts,
	writer func(io.Writer, *ast.Node, *WriteOpts) error,
) error {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if err := writer(w, c, opts); err != nil {
			return err
		}
	}
	return nil
}

const MaxWidth int = 74

// TokenStringer is a function type for converting tokens to strings
type TokenStringer func(t *lexer.Token, inMath bool) string

// Val converts a token to its string representation for Lit format
func Val(t *lexer.Token, inMath bool) string {
	switch t.Type {
	case lexer.WordToken, lexer.PunctuationToken:
		return t.Value
	case lexer.SymbolToken:
		if t.Value == "␣" {
			return " "
		}
		return t.Value
	case lexer.OpaqueToken:
		return t.Value
	default:
		return t.Value
	}
}

// HTMLVal converts a token to its HTML representation
func HTMLVal(t *lexer.Token, inMath bool) string {
	switch t.Type {
	case lexer.WordToken:
		return t.Value
	case lexer.PunctuationToken:
		switch t.Value {
		case "<":
			return "&lt;"
		case ">":
			return "&gt;"
		case "&":
			return "&amp;"
		case "‹":
			return "<i>"
		case "›":
			return "</i>"
		case "«":
			return "<b>"
		case "»":
			return "</b>"
		case "❮":
			return "<b>"
		case "❯":
			return "</b>"
		case "⸤":
			return "<span style='font-variant:small-caps'>"
		case "⸥":
			return "</span>"
		default:
			return t.Value
		}
	case lexer.SymbolToken:
		if t.Value == "␣" {
			return " "
		}
		return t.Value
	case lexer.OpaqueToken:
		return t.Value
	default:
		return t.Value
	}
}

// IsSpace returns true if the token represents a space
func IsSpace(t *lexer.Token) bool {
	return t.Type == lexer.SymbolToken && t.Value == "␣"
}

// LineBlocks breaks tokens into lines of a given width
func LineBlocks(ts []*lexer.Token, v TokenStringer, opts *WriteOpts, shouldEscapeInMath bool, width int) []string {
	var lines []string
	var currentLine strings.Builder
	var currentLineWidth int

	for i, t := range ts {
		val := v(t, opts.InMath)

		// Handle TeX-specific escaping if needed
		if shouldEscapeInMath {
			// Use TeX conversion if needed
			val = tex.ToTeX(t, opts.InMath)
		}

		// Calculate display width (rough approximation)
		tokenWidth := utf8.RuneCountInString(val)

		// If adding this token would exceed the width, start a new line
		if currentLineWidth > 0 && currentLineWidth+tokenWidth > width {
			lines = append(lines, currentLine.String())
			currentLine.Reset()
			currentLineWidth = 0
		}

		// Add the token to current line
		currentLine.WriteString(val)
		currentLineWidth += tokenWidth

		// If this is not the last token and the next token isn't a space,
		// we might need to add space or handle word boundaries
		if i < len(ts)-1 && !IsSpace(ts[i+1]) && !IsSpace(t) {
			// Add implicit space handling if needed
			if t.Type == lexer.WordToken && ts[i+1].Type == lexer.WordToken {
				currentLine.WriteString(" ")
				currentLineWidth++
			}
		}
	}

	// Add the final line if it has content
	if currentLine.Len() > 0 {
		lines = append(lines, currentLine.String())
	}

	return lines
}

// WriteLines writes lines to a writer with optional prefix
func WriteLines(w io.Writer, lines []string, prefix string, prefixFirst bool) {
	for i, line := range lines {
		if i == 0 && prefixFirst {
			io.WriteString(w, prefix)
		} else if i > 0 && !prefixFirst {
			io.WriteString(w, prefix)
		}
		io.WriteString(w, line)
		if i < len(lines)-1 {
			io.WriteString(w, "\n")
		}
	}
}

// TokenBlockStartingAt extracts a sequence of tokens starting from the given node
func TokenBlockStartingAt(c *ast.Node) (block []*lexer.Token, last *ast.Node) {
	for c != nil && c.Type == ast.TokenNode {
		block = append(block, c.Token)
		last = c
		c = c.NextSibling
	}
	return
}
