package runtime

import "time"

// RenderStats captures timing and dirty-region data for a render pass.
type RenderStats struct {
	Frame          int64
	Started        time.Time
	Ended          time.Time
	TotalDuration  time.Duration
	RenderDuration time.Duration
	FlushDuration  time.Duration
	DirtyCells     int
	FlushedCells   int
	TotalCells     int
	FullRedraw     bool
	DirtyRect      Rect
	LayerCount     int
}

// RenderObserver receives render timing and dirty stats.
type RenderObserver interface {
	ObserveRender(stats RenderStats)
}

// RenderObserverFunc adapts a function into a RenderObserver.
type RenderObserverFunc func(stats RenderStats)

// ObserveRender invokes the wrapped function.
func (f RenderObserverFunc) ObserveRender(stats RenderStats) {
	if f != nil {
		f(stats)
	}
}
