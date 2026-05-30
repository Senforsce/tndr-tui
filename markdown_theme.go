package tui

// MarkdownTheme controls how a Markdown component styles each construct. It is a
// flat struct of Style fields plus a few non-style extras. Construct a sensible
// default with DefaultMarkdownTheme and override fields as needed.
type MarkdownTheme struct {
	// Heading holds per-level heading styles, indexed by (level-1) for levels 1..6.
	Heading   [6]Style
	Paragraph Style
	// Bold, Italic, CodeSpan, and Link are layered over the surrounding text via
	// the inline scanner's flags. Their Attrs OR in and non-default colors replace.
	Bold     Style
	Italic   Style
	CodeSpan Style // inline `code`
	Link     Style

	CodeBlockText   Style
	CodeBlockBg     Color       // default => no fill
	CodeBlockBorder BorderStyle // a real full-box border around the code element

	// Tables reuse the existing table layout (a 1-char column gap, not grid lines).
	// TableBorder draws an outline around the whole table (BorderNone = no outline).
	// v1 styles the header and, optionally, a separator row under it.
	TableHeader        Style
	TableBorder        BorderStyle
	TableSeparator     bool
	TableSeparatorChar rune

	// Blockquotes render a 1-wide glyph column (borders draw full boxes, so a
	// BorderStyle cannot be used for a left bar).
	BlockquoteBar      rune
	BlockquoteBarStyle Style
	BlockquoteText     Style

	BulletMarker string // unordered-list marker, e.g. "• "
}

// DefaultMarkdownTheme returns a glow-inspired theme that reads well on dark and
// light terminals using only attributes and a couple of muted colors.
func DefaultMarkdownTheme() MarkdownTheme {
	return MarkdownTheme{
		Heading: [6]Style{
			NewStyle().Bold().Underline(), // h1: bold + underline
			NewStyle().Bold(),             // h2: bold
			NewStyle().Italic(),           // h3: italic
			NewStyle().Bold(),             // h4
			NewStyle().Bold(),             // h5
			NewStyle().Bold(),             // h6
		},
		Paragraph: NewStyle(),
		Bold:      NewStyle().Bold(),
		Italic:    NewStyle().Italic(),
		CodeSpan:  NewStyle().Foreground(BrightMagenta),
		Link:      NewStyle().Underline().Foreground(BrightBlue),

		CodeBlockText:   NewStyle().Foreground(BrightWhite),
		CodeBlockBg:     DefaultColor(),
		CodeBlockBorder: BorderRounded,

		TableHeader:        NewStyle().Bold(),
		TableBorder:        BorderRounded,
		TableSeparator:     false,
		TableSeparatorChar: '-',

		BlockquoteBar:      '│',
		BlockquoteBarStyle: NewStyle().Foreground(BrightBlack),
		BlockquoteText:     NewStyle().Italic().Foreground(BrightBlack),

		BulletMarker: "• ",
	}
}
