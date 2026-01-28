package fur

import (
	"context"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/odvcencio/fluffyui/compositor"
)

// Live provides updating display without full TUI.
type Live struct {
	mu         sync.Mutex
	renderable Renderable
	console    *Console
	rate       time.Duration
	transient  bool
	stopOnce   sync.Once
	stopCh     chan struct{}
	refreshCh  chan struct{}
}

// NewLive creates a live display.
func NewLive(r Renderable) *Live {
	return &Live{
		renderable: r,
		console:    Default(),
		rate:       100 * time.Millisecond,
		stopCh:     make(chan struct{}),
		refreshCh:  make(chan struct{}, 1),
	}
}

// WithConsole sets the console for live updates.
func (l *Live) WithConsole(c *Console) *Live {
	if l == nil {
		return l
	}
	l.console = c
	return l
}

// WithRate sets the refresh rate.
func (l *Live) WithRate(d time.Duration) *Live {
	if l == nil {
		return l
	}
	if d > 0 {
		l.rate = d
	}
	return l
}

// WithTransient clears output on stop.
func (l *Live) WithTransient(t bool) *Live {
	if l == nil {
		return l
	}
	l.transient = t
	return l
}

// Start begins live display. Blocks until Stop or context done.
func (l *Live) Start(ctx context.Context) error {
	if l == nil {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	ticker := time.NewTicker(l.rate)
	defer ticker.Stop()
	linesRendered := 0
	for {
		l.draw(&linesRendered)
		select {
		case <-ctx.Done():
			if l.transient {
				l.clear(linesRendered)
			}
			return ctx.Err()
		case <-l.stopCh:
			if l.transient {
				l.clear(linesRendered)
			}
			return nil
		case <-ticker.C:
			continue
		case <-l.refreshCh:
			continue
		}
	}
}

// Stop ends live display.
func (l *Live) Stop() {
	if l == nil {
		return
	}
	l.stopOnce.Do(func() {
		close(l.stopCh)
	})
}

// Update replaces the renderable.
func (l *Live) Update(r Renderable) {
	if l == nil {
		return
	}
	l.mu.Lock()
	l.renderable = r
	l.mu.Unlock()
	l.signalRefresh()
}

// Refresh forces immediate redraw.
func (l *Live) Refresh() {
	if l == nil {
		return
	}
	l.signalRefresh()
}

func (l *Live) signalRefresh() {
	select {
	case l.refreshCh <- struct{}{}:
	default:
	}
}

func (l *Live) currentRenderable() Renderable {
	l.mu.Lock()
	r := l.renderable
	l.mu.Unlock()
	return r
}

func (l *Live) draw(linesRendered *int) {
	c := l.console
	if c == nil {
		c = Default()
	}
	r := l.currentRenderable()
	if r == nil {
		return
	}
	lines := r.Render(c.Width())
	maxLines := len(lines)
	if *linesRendered > maxLines {
		maxLines = *linesRendered
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	var buf strings.Builder
	if *linesRendered > 0 {
		buf.WriteString(compositor.CursorUp(*linesRendered))
	}
	buf.WriteString(renderLinesToANSI(lines, maxLines, c.noColor))
	_, _ = io.WriteString(c.out, buf.String())
	*linesRendered = len(lines)
}

func (l *Live) clear(linesRendered int) {
	if linesRendered <= 0 {
		return
	}
	c := l.console
	if c == nil {
		c = Default()
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	var buf strings.Builder
	buf.WriteString(compositor.CursorUp(linesRendered))
	for i := 0; i < linesRendered; i++ {
		buf.WriteString(compositor.ANSIClearLine)
		if i < linesRendered-1 {
			buf.WriteByte('\n')
		}
	}
	if linesRendered > 1 {
		buf.WriteString(compositor.CursorUp(linesRendered - 1))
	}
	_, _ = io.WriteString(c.out, buf.String())
}

func renderLinesToANSI(lines []Line, totalLines int, noColor bool) string {
	var buf strings.Builder
	current := DefaultStyle()
	styleSet := false
	for i := 0; i < totalLines; i++ {
		buf.WriteString(compositor.ANSIClearLine)
		if i < len(lines) {
			for _, span := range lines[i] {
				if noColor {
					buf.WriteString(span.Text)
					continue
				}
				if !styleSet || !current.Equal(span.Style) {
					buf.WriteString(compositor.StyleToANSI(span.Style.ToCompositor()))
					current = span.Style
					styleSet = true
				}
				buf.WriteString(span.Text)
			}
		}
		if !noColor && styleSet {
			buf.WriteString(compositor.ANSIReset)
			current = DefaultStyle()
			styleSet = false
		}
		if i < totalLines-1 {
			buf.WriteByte('\n')
		}
	}
	return buf.String()
}
