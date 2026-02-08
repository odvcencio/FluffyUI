package runtime

// MeasureCache stores cached measurement results for widgets, keyed by
// widget identity and constraints. Entries are accepted from the current
// generation or the immediately preceding one, allowing measurements to
// persist across a single relayout pass without growing stale.
type MeasureCache struct {
	entries    map[Widget]measureEntry
	generation int64
}

// measureEntry holds a single cached measurement result.
type measureEntry struct {
	constraints Constraints
	result      Size
	generation  int64
}

// NewMeasureCache creates a new empty measure cache starting at generation 1.
func NewMeasureCache() *MeasureCache {
	return &MeasureCache{
		entries:    make(map[Widget]measureEntry),
		generation: 1,
	}
}

// Get returns the cached measurement if constraints match and the entry's
// generation is current (gen) or immediately prior (gen-1).
func (c *MeasureCache) Get(w Widget, constraints Constraints, gen int64) (Size, bool) {
	if c == nil {
		return Size{}, false
	}
	entry, ok := c.entries[w]
	if !ok {
		return Size{}, false
	}
	if entry.constraints != constraints {
		return Size{}, false
	}
	if entry.generation < gen-1 {
		return Size{}, false
	}
	return entry.result, true
}

// Set stores a measurement result for a widget with the given constraints
// and generation.
func (c *MeasureCache) Set(w Widget, constraints Constraints, result Size, gen int64) {
	if c == nil {
		return
	}
	c.entries[w] = measureEntry{
		constraints: constraints,
		result:      result,
		generation:  gen,
	}
}

// Invalidate removes a specific widget's cached measurement.
func (c *MeasureCache) Invalidate(w Widget) {
	if c == nil {
		return
	}
	delete(c.entries, w)
}

// BumpGeneration increments the generation counter and returns the new value.
func (c *MeasureCache) BumpGeneration() int64 {
	if c == nil {
		return 0
	}
	c.generation++
	return c.generation
}

// Generation returns the current generation counter.
func (c *MeasureCache) Generation() int64 {
	if c == nil {
		return 0
	}
	return c.generation
}

// Clear removes all cached entries.
func (c *MeasureCache) Clear() {
	if c == nil {
		return
	}
	c.entries = make(map[Widget]measureEntry)
}

// Len returns the number of cached entries (used for testing/diagnostics).
func (c *MeasureCache) Len() int {
	if c == nil {
		return 0
	}
	return len(c.entries)
}

// injectMeasureCache walks the widget tree rooted at w and sets the
// MeasureCache field on any Flex containers it finds.
func injectMeasureCache(w Widget, cache *MeasureCache) {
	if w == nil {
		return
	}
	if f, ok := w.(*Flex); ok {
		f.MeasureCache = cache
	}
	if cp, ok := w.(ChildProvider); ok {
		for _, child := range cp.ChildWidgets() {
			injectMeasureCache(child, cache)
		}
	}
}
