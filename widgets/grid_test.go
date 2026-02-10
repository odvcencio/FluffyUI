package widgets

import (
	"testing"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/runtime"
)

// gridTestWidget is a minimal widget for grid tests with a configurable preferred size.
type gridTestWidget struct {
	Base
	preferredSize runtime.Size
	layoutBounds  runtime.Rect
}

func (w *gridTestWidget) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.Constrain(w.preferredSize)
}

func (w *gridTestWidget) Layout(bounds runtime.Rect) {
	w.layoutBounds = bounds
	w.Base.Layout(bounds)
}

func (w *gridTestWidget) Render(ctx runtime.RenderContext) {}

func (w *gridTestWidget) HandleMessage(msg runtime.Message) runtime.HandleResult {
	return runtime.Unhandled()
}

// --- resolveTracks tests ---

func TestResolveTracks_AllFr(t *testing.T) {
	tracks := []TrackSize{Fr(1), Fr(2), Fr(1)}
	sizes := resolveTracks(tracks, 40, 0, nil)
	if len(sizes) != 3 {
		t.Fatalf("got %d sizes, want 3", len(sizes))
	}
	// 1fr=10, 2fr=20, 1fr=10
	if sizes[0] != 10 || sizes[1] != 20 || sizes[2] != 10 {
		t.Errorf("got %v, want [10 20 10]", sizes)
	}
}

func TestResolveTracks_MixedFixedFr(t *testing.T) {
	tracks := []TrackSize{Px(20), Fr(1), Fr(1)}
	sizes := resolveTracks(tracks, 60, 0, nil)
	if len(sizes) != 3 {
		t.Fatalf("got %d sizes, want 3", len(sizes))
	}
	// Fixed=20, remaining=40, each fr=20
	if sizes[0] != 20 || sizes[1] != 20 || sizes[2] != 20 {
		t.Errorf("got %v, want [20 20 20]", sizes)
	}
}

func TestResolveTracks_WithGap(t *testing.T) {
	tracks := []TrackSize{Fr(1), Fr(1)}
	sizes := resolveTracks(tracks, 41, 1, nil)
	if len(sizes) != 2 {
		t.Fatalf("got %d sizes, want 2", len(sizes))
	}
	// available=41, gap=1, remaining=40, each fr=20
	if sizes[0] != 20 || sizes[1] != 20 {
		t.Errorf("got %v, want [20 20]", sizes)
	}
}

func TestResolveTracks_AllFixed(t *testing.T) {
	tracks := []TrackSize{Px(10), Px(20), Px(30)}
	sizes := resolveTracks(tracks, 100, 0, nil)
	if sizes[0] != 10 || sizes[1] != 20 || sizes[2] != 30 {
		t.Errorf("got %v, want [10 20 30]", sizes)
	}
}

func TestResolveTracks_Auto(t *testing.T) {
	tracks := []TrackSize{AutoTrack(), AutoTrack()}
	content := []int{15, 25}
	sizes := resolveTracks(tracks, 100, 0, content)
	if sizes[0] != 15 || sizes[1] != 25 {
		t.Errorf("got %v, want [15 25]", sizes)
	}
}

func TestResolveTracks_MixedAutoFr(t *testing.T) {
	tracks := []TrackSize{AutoTrack(), Fr(1)}
	content := []int{20}
	sizes := resolveTracks(tracks, 60, 0, content)
	// Auto=20, remaining=40, fr=40
	if sizes[0] != 20 || sizes[1] != 40 {
		t.Errorf("got %v, want [20 40]", sizes)
	}
}

func TestResolveTracks_Empty(t *testing.T) {
	sizes := resolveTracks(nil, 100, 0, nil)
	if sizes != nil {
		t.Errorf("got %v, want nil", sizes)
	}
}

func TestResolveTracks_NoNegative(t *testing.T) {
	// Fixed tracks exceed available space
	tracks := []TrackSize{Px(50), Px(50), Fr(1)}
	sizes := resolveTracks(tracks, 80, 0, nil)
	// Fixed=50+50=100 > 80, remaining=-20, fr should be 0
	if sizes[2] < 0 {
		t.Errorf("fr track should not be negative, got %d", sizes[2])
	}
}

func TestResolveTracks_WithGapAndFixed(t *testing.T) {
	tracks := []TrackSize{Px(10), Fr(1), Px(10)}
	sizes := resolveTracks(tracks, 42, 1, nil)
	// gaps = 2*1 = 2, fixed = 20, remaining = 42-2-20 = 20, fr = 20
	if sizes[0] != 10 || sizes[1] != 20 || sizes[2] != 10 {
		t.Errorf("got %v, want [10 20 10]", sizes)
	}
}

// --- trackOffsets tests ---

func TestTrackOffsets(t *testing.T) {
	sizes := []int{10, 20, 30}
	offsets := trackOffsets(sizes, 0)
	if offsets[0] != 0 || offsets[1] != 10 || offsets[2] != 30 {
		t.Errorf("got %v, want [0 10 30]", offsets)
	}
}

func TestTrackOffsets_WithGap(t *testing.T) {
	sizes := []int{10, 20, 30}
	offsets := trackOffsets(sizes, 2)
	// 0, 10+2=12, 12+20+2=34
	if offsets[0] != 0 || offsets[1] != 12 || offsets[2] != 34 {
		t.Errorf("got %v, want [0 12 34]", offsets)
	}
}

// --- NewTrackGrid tests ---

func TestNewTrackGrid(t *testing.T) {
	g := NewTrackGrid([]TrackSize{Fr(1), Fr(2)}, []TrackSize{AutoTrack()})
	if g == nil {
		t.Fatal("NewTrackGrid returned nil")
	}
	if g.Cols != 2 {
		t.Errorf("Cols = %d, want 2", g.Cols)
	}
	if g.Rows != 1 {
		t.Errorf("Rows = %d, want 1", g.Rows)
	}
	if len(g.ColTracks) != 2 {
		t.Errorf("ColTracks len = %d, want 2", len(g.ColTracks))
	}
	if len(g.RowTracks) != 1 {
		t.Errorf("RowTracks len = %d, want 1", len(g.RowTracks))
	}
	if g.AccessibleRole() != accessibility.RoleGrid {
		t.Errorf("role = %q, want %q", g.AccessibleRole(), accessibility.RoleGrid)
	}
}

// --- TrackGridBuilder tests ---

func TestTrackGridBuilder_Basic(t *testing.T) {
	w1 := &gridTestWidget{preferredSize: runtime.Size{Width: 10, Height: 5}}
	w2 := &gridTestWidget{preferredSize: runtime.Size{Width: 10, Height: 5}}

	grid := NewTrackGridBuilder().
		Columns(Fr(1), Fr(1)).
		GapSize(0).
		Add(w1, 0, 0).
		Add(w2, 0, 1).
		Build()

	if grid.Cols != 2 {
		t.Fatalf("Cols = %d, want 2", grid.Cols)
	}
	if len(grid.Children) != 2 {
		t.Fatalf("children = %d, want 2", len(grid.Children))
	}
	// Rows should be auto-detected as 1
	if grid.Rows != 1 {
		t.Errorf("Rows = %d, want 1", grid.Rows)
	}
}

func TestTrackGridBuilder_AutoRows(t *testing.T) {
	w1 := &gridTestWidget{preferredSize: runtime.Size{Width: 10, Height: 5}}
	w2 := &gridTestWidget{preferredSize: runtime.Size{Width: 10, Height: 5}}
	w3 := &gridTestWidget{preferredSize: runtime.Size{Width: 10, Height: 5}}

	grid := NewTrackGridBuilder().
		Columns(Fr(1), Fr(1)).
		GapSize(0).
		Add(w1, 0, 0).
		Add(w2, 0, 1).
		Add(w3, 1, 0).
		Build()

	if grid.Rows != 2 {
		t.Errorf("auto-detected Rows = %d, want 2", grid.Rows)
	}
}

func TestTrackGridBuilder_ExplicitRows(t *testing.T) {
	grid := NewTrackGridBuilder().
		Columns(Fr(1)).
		Rows(Px(10), Px(20)).
		Build()

	if grid.Rows != 2 {
		t.Errorf("Rows = %d, want 2", grid.Rows)
	}
	if grid.RowTracks[0].Mode != TrackFixed || grid.RowTracks[0].Value != 10 {
		t.Errorf("row 0: got %+v, want Px(10)", grid.RowTracks[0])
	}
}

func TestTrackGridBuilder_WithSpan(t *testing.T) {
	w := &gridTestWidget{preferredSize: runtime.Size{Width: 20, Height: 5}}

	grid := NewTrackGridBuilder().
		Columns(Fr(1), Fr(1), Fr(1)).
		GapSize(0).
		AddSpan(w, 0, 0, 1, 2).
		Build()

	if len(grid.Children) != 1 {
		t.Fatalf("children = %d, want 1", len(grid.Children))
	}
	c := grid.Children[0]
	if c.ColSpan != 2 {
		t.Errorf("ColSpan = %d, want 2", c.ColSpan)
	}
}

func TestTrackGridBuilder_GapSize(t *testing.T) {
	grid := NewTrackGridBuilder().
		Columns(Fr(1), Fr(1)).
		GapSize(5).
		Build()

	if grid.Gap != 5 {
		t.Errorf("Gap = %d, want 5", grid.Gap)
	}
}

// --- Track-based layout tests ---

func TestGrid_TrackLayout_TwoFrColumns(t *testing.T) {
	w1 := &gridTestWidget{preferredSize: runtime.Size{Width: 10, Height: 5}}
	w2 := &gridTestWidget{preferredSize: runtime.Size{Width: 10, Height: 5}}

	grid := NewTrackGridBuilder().
		Columns(Fr(1), Fr(1)).
		Rows(Fr(1)).
		GapSize(0).
		Add(w1, 0, 0).
		Add(w2, 0, 1).
		Build()

	grid.Layout(runtime.Rect{X: 0, Y: 0, Width: 40, Height: 10})

	if w1.layoutBounds.X != 0 || w1.layoutBounds.Width != 20 {
		t.Errorf("w1: x=%d width=%d, want x=0 width=20", w1.layoutBounds.X, w1.layoutBounds.Width)
	}
	if w2.layoutBounds.X != 20 || w2.layoutBounds.Width != 20 {
		t.Errorf("w2: x=%d width=%d, want x=20 width=20", w2.layoutBounds.X, w2.layoutBounds.Width)
	}
}

func TestGrid_TrackLayout_FixedAndFr(t *testing.T) {
	w1 := &gridTestWidget{preferredSize: runtime.Size{Width: 10, Height: 5}}
	w2 := &gridTestWidget{preferredSize: runtime.Size{Width: 10, Height: 5}}

	grid := NewTrackGridBuilder().
		Columns(Px(10), Fr(1)).
		Rows(Fr(1)).
		GapSize(0).
		Add(w1, 0, 0).
		Add(w2, 0, 1).
		Build()

	grid.Layout(runtime.Rect{X: 0, Y: 0, Width: 50, Height: 10})

	if w1.layoutBounds.Width != 10 {
		t.Errorf("w1 width=%d, want 10", w1.layoutBounds.Width)
	}
	if w2.layoutBounds.X != 10 || w2.layoutBounds.Width != 40 {
		t.Errorf("w2: x=%d width=%d, want x=10 width=40", w2.layoutBounds.X, w2.layoutBounds.Width)
	}
}

func TestGrid_TrackLayout_WithGap(t *testing.T) {
	w1 := &gridTestWidget{preferredSize: runtime.Size{Width: 10, Height: 5}}
	w2 := &gridTestWidget{preferredSize: runtime.Size{Width: 10, Height: 5}}

	grid := NewTrackGridBuilder().
		Columns(Fr(1), Fr(1)).
		Rows(Fr(1)).
		GapSize(2).
		Add(w1, 0, 0).
		Add(w2, 0, 1).
		Build()

	grid.Layout(runtime.Rect{X: 0, Y: 0, Width: 42, Height: 10})

	// gap=2, available for cols = 42-2=40, each fr=20
	if w1.layoutBounds.Width != 20 {
		t.Errorf("w1 width=%d, want 20", w1.layoutBounds.Width)
	}
	if w2.layoutBounds.X != 22 {
		t.Errorf("w2 x=%d, want 22 (20 + 2 gap)", w2.layoutBounds.X)
	}
	if w2.layoutBounds.Width != 20 {
		t.Errorf("w2 width=%d, want 20", w2.layoutBounds.Width)
	}
}

func TestGrid_TrackLayout_ColSpan(t *testing.T) {
	w1 := &gridTestWidget{preferredSize: runtime.Size{Width: 10, Height: 5}}

	grid := NewTrackGridBuilder().
		Columns(Fr(1), Fr(1), Fr(1)).
		Rows(Fr(1)).
		GapSize(0).
		AddSpan(w1, 0, 0, 1, 2).
		Build()

	grid.Layout(runtime.Rect{X: 0, Y: 0, Width: 30, Height: 10})

	// 3 cols of 10 each, span 2 = 20
	if w1.layoutBounds.Width != 20 {
		t.Errorf("w1 width=%d, want 20 (spanning 2 columns)", w1.layoutBounds.Width)
	}
}

func TestGrid_TrackLayout_ColSpanWithGap(t *testing.T) {
	w1 := &gridTestWidget{preferredSize: runtime.Size{Width: 10, Height: 5}}

	grid := NewTrackGridBuilder().
		Columns(Fr(1), Fr(1), Fr(1)).
		Rows(Fr(1)).
		GapSize(2).
		AddSpan(w1, 0, 0, 1, 2).
		Build()

	grid.Layout(runtime.Rect{X: 0, Y: 0, Width: 34, Height: 10})

	// gaps = 2*2 = 4, remaining = 30, each col = 10
	// span 2 cols = 10 + 10 + gap = 22
	if w1.layoutBounds.Width != 22 {
		t.Errorf("w1 width=%d, want 22 (2 cols of 10 + 2 gap)", w1.layoutBounds.Width)
	}
}

// --- Legacy grid backward compatibility ---

func TestGrid_LegacyUniform(t *testing.T) {
	w1 := &gridTestWidget{preferredSize: runtime.Size{Width: 10, Height: 5}}
	w2 := &gridTestWidget{preferredSize: runtime.Size{Width: 10, Height: 5}}

	grid := NewGrid(1, 2)
	grid.Add(w1, 0, 0, 1, 1)
	grid.Add(w2, 0, 1, 1, 1)

	grid.Layout(runtime.Rect{X: 0, Y: 0, Width: 40, Height: 10})

	if w1.layoutBounds.Width != 20 {
		t.Errorf("w1 width=%d, want 20", w1.layoutBounds.Width)
	}
	if w2.layoutBounds.X != 20 {
		t.Errorf("w2 x=%d, want 20", w2.layoutBounds.X)
	}
}

// --- TrackSize constructors ---

func TestTrackSizeConstructors(t *testing.T) {
	f := Fr(2.5)
	if f.Mode != TrackFr || f.Value != 2.5 {
		t.Errorf("Fr(2.5) = %+v", f)
	}

	p := Px(42)
	if p.Mode != TrackFixed || p.Value != 42 {
		t.Errorf("Px(42) = %+v", p)
	}

	a := AutoTrack()
	if a.Mode != TrackAuto {
		t.Errorf("AutoTrack() = %+v", a)
	}
}
