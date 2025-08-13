package parser

import (
	"encoding/csv"
	"io"
	"strings"

	"github.com/nlandolfi/lit/internal/ast"
	"github.com/nlandolfi/lit/internal/lexer"
)

// ParseCSV parses CSV data and converts it to the AST
func ParseCSV(s string) (*ast.Node, error) {
	fragment := ast.Node{Type: ast.FragmentNode}

	r := csv.NewReader(strings.NewReader(s))
	for {
		list := &ast.Node{Type: ast.ListNode}
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		for _, field := range record {
			li := &ast.Node{Type: ast.ListItemNode}
			var ts []*lexer.Token
			ts, err = lexer.Lex(field)
			if err != nil {
				return nil, err
			}

			for _, t := range ts {
				tn := &ast.Node{Type: ast.TokenNode, Token: t}
				li.AppendChild(tn)
			}
			list.AppendChild(li)
		}
		fragment.AppendChild(list)
	}

	return &fragment, nil
}
