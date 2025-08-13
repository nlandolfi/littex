package parser

import "github.com/nlandolfi/lit/internal/ast"

// Must is a utility function like template.Must in std lib
// Use like parser.Must(parser.ParseLit(...))
func Must(n *ast.Node, err error) *ast.Node {
	if err != nil {
		panic(err)
	}
	return n
}
