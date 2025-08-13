package tex

import (
	"io"
	"strconv"
	"strings"

	"github.com/nlandolfi/lit/internal/ast"
	"github.com/nlandolfi/lit/internal/tex"
	"github.com/nlandolfi/lit/internal/writer"
)

// WriteTex writes an AST node in TeX format
func WriteTex(w io.Writer, n *ast.Node, opts *writer.WriteOpts) {
	if opts == nil {
		opts = writer.DefaultWriteOpts
	}
	writeTex(w, n, opts)
}

func writeTex(w io.Writer, n *ast.Node, opts *writer.WriteOpts) {
	switch n.Type {
	case ast.ErrorNode:
		io.WriteString(w, "% ERROR: "+n.Data+"\n")

	case ast.FragmentNode:
		writer.WriteKids(w, n, opts, writeTex)

	case ast.ParagraphNode:
		io.WriteString(w, "\n")
		writeTokensForTex(w, n, opts)
		io.WriteString(w, "\n\n")

	case ast.FootnoteNode:
		io.WriteString(w, "\\footnote{")
		writeTokensForTex(w, n, opts)
		io.WriteString(w, "}")

	case ast.DisplayMathNode:
		io.WriteString(w, "\n\\begin{displaymath}\n")
		writeTokensForTex(w, n, writer.InMath(opts))
		io.WriteString(w, "\n\\end{displaymath}\n")

	case ast.RunNode:
		writeTokensForTex(w, n, opts)

	case ast.TokenNode:
		if n.Token != nil {
			io.WriteString(w, tex.ToTeX(n.Token, opts.InMath))
		}

	case ast.TextNode:
		// Process children tokens
		writer.WriteKids(w, n, opts, writeTex)

	case ast.ListNode:
		listType := "itemize"
		if n.GetAttr("list-type") == "ordered" {
			listType = "enumerate"
		}
		io.WriteString(w, "\n\\begin{"+listType+"}\n")
		writer.WriteKids(w, n, writer.Indented(opts), writeTex)
		io.WriteString(w, "\\end{"+listType+"}\n")

	case ast.ListItemNode:
		io.WriteString(w, opts.Prefix+"\\item ")
		writer.WriteKids(w, n, writer.NoPrefix(opts), writeTex)
		io.WriteString(w, "\n")

	case ast.SectionNode:
		level := n.GetAttr("section-level")
		numbered := n.GetAttr("section-numbered") == "true"

		var sectionCmd string
		switch level {
		case "1":
			if numbered {
				sectionCmd = "\\section"
			} else {
				sectionCmd = "\\section*"
			}
		case "2":
			if numbered {
				sectionCmd = "\\subsection"
			} else {
				sectionCmd = "\\subsection*"
			}
		case "3":
			if numbered {
				sectionCmd = "\\subsubsection"
			} else {
				sectionCmd = "\\subsubsection*"
			}
		default:
			levelNum, _ := strconv.Atoi(level)
			if levelNum > 3 {
				// For deeper levels, use paragraph
				if numbered {
					sectionCmd = "\\paragraph"
				} else {
					sectionCmd = "\\paragraph*"
				}
			} else {
				sectionCmd = "\\section"
			}
		}

		io.WriteString(w, "\n"+sectionCmd+"{")
		writeTokensForTex(w, n, opts)
		io.WriteString(w, "}\n")

	case ast.CommentNode:
		io.WriteString(w, "% "+n.Data+"\n")

	case ast.TexOnlyNode:
		// Include TeX-only content directly
		writer.WriteKids(w, n, opts, writeTex)

	case ast.CenterAlignNode:
		io.WriteString(w, "\n\\begin{center}\n")
		writer.WriteKids(w, n, writer.Indented(opts), writeTex)
		io.WriteString(w, "\n\\end{center}\n")

	case ast.RightAlignNode:
		io.WriteString(w, "\n\\begin{flushright}\n")
		writer.WriteKids(w, n, writer.Indented(opts), writeTex)
		io.WriteString(w, "\n\\end{flushright}\n")

	case ast.EquationNode:
		io.WriteString(w, "\n\\begin{equation}\n")
		writer.WriteKids(w, n, writer.InMath(writer.Indented(opts)), writeTex)
		io.WriteString(w, "\n\\end{equation}\n")

	case ast.ImageNode:
		src := n.GetAttr("src")
		if src != "" {
			io.WriteString(w, "\\includegraphics{"+src+"}")
		}

	case ast.LinkNode:
		href := n.GetAttr("href")
		if href != "" {
			io.WriteString(w, "\\href{"+href+"}{")
			writer.WriteKids(w, n, opts, writeTex)
			io.WriteString(w, "}")
		} else {
			writer.WriteKids(w, n, opts, writeTex)
		}

	case ast.CodeNode:
		io.WriteString(w, "\\texttt{")
		writer.WriteKids(w, n, opts, writeTex)
		io.WriteString(w, "}")

	case ast.PreNode:
		io.WriteString(w, "\n\\begin{verbatim}\n")
		writer.WriteKids(w, n, opts, writeTex)
		io.WriteString(w, "\n\\end{verbatim}\n")

	default:
		// For other node types, just process children
		writer.WriteKids(w, n, opts, writeTex)
	}
}

// writeTokensForTex writes tokens in TeX format with proper line breaking
func writeTokensForTex(w io.Writer, n *ast.Node, opts *writer.WriteOpts) {
	// Collect all tokens from the node and its children
	var tokens []*ast.Node
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == ast.TokenNode {
			tokens = append(tokens, c)
		}
	}

	// Write tokens with proper formatting
	for i, token := range tokens {
		if token.Token != nil {
			io.WriteString(w, tex.ToTeX(token.Token, opts.InMath))
		}

		// Add space between tokens if needed
		if i < len(tokens)-1 &&
			token.Token != nil && tokens[i+1].Token != nil &&
			!writer.IsSpace(token.Token) && !writer.IsSpace(tokens[i+1].Token) {
			// Add implicit space for readability
			if strings.HasSuffix(tex.ToTeX(token.Token, opts.InMath), " ") {
				// Space already added by TeX conversion
			} else {
				io.WriteString(w, " ")
			}
		}
	}
}
