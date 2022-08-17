package lit

import (
	"fmt"
	"html"
	"io"
	"log"
	"strings"
	"unicode/utf8"
)

// WriteDebug prints the node tree in a pretty format.
func WriteDebug(w io.Writer, n *Node, prefix, indent string) {
	fmt.Fprintf(w, prefix+"%s", n.Type)
	if n.Type == TokenNode {
		fmt.Fprintf(w, ":%v", n.Token)
	}
	fmt.Fprintf(w, "\n")
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		WriteDebug(w, c, prefix+indent, indent)
	}
}

func WriteLit(w io.Writer, n *Node, prefix, indent string) {
	switch n.Type {
	case FragmentNode:
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			WriteLit(w, c, prefix, indent)
		}
	case ParagraphNode, ListNode:
		/*
			if n.FirstChild == nil {
				log.Printf("skipping empty paragraph or list")
				return // just skip!
			}
		*/
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
			if n.PrevSibling.Type == ParagraphNode {
				w.Write([]byte("\n"))
			}
		}
		switch n.Type {
		case ParagraphNode:
			w.Write([]byte(prefix + "¶ ⦊\n"))
		case ListNode:
			switch getAttr(n.Attr, "list-type") {
			case "ordered":
				w.Write([]byte(prefix + "𝍫 ⦊\n"))
			default: // includes unordered
				w.Write([]byte(prefix + "⁝ ⦊\n"))
			}
		default:
			panic("not reached")
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			WriteLit(w, c, prefix+indent, indent)
		}
		w.Write([]byte("\n" + prefix + "⦉"))
	case FootnoteNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}
		w.Write([]byte(prefix + "† ⦊\n"))
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			WriteLit(w, c, prefix+indent, indent)
		}
		w.Write([]byte("\n" + prefix + "⦉"))
	case DisplayMathNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}
		w.Write([]byte(prefix + "◇ ⦊\n"))
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			WriteLit(w, c, prefix+indent, indent)
		}
		w.Write([]byte("\n" + prefix + "⦉"))
	case RunNode, ListItemNode, SectionNode:
		/*
			if n.FirstChild == nil {
				log.Printf("skipping empty run")
				return // just skip!
			}
		*/
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}

		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}

		var out string
		switch n.Type {
		case RunNode:
			if n.PrevSibling == nil && n.Parent != nil && n.Parent.Type == ListItemNode {
				out = "‖ "
			} else {
				out = prefix + "‖ "
			}
		case ListItemNode:
			out = prefix + "‣ "
		case SectionNode:
			w.Write([]byte(prefix))
			if getAttr(n.Attr, "section-numbered") == "true" {
				w.Write([]byte("#"))
			}

			switch getAttr(n.Attr, "section-level") {
			case "1":
				w.Write([]byte("§ "))
			case "2":
				w.Write([]byte("§§ "))
			case "3":
				w.Write([]byte("§§§ "))
			default:
				w.Write([]byte("§ "))
			}
		default:
			panic("not reached")
		}

		w.Write([]byte(out))

		offset := utf8.RuneCountInString(out)
		//log.Printf("offset; %d", offset)

		//		lineBuffer = ""
		//relOffset := 0

		var afterFirstLine bool
		// Looping over the children.
		for c := n.FirstChild; c != nil; afterFirstLine = true {
			//log.Printf("RUN NODE: %s", c.Type)
			switch c.Type {
			case TokenNode:
				if c.PrevSibling != nil {
					w.Write([]byte("\n"))
				}

				// in case its a token, go a find all tokens to next non-token
				block, lastTokenNode := tokenBlockStartingAt(c)
				allowedWidth := maxWidth - offset
				lines := lineBlocks(block, Val, allowedWidth)
				if len(lines) > 0 {
					writeLines(w, lines, prefix+indent, afterFirstLine)
					afterFirstLine = true
				}

				c = lastTokenNode.NextSibling
			default:
				WriteLit(w, c, prefix+indent, indent)
				c = c.NextSibling
			}
		}

		if n.LastChild != nil && n.LastChild.Type == TokenNode {
			w.Write([]byte(" "))
		}
		// will need to do overflow check
		w.Write([]byte("⦉"))
	case TextNode:
		log.Printf("text nodes should not appear...")
		lines := strings.Split(n.Data, "\n")
		for i, l := range lines {
			lines[i] = prefix + l
		}
		out := strings.Join(lines, "\n")
		w.Write([]byte(out + "\n"))
	case CommentNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}
		w.Write([]byte(prefix + "<!--" + n.Data + "-->"))
	case TexOnlyNode, RightAlignNode, CenterAlignNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}
		if n.PrevSibling != nil && (n.PrevSibling.Type == ParagraphNode || n.PrevSibling.Type == ListNode) {
			w.Write([]byte("\n"))
		}
		var dataatom string
		switch n.Type {
		case RightAlignNode:
			dataatom = "right"
		case CenterAlignNode:
			dataatom = "center"
		case TexOnlyNode:
			dataatom = "tex"
		default:
			panic("not reached")
		}
		w.Write([]byte(prefix + "<" + dataatom + ">\n"))
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			WriteLit(w, c, prefix+indent, indent)
		}
		w.Write([]byte("\n" + prefix + "</" + dataatom + ">"))
	case EquationNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}
		if n.PrevSibling != nil && (n.PrevSibling.Type == ParagraphNode || n.PrevSibling.Type == ListNode) {
			w.Write([]byte("\n"))
		}
		w.Write([]byte(prefix + "<equation"))
		if id := getAttr(n.Attr, "id"); id != "" {
			w.Write([]byte(" " + "id='" + id + "'"))
		}
		w.Write([]byte(">\n"))
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			WriteLit(w, c, prefix+indent, indent)
		}
		w.Write([]byte("\n" + prefix + "</equation>"))
	case ImageNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}
		if n.PrevSibling != nil && (n.PrevSibling.Type == ParagraphNode || n.PrevSibling.Type == ListNode || n.PrevSibling.Type == RunNode) {
			w.Write([]byte("\n"))
		}
		w.Write([]byte(prefix + fmt.Sprintf("<img src=\"%s\"", getAttr(n.Attr, "src"))))
		if width := getAttr(n.Attr, "width"); width != "" {
			w.Write([]byte(fmt.Sprintf(" width=\"%s\"", width)))
		}
		w.Write([]byte("/>"))
	case StatementNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}
		if n.PrevSibling != nil && (n.PrevSibling.Type == ParagraphNode || n.PrevSibling.Type == ListNode) {
			w.Write([]byte("\n"))
		}
		w.Write([]byte(prefix + "<statement"))
		if id := getAttr(n.Attr, "id"); id != "" {
			w.Write([]byte(" " + "id='" + id + "'"))
		}
		if t := getAttr(n.Attr, "type"); t != "" {
			w.Write([]byte(" " + "type='" + t + "'"))
		}
		if text := getAttr(n.Attr, "text"); text != "" {
			w.Write([]byte(" " + "text='" + text + "'"))
		}
		w.Write([]byte(">\n"))
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			WriteLit(w, c, prefix+indent, indent)
		}
		w.Write([]byte("\n" + prefix + "</statement>"))
	case ProofNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}
		if n.PrevSibling != nil && (n.PrevSibling.Type == ParagraphNode || n.PrevSibling.Type == ListNode) {
			w.Write([]byte("\n"))
		}
		w.Write([]byte(prefix + "<proof"))
		w.Write([]byte(">\n"))
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			WriteLit(w, c, prefix+indent, indent)
		}
		w.Write([]byte("\n" + prefix + "</proof>"))
	case LinkNode:
		panic("TODO: links not implemented")
	default:
		log.Printf("prev: %v; cur: %v; next: %v", n.PrevSibling, n, n.NextSibling)
		panic(fmt.Sprintf("unhandled node type: %s", n.Type))
	}
}

func WriteTex(w io.Writer, n *Node, prefix, indent string) {
	switch n.Type {
	case FragmentNode:
		writeKids(w, n, prefix, indent, WriteTex)
	case ParagraphNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}
		writeKids(w, n, prefix+indent, indent, WriteTex)
		w.Write([]byte("\n"))
	case ListNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}
		lt := getAttr(n.Attr, "list-type")
		switch lt {
		case "ordered":
			w.Write([]byte(prefix + "\\begin{enumerate}\n"))
		default: // unordered
			w.Write([]byte(prefix + "\\begin{itemize}\n"))
		}
		writeKids(w, n, prefix+indent, indent, WriteTex)
		switch lt {
		case "ordered":
			w.Write([]byte("\n" + prefix + "\\end{enumerate}"))
		default: // unordered
			w.Write([]byte("\n" + prefix + "\\end{itemize}"))
		}
	case FootnoteNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}
		// the first little bit here removes the space.
		// it means there is no way to have a space in between
		// text and a footnote, but I don't think there's a way
		// to specify that within .gba files.
		// Could in the future check if the previous child here was a token
		// and a non implicit space token
		w.Write([]byte(prefix + "\\ifhmode\\unskip\\fi\\footnote{\n"))
		writeKids(w, n, prefix+indent, indent, WriteTex)
		w.Write([]byte("\n" + prefix + "}"))
	case DisplayMathNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}
		w.Write([]byte(prefix + "\\[\n"))
		writeKids(w, n, prefix+indent, indent, WriteTex)
		w.Write([]byte("\n" + prefix + "\\]"))
	case RunNode, ListItemNode, SectionNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}

		var offset int
		if n.Type == ListItemNode {
			out := prefix + "\\item "
			w.Write([]byte(out))
			offset = utf8.RuneCountInString(out)
		}
		if n.Type == SectionNode {
			switch getAttr(n.Attr, "section-level") {
			case "1":
				w.Write([]byte(indent + "\\section"))
			case "2":
				w.Write([]byte(indent + "\\subsection"))
			case "3":
				w.Write([]byte(indent + "\\subsubsection"))
			default:
				w.Write([]byte(indent + "\\section"))
			}
			if getAttr(n.Attr, "section-numbered") == "false" {
				w.Write([]byte("*"))
			}
			w.Write([]byte(indent + "{"))
		}

		var afterFirstLine bool
		// Looping over the children.
		for c := n.FirstChild; c != nil; afterFirstLine = true {
			//log.Printf("RUN NODE: %s", c.Type)
			switch c.Type {
			case TokenNode:
				if c.PrevSibling != nil && c.PrevSibling.Type != RunNode {
					w.Write([]byte("\n"))
				}

				// in case its a token, go a find all tokens to next non-token
				block, lastTokenNode := tokenBlockStartingAt(c)
				allowedWidth := maxWidth - offset
				lines := lineBlocks(block, Tex, allowedWidth)
				if len(lines) > 0 {
					//					writeLines(w, lines, prefix+indent, afterFirstLine)
					w.Write([]byte(strings.Join(lines, " ")))
					if afterFirstLine {
						afterFirstLine = true
					}
					afterFirstLine = true
				}

				c = lastTokenNode.NextSibling
			default:
				WriteTex(w, c, prefix+indent, indent)
				c = c.NextSibling
			}
		}
		if n.Type == SectionNode {
			w.Write([]byte(indent + "}\n"))
		}
	case CommentNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}
		if n.PrevSibling != nil && (n.PrevSibling.Type == ParagraphNode || n.PrevSibling.Type == ListNode) {
			w.Write([]byte("\n"))
		}
		for _, line := range strings.Split(n.Data, "\n") {
			if strings.TrimSpace(line) == "" {
				continue
			}
			w.Write([]byte(indent + "%" + line + "\n"))
		}
	case TexOnlyNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			WriteTex(w, c, indent, indent) // intentionally don't increase indent
		}
	case CenterAlignNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}
		w.Write([]byte("\\begin{center}"))
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			WriteTex(w, c, indent, indent) // intentionally don't increase indent
		}
		w.Write([]byte("\\end{center}"))
	case RightAlignNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}
		w.Write([]byte("\\begin{flushright}"))
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			WriteTex(w, c, indent, indent) // intentionally don't increase indent
		}
		w.Write([]byte("\\end{flushright}"))
	case EquationNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}
		w.Write([]byte("\\begin{equation}"))
		if id := getAttr(n.Attr, "id"); id != "" {
			w.Write([]byte("\n" + prefix + "\\label{" + id + "}"))
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			WriteTex(w, c, indent, indent) // intentionally don't increase indent
		}
		w.Write([]byte("\\end{equation}"))
	case ImageNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}
		if n.PrevSibling != nil && (n.PrevSibling.Type == ParagraphNode || n.PrevSibling.Type == ListNode || n.PrevSibling.Type == RunNode) {
			w.Write([]byte("\n"))
		}
		w.Write([]byte(prefix + fmt.Sprintf("\\includegraphics")))
		if width := getAttr(n.Attr, "width"); width != "" {
			if width[len(width)-1] == '%' {
				width = "0." + width[:len(width)-1] + "\\textwidth"
			}
			w.Write([]byte(fmt.Sprintf("[width=%s]", width)))
		}
		w.Write([]byte(fmt.Sprintf("{%s}", getAttr(n.Attr, "src"))))
	case StatementNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}
		t := "statement"
		if tt := getAttr(n.Attr, "type"); tt != "" {
			t = tt
		}
		w.Write([]byte(fmt.Sprintf("\\begin{%s}", t)))
		if text := getAttr(n.Attr, "text"); text != "" {
			w.Write([]byte(fmt.Sprintf("[%s]", text)))
		}
		if id := getAttr(n.Attr, "id"); id != "" {
			w.Write([]byte("\n" + prefix + "\\label{" + id + "}"))
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			WriteTex(w, c, indent, indent) // intentionally don't increase indent
		}
		w.Write([]byte(fmt.Sprintf("\\end{%s}", t)))
	case ProofNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}
		w.Write([]byte("\\begin{proof}"))
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			WriteTex(w, c, indent, indent) // intentionally don't increase indent
		}
		w.Write([]byte("\\end{proof}"))
	case LinkNode:
		panic("TODO: links not implemented")
	default:
		panic(fmt.Sprintf("unhandled node type: %s", n.Type))
	}
}

type nodeWriter func(w io.Writer, n *Node, pr, in string)

func writeKids(
	w io.Writer, n *Node, in, pr string,
	write nodeWriter,
) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		write(w, c, pr, in)
	}
}

const maxWidth int = 74

func Val(t *Token) string {
	if t.Type == OpaqueToken {
		return string(OpaqueOpenRune) + t.Value + string(OpaqueCloseRune)
	}
	out := t.Value
	if t.Implicit && isSpace(t) {
		out = " "
	}
	return out
}

func HTMLVal(t *Token) string {
	switch t.Value {
	case "᜶":
		return "<br />"
	case "«":
		return "<b>"
	case "»":
		return "</b>"
	case "‹":
		return "<i>"
	case "›":
		return "</i>"
	case "⸤":
		return "<span class='smallcaps'>"
	case "⸥":
		return "</span>"
	case "❬":
		return "<span class='term'>"
	case "❭":
		return "</span>"
	case "\\begin{flushright}":
		return "<span class='flushright'>"
	case "\\end{flushright}":
		return "</span>"
	case "↦":
		return "&nbsp;&nbsp;&nbsp;&nbsp;"
	case "＆":
		return "\\&"
	case "⁅":
		return "<span class='typewriter'>"
	case "⁆":
		return "</span>"
	}

	if t.Type == OpaqueToken {
		return t.Value
	}

	return html.EscapeString(Val(t))
}

func isSpace(t *Token) bool {
	return t.Type == PunctuationToken && t.Value == "␣"
}

type tokenStringer func(*Token) string

func lineBlocks(ts []*Token, v tokenStringer, width int) []string {
	var pieces = []string{""}
	var spaces []*Token
	for _, t := range ts {
		if isSpace(t) {
			spaces = append(spaces, t)
			pieces = append(pieces, "")
		} else {
			pieces[len(pieces)-1] = pieces[len(pieces)-1] + v(t)
		}
	}

	var lines []string = []string{""}
	var curRuneCount = 0
	var onePieceOnLine bool
	for i, p := range pieces {
		var lastPiece = (len(pieces)-1 == i)

		c := utf8.RuneCountInString(p) + 1 // for the space
		if lastPiece {
			c -= 1 // except the last piece
		}

		if curRuneCount+c > width {
			lines = append(lines, p)
			curRuneCount = c
			onePieceOnLine = true
		} else {
			nl := lines[len(lines)-1]
			if onePieceOnLine {
				nl += v(spaces[i-1])
			}
			nl += p
			if !lastPiece {
				nl += v(spaces[i])
			}
			lines[len(lines)-1] = nl
			curRuneCount += c + 1
			onePieceOnLine = false
		}
	}

	var o []string = make([]string, 0, len(lines))
	for _, line := range lines {
		// these are sort of quick fix hacks
		// to avoid end of line spaces and
		// empty lines, but really the code above
		// should ultimately be fixed
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		o = append(o, line)
	}

	return o
}

func writeLines(w io.Writer, lines []string, prefix string, prefixFirst bool) {
	for i, line := range lines {
		lastLine := (i == len(lines)-1)
		if i == 0 && !prefixFirst {
			w.Write([]byte(line))
		} else {
			w.Write([]byte(prefix + line))
		}

		if !lastLine {
			w.Write([]byte("\n"))
		}
	}
}

func tokenBlockStartingAt(c *Node) (block []*Token, last *Node) {
	block = append(block, c.Token)
	last = c

	var cc *Node
	for cc = c.NextSibling; cc != nil && cc.Type == TokenNode; cc = cc.NextSibling {
		block = append(block, cc.Token)
		last = cc
	}
	return block, last
}

func WriteHTMLInBody(w io.Writer, n *Node, prefix, indent string) {
	w.Write([]byte("<!DOCTYPE html>\n"))
	w.Write([]byte(`<meta charset="utf-8"/>`))
	w.Write([]byte("\n" + indent + "<body>\n"))
	WriteHTML(w, n, prefix+indent+indent, indent)
	w.Write([]byte("\n" + indent + "</body>\n"))
	w.Write([]byte("</html>"))
}

func WriteHTML(w io.Writer, n *Node, prefix, indent string) error {
	s := new(htmlWriteState)
	writeHTML(HTMLVal, s, w, n, prefix, indent)

	fmt.Fprintf(w, "<ol class='footnotes'>")
	for i, f := range s.footnotes {
		fmt.Fprintf(w, "<li id='footnote-%d'>", i+1)
		for c := f.FirstChild; c != nil; c = c.NextSibling {
			writeHTML(HTMLVal, nil, w, c, prefix+indent, indent)
		}
		fmt.Fprintf(w, "<a href='#footnote-%d-reference'>↩︎</a>", i+1)
		fmt.Fprintf(w, "</li>")
	}
	fmt.Fprintf(w, "</ol>")
	return nil
}

type htmlWriteState struct {
	footnotes []*Node
}

func writeHTML(val func(t *Token) string, s *htmlWriteState, w io.Writer, n *Node, prefix, indent string) error {
	switch n.Type {
	case FragmentNode:
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			writeHTML(val, s, w, c, prefix, indent)
		}
	case ParagraphNode, ListNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}
		if n.PrevSibling != nil && (n.PrevSibling.Type == ParagraphNode || n.PrevSibling.Type == ListNode) {
			w.Write([]byte("\n"))
		}
		switch n.Type {
		case ParagraphNode:
			w.Write([]byte(prefix + "<p>\n"))
		case ListNode:
			switch getAttr(n.Attr, "list-type") {
			case "ordered":
				w.Write([]byte(prefix + "<ol>\n"))
			default: // includes unordered
				w.Write([]byte(prefix + "<ul>\n"))
			}
		default:
			panic("not reached")
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			writeHTML(val, s, w, c, prefix+indent, indent)
		}
		switch n.Type {
		case ParagraphNode:
			w.Write([]byte("\n" + prefix + "</p>\n"))
		case ListNode:
			switch getAttr(n.Attr, "list-type") {
			case "ordered":
				w.Write([]byte(prefix + "</ol>\n"))
			default: // includes unordered
				w.Write([]byte(prefix + "</ul>\n"))
			}
		default:
			panic("not reached")
		}
	case FootnoteNode:
		s.footnotes = append(s.footnotes, n)
		d := len(s.footnotes)
		fmt.Fprintf(w, "<sup id='footnote-%d-reference'><a href='#footnote-%d'>%d</a></sup>", d, d, d)
	case DisplayMathNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}
		w.Write([]byte(prefix + "<p>\\[\n"))
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			writeHTML(Tex, s, w, c, prefix+indent, indent)
		}
		w.Write([]byte("\n" + prefix + "\\]</p>"))
	case RunNode, ListItemNode, SectionNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}

		if n.PrevSibling != nil && (n.PrevSibling.Type == RunNode || n.PrevSibling.Type == ListItemNode) {
			w.Write([]byte("\n"))
		}

		var out string
		switch n.Type {
		case RunNode:
			if n.PrevSibling == nil && n.Parent != nil && n.Parent.Type == ListItemNode {
				if n.Parent != nil && (n.Parent.Type == DisplayMathNode || n.Parent.Type == EquationNode) {
					out = ""
				} else {
					out = "<span class='run'>"
				}
			} else {
				if n.Parent != nil && (n.Parent.Type == DisplayMathNode || n.Parent.Type == EquationNode) {
					out = prefix
				} else {
					out = prefix + "<span class='run'>"
				}
			}
		case ListItemNode:
			out = prefix + "<li>"
		case SectionNode:
			switch getAttr(n.Attr, "section-level") {
			case "1":
				w.Write([]byte(prefix + "<h1>\n"))
			case "2":
				w.Write([]byte(prefix + "<h2>\n"))
			case "3":
				w.Write([]byte(prefix + "<h3>\n"))
			default:
				w.Write([]byte(prefix + "<h1>\n"))
			}
		}

		w.Write([]byte(out))

		offset := utf8.RuneCountInString(out)
		//log.Printf("offset; %d", offset)

		//		lineBuffer = ""
		//relOffset := 0

		var afterFirstLine bool
		// Looping over the children.
		for c := n.FirstChild; c != nil; afterFirstLine = true {
			//log.Printf("RUN NODE: %s", c.Type)
			switch c.Type {
			case TokenNode:
				if c.PrevSibling != nil {
					w.Write([]byte("\n"))
				}

				// in case its a token, go a find all tokens to next non-token
				block, lastTokenNode := tokenBlockStartingAt(c)
				allowedWidth := maxWidth - offset
				lines := lineBlocks(block, val, allowedWidth)
				if len(lines) > 0 {
					writeLines(w, lines, prefix+indent, afterFirstLine)
					afterFirstLine = true
				}

				c = lastTokenNode.NextSibling
			default:
				writeHTML(val, s, w, c, prefix+indent, indent)
				c = c.NextSibling
			}
		}

		if n.LastChild != nil && n.LastChild.Type == TokenNode {
			w.Write([]byte(" "))
		}
		// will need to do overflow check
		switch n.Type {
		case RunNode:
			if n.Parent != nil && (n.Parent.Type == DisplayMathNode || n.Parent.Type == EquationNode) {
			} else {
				w.Write([]byte("</span>"))
			}
		case ListItemNode:
			w.Write([]byte("</li>"))
		case SectionNode:
			switch getAttr(n.Attr, "section-level") {
			case "1":
				w.Write([]byte(prefix + "</h1>\n"))
			case "2":
				w.Write([]byte(prefix + "</h2>\n"))
			case "3":
				w.Write([]byte(prefix + "</h3>\n"))
			default:
				w.Write([]byte(prefix + "</h1>\n"))
			}
		default:
			panic("not reached")
		}
	case TextNode:
		log.Printf("text nodes should not appear...")
		lines := strings.Split(n.Data, "\n")
		for i, l := range lines {
			lines[i] = prefix + l
		}
		out := strings.Join(lines, "\n")
		w.Write([]byte(out + "\n"))
	case CommentNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}
		if n.PrevSibling != nil && (n.PrevSibling.Type == ParagraphNode || n.PrevSibling.Type == ListNode) {
			w.Write([]byte("\n"))
		}
		w.Write([]byte(prefix + "<!--" + n.Data + "-->\n"))
	case TexOnlyNode:
	case CenterAlignNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}
		w.Write([]byte("<div style='text-align:center'>"))
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			writeHTML(val, s, w, c, indent, indent) // intentionally don't increase indent
		}
		w.Write([]byte("</div>"))
	case RightAlignNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}
		w.Write([]byte("<div style='text-align:right'>"))
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			writeHTML(val, s, w, c, indent, indent) // intentionally don't increase indent
		}
		w.Write([]byte("</div>"))
	case EquationNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}
		w.Write([]byte(prefix + "<div style='equation'>"))
		w.Write([]byte(prefix + "\\begin{equation}"))
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			writeHTML(Tex, s, w, c, prefix+indent, indent) // intentionally don't increase indent
		}
		if id := getAttr(n.Attr, "id"); id != "" {
			w.Write([]byte(prefix + indent + "\\label{" + id + "}"))
		}
		w.Write([]byte(prefix + "\\end{equation}"))
		w.Write([]byte(prefix + "</div>"))
	case ImageNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}
		if n.PrevSibling != nil && (n.PrevSibling.Type == ParagraphNode || n.PrevSibling.Type == ListNode || n.PrevSibling.Type == RunNode) {
			w.Write([]byte("\n"))
		}
		w.Write([]byte(prefix + fmt.Sprintf("<img src=\"%s\"", getAttr(n.Attr, "src"))))
		if width := getAttr(n.Attr, "width"); width != "" {
			w.Write([]byte(fmt.Sprintf(" width=\"%s\"", width)))
		}
		w.Write([]byte("/>"))
	case StatementNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}
		t := "statement"
		if tt := getAttr(n.Attr, "type"); tt != "" {
			t = tt
		}
		w.Write([]byte(prefix + fmt.Sprintf("<div class='%s'", t)))
		if text := getAttr(n.Attr, "text"); text != "" {
			w.Write([]byte(fmt.Sprintf(" text='%s'", text)))
		}
		if id := getAttr(n.Attr, "id"); id != "" {
			w.Write([]byte(fmt.Sprintf(" id='%s'", id)))
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			writeHTML(val, s, w, c, prefix+indent, indent) // intentionally don't increase indent
		}
		w.Write([]byte(prefix + "</div>"))
	case ProofNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}
		w.Write([]byte(prefix + "<div class='proof'>"))
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			writeHTML(val, s, w, c, prefix+indent, indent) // intentionally don't increase indent
		}
		w.Write([]byte(prefix + "</div>"))
	case LinkNode:
		w.Write([]byte(prefix + fmt.Sprintf("<a href='%s'", getAttr(n.Attr, "href"))))
		w.Write([]byte("/>"))
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			writeHTML(Tex, s, w, c, prefix+indent, indent) // intentionally don't increase indent
		}
		w.Write([]byte(prefix + "</a>"))
	default:
		return fmt.Errorf("unhandled node type %s, prev: %v; cur: %v; next: %v", n.Type, n.PrevSibling, n, n.NextSibling)
	}
	return nil
}
