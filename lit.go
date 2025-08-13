// Package lit provides a clean API for the LitTex archival markup language.
// It compiles to LaTeX and HTML and is designed for mathematical and classical texts.
package lit

import (
	"io"

	"github.com/nlandolfi/lit/internal/ast"
	"github.com/nlandolfi/lit/internal/parser"
	"github.com/nlandolfi/lit/internal/writer"
	"github.com/nlandolfi/lit/internal/writer/html"
	"github.com/nlandolfi/lit/internal/writer/lit"
	"github.com/nlandolfi/lit/internal/writer/tex"
)

// Node represents a node in the Abstract Syntax Tree
type Node = ast.Node

// WriteOpts contains options for writing output
type WriteOpts = writer.WriteOpts

// DefaultWriteOpts provides default options for writing
var DefaultWriteOpts = writer.DefaultWriteOpts

// Parsing functions

// ParseLit parses a LitTex format string and returns the AST
func ParseLit(s string) (*Node, error) {
	return parser.ParseLit(s)
}

// ParseHTML parses an HTML string and returns the AST
func ParseHTML(s string) (*Node, error) {
	return parser.ParseHTML(s)
}

// ParseTex parses a LaTeX string and returns the AST
func ParseTex(s string) (*Node, error) {
	return parser.ParseTex(s)
}

// ParseCSV parses a CSV string and returns the AST
func ParseCSV(s string) (*Node, error) {
	return parser.ParseCSV(s)
}

// Must is a utility function like template.Must in std lib
func Must(n *Node, err error) *Node {
	return parser.Must(n, err)
}

// Writing functions

// WriteLit writes the AST as LitTex format
func WriteLit(w io.Writer, n *Node, opts *WriteOpts) error {
	return lit.WriteLit(w, n, opts)
}

// WriteTex writes the AST as LaTeX format
func WriteTex(w io.Writer, n *Node, opts *WriteOpts) {
	tex.WriteTex(w, n, opts)
}

// WriteHTML writes the AST as HTML format
func WriteHTML(w io.Writer, n *Node, opts *WriteOpts) error {
	return html.WriteHTML(w, n, opts)
}

// WriteHTMLInBody writes HTML content for inclusion in a body tag
func WriteHTMLInBody(w io.Writer, n *Node, opts *WriteOpts) {
	html.WriteHTMLInBody(w, n, opts)
}

// WriteDebug writes a debug representation of the AST
func WriteDebug(w io.Writer, n *Node, opts *WriteOpts) {
	writer.WriteDebug(w, n, opts)
}

// Utility functions for WriteOpts

// InMath returns a copy of WriteOpts with InMath set to true
func InMath(o *WriteOpts) *WriteOpts {
	return writer.InMath(o)
}

// Indented returns a copy of WriteOpts with increased indentation
func Indented(o *WriteOpts) *WriteOpts {
	return writer.Indented(o)
}

// NoPrefix returns a copy of WriteOpts with no prefix
func NoPrefix(o *WriteOpts) *WriteOpts {
	return writer.NoPrefix(o)
}
