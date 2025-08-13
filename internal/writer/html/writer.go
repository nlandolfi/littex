package html

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log"
	"path"
	"strconv"
	"strings"

	"github.com/nlandolfi/lit/internal/ast"
	"github.com/nlandolfi/lit/internal/writer"
	"gopkg.in/yaml.v3"
)

type htmlWriteState struct {
	headerIds []string
}

// WriteHTML writes an AST node in HTML format
func WriteHTML(w io.Writer, n *ast.Node, opts *writer.WriteOpts) error {
	if opts == nil {
		opts = writer.DefaultWriteOpts
	}

	state := &htmlWriteState{}
	return writeHTML(writer.HTMLVal, state, w, n, opts)
}

// WriteHTMLInBody writes HTML content that should go inside a body tag
func WriteHTMLInBody(w io.Writer, n *ast.Node, opts *writer.WriteOpts) {
	if opts == nil {
		opts = writer.DefaultWriteOpts
	}

	state := &htmlWriteState{}
	writeHTML(writer.HTMLVal, state, w, n, opts)
}

func writeHTML(val writer.TokenStringer, s *htmlWriteState, w io.Writer, n *ast.Node, opts *writer.WriteOpts) error {
	switch n.Type {
	case ast.ErrorNode:
		io.WriteString(w, "<div class='error'>ERROR: ")
		io.WriteString(w, html.EscapeString(n.Data))
		io.WriteString(w, "</div>")

	case ast.FragmentNode:
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if err := writeHTML(val, s, w, c, opts); err != nil {
				return err
			}
		}

	case ast.ParagraphNode:
		io.WriteString(w, "<p>")
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if err := writeHTML(val, s, w, c, opts); err != nil {
				return err
			}
		}
		io.WriteString(w, "</p>")

	case ast.FootnoteNode:
		io.WriteString(w, "<footnote>")
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if err := writeHTML(val, s, w, c, opts); err != nil {
				return err
			}
		}
		io.WriteString(w, "</footnote>")

	case ast.DisplayMathNode:
		io.WriteString(w, "<div class='displaymath'>$$")
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if err := writeHTML(val, s, w, c, writer.InMath(opts)); err != nil {
				return err
			}
		}
		io.WriteString(w, "$$</div>")

	case ast.RunNode:
		io.WriteString(w, "<span class='run'>")
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if err := writeHTML(val, s, w, c, opts); err != nil {
				return err
			}
		}
		io.WriteString(w, "</span>")

	case ast.TokenNode:
		if n.Token != nil {
			io.WriteString(w, val(n.Token, opts.InMath))
		}

	case ast.TextNode:
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if err := writeHTML(val, s, w, c, opts); err != nil {
				return err
			}
		}

	case ast.ListNode:
		listType := "ul"
		if n.GetAttr("list-type") == "ordered" {
			listType = "ol"
		}
		io.WriteString(w, "<"+listType+">")
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if err := writeHTML(val, s, w, c, opts); err != nil {
				return err
			}
		}
		io.WriteString(w, "</"+listType+">")

	case ast.ListItemNode:
		io.WriteString(w, "<li>")
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if err := writeHTML(val, s, w, c, opts); err != nil {
				return err
			}
		}
		io.WriteString(w, "</li>")

	case ast.SectionNode:
		level := n.GetAttr("section-level")
		levelNum, _ := strconv.Atoi(level)
		if levelNum < 1 || levelNum > 6 {
			levelNum = 1
		}

		headerTag := fmt.Sprintf("h%d", levelNum)
		io.WriteString(w, "<"+headerTag)

		// Generate header ID for linking
		headerId := generateHeaderId(n, s)
		if headerId != "" {
			io.WriteString(w, " id='"+headerId+"'")
		}

		io.WriteString(w, ">")
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if err := writeHTML(val, s, w, c, opts); err != nil {
				return err
			}
		}
		io.WriteString(w, "</"+headerTag+">")

	case ast.CommentNode:
		io.WriteString(w, "<!-- ")
		io.WriteString(w, html.EscapeString(n.Data))
		io.WriteString(w, " -->")

	case ast.TexOnlyNode:
		// Skip TeX-only content in HTML

	case ast.CenterAlignNode:
		io.WriteString(w, "<div class='center'>")
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if err := writeHTML(val, s, w, c, opts); err != nil {
				return err
			}
		}
		io.WriteString(w, "</div>")

	case ast.RightAlignNode:
		io.WriteString(w, "<div class='right'>")
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if err := writeHTML(val, s, w, c, opts); err != nil {
				return err
			}
		}
		io.WriteString(w, "</div>")

	case ast.EquationNode:
		io.WriteString(w, "<div class='equation'>$$")
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if err := writeHTML(val, s, w, c, writer.InMath(opts)); err != nil {
				return err
			}
		}
		io.WriteString(w, "$$</div>")

	case ast.ImageNode:
		src := n.GetAttr("src")
		alt := n.GetAttr("alt")
		if src != "" {
			io.WriteString(w, "<img src='"+html.EscapeString(src)+"'")
			if alt != "" {
				io.WriteString(w, " alt='"+html.EscapeString(alt)+"'")
			}
			io.WriteString(w, "/>")
		}

	case ast.LinkNode:
		href := n.GetAttr("href")
		if href != "" {
			io.WriteString(w, "<a href='"+html.EscapeString(href)+"'>")
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if err := writeHTML(val, s, w, c, opts); err != nil {
					return err
				}
			}
			io.WriteString(w, "</a>")
		} else {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if err := writeHTML(val, s, w, c, opts); err != nil {
					return err
				}
			}
		}

	case ast.DivNode:
		io.WriteString(w, "<div>")
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if err := writeHTML(val, s, w, c, opts); err != nil {
				return err
			}
		}
		io.WriteString(w, "</div>")

	case ast.CodeNode:
		io.WriteString(w, "<code>")
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if err := writeHTML(val, s, w, c, opts); err != nil {
				return err
			}
		}
		io.WriteString(w, "</code>")

	case ast.PreNode:
		io.WriteString(w, "<pre>")
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if err := writeHTML(val, s, w, c, opts); err != nil {
				return err
			}
		}
		io.WriteString(w, "</pre>")

	case ast.JSONNode:
		if n.JSON != nil {
			if n.IsComment {
				io.WriteString(w, "<!-- ")
				encoder := json.NewEncoder(w)
				encoder.SetIndent(opts.Prefix, opts.Indent)
				_ = encoder.Encode(n.JSON)
				io.WriteString(w, " -->")
			} else {
				io.WriteString(w, "<script type='application/json'>")
				encoder := json.NewEncoder(w)
				encoder.SetIndent(opts.Prefix, opts.Indent)
				_ = encoder.Encode(n.JSON)
				io.WriteString(w, "</script>")
			}
		}

	case ast.YAMLNode:
		if n.YAML != nil {
			if n.IsComment {
				io.WriteString(w, "<!-- ")
				yamlBytes, err := yaml.Marshal(n.YAML)
				if err != nil {
					log.Printf("Error marshaling YAML: %v", err)
				} else {
					_, _ = w.Write(yamlBytes)
				}
				io.WriteString(w, " -->")
			} else {
				io.WriteString(w, "<script type='application/yaml'>")
				yamlBytes, err := yaml.Marshal(n.YAML)
				if err != nil {
					log.Printf("Error marshaling YAML: %v", err)
				} else {
					_, _ = w.Write(yamlBytes)
				}
				io.WriteString(w, "</script>")
			}
		}

	default:
		// For other node types, just process children
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if err := writeHTML(val, s, w, c, opts); err != nil {
				return err
			}
		}
	}

	return nil
}

// generateHeaderId generates a unique ID for a header
func generateHeaderId(n *ast.Node, s *htmlWriteState) string {
	// Collect text from the header
	var textBuilder strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == ast.TokenNode && c.Token != nil {
			textBuilder.WriteString(c.Token.Value)
		}
	}

	text := textBuilder.String()
	if text == "" {
		return ""
	}

	// Convert to slug
	slug := strings.ToLower(text)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = path.Clean(slug)

	// Ensure uniqueness
	originalSlug := slug
	counter := 1
	for contains(s.headerIds, slug) {
		slug = originalSlug + "-" + strconv.Itoa(counter)
		counter++
	}

	s.headerIds = append(s.headerIds, slug)
	return slug
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
