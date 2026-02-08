package runtime

import (
	"math"
	"strconv"
)

// FlexDirection specifies the main axis of a flex container.
type FlexDirection int

const (
	Column FlexDirection = iota // Vertical (VBox)
	Row                         // Horizontal (HBox)
)

// FlexChild wraps a widget with flex layout properties.
type FlexChild struct {
	Widget Widget
	Grow   float64 // How much to grow (0 = fixed, 1+ = proportional)
	Shrink float64 // How much to shrink (0 = fixed, 1+ = proportional)
	Basis  int     // Base size (-1 = use measured size)
}

// Fixed creates a child that doesn't grow or shrink.
func Fixed(w Widget) FlexChild {
	return FlexChild{Widget: w, Grow: 0, Shrink: 0, Basis: -1}
}

// Flexible creates a child that grows with the given factor.
func Flexible(w Widget, grow float64) FlexChild {
	return FlexChild{Widget: w, Grow: grow, Shrink: 1, Basis: -1}
}

// Expanded creates a child that grows to fill available space (Grow=1).
func Expanded(w Widget) FlexChild {
	return FlexChild{Widget: w, Grow: 1, Shrink: 1, Basis: -1}
}

// Sized creates a child with a fixed basis size.
func Sized(w Widget, basis int) FlexChild {
	return FlexChild{Widget: w, Grow: 0, Shrink: 0, Basis: basis}
}

// Flex is a container that lays out children along an axis.
type Flex struct {
	Direction    FlexDirection
	Children     []FlexChild
	Gap          int           // Space between children
	MeasureCache *MeasureCache // Optional measurement cache (set by Screen)

	// Cached layout
	bounds      Rect
	childBounds []Rect
	measured    Size

	// Reusable buffers to reduce allocations
	childSizesCache   []Size
	baseSizesCache    []int
	growWeightsCache  []float64
	shrinkWeightsCache []float64
	sizesCache        []int
}

// VBox creates a vertical flex container.
func VBox(children ...FlexChild) *Flex {
	return &Flex{Direction: Column, Children: children}
}

// HBox creates a horizontal flex container.
func HBox(children ...FlexChild) *Flex {
	return &Flex{Direction: Row, Children: children}
}

// WithGap sets the gap between children.
func (f *Flex) WithGap(gap int) *Flex {
	f.Gap = gap
	return f
}

// Add appends a child to the flex container.
func (f *Flex) Add(child FlexChild) {
	f.Children = append(f.Children, child)
}

// Measure calculates the desired size of the flex container.
func (f *Flex) Measure(constraints Constraints) Size {
	if len(f.Children) == 0 {
		f.measured = constraints.MinSize()
		return f.measured
	}

	// Measure all children with loose constraints on the main axis
	n := len(f.Children)
	if cap(f.childSizesCache) < n {
		f.childSizesCache = make([]Size, n)
	} else {
		f.childSizesCache = f.childSizesCache[:n]
	}
	childSizes := f.childSizesCache
	totalMain := 0
	maxCross := 0

	for i, child := range f.Children {
		var childConstraints Constraints
		if f.Direction == Column {
			// For column: constrain width, free height
			childConstraints = Constraints{
				MinWidth:  constraints.MinWidth,
				MaxWidth:  constraints.MaxWidth,
				MinHeight: 0,
				MaxHeight: maxInt,
			}
		} else {
			// For row: free width, constrain height
			childConstraints = Constraints{
				MinWidth:  0,
				MaxWidth:  maxInt,
				MinHeight: constraints.MinHeight,
				MaxHeight: constraints.MaxHeight,
			}
		}
		WarnInvalidConstraints("runtime.Flex.Measure.child", childConstraints)

		if child.Basis >= 0 {
			childSizes[i] = f.sizeWithBasis(child.Basis)
		} else {
			childSizes[i] = f.cachedMeasure(child.Widget, childConstraints)
		}
		WarnZeroMeasure("runtime.Flex.Measure.child", childConstraints, childSizes[i])

		if f.Direction == Column {
			totalMain += childSizes[i].Height
			maxCross = max(maxCross, childSizes[i].Width)
		} else {
			totalMain += childSizes[i].Width
			maxCross = max(maxCross, childSizes[i].Height)
		}
	}

	// Add gaps
	if len(f.Children) > 1 {
		totalMain += f.Gap * (len(f.Children) - 1)
	}

	if f.Direction == Column {
		f.measured = constraints.Constrain(Size{Width: maxCross, Height: totalMain})
	} else {
		f.measured = constraints.Constrain(Size{Width: totalMain, Height: maxCross})
	}
	return f.measured
}

// Layout positions all children within the given bounds.
func (f *Flex) Layout(bounds Rect) {
	f.bounds = bounds
	n := len(f.Children)
	if cap(f.childBounds) < n {
		f.childBounds = make([]Rect, n)
	} else {
		f.childBounds = f.childBounds[:n]
	}

	if n == 0 {
		return
	}

	// Measure children to get their preferred sizes
	if cap(f.childSizesCache) < n {
		f.childSizesCache = make([]Size, n)
	} else {
		f.childSizesCache = f.childSizesCache[:n]
	}
	childSizes := f.childSizesCache

	if cap(f.baseSizesCache) < n {
		f.baseSizesCache = make([]int, n)
	} else {
		f.baseSizesCache = f.baseSizesCache[:n]
	}
	baseSizes := f.baseSizesCache

	if cap(f.growWeightsCache) < n {
		f.growWeightsCache = make([]float64, n)
	} else {
		f.growWeightsCache = f.growWeightsCache[:n]
		for i := range f.growWeightsCache {
			f.growWeightsCache[i] = 0
		}
	}
	growWeights := f.growWeightsCache

	if cap(f.shrinkWeightsCache) < n {
		f.shrinkWeightsCache = make([]float64, n)
	} else {
		f.shrinkWeightsCache = f.shrinkWeightsCache[:n]
		for i := range f.shrinkWeightsCache {
			f.shrinkWeightsCache[i] = 0
		}
	}
	shrinkWeights := f.shrinkWeightsCache

	totalBase := 0
	totalGrow := 0.0
	totalShrink := 0.0

	for i, child := range f.Children {
		var childConstraints Constraints
		if f.Direction == Column {
			childConstraints = Loose(bounds.Width, maxInt)
		} else {
			childConstraints = Loose(maxInt, bounds.Height)
		}
		WarnInvalidConstraints("runtime.Flex.Layout.child", childConstraints)

		if child.Basis >= 0 {
			childSizes[i] = f.sizeWithBasis(child.Basis)
		} else {
			childSizes[i] = f.cachedMeasure(child.Widget, childConstraints)
		}
		WarnZeroMeasure("runtime.Flex.Layout.child", childConstraints, childSizes[i])

		mainSize := f.mainSize(childSizes[i])
		baseSizes[i] = mainSize
		totalBase += mainSize
		if child.Grow > 0 {
			growWeights[i] = child.Grow
			totalGrow += child.Grow
		}
		if child.Shrink > 0 {
			shrinkWeights[i] = child.Shrink * float64(mainSize)
			totalShrink += shrinkWeights[i]
		}
	}

	// Add gaps to fixed space
	gaps := 0
	if len(f.Children) > 1 {
		gaps = f.Gap * (len(f.Children) - 1)
	}
	containerMain := f.mainSize(bounds.Size())
	available := containerMain - gaps - totalBase

	if cap(f.sizesCache) < n {
		f.sizesCache = make([]int, n)
	} else {
		f.sizesCache = f.sizesCache[:n]
	}
	sizes := f.sizesCache
	copy(sizes, baseSizes)
	if available > 0 && totalGrow > 0 {
		extras := distributeFlexSpace(available, growWeights)
		for i := range sizes {
			sizes[i] += extras[i]
		}
	} else if available < 0 && totalShrink > 0 {
		shrinks := distributeFlexShrink(-available, shrinkWeights, sizes)
		for i := range sizes {
			sizes[i] -= shrinks[i]
			if sizes[i] < 0 {
				sizes[i] = 0
			}
		}
	}

	// Position children
	offset := 0
	for i, child := range f.Children {
		// Calculate size
		mainSize := sizes[i]

		// Create bounds for this child
		var childBounds Rect
		if f.Direction == Column {
			childBounds = Rect{
				X:      bounds.X,
				Y:      bounds.Y + offset,
				Width:  bounds.Width,
				Height: mainSize,
			}
		} else {
			childBounds = Rect{
				X:      bounds.X + offset,
				Y:      bounds.Y,
				Width:  mainSize,
				Height: bounds.Height,
			}
		}

		f.childBounds[i] = childBounds
		child.Widget.Layout(childBounds)

		offset += mainSize + f.Gap
	}
}

func distributeFlexSpace(available int, weights []float64) []int {
	out := make([]int, len(weights))
	if available <= 0 {
		return out
	}
	total := 0.0
	for _, w := range weights {
		total += w
	}
	if total <= 0 {
		return out
	}
	fractions := make([]float64, len(weights))
	used := 0
	for i, w := range weights {
		if w <= 0 {
			continue
		}
		share := float64(available) * (w / total)
		base := int(math.Floor(share))
		out[i] = base
		used += base
		fractions[i] = share - float64(base)
	}
	remaining := available - used
	for remaining > 0 {
		idx := -1
		best := -1.0
		for i, frac := range fractions {
			if frac > best {
				best = frac
				idx = i
			}
		}
		if idx == -1 {
			break
		}
		out[idx]++
		fractions[idx] = 0
		remaining--
	}
	return out
}

func distributeFlexShrink(need int, weights []float64, sizes []int) []int {
	out := make([]int, len(weights))
	if need <= 0 {
		return out
	}
	remaining := need
	for rounds := 0; rounds < len(weights) && remaining > 0; rounds++ {
		total := 0.0
		for i, w := range weights {
			if w > 0 && sizes[i]-out[i] > 0 {
				total += w
			}
		}
		if total <= 0 {
			break
		}
		fractions := make([]float64, len(weights))
		used := 0
		for i, w := range weights {
			capacity := sizes[i] - out[i]
			if w <= 0 || capacity <= 0 {
				continue
			}
			share := float64(remaining) * (w / total)
			base := int(math.Floor(share))
			if base > capacity {
				base = capacity
			}
			out[i] += base
			used += base
			fractions[i] = share - float64(base)
		}
		remaining -= used
		if remaining <= 0 {
			break
		}
		progress := false
		for remaining > 0 {
			idx := -1
			best := -1.0
			for i, frac := range fractions {
				if frac > best && sizes[i]-out[i] > 0 {
					best = frac
					idx = i
				}
			}
			if idx == -1 {
				break
			}
			out[idx]++
			remaining--
			progress = true
		}
		if !progress {
			break
		}
	}
	return out
}

// Bounds returns the assigned bounds for the flex container.
func (f *Flex) Bounds() Rect {
	return f.bounds
}

// ChildWidgets returns the flex container's child widgets.
func (f *Flex) ChildWidgets() []Widget {
	if len(f.Children) == 0 {
		return nil
	}
	children := make([]Widget, 0, len(f.Children))
	for _, child := range f.Children {
		if child.Widget != nil {
			children = append(children, child.Widget)
		}
	}
	return children
}

// PathSegment returns a debug path segment for the given child.
func (f *Flex) PathSegment(child Widget) string {
	if f == nil {
		return "Flex"
	}
	for i, entry := range f.Children {
		if entry.Widget == child {
			return "Flex[" + strconv.Itoa(i) + "]"
		}
	}
	return "Flex"
}

// Render draws all children.
func (f *Flex) Render(ctx RenderContext) {
	for i, child := range f.Children {
		if i >= len(f.childBounds) {
			continue
		}
		bounds := f.childBounds[i]
		if bounds.Width <= 0 || bounds.Height <= 0 {
			continue
		}
		if child.Widget == nil {
			continue
		}
		if childCtx, ok := ctx.SubVisible(bounds); ok {
			child.Widget.Render(childCtx)
		}
	}
}

// HandleMessage dispatches to children.
// Messages go to all children; first handler wins.
func (f *Flex) HandleMessage(msg Message) HandleResult {
	for _, child := range f.Children {
		result := child.Widget.HandleMessage(msg)
		if result.Handled {
			return result
		}
	}
	return Unhandled()
}

// cachedMeasure measures a child widget, using the cache when available.
// If the cache is nil or the entry misses, it falls back to a direct
// Measure call and stores the result for future lookups.
func (f *Flex) cachedMeasure(w Widget, constraints Constraints) Size {
	if f.MeasureCache == nil {
		return w.Measure(constraints)
	}
	gen := f.MeasureCache.Generation()
	if cached, ok := f.MeasureCache.Get(w, constraints, gen); ok {
		return cached
	}
	result := w.Measure(constraints)
	f.MeasureCache.Set(w, constraints, result, gen)
	return result
}

// mainSize returns the size along the main axis.
func (f *Flex) mainSize(s Size) int {
	if f.Direction == Column {
		return s.Height
	}
	return s.Width
}

// sizeWithBasis creates a size with the basis on the main axis.
func (f *Flex) sizeWithBasis(basis int) Size {
	if f.Direction == Column {
		return Size{Width: 0, Height: basis}
	}
	return Size{Width: basis, Height: 0}
}

// Spacer is a flexible empty widget for adding space in flex layouts.
type Spacer struct {
	bounds Rect
}

// NewSpacer creates a spacer widget.
func NewSpacer() *Spacer {
	return &Spacer{}
}

func (s *Spacer) Measure(constraints Constraints) Size {
	return constraints.MinSize()
}

func (s *Spacer) Layout(bounds Rect) {
	s.bounds = bounds
}

// Bounds returns the assigned bounds for the spacer.
func (s *Spacer) Bounds() Rect {
	return s.bounds
}

func (s *Spacer) Render(ctx RenderContext) {
	// Spacer is invisible
}

func (s *Spacer) HandleMessage(msg Message) HandleResult {
	return Unhandled()
}

// Space creates a flexible spacer that expands to fill available space.
func Space() FlexChild {
	return Expanded(NewSpacer())
}

// FixedSpace creates a fixed-size spacer.
func FixedSpace(size int) FlexChild {
	return Sized(NewSpacer(), size)
}
