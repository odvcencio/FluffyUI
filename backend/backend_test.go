package backend

import "testing"

type recordTarget struct {
	width  int
	height int
	calls  []setCall
}

type setCall struct {
	x     int
	y     int
	mainc rune
	comb  []rune
	style Style
}

func (r *recordTarget) Size() (int, int) {
	return r.width, r.height
}

func (r *recordTarget) SetContent(x, y int, mainc rune, comb []rune, style Style) {
	r.calls = append(r.calls, setCall{x: x, y: y, mainc: mainc, comb: comb, style: style})
}

func TestColorRGBAndRGB(t *testing.T) {
	color := ColorRGB(10, 20, 30)
	if !color.IsRGB() {
		t.Fatalf("expected RGB color")
	}
	r, g, b := color.RGB()
	if r != 10 || g != 20 || b != 30 {
		t.Fatalf("unexpected RGB values: (%d, %d, %d)", r, g, b)
	}

	nonRGB := ColorRed
	if nonRGB.IsRGB() {
		t.Fatalf("expected palette color to not be RGB")
	}
	r, g, b = nonRGB.RGB()
	if r != 0 || g != 0 || b != 0 {
		t.Fatalf("expected non-RGB color to return zeros, got (%d, %d, %d)", r, g, b)
	}
}

func TestStyleAttributesAndAccessors(t *testing.T) {
	style := DefaultStyle()
	if style.FG() != ColorDefault || style.BG() != ColorDefault {
		t.Fatalf("expected default FG/BG colors")
	}

	style = style.
		Foreground(ColorRed).
		Background(ColorBlue).
		Bold(true).
		Italic(true).
		Dim(true).
		Underline(true).
		Reverse(true).
		Blink(true).
		StrikeThrough(true)

	wantAttrs := AttrBold | AttrItalic | AttrDim | AttrUnderline | AttrReverse | AttrBlink | AttrStrikeThrough
	if style.Attributes() != wantAttrs {
		t.Fatalf("unexpected attributes: got %v want %v", style.Attributes(), wantAttrs)
	}
	fg, bg, attrs := style.Decompose()
	if fg != ColorRed || bg != ColorBlue || attrs != wantAttrs {
		t.Fatalf("unexpected decompose: fg=%v bg=%v attrs=%v", fg, bg, attrs)
	}

	style = style.
		Bold(false).
		Italic(false).
		Dim(false).
		Underline(false).
		Reverse(false).
		Blink(false).
		StrikeThrough(false)
	if style.Attributes() != 0 {
		t.Fatalf("expected attributes cleared, got %v", style.Attributes())
	}
}

func TestSubTargetSetContent(t *testing.T) {
	parent := &recordTarget{width: 20, height: 10}
	sub := NewSubTarget(parent, 5, 7, 3, 2)

	w, h := sub.Size()
	if w != 3 || h != 2 {
		t.Fatalf("unexpected subtarget size: %dx%d", w, h)
	}

	style := DefaultStyle().Foreground(ColorGreen)
	sub.SetContent(1, 1, 'x', []rune{'y'}, style)
	if len(parent.calls) != 1 {
		t.Fatalf("expected 1 parent call, got %d", len(parent.calls))
	}
	call := parent.calls[0]
	if call.x != 6 || call.y != 8 {
		t.Fatalf("unexpected coordinates: (%d, %d)", call.x, call.y)
	}
	if call.mainc != 'x' {
		t.Fatalf("unexpected rune: %q", call.mainc)
	}
	if len(call.comb) != 1 || call.comb[0] != 'y' {
		t.Fatalf("unexpected comb: %v", call.comb)
	}
	if call.style.ForegroundColor() != ColorGreen {
		t.Fatalf("unexpected style foreground: %v", call.style.ForegroundColor())
	}

	sub.SetContent(-1, 0, 'a', nil, style)
	sub.SetContent(3, 0, 'b', nil, style)
	sub.SetContent(0, 2, 'c', nil, style)
	if len(parent.calls) != 1 {
		t.Fatalf("expected out-of-bounds writes to be ignored")
	}
}
