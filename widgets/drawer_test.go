package widgets

import (
	"testing"

	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/state"
	"github.com/odvcencio/fluffyui/terminal"
)

func TestDrawer_New(t *testing.T) {
	open := state.NewSignal(false)
	child := NewBadge("content")
	d := NewDrawer(child, open)
	if d == nil {
		t.Fatal("expected non-nil Drawer")
	}
	if d.IsOpen() {
		t.Fatal("expected drawer to start closed")
	}
	if d.CanFocus() {
		t.Fatal("expected drawer (Base) to not be focusable")
	}
}

func TestDrawer_Role(t *testing.T) {
	open := state.NewSignal(false)
	d := NewDrawer(nil, open)
	if d.AccessibleRole() != "complementary" {
		t.Fatalf("expected role complementary, got %q", d.AccessibleRole())
	}
}

func TestDrawer_WithOptions(t *testing.T) {
	open := state.NewSignal(false)
	d := NewDrawer(nil, open,
		WithDrawerSide(DrawerRight),
		WithDrawerWidth(40),
		WithDrawerHeight(20),
	)
	if d.side != DrawerRight {
		t.Fatalf("expected DrawerRight, got %d", d.side)
	}
	if d.width != 40 {
		t.Fatalf("expected width 40, got %d", d.width)
	}
	if d.height != 20 {
		t.Fatalf("expected height 20, got %d", d.height)
	}
}

func TestDrawer_OpenClose(t *testing.T) {
	open := state.NewSignal(false)
	d := NewDrawer(nil, open)

	d.SetOpen(true)
	if !d.IsOpen() {
		t.Fatal("expected drawer to be open")
	}

	d.SetOpen(false)
	if d.IsOpen() {
		t.Fatal("expected drawer to be closed")
	}
}

func TestDrawer_EscapeCloses(t *testing.T) {
	open := state.NewSignal(true)
	d := NewDrawer(nil, open)

	result := d.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEscape})
	if !result.Handled {
		t.Fatal("expected escape to be handled")
	}
	if d.IsOpen() {
		t.Fatal("expected drawer to be closed after escape")
	}
}

func TestDrawer_ClosedUnhandled(t *testing.T) {
	open := state.NewSignal(false)
	d := NewDrawer(nil, open)

	result := d.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEscape})
	if result.Handled {
		t.Fatal("expected unhandled when drawer is closed")
	}
}

func TestDrawer_MeasureClosed(t *testing.T) {
	open := state.NewSignal(false)
	d := NewDrawer(nil, open, WithDrawerWidth(30))
	size := d.Measure(runtime.Constraints{MaxWidth: 80, MaxHeight: 24})
	if size.Width != 0 || size.Height != 0 {
		t.Fatalf("expected zero size when closed, got %dx%d", size.Width, size.Height)
	}
}

func TestDrawer_MeasureOpenLeft(t *testing.T) {
	open := state.NewSignal(true)
	d := NewDrawer(nil, open, WithDrawerSide(DrawerLeft), WithDrawerWidth(20))
	size := d.Measure(runtime.Constraints{MaxWidth: 80, MaxHeight: 24})
	if size.Width != 20 {
		t.Fatalf("expected width 20 when open left, got %d", size.Width)
	}
	if size.Height != 24 {
		t.Fatalf("expected height 24 when open left, got %d", size.Height)
	}
}

func TestDrawer_MeasureOpenBottom(t *testing.T) {
	open := state.NewSignal(true)
	d := NewDrawer(nil, open, WithDrawerSide(DrawerBottom), WithDrawerHeight(8))
	size := d.Measure(runtime.Constraints{MaxWidth: 80, MaxHeight: 24})
	if size.Width != 80 {
		t.Fatalf("expected width 80 when open bottom, got %d", size.Width)
	}
	if size.Height != 8 {
		t.Fatalf("expected height 8 when open bottom, got %d", size.Height)
	}
}

func TestDrawer_ChildWidgets(t *testing.T) {
	open := state.NewSignal(false)
	child := NewBadge("test")
	d := NewDrawer(child, open)

	// Closed: no children.
	children := d.ChildWidgets()
	if len(children) != 0 {
		t.Fatalf("expected 0 children when closed, got %d", len(children))
	}

	// Open: child returned.
	d.SetOpen(true)
	children = d.ChildWidgets()
	if len(children) != 1 {
		t.Fatalf("expected 1 child when open, got %d", len(children))
	}
}

func TestDrawer_A11yState(t *testing.T) {
	open := state.NewSignal(false)
	d := NewDrawer(nil, open)

	st := d.AccessibleState()
	if st.Expanded == nil {
		t.Fatal("expected Expanded to be set")
	}
	if *st.Expanded {
		t.Fatal("expected Expanded false when closed")
	}
	if !st.Hidden {
		t.Fatal("expected Hidden true when closed")
	}

	d.SetOpen(true)
	st = d.AccessibleState()
	if !*st.Expanded {
		t.Fatal("expected Expanded true when open")
	}
	if st.Hidden {
		t.Fatal("expected Hidden false when open")
	}
}
