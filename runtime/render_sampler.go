package runtime

import (
	"sync"
	"time"
)

// RenderSampler collects recent render stats for quick profiling summaries.
type RenderSampler struct {
	mu      sync.Mutex
	window  int
	frames  int64
	samples []RenderStats
}

// RenderSummary aggregates a window of render samples.
type RenderSummary struct {
	Frames        int64
	Samples       int
	Window        int
	Last          RenderStats
	AvgTotal      time.Duration
	AvgRender     time.Duration
	AvgFlush      time.Duration
	AvgDirtyRatio float64
	MaxTotal      time.Duration
	MaxRender     time.Duration
	MaxFlush      time.Duration
}

// NewRenderSampler creates a sampler retaining the last N samples.
func NewRenderSampler(window int) *RenderSampler {
	if window <= 0 {
		window = 120
	}
	return &RenderSampler{window: window}
}

// ObserveRender records a render sample.
func (r *RenderSampler) ObserveRender(stats RenderStats) {
	if r == nil {
		return
	}
	r.mu.Lock()
	r.frames++
	r.samples = append(r.samples, stats)
	if r.window > 0 && len(r.samples) > r.window {
		r.samples = r.samples[len(r.samples)-r.window:]
	}
	r.mu.Unlock()
}

// Summary returns aggregate stats for the current sample window.
func (r *RenderSampler) Summary() RenderSummary {
	if r == nil {
		return RenderSummary{}
	}
	r.mu.Lock()
	samples := make([]RenderStats, len(r.samples))
	copy(samples, r.samples)
	frames := r.frames
	window := r.window
	r.mu.Unlock()

	if len(samples) == 0 {
		return RenderSummary{Frames: frames, Samples: 0, Window: window}
	}

	var totalTotal time.Duration
	var totalRender time.Duration
	var totalFlush time.Duration
	var totalRatio float64
	var ratioCount int
	var maxTotal time.Duration
	var maxRender time.Duration
	var maxFlush time.Duration
	for _, sample := range samples {
		totalTotal += sample.TotalDuration
		totalRender += sample.RenderDuration
		totalFlush += sample.FlushDuration
		if sample.TotalCells > 0 {
			totalRatio += float64(sample.DirtyCells) / float64(sample.TotalCells)
			ratioCount++
		}
		if sample.TotalDuration > maxTotal {
			maxTotal = sample.TotalDuration
		}
		if sample.RenderDuration > maxRender {
			maxRender = sample.RenderDuration
		}
		if sample.FlushDuration > maxFlush {
			maxFlush = sample.FlushDuration
		}
	}

	avgDirty := 0.0
	if ratioCount > 0 {
		avgDirty = totalRatio / float64(ratioCount)
	}

	count := time.Duration(len(samples))
	return RenderSummary{
		Frames:        frames,
		Samples:       len(samples),
		Window:        window,
		Last:          samples[len(samples)-1],
		AvgTotal:      totalTotal / count,
		AvgRender:     totalRender / count,
		AvgFlush:      totalFlush / count,
		AvgDirtyRatio: avgDirty,
		MaxTotal:      maxTotal,
		MaxRender:     maxRender,
		MaxFlush:      maxFlush,
	}
}
