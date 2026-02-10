// Package i18n cursor.go provides cursor position mapping between logical and
// visual positions in bidirectional text.
package i18n

// CursorMap maps between logical (editing) and visual (display) positions
// in bidirectional text.
type CursorMap struct {
	logicalToVisual []int
	visualToLogical []int
}

// BuildCursorMap creates a mapping between logical text order and visual
// (display) order after BiDi reordering. Both strings must have the same
// rune count.
func BuildCursorMap(logical, visual string) CursorMap {
	logRunes := []rune(logical)
	visRunes := []rune(visual)
	if len(logRunes) != len(visRunes) {
		// Fallback: identity mapping
		m := make([]int, len(logRunes))
		for i := range m {
			m[i] = i
		}
		return CursorMap{logicalToVisual: m, visualToLogical: m}
	}

	l2v := make([]int, len(logRunes))
	v2l := make([]int, len(visRunes))
	used := make([]bool, len(visRunes))

	for i, lr := range logRunes {
		for j, vr := range visRunes {
			if !used[j] && lr == vr {
				l2v[i] = j
				v2l[j] = i
				used[j] = true
				break
			}
		}
	}
	return CursorMap{logicalToVisual: l2v, visualToLogical: v2l}
}

// LogicalToVisual converts a logical cursor position to its visual position.
func (m CursorMap) LogicalToVisual(pos int) int {
	if len(m.logicalToVisual) == 0 {
		return pos
	}
	if pos >= len(m.logicalToVisual) {
		return len(m.logicalToVisual)
	}
	if pos < 0 {
		return 0
	}
	return m.logicalToVisual[pos]
}

// VisualToLogical converts a visual cursor position to its logical position.
func (m CursorMap) VisualToLogical(pos int) int {
	if len(m.visualToLogical) == 0 {
		return pos
	}
	if pos >= len(m.visualToLogical) {
		return len(m.visualToLogical)
	}
	if pos < 0 {
		return 0
	}
	return m.visualToLogical[pos]
}
