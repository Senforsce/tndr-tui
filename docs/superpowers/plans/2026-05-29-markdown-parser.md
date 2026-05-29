# Markdown Parser Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: superpowers:executing-plans or subagent-driven-development. Steps use `- [ ]`.

**Goal:** A zero-dependency markdown parser in `internal/markdown` that turns a markdown string into a recursive `[]Block` tree, consumable later by the markdown component (Plan 4). No dependency on the `tui` package.

**Architecture:** Line-oriented block dispatcher (`Parse` → `parseBlocks`) with per-construct helpers, plus a rune-level inline scanner (`parseInline`). Blockquotes and lists nest recursively. Tables and code fences are multi-line state machines.

**Tech Stack:** Go 1.25, stdlib `strings` only. Table-driven tests.

**Plan sequence:** Plan **3 of 4** for issue #62. Independent of Plans 1–2 (no `tui` import). Plan 4 composes all three.

**Feature scope:** ATX + setext headings, bold (`**`/`__`), italic (`*`/`_`), inline code (`` ` ``), links (`[label](url)`), fenced code blocks, pipe tables, ordered/unordered lists with nesting, blockquotes (recursive). Anything else degrades to a paragraph of text.

**Conventions:** `gcommit -m "..."`, conventional commits. `go test ./internal/markdown/` per task, `go test -race ./...` at the end. Work on a branch.

---

## File Structure

- Create: `internal/markdown/markdown.go` — types (`Block`, `BlockKind`, `Inline`, `TableCell`), `Parse`, `parseBlocks` dispatcher + leaf helpers (heading/paragraph/fence/table).
- Create: `internal/markdown/inline.go` — `parseInline`, `parseLink`.
- Create: `internal/markdown/list.go` — list + blockquote parsing.
- Create: `internal/markdown/markdown_test.go`, `internal/markdown/inline_test.go`, `internal/markdown/list_test.go`.

---

## Task 1: Types, Parse skeleton, inline scanner

**Files:** `internal/markdown/markdown.go`, `internal/markdown/inline.go`, `internal/markdown/inline_test.go`

- [ ] **Step 1: Create the types and Parse skeleton** in `markdown.go`:

```go
// Package markdown parses a small, well-scoped subset of markdown into a block
// tree for terminal rendering. It has no dependency on the tui package.
package markdown

import "strings"

type BlockKind int

const (
	KindParagraph BlockKind = iota
	KindHeading
	KindCodeFence
	KindTable
	KindList
	KindListItem
	KindBlockquote
)

// Inline is a styled run of text. Code spans set Code; links set Link (with Text
// as the label). Bold/Italic may combine.
type Inline struct {
	Text   string
	Bold   bool
	Italic bool
	Code   bool
	Link   string // non-empty => hyperlink target
}

// TableCell holds one cell's inline content.
type TableCell struct {
	Inline []Inline
}

// Block is one node in the document tree.
type Block struct {
	Kind     BlockKind
	Level    int           // heading level (1-6); unused otherwise
	Ordered  bool          // ordered list
	Lang     string        // code-fence info string
	Inline   []Inline      // leaf inline content (heading, paragraph, list item)
	Lines    []string      // raw code-fence lines
	Rows     [][]TableCell // table rows; row 0 is the header
	Children []Block       // nested blocks (list items, blockquote/list contents)
}

// Parse parses markdown source into a block tree.
func Parse(src string) []Block {
	lines := strings.Split(strings.ReplaceAll(src, "\r\n", "\n"), "\n")
	return parseBlocks(lines)
}
```

- [ ] **Step 2: Implement the inline scanner** in `inline.go`:

```go
package markdown

// parseInline scans text into styled segments: **bold**/__bold__, *italic*/_italic_,
// `code`, and [label](url). Unmatched markers are literal. Code spans and link
// labels are not further parsed.
func parseInline(s string) []Inline {
	var out []Inline
	var buf []rune
	bold, italic := false, false

	flush := func() {
		if len(buf) > 0 {
			out = append(out, Inline{Text: string(buf), Bold: bold, Italic: italic})
			buf = buf[:0]
		}
	}

	rs := []rune(s)
	for i := 0; i < len(rs); {
		switch r := rs[i]; {
		case r == '`':
			j := i + 1
			for j < len(rs) && rs[j] != '`' {
				j++
			}
			if j < len(rs) {
				flush()
				out = append(out, Inline{Text: string(rs[i+1 : j]), Code: true})
				i = j + 1
				continue
			}
			buf = append(buf, r)
			i++
		case r == '[':
			if label, url, n, ok := parseLink(rs[i:]); ok {
				flush()
				out = append(out, Inline{Text: label, Link: url})
				i += n
				continue
			}
			buf = append(buf, r)
			i++
		case r == '*' || r == '_':
			flush()
			if i+1 < len(rs) && rs[i+1] == r {
				bold = !bold
				i += 2
			} else {
				italic = !italic
				i++
			}
		default:
			buf = append(buf, r)
			i++
		}
	}
	flush()
	return out
}

// parseLink parses [label](url) at rs[0]=='['. Returns label, url, runes consumed, ok.
func parseLink(rs []rune) (label, url string, n int, ok bool) {
	if len(rs) == 0 || rs[0] != '[' {
		return "", "", 0, false
	}
	closeIdx := -1
	for i := 1; i < len(rs); i++ {
		if rs[i] == ']' {
			closeIdx = i
			break
		}
	}
	if closeIdx < 0 || closeIdx+1 >= len(rs) || rs[closeIdx+1] != '(' {
		return "", "", 0, false
	}
	parenIdx := -1
	for i := closeIdx + 2; i < len(rs); i++ {
		if rs[i] == ')' {
			parenIdx = i
			break
		}
	}
	if parenIdx < 0 {
		return "", "", 0, false
	}
	return string(rs[1:closeIdx]), string(rs[closeIdx+2 : parenIdx]), parenIdx + 1, true
}
```

`parseBlocks` does not exist yet, so `markdown.go` won't compile until Task 2. Write `inline_test.go` covering: plain text; `**bold**`; `*italic*`; `` `code` ``; `[x](http://y)`; mixed `a **b** c`; unmatched `[oops`. Run `go test ./internal/markdown/ -run TestParseInline` after Task 2 builds.

- [ ] **Step 3: Commit** after Task 2 compiles (types + inline land together with the dispatcher).

---

## Task 2: Block dispatcher — headings, paragraphs, fenced code, tables

**Files:** `internal/markdown/markdown.go`, `internal/markdown/markdown_test.go`

- [ ] **Step 1: Add the dispatcher and leaf helpers** to `markdown.go`:

```go
func parseBlocks(lines []string) []Block {
	var blocks []Block
	for i := 0; i < len(lines); {
		line := lines[i]
		if strings.TrimSpace(line) == "" {
			i++
			continue
		}
		switch {
		case isFence(line):
			b, next := parseFence(lines, i)
			blocks, i = append(blocks, b), next
		case isATXHeading(line):
			blocks, i = append(blocks, parseATX(line)), i+1
		case isBlockquote(line):
			b, next := parseBlockquote(lines, i)
			blocks, i = append(blocks, b), next
		case isListLine(line):
			b, next := parseList(lines, i, listIndent(line))
			blocks, i = append(blocks, b), next
		case isTableStart(lines, i):
			b, next := parseTable(lines, i)
			blocks, i = append(blocks, b), next
		default:
			b, next := parseParagraphOrSetext(lines, i)
			blocks, i = append(blocks, b), next
		}
	}
	return blocks
}

func isFence(line string) bool {
	return strings.HasPrefix(strings.TrimSpace(line), "```")
}

func parseFence(lines []string, i int) (Block, int) {
	info := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(lines[i]), "```"))
	var body []string
	j := i + 1
	for j < len(lines) && !strings.HasPrefix(strings.TrimSpace(lines[j]), "```") {
		body = append(body, lines[j])
		j++
	}
	if j < len(lines) {
		j++ // consume closing fence
	}
	return Block{Kind: KindCodeFence, Lang: info, Lines: body}, j
}

func isATXHeading(line string) bool {
	t := strings.TrimLeft(line, " ")
	n := 0
	for n < len(t) && t[n] == '#' {
		n++
	}
	return n >= 1 && n <= 6 && n < len(t) && t[n] == ' '
}

func parseATX(line string) Block {
	t := strings.TrimLeft(line, " ")
	n := 0
	for n < len(t) && t[n] == '#' {
		n++
	}
	text := strings.TrimSpace(strings.TrimRight(strings.TrimSpace(t[n:]), "#"))
	return Block{Kind: KindHeading, Level: n, Inline: parseInline(text)}
}

func isSetextUnderline(s string, ch byte) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] != ch {
			return false
		}
	}
	return true
}

func parseParagraphOrSetext(lines []string, i int) (Block, int) {
	// Single-line setext: a text line followed by an all-'=' or all-'-' underline.
	if i+1 < len(lines) {
		if isSetextUnderline(lines[i+1], '=') {
			return Block{Kind: KindHeading, Level: 1, Inline: parseInline(strings.TrimSpace(lines[i]))}, i + 2
		}
		if isSetextUnderline(lines[i+1], '-') {
			return Block{Kind: KindHeading, Level: 2, Inline: parseInline(strings.TrimSpace(lines[i]))}, i + 2
		}
	}
	var parts []string
	j := i
	for j < len(lines) {
		l := lines[j]
		if strings.TrimSpace(l) == "" || isFence(l) || isATXHeading(l) || isBlockquote(l) || isListLine(l) {
			break
		}
		parts = append(parts, strings.TrimSpace(l))
		j++
	}
	return Block{Kind: KindParagraph, Inline: parseInline(strings.Join(parts, " "))}, j
}

func isTableStart(lines []string, i int) bool {
	return i+1 < len(lines) && strings.Contains(lines[i], "|") && isTableSeparator(lines[i+1])
}

func isTableSeparator(line string) bool {
	s := strings.TrimSpace(line)
	if s == "" || !strings.Contains(s, "-") {
		return false
	}
	for _, r := range s {
		if r != '-' && r != '|' && r != ':' && r != ' ' {
			return false
		}
	}
	return true
}

func parseTable(lines []string, i int) (Block, int) {
	rows := [][]TableCell{splitRow(lines[i])}
	j := i + 2 // skip header + separator
	for j < len(lines) && strings.TrimSpace(lines[j]) != "" && strings.Contains(lines[j], "|") {
		rows = append(rows, splitRow(lines[j]))
		j++
	}
	return Block{Kind: KindTable, Rows: rows}, j
}

func splitRow(line string) []TableCell {
	s := strings.TrimSpace(line)
	s = strings.TrimSuffix(strings.TrimPrefix(s, "|"), "|")
	parts := strings.Split(s, "|")
	cells := make([]TableCell, 0, len(parts))
	for _, p := range parts {
		cells = append(cells, TableCell{Inline: parseInline(strings.TrimSpace(p))})
	}
	return cells
}
```

(`isBlockquote`, `isListLine`, `listIndent`, `parseList`, `parseBlockquote` come from Task 3's `list.go`; the package compiles only once Task 3 lands, so write Task 2 + Task 3 code before running tests. Alternatively stub them in Task 2 and fill in Task 3 — but since this is a single executor, write both files then test.)

- [ ] **Step 2: Write `markdown_test.go`** covering ATX h1–h3, setext h1/h2, a paragraph joining two lines, a fenced code block (with lang + body lines), and a 2-column table (header + 2 rows). Assert `Kind`, `Level`, `Lang`, `Lines`, and `Rows` shapes.

- [ ] **Step 3: Run** `go test ./internal/markdown/` (after Task 3 lands so it builds) and **commit**.

---

## Task 3: Lists (nested) and blockquotes (recursive)

**Files:** `internal/markdown/list.go`, `internal/markdown/list_test.go`

- [ ] **Step 1: Implement** `list.go`:

```go
package markdown

import "strings"

func isBlockquote(line string) bool {
	return strings.HasPrefix(strings.TrimLeft(line, " "), ">")
}

func parseBlockquote(lines []string, i int) (Block, int) {
	var inner []string
	j := i
	for j < len(lines) && isBlockquote(lines[j]) {
		l := strings.TrimPrefix(strings.TrimLeft(lines[j], " "), ">")
		inner = append(inner, strings.TrimPrefix(l, " "))
		j++
	}
	return Block{Kind: KindBlockquote, Children: parseBlocks(inner)}, j
}

func listIndent(line string) int {
	n := 0
	for n < len(line) && line[n] == ' ' {
		n++
	}
	return n
}

func isListLine(line string) bool {
	t := strings.TrimLeft(line, " ")
	if strings.HasPrefix(t, "- ") || strings.HasPrefix(t, "* ") || strings.HasPrefix(t, "+ ") {
		return true
	}
	k := 0
	for k < len(t) && t[k] >= '0' && t[k] <= '9' {
		k++
	}
	return k > 0 && k+1 < len(t) && t[k] == '.' && t[k+1] == ' '
}

// listMarkerInfo reports whether the item is ordered and the byte offset where
// its content starts (after the marker and one space).
func listMarkerInfo(line string) (ordered bool, contentOffset int) {
	t := strings.TrimLeft(line, " ")
	lead := len(line) - len(t)
	if strings.HasPrefix(t, "- ") || strings.HasPrefix(t, "* ") || strings.HasPrefix(t, "+ ") {
		return false, lead + 2
	}
	k := 0
	for k < len(t) && t[k] >= '0' && t[k] <= '9' {
		k++
	}
	return true, lead + k + 2 // digits + ". "
}

// parseList consumes consecutive list lines at the given indent into a List of
// ListItems. Lines indented further attach as a nested List on the previous item.
func parseList(lines []string, i, indent int) (Block, int) {
	ordered, _ := listMarkerInfo(lines[i])
	list := Block{Kind: KindList, Ordered: ordered}
	j := i
	for j < len(lines) {
		if strings.TrimSpace(lines[j]) == "" || !isListLine(lines[j]) {
			break
		}
		ind := listIndent(lines[j])
		if ind < indent {
			break // belongs to an outer list
		}
		if ind > indent {
			child, next := parseList(lines, j, ind)
			if n := len(list.Children); n > 0 {
				list.Children[n-1].Children = append(list.Children[n-1].Children, child)
			} else {
				list.Children = append(list.Children, Block{Kind: KindListItem, Children: []Block{child}})
			}
			j = next
			continue
		}
		_, off := listMarkerInfo(lines[j])
		text := strings.TrimSpace(lines[j][off:])
		list.Children = append(list.Children, Block{Kind: KindListItem, Inline: parseInline(text)})
		j++
	}
	return list, j
}
```

- [ ] **Step 2: Write `list_test.go`** covering: an unordered list of 3 items (assert 3 `KindListItem` children, `Ordered==false`); an ordered list (`Ordered==true`); a nested list (a deeper-indented item attaches as a `KindList` child of the previous item); a blockquote containing a paragraph (assert `Children[0].Kind==KindParagraph`); a blockquote containing a nested list.

- [ ] **Step 3: Run** `go test ./internal/markdown/` and **commit** (Tasks 2 + 3 together — the package first compiles here).

---

## Task 4: Integration + race

**Files:** `internal/markdown/markdown_test.go`

- [ ] **Step 1: Add a full-document test** parsing a string that exercises every construct in order (heading, paragraph with bold/italic/code/link, fenced code, table, nested list, blockquote) and assert the top-level `Kind` sequence and a couple of deep fields (e.g. the link URL inside the paragraph, the nested list item text).

- [ ] **Step 2: Run** `go test ./internal/markdown/ -v` and `go test -race ./...`. Expected: all PASS.

- [ ] **Step 3: Commit.**

---

## Known v1 limitations (acceptable, documented)

- Multi-line paragraph immediately followed by a setext underline: only single-line setext is recognized; otherwise the `---`/`===` is swallowed into the paragraph. Rare; documented.
- Link labels and code spans are not parsed for nested inline formatting.
- A blank line ends a list (no loose-list/multi-paragraph-item support).
- Table column alignment markers (`:--`, `--:`) are parsed-and-ignored (all cells left).
- No thematic break (`---` standalone), images, footnotes, or HTML.

## Self-Review
- Every branch in `parseBlocks` advances `i` (each helper returns `next > i`, paragraphs always consume ≥1 line), so no infinite loop.
- All code is stdlib-only; no `tui` import — keeps the package independent.
- Names: `Parse`/`parseBlocks`/`parseInline`/`parseList`/`parseBlockquote`/`parseTable`/`parseFence`/`parseATX`; types `Block`/`BlockKind`/`Inline`/`TableCell`. Consistent across files.
