// Package lit provides parsing and rendering capabilities for the LitTex markup language.
package lit

// Unicode symbols used throughout the LitTex syntax.
// These constants define the special characters that give LitTex its semantic structure.

// Node type symbols - These represent different structural elements in LitTex
const (
	// ParagraphSymbol represents a paragraph block in LitTex
	ParagraphSymbol = '¶'
	// FootnoteSymbol represents a footnote in LitTex
	FootnoteSymbol = '†'
	// DisplayMathSymbol represents a display math block in LitTex
	DisplayMathSymbol = '◇'
	// RunSymbol represents a text run in LitTex
	RunSymbol = '‖'
	// ListSymbol represents a list in LitTex
	ListSymbol = '⁝'
	// ListItemSymbol represents a list item in LitTex
	ListItemSymbol = '‣'
	// SectionSymbol represents a section in LitTex
	SectionSymbol = '§'
)

// Block delimiter symbols - These define the boundaries of blocks in LitTex
const (
	// OpenBlockDelimiter marks the beginning of a block
	OpenBlockDelimiter = '⦊'
	// CloseBlockDelimiter marks the end of a block
	CloseBlockDelimiter = '⦉'
)

// Text formatting symbols - These define text styling in LitTex
const (
	// EmphasisOpenDelimiter marks the beginning of emphasized text
	EmphasisOpenDelimiter = '‹'
	// EmphasisCloseDelimiter marks the end of emphasized text
	EmphasisCloseDelimiter = '›'
	
	// BoldOpenDelimiter marks the beginning of bold text
	BoldOpenDelimiter = '«'
	// BoldCloseDelimiter marks the end of bold text
	BoldCloseDelimiter = '»'
	
	// TeXOpenDelimiter1 marks the beginning of TeX-specific content (first style)
	TeXOpenDelimiter1 = '❬'
	// TeXCloseDelimiter1 marks the end of TeX-specific content (first style)
	TeXCloseDelimiter1 = '❭'
	
	// CustomOpenDelimiter1 marks the beginning of custom-styled content (first style)
	CustomOpenDelimiter1 = '⁅'
	// CustomCloseDelimiter1 marks the end of custom-styled content (first style)
	CustomCloseDelimiter1 = '⁆'
	
	// BoldOpenDelimiter2 marks the beginning of bold text (alternate style)
	BoldOpenDelimiter2 = '❮'
	// BoldCloseDelimiter2 marks the end of bold text (alternate style)
	BoldCloseDelimiter2 = '❯'
	
	// TeXOpenDelimiter2 marks the beginning of TeX-specific content (second style)
	TeXOpenDelimiter2 = '⧼'
	// TeXCloseDelimiter2 marks the end of TeX-specific content (second style)
	TeXCloseDelimiter2 = '⧽'
	
	// SmallCapsOpenDelimiter marks the beginning of small caps text
	SmallCapsOpenDelimiter = '⸤'
	// SmallCapsCloseDelimiter marks the end of small caps text
	SmallCapsCloseDelimiter = '⸥'
)

// Math delimiter symbols - These define mathematical expressions in LitTex
const (
	// MathOpenDelimiter marks the beginning of inline math
	MathOpenDelimiter = '❲'
	// MathCloseDelimiter marks the end of inline math
	MathCloseDelimiter = '❳'
)

// Special symbols - These are used for specific semantic purposes in LitTex
const (
	// LineBreakSymbol represents a line break in LitTex
	LineBreakSymbol = '᜶'
	// MapsToSymbol represents a mathematical "maps to" relation
	MapsToSymbol = '↦'
	// AmpersandSymbol represents a standard ampersand
	AmpersandSymbol = '&'
	// SafeAmpersandSymbol represents an ampersand that won't be escaped in TeX
	SafeAmpersandSymbol = '＆'
)

// String versions of symbols for use in string operations
const (
	// ParagraphStr is the string version of ParagraphSymbol
	ParagraphStr = "¶"
	// OpenBlockDelimiterStr is the string version of OpenBlockDelimiter
	OpenBlockDelimiterStr = "⦊"
	// CloseBlockDelimiterStr is the string version of CloseBlockDelimiter
	CloseBlockDelimiterStr = "⦉"
	// RunStr is the string version of RunSymbol
	RunStr = "‖"
	// DisplayMathStr is the string version of DisplayMathSymbol
	DisplayMathStr = "◇"
	// FootnoteStr is the string version of FootnoteSymbol
	FootnoteStr = "†"
)
