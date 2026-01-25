package fur

import "testing"

func TestStyleForeground(t *testing.T) {
	s := DefaultStyle().Foreground(ColorRed)
	if s.fg != ColorRed {
		t.Errorf("expected red foreground, got %v", s.fg)
	}
}

func TestStyleBackground(t *testing.T) {
	s := DefaultStyle().Background(ColorBlue)
	if s.bg != ColorBlue {
		t.Errorf("expected blue background, got %v", s.bg)
	}
}

func TestStyleBold(t *testing.T) {
	s := DefaultStyle().Bold()
	if !s.bold {
		t.Error("expected bold")
	}
}

func TestStyleItalic(t *testing.T) {
	s := DefaultStyle().Italic()
	if !s.italic {
		t.Error("expected italic")
	}
}

func TestStyleUnderline(t *testing.T) {
	s := DefaultStyle().Underline()
	if !s.underline {
		t.Error("expected underline")
	}
}

func TestStyleDim(t *testing.T) {
	s := DefaultStyle().Dim()
	if !s.dim {
		t.Error("expected dim")
	}
}

func TestStyleChaining(t *testing.T) {
	s := DefaultStyle().Bold().Italic().Foreground(ColorCyan)
	if !s.bold {
		t.Error("expected bold")
	}
	if !s.italic {
		t.Error("expected italic")
	}
	if s.fg != ColorCyan {
		t.Errorf("expected cyan, got %v", s.fg)
	}
}

func TestStyleEqual(t *testing.T) {
	s1 := DefaultStyle().Bold().Foreground(ColorRed)
	s2 := DefaultStyle().Bold().Foreground(ColorRed)
	s3 := DefaultStyle().Bold().Foreground(ColorBlue)

	if !s1.Equal(s2) {
		t.Error("expected s1 == s2")
	}
	if s1.Equal(s3) {
		t.Error("expected s1 != s3")
	}
}

func TestColorHelpers(t *testing.T) {
	rgb := RGB(255, 128, 0)
	if rgb.Value != 0xFF8000 {
		t.Errorf("RGB value wrong: %x", rgb.Value)
	}

	hex := Hex(0xAABBCC)
	if hex.Value != 0xAABBCC {
		t.Errorf("Hex value wrong: %x", hex.Value)
	}

	c256 := Color256(100)
	if c256.Value != 100 {
		t.Errorf("Color256 value wrong: %d", c256.Value)
	}
}
