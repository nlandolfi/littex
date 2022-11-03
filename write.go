package lit

import (
	"bytes"
	"fmt"
	"html"
	"io"
	"log"
	"path"
	"strings"
	"unicode/utf8"
)

type WriteOpts struct {
	Prefix, Indent string
	InMath         bool
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
func WriteDebug(w io.Writer, n *Node, opts *WriteOpts) {
	fmt.Fprintf(w, opts.Prefix+"%s", n.Type)
	switch n.Type {
	case TokenNode:
		fmt.Fprintf(w, ":%v", n.Token)
	}
	if len(n.Attr) > 0 {
		fmt.Fprintf(w, "(")
		for i, a := range n.Attr {
			fmt.Fprintf(w, "%s=%q", a.Key, a.Val)
			if i != len(n.Attr)-1 {
				fmt.Fprintf(w, ", ")
			}
		}
		fmt.Fprintf(w, ")")
	}
	fmt.Fprintf(w, "\n")
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		WriteDebug(w, c, &WriteOpts{
			Prefix: opts.Prefix + opts.Indent,
			Indent: opts.Indent,
		})
	}
}

func WriteLit(w io.Writer, n *Node, opts *WriteOpts) {
	switch n.Type {
	case FragmentNode:
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			WriteLit(w, c, opts)
		}
	case ParagraphNode, ListNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
			if n.PrevSibling.Type == ParagraphNode {
				w.Write([]byte("\n"))
			}
		}
		switch n.Type {
		case ParagraphNode:
			w.Write([]byte(opts.Prefix + "¶ ⦊\n"))
		case ListNode:
			switch getAttr(n.Attr, "list-type") {
			case "ordered":
				w.Write([]byte(opts.Prefix + "𝍫 ⦊\n"))
			default: // includes unordered
				w.Write([]byte(opts.Prefix + "⁝ ⦊\n"))
			}
		default:
			panic("not reached")
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			WriteLit(w, c, Indented(opts))
		}
		w.Write([]byte("\n" + opts.Prefix + "⦉"))
	case FootnoteNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}
		w.Write([]byte(opts.Prefix + "† ⦊\n"))
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			WriteLit(w, c, Indented(opts))
		}
		w.Write([]byte("\n" + opts.Prefix + "⦉"))
	case DisplayMathNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}
		w.Write([]byte(opts.Prefix + "◇ ⦊\n"))
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			WriteLit(w, c, Indented(opts))
		}
		w.Write([]byte("\n" + opts.Prefix + "⦉"))
	case RunNode, ListItemNode, SectionNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n\n"))
		}

		var out string
		switch n.Type {
		case RunNode:
			if n.PrevSibling == nil && n.Parent != nil && n.Parent.Type == ListItemNode {
				out = "‖ "
			} else {
				out = opts.Prefix + "‖ "
			}
		case ListItemNode:
			out = opts.Prefix + "‣ "
		case SectionNode:
			w.Write([]byte(opts.Prefix))
			if n.SectionNumbered() {
				w.Write([]byte("#"))
			}

			// maybe don't bother with these helper functions Section*
			switch n.SectionLevel() {
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
				lines := lineBlocks(block, Val, opts, false, allowedWidth)
				if len(lines) > 0 {
					writeLines(w, lines, opts.Prefix+opts.Indent, afterFirstLine)
					afterFirstLine = true
				}

				c = lastTokenNode.NextSibling
			default:
				WriteLit(w, c, Indented(opts))
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
			lines[i] = opts.Prefix + l
		}
		out := strings.Join(lines, "\n")
		w.Write([]byte(out + "\n"))
	case CommentNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n\n"))
		}
		w.Write([]byte(opts.Prefix + "<!--" + n.Data + "-->"))
	case TexOnlyNode, RightAlignNode, CenterAlignNode, TableNode, TableHeadNode, TableBodyNode, TableRowNode, THNode, TDNode, SubequationsNode, QuoteNode, DivNode, CodeNode:
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
		case TableNode:
			dataatom = "table"
		case TableHeadNode:
			dataatom = "thead"
		case TableBodyNode:
			dataatom = "tbody"
		case TableRowNode:
			dataatom = "tr"
		case THNode:
			dataatom = "th"
		case TDNode:
			dataatom = "td"
		case SubequationsNode:
			dataatom = "subequations"
		case QuoteNode:
			dataatom = "quote"
		case DivNode:
			dataatom = "div"
		case CodeNode:
			dataatom = "code"
		default:
			panic("not reached")
		}
		w.Write([]byte(opts.Prefix + "<" + dataatom))
		switch n.Type {
		case TableNode, TableHeadNode, TableBodyNode, TableRowNode, THNode, TDNode, DivNode, CodeNode:
			for _, a := range n.Attr {
				w.Write([]byte(fmt.Sprintf(" %s='%s'", a.Key, a.Val)))
			}
		}
		w.Write([]byte(">"))
		if n.FirstChild != nil {
			w.Write([]byte("\n"))
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			WriteLit(w, c, Indented(opts))
		}
		if n.FirstChild != nil {
			w.Write([]byte("\n" + opts.Prefix))
		}
		w.Write([]byte("</" + dataatom + ">"))
	case EquationNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}
		if n.PrevSibling != nil && (n.PrevSibling.Type == ParagraphNode || n.PrevSibling.Type == ListNode) {
			w.Write([]byte("\n"))
		}
		w.Write([]byte(opts.Prefix + "<equation"))
		if id := getAttr(n.Attr, "id"); id != "" {
			w.Write([]byte(" " + "id='" + id + "'"))
		}
		w.Write([]byte(">\n"))
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			WriteLit(w, c, Indented(opts))
		}
		w.Write([]byte("\n" + opts.Prefix + "</equation>"))
	case ImageNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}
		if n.PrevSibling != nil && (n.PrevSibling.Type == ParagraphNode || n.PrevSibling.Type == ListNode || n.PrevSibling.Type == RunNode) {
			w.Write([]byte("\n"))
		}
		w.Write([]byte(opts.Prefix + fmt.Sprintf("<img src='%s'", getAttr(n.Attr, "src"))))
		if width := getAttr(n.Attr, "width"); width != "" {
			w.Write([]byte(fmt.Sprintf(" width='%s'", width)))
		}
		w.Write([]byte("/>"))
	case StatementNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}
		if n.PrevSibling != nil && (n.PrevSibling.Type == ParagraphNode || n.PrevSibling.Type == ListNode) {
			w.Write([]byte("\n"))
		}
		w.Write([]byte(opts.Prefix + "<statement"))
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
			WriteLit(w, c, Indented(opts))
		}
		w.Write([]byte("\n" + opts.Prefix + "</statement>"))
	case ProofNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}
		if n.PrevSibling != nil && (n.PrevSibling.Type == ParagraphNode || n.PrevSibling.Type == ListNode) {
			w.Write([]byte("\n"))
		}
		w.Write([]byte(opts.Prefix + "<proof>\n"))
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			WriteLit(w, c, Indented(opts))
		}
		w.Write([]byte("\n" + opts.Prefix + "</proof>"))
	case LinkNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}
		w.Write([]byte(opts.Prefix + fmt.Sprintf("<a href='%s'>\n", getAttr(n.Attr, "href"))))
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			WriteLit(w, c, Indented(opts))
		}
		w.Write([]byte("\n" + opts.Prefix + "</a>"))
	default:
		log.Print("WriteLit")
		log.Printf("prev: %v; cur: %v; next: %v", n.PrevSibling, n, n.NextSibling)
		log.Fatalf("unhandled node type: %s", n.Type)
		//		panic(fmt.Sprintf("unhandled node type: %s", n.Type))
	}
}

func WriteTex(w io.Writer, n *Node, opts *WriteOpts) {
	switch n.Type {
	case FragmentNode:
		writeKids(w, n, opts, WriteTex)
	case ParagraphNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}
		writeKids(w, n, Indented(opts), WriteTex)
		w.Write([]byte("\n"))
	case ListNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}
		lt := getAttr(n.Attr, "list-type")
		switch lt {
		case "ordered":
			w.Write([]byte(opts.Prefix + "\\begin{enumerate}\n"))
		default: // unordered
			w.Write([]byte(opts.Prefix + "\\begin{itemize}\n"))
		}
		writeKids(w, n, Indented(opts), WriteTex)
		switch lt {
		case "ordered":
			w.Write([]byte("\n" + opts.Prefix + "\\end{enumerate}"))
		default: // unordered
			w.Write([]byte("\n" + opts.Prefix + "\\end{itemize}"))
		}
	case FootnoteNode:
		// the first little bit here removes the space.
		// it means there is no way to have a space in between
		// text and a footnote, but I don't think there's a way
		// to specify that within .gba files.
		// Could in the future check if the previous child here was a token
		// and a non implicit space token
		w.Write([]byte("\\footnote{"))
		writeKids(w, n, opts, WriteTex)
		w.Write([]byte("}"))
	case DisplayMathNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}
		w.Write([]byte(opts.Prefix + "\\[\n"))
		writeKids(w, n, InMath(Indented(opts)), WriteTex)
		w.Write([]byte("\n" + opts.Prefix + "\\]"))
	case RunNode, ListItemNode, SectionNode:
		if n.PrevSibling != nil && n.PrevSibling.Type != LinkNode {
			w.Write([]byte("\n"))
		}

		var offset int
		if n.Type == ListItemNode {
			out := opts.Prefix + "\\item "
			w.Write([]byte(out))
			offset = utf8.RuneCountInString(out)
		}
		if n.Type == SectionNode {
			switch getAttr(n.Attr, "section-level") {
			case "1":
				w.Write([]byte(opts.Prefix + "\\section"))
			case "2":
				w.Write([]byte(opts.Prefix + "\\subsection"))
			case "3":
				w.Write([]byte(opts.Prefix + "\\subsubsection"))
			default:
				w.Write([]byte(opts.Prefix + "\\section"))
			}
			if getAttr(n.Attr, "section-numbered") == "false" {
				w.Write([]byte("*"))
			}
			w.Write([]byte("{"))
		}

		var afterFirstLine bool
		// Looping over the children.
		for c := n.FirstChild; c != nil; afterFirstLine = true {
			//log.Printf("RUN NODE: %s", c.Type)
			switch c.Type {
			case TokenNode:
				if c.PrevSibling != nil && c.PrevSibling.Type != RunNode && c.PrevSibling.Type != LinkNode {
					w.Write([]byte("\n"))
				}

				// in case its a token, go a find all tokens to next non-token
				block, lastTokenNode := tokenBlockStartingAt(c)
				allowedWidth := maxWidth - offset
				lines := lineBlocks(block, Tex, opts, true, allowedWidth)
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
				WriteTex(w, c, Indented(opts))
				c = c.NextSibling
			}
		}
		if n.Type == SectionNode {
			w.Write([]byte("}\n"))
		}
	case CommentNode:
		for _, line := range strings.Split(n.Data, "\n") {
			if line == "" {
				continue
			}
			w.Write([]byte("\n%" + line))
		}
		w.Write([]byte("\n"))
	case TexOnlyNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			WriteTex(w, c, opts)
		}
	case DivNode:
		w.Write([]byte("{"))
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			WriteTex(w, c, opts)
		}
		w.Write([]byte("}"))
	case CodeNode:
		w.Write([]byte("\\texttt{"))
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			WriteTex(w, c, opts)
		}
		w.Write([]byte("}"))
	case CenterAlignNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}
		w.Write([]byte("\\begin{center}"))
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			WriteTex(w, c, opts)
		}
		w.Write([]byte("\\end{center}"))
	case RightAlignNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}
		w.Write([]byte("\\begin{flushright}"))
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			WriteTex(w, c, opts)
		}
		w.Write([]byte("\\end{flushright}"))
	case QuoteNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}
		w.Write([]byte("\\begin{quote}"))
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			WriteTex(w, c, opts)
		}
		w.Write([]byte("\\end{quote}"))
	case EquationNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}
		w.Write([]byte("\\begin{equation}"))
		if id := getAttr(n.Attr, "id"); id != "" {
			w.Write([]byte("\n" + opts.Prefix + "\\label{" + id + "}"))
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			WriteTex(w, c, InMath(opts))
		}
		w.Write([]byte("\\end{equation}"))
	case SubequationsNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}
		w.Write([]byte("\\begin{subequations}"))
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			WriteTex(w, c, InMath(opts))
		}
		w.Write([]byte("\\end{subequations}"))
	case ImageNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}
		if n.PrevSibling != nil && (n.PrevSibling.Type == ParagraphNode || n.PrevSibling.Type == ListNode || n.PrevSibling.Type == RunNode) {
			w.Write([]byte("\n"))
		}
		w.Write([]byte(opts.Prefix + fmt.Sprintf("\\includegraphics")))
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
		w.Write([]byte("\n"))
		if id := getAttr(n.Attr, "id"); id != "" {
			w.Write([]byte("\n" + opts.Prefix + "\\label{" + id + "}"))
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			WriteTex(w, c, opts)
		}
		w.Write([]byte(fmt.Sprintf("\\end{%s}", t)))
	case ProofNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}
		w.Write([]byte("\\begin{proof}"))
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			WriteTex(w, c, opts)
		}
		w.Write([]byte("\\end{proof}"))
	case LinkNode:
		href := getAttr(n.Attr, "href")
		if strings.HasPrefix(href, "/sheets/") {
			sheetname := strings.TrimSuffix(path.Base(href), path.Ext(href))
			w.Write([]byte(fmt.Sprintf(" \\sheetref{%s}{", sheetname)))

		} else {
			w.Write([]byte(fmt.Sprintf(" \\href{%s}{", href)))
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			WriteTex(w, c, opts)
		}
		w.Write([]byte("}"))
	case TableNode:
		//		w.Write([]byte(fmt.Sprintf("\\begin{table}\n")))
		// TODO: one day write table...
		w.Write([]byte(fmt.Sprintf("\n\\vspace{0.3cm}\n\\begin{tabular}{%s}\n", getAttr(n.Attr, "tex"))))
		// TODO write the alignments at some point

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			WriteTex(w, c, opts)
		}
		w.Write([]byte("\\end{tabular}\n\\vspace{0.1cm}\n"))
		//		w.Write([]byte(fmt.Sprintf("\\end{table}\n")))
	case TableHeadNode:
		// TODO write some rules \toprule, \bottomrule, etc
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			WriteTex(w, c, opts)
		}
	case TableBodyNode:
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			WriteTex(w, c, opts)
		}
	case TableRowNode:
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			WriteTex(w, c, opts)
		}
		w.Write([]byte("\\\\\n"))
	case THNode, TDNode:
		if n.PrevSibling != nil && (n.PrevSibling.Type == THNode || n.PrevSibling.Type == TDNode) {
			w.Write([]byte(" & "))
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			WriteTex(w, c, opts)
		}
	default:
		log.Fatalf("unhandled node type: %s", n.Type)
	}
}

type nodeWriter func(w io.Writer, n *Node, opts *WriteOpts)

func writeKids(
	w io.Writer, n *Node, opts *WriteOpts,
	write nodeWriter,
) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		write(w, c, opts)
	}
}

const maxWidth int = 74

func Val(t *Token, inMath bool) string {
	if inMath {
		return Tex(t, inMath)
	}
	if t.Type == OpaqueToken {
		return string(OpaqueOpenRune) + t.Value + string(OpaqueCloseRune)
	}
	out := t.Value
	if t.Implicit && isSpace(t) {
		out = " "
	}
	switch out {
	case "¶", "‖", "†", "◇", "⁝", "𝍫", "‣", "§", "⦉":
		out = "\\" + out
		/*
			case "<", ">":
				if !inMath {
					out = "\\" + out
				}
		*/
	}
	return out
}

func HTMLVal(t *Token, inMath bool) string {
	if isSpace(t) && !t.Implicit {
		return "&nbsp;"
	}
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
	case "↦":
		if inMath {
			return "\\mapsto"
		} else {
			return "&nbsp;&nbsp;&nbsp;&nbsp;"
		}
	case "＆": // This is a full-width &
		return "\\&"
	case "⁅":
		return "<span class='typewriter'>"
	case "⁆":
		return "</span>"
	}

	if t.Type == OpaqueToken {
		return t.Value
	}

	switch t.Value {
	case "¶", "‖", "◇", "†", "⁝", "‣", "𝍫", "§", "⦉":
		return t.Value
		/*
			case "<", ">":
				return html.EscapeString(t.Value)
		*/
	}

	return html.EscapeString(Val(t, inMath))
}

func isSpace(t *Token) bool {
	return t.Type == SymbolToken && t.Value == "␣"
}

type tokenStringer func(t *Token, inMath bool) string

func lineBlocks(ts []*Token, v tokenStringer, opts *WriteOpts, shouldEscapeInMath bool, width int) []string {
	var pieces = []string{""}
	var spaces []*Token
	var inMath bool = opts.InMath
	for _, t := range ts {
		if isSpace(t) {
			spaces = append(spaces, t)
			pieces = append(pieces, "")
		} else {
			pieces[len(pieces)-1] = pieces[len(pieces)-1] + v(t, inMath && shouldEscapeInMath)
		}
		if t.Type == SymbolToken && t.Value == "$" {
			if opts.InMath {
				panic("$ in math")
			}
			inMath = !inMath
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
				nl += v(spaces[i-1], false) // assume it's always right to write as text
			}
			nl += p
			if !lastPiece {
				nl += v(spaces[i], false) // assume it's always right to write as text
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

func WriteHTMLInBody(w io.Writer, n *Node, opts *WriteOpts) {
	w.Write([]byte("<!DOCTYPE html>\n"))
	w.Write([]byte(`<meta charset="utf-8"/>`))
	w.Write([]byte("\n" + opts.Indent + "<body>\n"))
	WriteHTML(w, n, Indented(Indented(opts)))
	w.Write([]byte("\n" + opts.Indent + "</body>\n"))
	w.Write([]byte("</html>"))
}

func WriteHTML(w io.Writer, n *Node, opts *WriteOpts) error {
	s := new(htmlWriteState)
	writeHTML(HTMLVal, s, w, n, opts)

	if len(s.footnotes) > 0 {

		fmt.Fprintf(w, "<hr style='margin-top:0.5in'>")
		fmt.Fprintf(w, "<ol class='footnotes'>")
		for i, f := range s.footnotes {
			fmt.Fprintf(w, "<li id='footnote-%d'>", i+1)
			for c := f.FirstChild; c != nil; c = c.NextSibling {
				writeHTML(HTMLVal, nil, w, c, Indented(opts))
			}
			fmt.Fprintf(w, " <a href='#footnote-%d-reference'>↩︎</a>", i+1)
			fmt.Fprintf(w, "</li>")
		}
		fmt.Fprintf(w, "</ol>")
	}
	return nil
}

type htmlWriteState struct {
	footnotes []*Node
}

func writeHTML(val tokenStringer, s *htmlWriteState, w io.Writer, n *Node, opts *WriteOpts) error {
	switch n.Type {
	case FragmentNode:
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			writeHTML(val, s, w, c, opts)
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
			w.Write([]byte(opts.Prefix + "<p>\n"))
		case ListNode:
			switch getAttr(n.Attr, "list-type") {
			case "ordered":
				w.Write([]byte(opts.Prefix + "<ol>\n"))
			default: // includes unordered
				w.Write([]byte(opts.Prefix + "<ul>\n"))
			}
		default:
			panic("not reached")
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			writeHTML(val, s, w, c, Indented(opts))
		}
		switch n.Type {
		case ParagraphNode:
			w.Write([]byte("\n" + opts.Prefix + "</p>\n"))
		case ListNode:
			switch getAttr(n.Attr, "list-type") {
			case "ordered":
				w.Write([]byte(opts.Prefix + "</ol>\n"))
			default: // includes unordered
				w.Write([]byte(opts.Prefix + "</ul>\n"))
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
		w.Write([]byte(opts.Prefix + "<p>\\[\n"))
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			writeHTML(Tex, s, w, c, InMath(Indented(opts)))
		}
		w.Write([]byte("\n" + opts.Prefix + "\\]</p>"))
	case RunNode, ListItemNode, SectionNode:
		if n.PrevSibling != nil && n.PrevSibling.Type != LinkNode {
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
					out = opts.Prefix
				} else {
					out = opts.Prefix + "<span class='run'>"
				}
			}
		case ListItemNode:
			out = opts.Prefix + "<li>"
		case SectionNode:
			switch getAttr(n.Attr, "section-level") {
			case "1":
				w.Write([]byte(opts.Prefix + "<h1 "))
			case "2":
				w.Write([]byte(opts.Prefix + "<h2 "))
			case "3":
				w.Write([]byte(opts.Prefix + "<h3 "))
			default:
				w.Write([]byte(opts.Prefix + "<h1 "))
			}
			w.Write([]byte(fmt.Sprintf("id='%s'", n.FirstChild.headerTokenString())))
			w.Write([]byte(">"))
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
				if c.PrevSibling != nil && c.PrevSibling.Type != LinkNode {
					w.Write([]byte("\n"))
				}

				// in case its a token, go a find all tokens to next non-token
				block, lastTokenNode := tokenBlockStartingAt(c)
				allowedWidth := maxWidth - offset
				lines := lineBlocks(block, val, opts, true, allowedWidth)
				if len(lines) > 0 {
					prefix := opts.Prefix + opts.Indent
					if c.PrevSibling != nil && c.PrevSibling.Type == LinkNode {
						prefix = ""
					}
					writeLines(w, lines, prefix, afterFirstLine)
					afterFirstLine = true
				}

				c = lastTokenNode.NextSibling
			default:
				writeHTML(val, s, w, c, Indented(opts))
				c = c.NextSibling
			}
		}

		// WHY IS THIS HERE? commenting out for now - NCL 9/1/22
		//if n.LastChild != nil && n.LastChild.Type == TokenNode {
		//	w.Write([]byte(" "))
		//}
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
				w.Write([]byte("</h1>\n"))
			case "2":
				w.Write([]byte("</h2>\n"))
			case "3":
				w.Write([]byte("</h3>\n"))
			default:
				w.Write([]byte("</h1>\n"))
			}
		default:
			panic("not reached")
		}
	case TextNode:
		log.Printf("text nodes should not appear...")
		lines := strings.Split(n.Data, "\n")
		for i, l := range lines {
			lines[i] = opts.Prefix + l
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
		w.Write([]byte(opts.Prefix + "<!--" + n.Data + "-->\n"))
	case TexOnlyNode:
	case DivNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}
		w.Write([]byte("<div"))
		for _, a := range n.Attr {
			w.Write([]byte(fmt.Sprintf(" %s='%s'", a.Key, a.Val)))
		}
		w.Write([]byte(">"))
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			writeHTML(val, s, w, c, opts)
		}
		w.Write([]byte("</div>"))
	case PreNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}
		w.Write([]byte("<code"))
		for _, a := range n.Attr {
			w.Write([]byte(fmt.Sprintf(" %s='%s'", a.Key, a.Val)))
		}
		w.Write([]byte("><pre>"))
		var b bytes.Buffer
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			WriteLit(&b, c, opts)
		}
		w.Write([]byte(html.EscapeString(b.String())))
		w.Write([]byte("</pre></code>"))
	case CodeNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}
		w.Write([]byte("<code"))
		for _, a := range n.Attr {
			w.Write([]byte(fmt.Sprintf(" %s='%s'", a.Key, a.Val)))
		}
		w.Write([]byte("><pre>"))
		var b bytes.Buffer
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			WriteLit(&b, c, opts)
		}
		w.Write([]byte(html.EscapeString(b.String())))
		w.Write([]byte("</pre></code>"))
	case CenterAlignNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}
		// TODO should the br be here
		w.Write([]byte("<div style='display: flex; flex-direction: row; justify-content: center;text-align:center'><div>"))
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			writeHTML(val, s, w, c, opts)
		}
		w.Write([]byte("</div></div>"))
	case RightAlignNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}
		w.Write([]byte("<div style='display: flex; flex-direction: row; justify-content: right;'><div>"))
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			writeHTML(val, s, w, c, opts)
		}
		w.Write([]byte("</div></div>"))
	case QuoteNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}
		// TODO: maybe don't use blockquote, use custom
		// quote class div
		w.Write([]byte("<blockquote>"))
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			writeHTML(val, s, w, c, opts)
		}
		w.Write([]byte("</blockquote>"))
	case TableNode, TableHeadNode, TableBodyNode, TableRowNode, THNode, TDNode:
		var dataatom string
		switch n.Type {
		case TableNode:
			dataatom = "table"
		case TableHeadNode:
			dataatom = "thead"
		case TableBodyNode:
			dataatom = "tbody"
		case TableRowNode:
			dataatom = "tr"
		case THNode:
			dataatom = "th"
		case TDNode:
			dataatom = "td"
		}
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}
		w.Write([]byte("<" + dataatom))
		for _, a := range n.Attr {
			w.Write([]byte(fmt.Sprintf(" %s='%s'", a.Key, a.Val)))
		}
		w.Write([]byte(">"))
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			writeHTML(val, s, w, c, opts)
		}
		w.Write([]byte("</" + dataatom + ">"))
	case EquationNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}
		w.Write([]byte(opts.Prefix + "<div style='equation'>"))
		w.Write([]byte(opts.Prefix + "\\begin{equation}"))
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			writeHTML(Tex, s, w, c, Indented(opts)) // intentionally don't increase indent
		}
		if id := getAttr(n.Attr, "id"); id != "" {
			w.Write([]byte(opts.Prefix + opts.Indent + "\\label{" + id + "}"))
		}
		w.Write([]byte(opts.Prefix + "\\end{equation}"))
		w.Write([]byte(opts.Prefix + "</div>"))
	case ImageNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}
		if n.PrevSibling != nil && (n.PrevSibling.Type == ParagraphNode || n.PrevSibling.Type == ListNode || n.PrevSibling.Type == RunNode) {
			w.Write([]byte("\n"))
		}
		w.Write([]byte(opts.Prefix + fmt.Sprintf("<img src=\"%s\"", getAttr(n.Attr, "src"))))
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
		w.Write([]byte(opts.Prefix + fmt.Sprintf("<div class='%s'", t)))
		if text := getAttr(n.Attr, "text"); text != "" {
			w.Write([]byte(fmt.Sprintf(" text='%s'", text)))
		}
		if id := getAttr(n.Attr, "id"); id != "" {
			w.Write([]byte(fmt.Sprintf(" id='%s'", id)))
		}
		w.Write([]byte(">\n"))
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			writeHTML(val, s, w, c, Indented(opts))
		}
		w.Write([]byte(opts.Prefix + "</div>"))
	case ProofNode:
		if n.PrevSibling != nil {
			w.Write([]byte("\n"))
		}
		w.Write([]byte(opts.Prefix + "<div class='proof'>"))
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			writeHTML(val, s, w, c, Indented(opts))
		}
		w.Write([]byte(opts.Prefix + "</div>"))
	case LinkNode:
		if opts.InMath {
			log.Fatal("can't be in a link node in math")
		}
		w.Write([]byte(fmt.Sprintf("<a href='%s'", getAttr(n.Attr, "href"))))
		w.Write([]byte("/>"))
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			writeHTML(val, s, w, c, NoPrefix(opts))
		}
		w.Write([]byte("</a>"))
	default:
		return fmt.Errorf("unhandled node type %s, prev: %v; cur: %v; next: %v", n.Type, n.PrevSibling, n, n.NextSibling)
	}
	return nil
}
