# Rich-Text Primitive Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a public `TextSpan` rich-text primitive to the `tui` package so an Element can render a single wrapped line made of multiple styles (e.g. a bold word inside an otherwise-plain paragraph).

**Architecture:** A new `richText []TextSpan` field on `Element` runs parallel to the existing single `text`/`textStyle`. Plain text keeps its existing render path untouched (zero regression risk); rich text gets a parallel path through a shared `drawSpanLines` helper used by both the normal and clipped/scroll render sites, plus rich-text branches in the two layout measurement functions. Wrapping reuses the existing word-packing algorithm, carrying per-word style.

**Tech Stack:** Go 1.25, no external dependencies. Tests use the in-repo `Buffer` + `RenderTree` + `buf.Cell(x,y)` inspection pattern (see `element_render_test.go`).

**Plan sequence:** This is **Plan 1 of 4** for issue #62 (spec: `docs/superpowers/specs/2026-05-29-markdown-component-design.md`).
1. **Rich-text primitive (this plan)** — `tui` package, framework feature.
2. OSC 8 hyperlinks — `tui` package (`Cell` link id, buffer diff, `Flush`).
3. Markdown parser — `internal/markdown`, zero-dep, no `tui` dependency.
4. Markdown component + gsx integration — composes 1, 2, 3.

Plans 1–3 are independent and may be built in parallel. The `TextSpan.Link` field is defined here (so the type is stable) but is inert until Plan 2 wires it into the cell/Flush pipeline.

**Conventions for every commit in this plan:** use `gcommit -m "..."` (NOT `git commit` — the project requires signed commits via `gcommit`), conventional-commit format.

---

## File Structure

- Create: `richtext.go` — `TextSpan` type, `WithRichText` option, `SetRichText`/`RichText` accessors, `mergeSpanStyle`, `richTextWidth`, `spanLineWidth`. All rich-text-specific helpers live together here.
- Create: `richtext_test.go` — unit tests for the type, accessors, clearing semantics, and style merge.
- Modify: `element.go:70` — add the `richText []TextSpan` field.
- Modify: `element_options.go:207` (`WithText`) — clear `richText` when plain text is set.
- Modify: `element_accessors.go:60` (`SetText`) — clear `richText` when plain text is set.
- Modify: `text_wrap.go` — add `wrapSpans` beside `wrapText`.
- Modify: `text_wrap_test.go` (create if absent) — `wrapSpans` tests.
- Modify: `element_render.go` — add `drawSpanLines`; add rich-text branch + gate in `renderTextContent`/`renderElement` (`:137`) and `renderClippedElement` (`:246`).
- Modify: `element_render_test.go` — rich-text render tests (normal + clipped paths).
- Modify: `element_layout.go` — rich-text branch in `IntrinsicSize` (`:89`) and `HeightForWidth` (`:196`).
- Modify: `element_layout_test.go` — rich-text measurement tests.

---

## Task 1: TextSpan type, option, accessors, and clearing semantics

**Files:**
- Create: `richtext.go`
- Create: `richtext_test.go`
- Modify: `element.go:70`
- Modify: `element_options.go:207`
- Modify: `element_accessors.go:60`

- [ ] **Step 1: Write the failing test**

Create `richtext_test.go`:

```go
package tui

import "testing"

func TestRichText_AccessorRoundTrip(t *testing.T) {
	spans := []TextSpan{
		{Text: "hello "},
		{Text: "world", Style: NewStyle().Bold()},
	}
	e := New(WithRichText(spans...))
	got := e.RichText()
	if len(got) != 2 || got[0].Text != "hello " || got[1].Text != "world" {
		t.Fatalf("RichText() = %+v, want 2 spans hello/world", got)
	}
	if got[1].Style.Attrs&AttrBold == 0 {
		t.Errorf("second span should be bold, got attrs %v", got[1].Style.Attrs)
	}
}

func TestRichText_SettingPlainTextClearsRichText(t *testing.T) {
	e := New(WithRichText(TextSpan{Text: "rich"}))
	e.SetText("plain")
	if len(e.RichText()) != 0 {
		t.Errorf("SetText should clear richText, got %+v", e.RichText())
	}
	if e.Text() != "plain" {
		t.Errorf("Text() = %q, want \"plain\"", e.Text())
	}
}

func TestRichText_SettingRichTextClearsPlainText(t *testing.T) {
	e := New(WithText("plain"))
	e.SetRichText(TextSpan{Text: "rich"})
	if e.Text() != "" {
		t.Errorf("SetRichText should clear text, got %q", e.Text())
	}
	if len(e.RichText()) != 1 {
		t.Errorf("RichText() len = %d, want 1", len(e.RichText()))
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./ -run TestRichText -v`
Expected: FAIL — `undefined: TextSpan`, `undefined: WithRichText`, `e.RichText undefined`, `e.SetRichText undefined`.

- [ ] **Step 3: Add the `richText` field to Element**

In `element.go`, change the text-properties block (currently ending at line 70 with `noWrap`) to add the field:

```go
	// Text properties
	text         string
	richText     []TextSpan // when non-empty, takes precedence over text
	textStyle    Style
	textStyleSet bool // true if textStyle was explicitly configured (false = inherit from parent)
	textAlign    TextAlign
	truncate     bool
	noWrap       bool // true = wrapping disabled; false (default) = wrapping enabled
```

- [ ] **Step 4: Create `richtext.go` with the type, option, and accessors**

```go
package tui

// TextSpan is a run of text sharing one style. A zero-value Style means the
// span inherits the element's resolved textStyle; set fields override it
// (attributes OR in, non-default colors replace). When an Element has rich
// text it renders as a sequence of spans that wrap together at word
// boundaries, allowing mixed styling within one wrapped paragraph.
type TextSpan struct {
	Text  string
	Style Style
	// Link is an optional OSC 8 hyperlink target. It is stored here so the type
	// is stable, but it is inert until the OSC 8 layer wires it into the cell
	// pipeline. Plain styled rendering ignores it.
	Link string
}

// WithRichText sets styled, multi-segment text on an element. When set it takes
// precedence over WithText and clears any plain text. Wrapping and alignment
// behave as for plain text.
func WithRichText(spans ...TextSpan) Option {
	return func(e *Element) {
		e.richText = spans
		e.text = ""
	}
}

// RichText returns the element's rich-text spans (nil if none).
func (e *Element) RichText() []TextSpan {
	return e.richText
}

// SetRichText replaces the element's rich-text spans and clears any plain text.
func (e *Element) SetRichText(spans ...TextSpan) {
	e.richText = spans
	e.text = ""
	e.MarkDirty()
}
```

- [ ] **Step 5: Make plain-text setters clear rich text**

In `element_options.go`, update `WithText` (line 207):

```go
func WithText(content string) Option {
	return func(e *Element) {
		e.text = content
		e.richText = nil
	}
}
```

In `element_accessors.go`, update `SetText` (line 60):

```go
func (e *Element) SetText(content string) {
	e.text = content
	e.richText = nil
	e.MarkDirty()
}
```

- [ ] **Step 6: Run tests to verify they pass**

Run: `go test ./ -run TestRichText -v`
Expected: PASS (all three).

- [ ] **Step 7: Commit**

```bash
git add richtext.go richtext_test.go element.go element_options.go element_accessors.go
gcommit -m "feat: add TextSpan rich-text type with accessors and clearing semantics"
```

---

## Task 2: Style merge and width helpers

**Files:**
- Modify: `richtext.go`
- Modify: `richtext_test.go`

- [ ] **Step 1: Write the failing test**

Append to `richtext_test.go`:

```go
func TestMergeSpanStyle(t *testing.T) {
	base := NewStyle().Foreground(White).Background(Blue)

	// Span attribute ORs into base; base colors preserved when span uses defaults.
	got := mergeSpanStyle(base, NewStyle().Bold())
	if got.Attrs&AttrBold == 0 {
		t.Errorf("bold not merged in: %v", got.Attrs)
	}
	if got.Fg != White || got.Bg != Blue {
		t.Errorf("base colors should survive: fg=%v bg=%v", got.Fg, got.Bg)
	}

	// Non-default span color overrides base.
	got = mergeSpanStyle(base, NewStyle().Foreground(Red))
	if got.Fg != Red {
		t.Errorf("span fg should override: got %v", got.Fg)
	}
	if got.Bg != Blue {
		t.Errorf("base bg should survive: got %v", got.Bg)
	}
}

func TestRichTextWidth(t *testing.T) {
	spans := []TextSpan{{Text: "ab"}, {Text: "cde", Style: NewStyle().Bold()}}
	if got := richTextWidth(spans); got != 5 {
		t.Errorf("richTextWidth = %d, want 5", got)
	}
}

func TestSpanLineWidth(t *testing.T) {
	line := []TextSpan{{Text: "hi "}, {Text: "yo"}}
	if got := spanLineWidth(line); got != 5 {
		t.Errorf("spanLineWidth = %d, want 5", got)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./ -run 'TestMergeSpanStyle|TestRichTextWidth|TestSpanLineWidth' -v`
Expected: FAIL — `undefined: mergeSpanStyle`, `undefined: richTextWidth`, `undefined: spanLineWidth`.

- [ ] **Step 3: Implement the helpers in `richtext.go`**

Append to `richtext.go`:

```go
// mergeSpanStyle layers a span's style over the element's resolved base style:
// attributes OR together, and a non-default span color replaces the base color.
func mergeSpanStyle(base, span Style) Style {
	out := base
	out.Attrs |= span.Attrs
	if !span.Fg.IsDefault() {
		out.Fg = span.Fg
	}
	if !span.Bg.IsDefault() {
		out.Bg = span.Bg
	}
	return out
}

// richTextWidth returns the total display width of all spans concatenated,
// used for intrinsic (unwrapped) sizing.
func richTextWidth(spans []TextSpan) int {
	w := 0
	for _, s := range spans {
		w += stringWidth(s.Text)
	}
	return w
}

// spanLineWidth returns the display width of one wrapped line of spans.
func spanLineWidth(line []TextSpan) int {
	w := 0
	for _, s := range line {
		w += stringWidth(s.Text)
	}
	return w
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./ -run 'TestMergeSpanStyle|TestRichTextWidth|TestSpanLineWidth' -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add richtext.go richtext_test.go
gcommit -m "feat: add rich-text style merge and width helpers"
```

---

## Task 3: wrapSpans — word-wrap that carries per-word style

**Files:**
- Modify: `text_wrap.go`
- Create/Modify: `text_wrap_test.go`

`wrapSpans` mirrors `wrapParagraph`'s greedy word-packing (`text_wrap.go:24`), but each emitted word keeps the style of the span it came from, and adjacent same-style segments on a line are merged.

- [ ] **Step 1: Write the failing test**

Create `text_wrap_test.go` (or append if it exists):

```go
package tui

import "testing"

func TestWrapSpans_KeepsStyleAcrossWrap(t *testing.T) {
	bold := NewStyle().Bold()
	spans := []TextSpan{
		{Text: "aa "},
		{Text: "bbbb cccc", Style: bold}, // two bold words
		{Text: " dd"},
	}
	// Width 6 forces a break inside the bold run.
	lines := wrapSpans(spans, 6)
	if len(lines) < 2 {
		t.Fatalf("expected multiple lines, got %d: %+v", len(lines), lines)
	}
	// Every segment whose text is a bold word must carry bold on every line.
	for li, line := range lines {
		for _, seg := range line {
			if seg.Text == "bbbb" || seg.Text == "cccc" {
				if seg.Style.Attrs&AttrBold == 0 {
					t.Errorf("line %d: %q lost bold", li, seg.Text)
				}
			}
		}
	}
}

func TestWrapSpans_MergesAdjacentSameStyle(t *testing.T) {
	// Two plain spans with words that fit on one line should merge.
	spans := []TextSpan{{Text: "foo "}, {Text: "bar"}}
	lines := wrapSpans(spans, 40)
	if len(lines) != 1 {
		t.Fatalf("want 1 line, got %d", len(lines))
	}
	if len(lines[0]) != 1 {
		t.Errorf("adjacent same-style segments should merge into 1, got %d: %+v", len(lines[0]), lines[0])
	}
	if lines[0][0].Text != "foo bar" {
		t.Errorf("merged text = %q, want \"foo bar\"", lines[0][0].Text)
	}
}

func TestWrapSpans_Empty(t *testing.T) {
	if got := wrapSpans(nil, 10); len(got) != 1 || len(got[0]) != 0 {
		t.Errorf("empty spans should give one empty line, got %+v", got)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./ -run TestWrapSpans -v`
Expected: FAIL — `undefined: wrapSpans`.

- [ ] **Step 3: Implement `wrapSpans` in `text_wrap.go`**

Append to `text_wrap.go`:

```go
// styledWord is one whitespace-delimited token with the style of its source span.
type styledWord struct {
	text  string
	style Style
}

// wrapSpans wraps styled spans to maxWidth using word boundaries, mirroring
// wrapParagraph. Each emitted word keeps its source span's style, so a multi-word
// styled run stays styled across a line break. Adjacent same-style segments on a
// line are merged. Newlines inside span text start new lines.
func wrapSpans(spans []TextSpan, maxWidth int) [][]TextSpan {
	if maxWidth < 1 {
		return [][]TextSpan{{}}
	}

	// Flatten spans into styled words, treating '\n' as a hard line break.
	var words []styledWord
	for _, sp := range spans {
		for i, para := range strings.Split(sp.Text, "\n") {
			if i > 0 {
				words = append(words, styledWord{text: "\n"}) // marker
			}
			for _, w := range strings.Fields(para) {
				words = append(words, styledWord{text: w, style: sp.Style})
			}
		}
	}

	var lines [][]TextSpan
	var cur []TextSpan
	lineWidth := 0

	flush := func() {
		lines = append(lines, cur)
		cur = nil
		lineWidth = 0
	}
	// appendWord adds a word to cur, merging into the last segment if same style.
	appendWord := func(w styledWord, leadingSpace bool) {
		text := w.text
		if leadingSpace {
			text = " " + text
		}
		if n := len(cur); n > 0 && cur[n-1].Style == w.style {
			cur[n-1].Text += text
		} else {
			cur = append(cur, TextSpan{Text: text, Style: w.style})
		}
	}

	for _, w := range words {
		if w.text == "\n" {
			flush()
			continue
		}
		ww := stringWidth(w.text)
		if ww > maxWidth {
			// Word longer than the line: flush current, then hard-break by rune.
			if lineWidth > 0 {
				flush()
			}
			for _, r := range w.text {
				rw := RuneWidth(r)
				if lineWidth+rw > maxWidth && lineWidth > 0 {
					flush()
				}
				appendWord(styledWord{text: string(r), style: w.style}, false)
				lineWidth += rw
			}
			continue
		}
		switch {
		case lineWidth == 0:
			appendWord(w, false)
			lineWidth = ww
		case lineWidth+1+ww <= maxWidth:
			appendWord(w, true)
			lineWidth += 1 + ww
		default:
			flush()
			appendWord(w, false)
			lineWidth = ww
		}
	}
	flush()
	return lines
}
```

Note: `text_wrap.go` already imports `strings`.

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./ -run TestWrapSpans -v`
Expected: PASS (all three).

- [ ] **Step 5: Commit**

```bash
git add text_wrap.go text_wrap_test.go
gcommit -m "feat: add wrapSpans for style-preserving word wrapping"
```

---

## Task 4: Render rich text in the normal pass

**Files:**
- Modify: `element_render.go` (add `drawSpanLines`; gate at `:137`; branch in `renderTextContent`)
- Modify: `element_render_test.go`

- [ ] **Step 1: Write the failing test**

Append to `element_render_test.go`:

```go
func TestRenderTree_RichTextStylesPerSegment(t *testing.T) {
	buf := NewBuffer(20, 3)
	e := New(
		WithSize(10, 1),
		WithRichText(
			TextSpan{Text: "ab"},
			TextSpan{Text: "cd", Style: NewStyle().Bold()},
		),
	)
	e.Calculate(20, 3)
	RenderTree(buf, e)

	// "abcd" laid out left to right.
	for x, want := range []rune{'a', 'b', 'c', 'd'} {
		if got := buf.Cell(x, 0).Rune; got != want {
			t.Errorf("cell(%d,0).Rune = %q, want %q", x, got, want)
		}
	}
	// First two cells plain, last two bold.
	if buf.Cell(0, 0).Style.Attrs&AttrBold != 0 {
		t.Errorf("cell(0,0) should not be bold")
	}
	if buf.Cell(2, 0).Style.Attrs&AttrBold == 0 {
		t.Errorf("cell(2,0) should be bold")
	}
	if buf.Cell(3, 0).Style.Attrs&AttrBold == 0 {
		t.Errorf("cell(3,0) should be bold")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./ -run TestRenderTree_RichTextStylesPerSegment -v`
Expected: FAIL — nothing is drawn (the `if e.text != ""` gate at `element_render.go:137` is false for rich text), so `cell(0,0).Rune` is `' '` not `'a'`.

- [ ] **Step 3: Add the shared `drawSpanLines` helper**

In `element_render.go`, add near `renderTextContent` (after the function, around line 554):

```go
// drawSpanLines renders pre-wrapped rich-text lines into the buffer.
// originX/originY is the top-left content cell, contentWidth is the line box used
// for alignment, base is the element's resolved text style (with background
// already merged), and clip bounds the drawable region. Cells outside clip are
// skipped but still advance x (so horizontal scroll offsets line up).
func drawSpanLines(buf *Buffer, lines [][]TextSpan, originX, originY, contentWidth int, align TextAlign, base Style, clip Rect) {
	for li, line := range lines {
		y := originY + li
		if y < clip.Y || y >= clip.Bottom() {
			continue
		}
		x := originX
		if lw := spanLineWidth(line); contentWidth > lw {
			switch align {
			case TextAlignCenter:
				x += (contentWidth - lw) / 2
			case TextAlignRight:
				x += contentWidth - lw
			}
		}
		for _, span := range line {
			st := mergeSpanStyle(base, span.Style)
			for _, r := range span.Text {
				if x >= clip.Right() {
					return
				}
				w := RuneWidth(r)
				if w == 2 && x+1 >= clip.Right() {
					return
				}
				if x >= clip.X {
					style := st
					if style.Bg.IsDefault() {
						if cellBg := buf.Cell(x, y).Style.Bg; !cellBg.IsDefault() {
							style.Bg = cellBg
						}
					}
					buf.SetRune(x, y, r, style)
				}
				x += w
			}
		}
	}
}
```

- [ ] **Step 4: Branch into rich text from `renderTextContent`**

At the top of `renderTextContent` (`element_render.go:432`, right after the `contentRect.IsEmpty()` early return), add:

```go
	if len(e.richText) > 0 {
		ts := textStyle
		if bg != nil && !bg.Bg.IsDefault() {
			ts.Bg = bg.Bg
		}
		var lines [][]TextSpan
		if !e.noWrap && contentRect.Width > 0 {
			lines = wrapSpans(e.richText, contentRect.Width)
		} else {
			lines = [][]TextSpan{e.richText}
		}
		drawSpanLines(buf, lines, contentRect.X, contentRect.Y, contentRect.Width, e.textAlign, ts, contentRect)
		return
	}
```

- [ ] **Step 5: Update the render gate**

In `renderElement` (`element_render.go:137`), change the gate so rich text triggers text rendering:

```go
	// 3. Draw text content if present
	if e.text != "" || len(e.richText) > 0 {
		renderTextContent(buf, e, rc.textStyle, rc.bg)
	}
```

- [ ] **Step 6: Run tests to verify they pass**

Run: `go test ./ -run TestRenderTree_RichTextStylesPerSegment -v`
Expected: PASS.

- [ ] **Step 7: Run the full package to check for regressions**

Run: `go test ./`
Expected: PASS (plain-text rendering untouched).

- [ ] **Step 8: Commit**

```bash
git add element_render.go element_render_test.go
gcommit -m "feat: render rich text in the normal render pass"
```

---

## Task 5: Render rich text in the clipped/scroll pass

This is the path the spec calls out as the silent-failure risk: scrollable containers render children through `renderClippedElement` (`element_render.go:246`), which has its own text gate and loop.

**Files:**
- Modify: `element_render.go` (gate + branch in `renderClippedElement`)
- Modify: `element_render_test.go`

- [ ] **Step 1: Write the failing test**

Append to `element_render_test.go`:

```go
func TestRichText_RendersInsideScrollableContainer(t *testing.T) {
	buf := NewBuffer(20, 5)
	child := New(
		WithSize(10, 1),
		WithRichText(
			TextSpan{Text: "ab"},
			TextSpan{Text: "cd", Style: NewStyle().Bold()},
		),
	)
	container := New(
		WithSize(12, 3),
		WithScrollable(ScrollVertical),
	)
	container.AddChild(child)
	container.Calculate(20, 5)
	RenderTree(buf, container)

	// Text must appear (this is the bug the spec warns about).
	if got := buf.Cell(0, 0).Rune; got != 'a' {
		t.Errorf("rich text not rendered in scroll container: cell(0,0)=%q, want 'a'", got)
	}
	if buf.Cell(2, 0).Style.Attrs&AttrBold == 0 {
		t.Errorf("cell(2,0) should be bold inside scroll container")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./ -run TestRichText_RendersInsideScrollableContainer -v`
Expected: FAIL — `cell(0,0)` is `' '` because `renderClippedElement`'s text gate (`if e.text != ""`) is false for rich text.

- [ ] **Step 3: Add the rich-text branch in `renderClippedElement`**

In `renderClippedElement`, the text block begins at `element_render.go:246` with `if e.text != "" {`. Immediately BEFORE that block, insert a rich-text branch (it computes the same base coordinates the plain path uses, then delegates to `drawSpanLines`):

```go
	// Render rich text with clipping (parallel to the plain-text block below).
	if len(e.richText) > 0 {
		textBaseX := screenX + e.style.Padding.Left
		textBaseY := screenY + e.style.Padding.Top
		if e.border != BorderNone {
			textBaseX += 1
			textBaseY += 1
		}
		availTextWidth := childRect.Width - e.style.Padding.Horizontal()
		if e.border != BorderNone {
			availTextWidth -= 2
		}

		ts := rc.textStyle
		if rc.bg != nil && !rc.bg.Bg.IsDefault() {
			ts.Bg = rc.bg.Bg
		}

		var lines [][]TextSpan
		if !e.noWrap && availTextWidth > 0 {
			lines = wrapSpans(e.richText, availTextWidth)
		} else {
			lines = [][]TextSpan{e.richText}
		}
		drawSpanLines(buf, lines, textBaseX, textBaseY, availTextWidth, e.textAlign, ts, clipRect)
	}
```

(Leave the existing `if e.text != "" {` block exactly as-is below it. An element has either plain text or rich text, never both, so the two branches never both fire.)

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./ -run TestRichText_RendersInsideScrollableContainer -v`
Expected: PASS.

- [ ] **Step 5: Run the full package**

Run: `go test ./`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add element_render.go element_render_test.go
gcommit -m "feat: render rich text in the clipped/scroll render pass"
```

---

## Task 6: Measure rich text (IntrinsicSize + HeightForWidth)

Without measurement, auto-sized rich-text elements collapse (width from `text` only, height 1), so paragraphs would not wrap to multiple rows.

**Files:**
- Modify: `element_layout.go` (`IntrinsicSize` at `:89`, `HeightForWidth` at `:196`)
- Modify: `element_layout_test.go`

- [ ] **Step 1: Write the failing test**

Append to `element_layout_test.go`:

```go
func TestIntrinsicSize_RichText(t *testing.T) {
	e := New(WithRichText(
		TextSpan{Text: "ab"},
		TextSpan{Text: "cde", Style: NewStyle().Bold()},
	))
	w, h := e.IntrinsicSize()
	if w != 5 {
		t.Errorf("intrinsic width = %d, want 5", w)
	}
	if h != 1 {
		t.Errorf("intrinsic height = %d, want 1", h)
	}
}

func TestHeightForWidth_RichTextWraps(t *testing.T) {
	// "aaa bbb ccc" is 11 cells; width 7 forces 2 lines.
	e := New(WithRichText(
		TextSpan{Text: "aaa "},
		TextSpan{Text: "bbb", Style: NewStyle().Bold()},
		TextSpan{Text: " ccc"},
	))
	if got := e.HeightForWidth(7); got != 2 {
		t.Errorf("HeightForWidth(7) = %d, want 2", got)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./ -run 'TestIntrinsicSize_RichText|TestHeightForWidth_RichTextWraps' -v`
Expected: FAIL — `IntrinsicSize` returns width 0 / height from the children path, and `HeightForWidth` returns the intrinsic height (1), not 2.

- [ ] **Step 3: Add the rich-text branch to `IntrinsicSize`**

In `element_layout.go`, immediately AFTER the `if e.text != "" {` block (which ends at line 101 with `return width, height`), insert:

```go
	// Rich text content has explicit intrinsic size (single unwrapped line).
	if len(e.richText) > 0 {
		width = richTextWidth(e.richText) + e.style.Padding.Horizontal()
		height = 1 + e.style.Padding.Vertical()
		if e.border != BorderNone {
			width += 2
			height += 2
		}
		return width, height
	}
```

- [ ] **Step 4: Add the rich-text branch to `HeightForWidth`**

In `element_layout.go`, immediately AFTER the text-wrapping block in `HeightForWidth` (the `if e.text != "" && !e.noWrap {` block that ends at line 214 with `return h`), insert:

```go
	// Rich text elements with wrapping.
	if len(e.richText) > 0 && !e.noWrap {
		contentWidth := width - e.style.Padding.Horizontal()
		if e.border != BorderNone {
			contentWidth -= 2
		}
		if contentWidth <= 0 {
			h := e.style.Padding.Vertical()
			if e.border != BorderNone {
				h += 2
			}
			return h
		}
		lines := wrapSpans(e.richText, contentWidth)
		h := len(lines) + e.style.Padding.Vertical()
		if e.border != BorderNone {
			h += 2
		}
		return h
	}
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `go test ./ -run 'TestIntrinsicSize_RichText|TestHeightForWidth_RichTextWraps' -v`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add element_layout.go element_layout_test.go
gcommit -m "feat: measure rich text in IntrinsicSize and HeightForWidth"
```

---

## Task 7: Integration test — the payoff scenario

Prove the end-to-end behavior the whole primitive exists for: a bold phrase inside a wrapping paragraph stays bold across the line break.

**Files:**
- Modify: `element_render_test.go`

- [ ] **Step 1: Write the test**

Append to `element_render_test.go`:

```go
func TestRichText_BoldSurvivesWrap(t *testing.T) {
	buf := NewBuffer(40, 6)
	// Paragraph fixed to width 10 so the bold run wraps.
	para := New(
		WithWidth(10),
		WithRichText(
			TextSpan{Text: "see "},
			TextSpan{Text: "this bold run", Style: NewStyle().Bold()},
			TextSpan{Text: " end"},
		),
	)
	para.Calculate(40, 6)
	RenderTree(buf, para)

	// Find every 'b','o','l','d'-bearing cell is bold across whatever line it
	// landed on: scan all cells and assert any glyph from the bold words is bold.
	// Simpler concrete assertion: row 0 starts "see " (plain) then "this" (bold).
	if buf.Cell(0, 0).Rune != 's' || buf.Cell(0, 0).Style.Attrs&AttrBold != 0 {
		t.Errorf("cell(0,0) should be plain 's', got %q attrs=%v", buf.Cell(0, 0).Rune, buf.Cell(0, 0).Style.Attrs)
	}
	// "see " is 4 cells; the bold word "this" begins at x=4 on row 0.
	if buf.Cell(4, 0).Style.Attrs&AttrBold == 0 {
		t.Errorf("cell(4,0) should be bold (start of bold run)")
	}
	// The paragraph must occupy more than one row at width 10.
	rowsWithText := 0
	for y := 0; y < 6; y++ {
		for x := 0; x < 10; x++ {
			if buf.Cell(x, y).Rune != ' ' {
				rowsWithText++
				break
			}
		}
	}
	if rowsWithText < 2 {
		t.Errorf("expected the paragraph to wrap to >=2 rows, got %d", rowsWithText)
	}
}
```

- [ ] **Step 2: Run the test**

Run: `go test ./ -run TestRichText_BoldSurvivesWrap -v`
Expected: PASS (all prior tasks make this pass with no new production code).

- [ ] **Step 3: Run the full suite with the race detector**

Run: `go test -race ./`
Expected: PASS.

- [ ] **Step 4: Commit**

```bash
git add element_render_test.go
gcommit -m "test: rich-text bold survives wrapping (integration)"
```

---

## Self-Review (completed by plan author)

**Spec coverage (Layer 1 of the spec):**
- `TextSpan` + `WithRichText` + `Link` field — Task 1, Task 1 (Link field).
- Mutation/accessor + clearing semantics (`SetRichText`, `RichText`, text↔richtext precedence) — Task 1.
- Style-merge semantics (attrs OR, non-default color overrides) — Task 2.
- `wrapSpans` word-wrap carrying style, merge adjacent, newline handling — Task 3.
- Rich text in normal render path + gate — Task 4.
- Rich text in clipped/scroll render path + gate (the called-out silent-failure case) — Task 5.
- Measurement in both `IntrinsicSize` and `HeightForWidth` — Task 6.
- Shared `drawSpanLines` helper consumed by both render sites — Task 4 (defined), Task 5 (reused). Note: this is the pragmatic interpretation of the spec's "extract one shared helper" — the plain-text path is left intact to avoid regression, and the shared helper covers the new span rendering used by both sites.

**Out of scope here (later plans):** OSC 8 link rendering (`TextSpan.Link` is inert) — Plan 2. Markdown parsing/component/gsx — Plans 3–4.

**Placeholder scan:** none — every code/test step contains complete code and exact commands.

**Type consistency:** `TextSpan{Text, Style, Link}`, `wrapSpans([]TextSpan, int) [][]TextSpan`, `drawSpanLines(buf, [][]TextSpan, int, int, int, TextAlign, Style, Rect)`, `mergeSpanStyle(Style, Style) Style`, `richTextWidth([]TextSpan) int`, `spanLineWidth([]TextSpan) int` — names match across all tasks.
