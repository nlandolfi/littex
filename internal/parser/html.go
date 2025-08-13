package parser

import (
	"bytes"
	"strconv"

	"github.com/nlandolfi/lit/internal/ast"
	"github.com/nlandolfi/lit/internal/lexer"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// ParseHTML parses an HTML string and returns the corresponding AST
func ParseHTML(s string) (*ast.Node, error) {
	var fragment html.Node = html.Node{
		Type:     html.ElementNode,
		DataAtom: atom.Div,
		Data:     "div",
		Attr: []html.Attribute{
			{Key: "data-littype", Val: "fragment"},
		},
	}

	ns, err := html.ParseFragment(bytes.NewBufferString(s), &fragment)
	if err != nil {
		return nil, err
	}
	for _, n := range ns {
		fragment.AppendChild(n)
	}
	nGBA, err := UnmarshalHTML(&fragment)
	return nGBA, err
}

// UnmarshalHTML converts an html.Node to an ast.Node
func UnmarshalHTML(in *html.Node) (*ast.Node, error) {
	return unmarshalHTML(in, nil)
}

func unmarshalHTMLText(in *html.Node) (tokens []*ast.Node, err error) {
	text := in.Data
	var ts []*lexer.Token
	ts, err = lexer.Lex(text)
	if err != nil {
		return
	}

	for _, t := range ts {
		tokens = append(tokens, &ast.Node{
			Type:  ast.TokenNode,
			Token: t,
		})
	}

	return
}

func unmarshalHTML(in *html.Node, parent *ast.Node) (*ast.Node, error) {
	var out *ast.Node

	switch in.Type {
	case html.ErrorNode:
		out = &ast.Node{
			Type: ast.ErrorNode,
			Data: in.Data,
		}
	case html.TextNode:
		tokens, err := unmarshalHTMLText(in)
		if err != nil {
			return nil, err
		}

		// If we have exactly one token, just return it
		if len(tokens) == 1 {
			return tokens[0], nil
		}

		// Otherwise, create a text node to hold them
		out = &ast.Node{
			Type: ast.TextNode,
		}
		for _, token := range tokens {
			// Clear parent relationship before appending
			token.Parent = nil
			token.PrevSibling = nil
			token.NextSibling = nil
			out.AppendChild(token)
		}
	case html.DocumentNode:
		out = &ast.Node{
			Type: ast.FragmentNode,
		}
	case html.ElementNode:
		littype := ast.LittypeOf(in)
		switch littype {
		case "error":
			out = &ast.Node{
				Type: ast.ErrorNode,
			}
		case "fragment":
			out = &ast.Node{
				Type: ast.FragmentNode,
			}
		case "paragraph":
			out = &ast.Node{
				Type: ast.ParagraphNode,
			}
		case "footnote":
			out = &ast.Node{
				Type: ast.FootnoteNode,
			}
		case "displaymath":
			out = &ast.Node{
				Type: ast.DisplayMathNode,
			}
		case "run":
			out = &ast.Node{
				Type: ast.RunNode,
			}
		case "list":
			out = &ast.Node{
				Type: ast.ListNode,
			}
			out.SetAttr("list-type", ast.LitlisttypeOf(in.Attr))
		case "list-item":
			out = &ast.Node{
				Type: ast.ListItemNode,
			}
		case "section":
			out = &ast.Node{
				Type: ast.SectionNode,
			}
			out.SetAttr("section-level", ast.LitsectionlevelOf(in.Attr))
			out.SetAttr("section-numbered", ast.LitsectionnumberedOf(in.Attr))
		case "comment":
			out = &ast.Node{
				Type: ast.CommentNode,
			}
		case "tex":
			out = &ast.Node{
				Type: ast.TexOnlyNode,
			}
		case "center":
			out = &ast.Node{
				Type: ast.CenterAlignNode,
			}
		case "right":
			out = &ast.Node{
				Type: ast.RightAlignNode,
			}
		case "equation":
			out = &ast.Node{
				Type: ast.EquationNode,
			}
		case "draw":
			out = &ast.Node{
				Type: ast.DrawNode,
			}
		case "json":
			out = &ast.Node{
				Type: ast.JSONNode,
			}
		case "yaml":
			out = &ast.Node{
				Type: ast.YAMLNode,
			}
		default:
			// Handle regular HTML elements
			switch in.DataAtom {
			case atom.Img:
				out = &ast.Node{
					Type: ast.ImageNode,
				}
				out.SetAttr("src", ast.LitimgsrcOf(in.Attr))
			case atom.A:
				out = &ast.Node{
					Type: ast.LinkNode,
				}
			case atom.Div:
				out = &ast.Node{
					Type: ast.DivNode,
				}
			case atom.Code:
				out = &ast.Node{
					Type: ast.CodeNode,
				}
			case atom.Pre:
				out = &ast.Node{
					Type: ast.PreNode,
				}
			default:
				out = &ast.Node{
					Type: ast.DivNode,
				}
			}
		}
		out.DataAtom = in.DataAtom
		out.Data = in.Data
		out.Attr = make([]ast.Attribute, len(in.Attr))
		copy(out.Attr, in.Attr)
	case html.CommentNode:
		out = &ast.Node{
			Type: ast.CommentNode,
			Data: in.Data,
		}
	case html.DoctypeNode:
		out = &ast.Node{
			Type: ast.FragmentNode,
			Data: in.Data,
		}
	default:
		out = &ast.Node{
			Type: ast.ErrorNode,
			Data: "unknown node type: " + strconv.Itoa(int(in.Type)),
		}
	}

	out.Parent = parent

	// Process children
	for c := in.FirstChild; c != nil; c = c.NextSibling {
		child, err := unmarshalHTML(c, out)
		if err != nil {
			return nil, err
		}
		if child != nil {
			// Clear any existing parent relationship before appending
			child.Parent = nil
			child.PrevSibling = nil
			child.NextSibling = nil
			out.AppendChild(child)
		}
	}

	return out, nil
}
