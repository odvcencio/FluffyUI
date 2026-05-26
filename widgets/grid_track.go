package widgets

import "m31labs.dev/fluffyui/runtime"

// TrackSizeMode defines how a grid track is sized.
type TrackSizeMode int

const (
	TrackFixed TrackSizeMode = iota // Fixed pixel size
	TrackFr                         // Fractional unit (CSS fr)
	TrackAuto                       // Size to content
)

// TrackSize defines the size of a grid column or row.
type TrackSize struct {
	Mode  TrackSizeMode
	Value float64
}

// Fr creates a fractional track size (like CSS 1fr, 2fr).
func Fr(n float64) TrackSize { return TrackSize{Mode: TrackFr, Value: n} }

// Px creates a fixed pixel track size.
func Px(n int) TrackSize { return TrackSize{Mode: TrackFixed, Value: float64(n)} }

// AutoTrack creates a track that sizes to its content.
func AutoTrack() TrackSize { return TrackSize{Mode: TrackAuto} }

// resolveTracks resolves track sizes given available space and content measurements.
// contentSizes[i] is the measured content size for track i (used for Auto tracks).
// Returns the resolved pixel size for each track.
func resolveTracks(tracks []TrackSize, available int, gap int, contentSizes []int) []int {
	n := len(tracks)
	if n == 0 {
		return nil
	}

	sizes := make([]int, n)
	totalGaps := gap * (n - 1)
	remaining := available - totalGaps

	// Pass 1: Resolve fixed and auto tracks
	var totalFr float64
	for i, t := range tracks {
		switch t.Mode {
		case TrackFixed:
			sizes[i] = int(t.Value)
			remaining -= sizes[i]
		case TrackAuto:
			if i < len(contentSizes) {
				sizes[i] = contentSizes[i]
			}
			remaining -= sizes[i]
		case TrackFr:
			totalFr += t.Value
		}
	}

	// Pass 2: Distribute remaining space to fr tracks
	if totalFr > 0 && remaining > 0 {
		for i, t := range tracks {
			if t.Mode == TrackFr {
				sizes[i] = int(float64(remaining) * t.Value / totalFr)
			}
		}

		// Fix rounding: give leftover pixels to first fr track
		var frTotal int
		for i, t := range tracks {
			if t.Mode == TrackFr {
				frTotal += sizes[i]
			}
			_ = t
		}
		if diff := remaining - frTotal; diff > 0 {
			for i, t := range tracks {
				if t.Mode == TrackFr {
					sizes[i] += diff
					break
				}
			}
		}
	}

	// Ensure no negative sizes
	for i := range sizes {
		if sizes[i] < 0 {
			sizes[i] = 0
		}
	}

	return sizes
}

// trackOffsets converts sizes to starting offsets (cumulative positions with gaps).
func trackOffsets(sizes []int, gap int) []int {
	offsets := make([]int, len(sizes))
	pos := 0
	for i, s := range sizes {
		offsets[i] = pos
		pos += s + gap
	}
	return offsets
}

// TrackGridBuilder provides a fluent API for constructing track-based grids
// with CSS Grid-like column and row definitions.
type TrackGridBuilder struct {
	cols     []TrackSize
	rows     []TrackSize
	gap      int
	children []GridChild
}

// NewTrackGridBuilder creates a new TrackGridBuilder.
func NewTrackGridBuilder() *TrackGridBuilder { return &TrackGridBuilder{} }

// Columns sets the column track definitions.
func (b *TrackGridBuilder) Columns(cols ...TrackSize) *TrackGridBuilder {
	b.cols = cols
	return b
}

// Rows sets the row track definitions.
func (b *TrackGridBuilder) Rows(rows ...TrackSize) *TrackGridBuilder {
	b.rows = rows
	return b
}

// GapSize sets the gap between grid cells.
func (b *TrackGridBuilder) GapSize(gap int) *TrackGridBuilder {
	b.gap = gap
	return b
}

// Add adds a widget at the given row and column with span 1x1.
func (b *TrackGridBuilder) Add(w runtime.Widget, row, col int) *TrackGridBuilder {
	b.children = append(b.children, GridChild{Widget: w, Row: row, Col: col, RowSpan: 1, ColSpan: 1})
	return b
}

// AddSpan adds a widget at the given row and column with the specified spans.
func (b *TrackGridBuilder) AddSpan(w runtime.Widget, row, col, rowSpan, colSpan int) *TrackGridBuilder {
	b.children = append(b.children, GridChild{Widget: w, Row: row, Col: col, RowSpan: rowSpan, ColSpan: colSpan})
	return b
}

// Build constructs the Grid from the builder configuration.
func (b *TrackGridBuilder) Build() *Grid {
	rows := b.rows
	if len(rows) == 0 {
		// Auto-detect needed rows from children
		maxRow := 0
		for _, c := range b.children {
			if end := c.Row + c.RowSpan; end > maxRow {
				maxRow = end
			}
		}
		if maxRow == 0 {
			maxRow = 1
		}
		rows = make([]TrackSize, maxRow)
		for i := range rows {
			rows[i] = AutoTrack()
		}
	}
	g := NewTrackGrid(b.cols, rows)
	g.Gap = b.gap
	g.Children = b.children
	return g
}
