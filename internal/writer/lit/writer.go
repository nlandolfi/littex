package lit

import (
	"io"

	"github.com/nlandolfi/lit/internal/ast"
	"github.com/nlandolfi/lit/internal/writer"
)

// WriteLit writes an AST node in Lit format
func WriteLit(w io.Writer, n *ast.Node, opts *writer.WriteOpts) error {
	if opts == nil {
		opts = writer.DefaultWriteOpts
	}

	return writeLit(w, n, opts)
}

func writeLit(w io.Writer, n *ast.Node, opts *writer.WriteOpts) error {
	switch n.Type {
	case ast.ErrorNode:
		io.WriteString(w, "ERROR: "+n.Data)
		return nil

	case ast.FragmentNode:
		return writer.WriteKidsWithError(w, n, opts, writeLit)

	case ast.ParagraphNode:
		io.WriteString(w, opts.Prefix)
		io.WriteString(w, "¶⦊")
		io.WriteString(w, "\n")
		if err := writer.WriteKidsWithError(w, n, writer.Indented(opts), writeLit); err != nil {
			return err
		}
		io.WriteString(w, "\n")
		io.WriteString(w, opts.Prefix)
		io.WriteString(w, "⦉")
		return nil

	case ast.FootnoteNode:
		io.WriteString(w, opts.Prefix)
		io.WriteString(w, "†⦊")
		io.WriteString(w, "\n")
		if err := writer.WriteKidsWithError(w, n, writer.Indented(opts), writeLit); err != nil {
			return err
		}
		io.WriteString(w, "\n")
		io.WriteString(w, opts.Prefix)
		io.WriteString(w, "⦉")
		return nil

	case ast.DisplayMathNode:
		io.WriteString(w, opts.Prefix)
		io.WriteString(w, "◇⦊")
		io.WriteString(w, "\n")
		if err := writer.WriteKidsWithError(w, n, writer.InMath(writer.Indented(opts)), writeLit); err != nil {
			return err
		}
		io.WriteString(w, "\n")
		io.WriteString(w, opts.Prefix)
		io.WriteString(w, "⦉")
		return nil

	case ast.RunNode:
		io.WriteString(w, "‖")
		if err := writer.WriteKidsWithError(w, n, opts, writeLit); err != nil {
			return err
		}
		io.WriteString(w, "⦉")
		return nil

	case ast.TokenNode:
		if n.Token != nil {
			io.WriteString(w, writer.Val(n.Token, opts.InMath))
		}
		return nil

	case ast.TextNode:
		// Extract tokens and format them properly
		block, _ := writer.TokenBlockStartingAt(n.FirstChild)
		lines := writer.LineBlocks(block, writer.Val, opts, false, writer.MaxWidth)
		writer.WriteLines(w, lines, opts.Prefix, true)
		return nil

	case ast.ListNode:
		io.WriteString(w, opts.Prefix)
		listType := "⁝" // unordered by default
		if n.GetAttr("list-type") == "ordered" {
			listType = "𝍫"
		}
		io.WriteString(w, listType+"⦊")
		io.WriteString(w, "\n")
		if err := writer.WriteKidsWithError(w, n, writer.Indented(opts), writeLit); err != nil {
			return err
		}
		io.WriteString(w, "\n")
		io.WriteString(w, opts.Prefix)
		io.WriteString(w, "⦉")
		return nil

	case ast.ListItemNode:
		io.WriteString(w, opts.Prefix)
		io.WriteString(w, "‣")
		io.WriteString(w, " ")
		if err := writer.WriteKidsWithError(w, n, writer.NoPrefix(opts), writeLit); err != nil {
			return err
		}
		io.WriteString(w, "\n")
		return nil

	case ast.SectionNode:
		io.WriteString(w, opts.Prefix)
		level := n.GetAttr("section-level")
		numbered := n.GetAttr("section-numbered") == "true"
		if numbered {
			io.WriteString(w, "#")
		}
		io.WriteString(w, "§"+level+" ")
		if err := writer.WriteKidsWithError(w, n, writer.NoPrefix(opts), writeLit); err != nil {
			return err
		}
		io.WriteString(w, "\n")
		return nil

	case ast.CommentNode:
		// Comments are typically not written in output
		return nil

	case ast.TexOnlyNode:
		// TeX-only content is skipped in Lit output
		return nil

	case ast.CenterAlignNode:
		io.WriteString(w, opts.Prefix)
		io.WriteString(w, "[center]\n")
		if err := writer.WriteKidsWithError(w, n, writer.Indented(opts), writeLit); err != nil {
			return err
		}
		io.WriteString(w, "\n")
		io.WriteString(w, opts.Prefix)
		io.WriteString(w, "[/center]")
		return nil

	case ast.RightAlignNode:
		io.WriteString(w, opts.Prefix)
		io.WriteString(w, "[right]\n")
		if err := writer.WriteKidsWithError(w, n, writer.Indented(opts), writeLit); err != nil {
			return err
		}
		io.WriteString(w, "\n")
		io.WriteString(w, opts.Prefix)
		io.WriteString(w, "[/right]")
		return nil

	case ast.EquationNode:
		io.WriteString(w, opts.Prefix)
		io.WriteString(w, "[equation]\n")
		if err := writer.WriteKidsWithError(w, n, writer.InMath(writer.Indented(opts)), writeLit); err != nil {
			return err
		}
		io.WriteString(w, "\n")
		io.WriteString(w, opts.Prefix)
		io.WriteString(w, "[/equation]")
		return nil

	default:
		// For other node types, just process children
		return writer.WriteKidsWithError(w, n, opts, writeLit)
	}
}
