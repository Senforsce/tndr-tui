package tui

import "testing"

func TestDefaultMarkdownTheme(t *testing.T) {
	th := DefaultMarkdownTheme()

	if a := th.Heading[0].Attrs; a&AttrBold == 0 || a&AttrUnderline == 0 || a&AttrItalic == 0 {
		t.Errorf("h1 should be bold + underline + italic, attrs=%v", a)
	}
	if a := th.Heading[1].Attrs; a&AttrBold == 0 || a&AttrItalic == 0 || a&AttrUnderline != 0 {
		t.Errorf("h2 should be bold + italic (no underline), attrs=%v", a)
	}
	if a := th.Heading[2].Attrs; a&AttrItalic == 0 || a&AttrBold != 0 {
		t.Errorf("h3 should be italic only, attrs=%v", a)
	}
	if th.BlockquoteText.Attrs&AttrItalic == 0 {
		t.Errorf("blockquote text should be italic, attrs=%v", th.BlockquoteText.Attrs)
	}
	if th.TableBorder == BorderNone {
		t.Errorf("table should be outlined by default")
	}
	if th.Bold.Attrs&AttrBold == 0 {
		t.Errorf("Bold style should set bold attr")
	}
	if th.Italic.Attrs&AttrItalic == 0 {
		t.Errorf("Italic style should set italic attr")
	}
	if th.Link.Attrs&AttrUnderline == 0 {
		t.Errorf("Link style should be underlined")
	}
	if th.BulletMarker == "" {
		t.Errorf("BulletMarker should have a default")
	}
	if th.BlockquoteBar == 0 {
		t.Errorf("BlockquoteBar should have a default glyph")
	}
}
