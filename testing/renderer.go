//go:build !js

package testing

import (
	"context"
	"errors"
	"sync"
	"time"

	"m31labs.dev/fluffyui/backend"
	"m31labs.dev/fluffyui/backend/sim"
	"m31labs.dev/fluffyui/runtime"
)

const defaultRendererStartTimeout = time.Second

// TestFrame is an immutable snapshot captured after a completed render pass.
type TestFrame struct {
	Stats  runtime.RenderStats
	Screen string
}

// TestRenderer runs an app against the simulation backend and exposes
// deterministic frame, capture, and visual-idle helpers for tests.
type TestRenderer struct {
	app     *runtime.App
	backend *sim.Backend
	sampler *runtime.RenderSampler
	cancel  context.CancelFunc
	done    chan struct{}
	frameCh chan struct{}

	mu     sync.RWMutex
	latest TestFrame
	runErr error
	close  sync.Once
}

// NewTestRenderer starts an app and waits for its initial frame.
func NewTestRenderer(ctx context.Context, root runtime.Widget, width, height int) (*TestRenderer, error) {
	return NewTestRendererWithConfig(ctx, runtime.AppConfig{Root: root}, width, height)
}

// NewTestRendererWithConfig starts an app with cfg against a simulation
// backend. Backend is replaced, while an existing render observer is preserved.
func NewTestRendererWithConfig(ctx context.Context, cfg runtime.AppConfig, width, height int) (*TestRenderer, error) {
	if width <= 0 || height <= 0 {
		return nil, errors.New("renderer dimensions must be positive")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	runCtx, cancel := context.WithCancel(ctx)
	renderer := &TestRenderer{
		backend: sim.New(width, height),
		sampler: runtime.NewRenderSampler(120),
		cancel:  cancel,
		done:    make(chan struct{}),
		frameCh: make(chan struct{}, 1),
	}
	observer := cfg.RenderObserver
	cfg.Backend = renderer.backend
	cfg.RenderObserver = runtime.RenderObserverFunc(func(stats runtime.RenderStats) {
		renderer.observe(stats)
		if observer != nil {
			observer.ObserveRender(stats)
		}
	})
	renderer.app = runtime.NewApp(cfg)

	go func() {
		err := renderer.app.Run(runCtx)
		renderer.mu.Lock()
		renderer.runErr = err
		renderer.mu.Unlock()
		close(renderer.done)
	}()

	startCtx, startCancel := context.WithTimeout(ctx, defaultRendererStartTimeout)
	defer startCancel()
	if _, err := renderer.waitAfter(startCtx, 0); err != nil {
		_ = renderer.Close()
		return nil, err
	}
	return renderer, nil
}

// App returns the running app for posting messages or invoking app-loop work.
func (r *TestRenderer) App() *runtime.App {
	if r == nil {
		return nil
	}
	return r.app
}

// Backend returns the simulation backend for input injection and cell checks.
func (r *TestRenderer) Backend() *sim.Backend {
	if r == nil {
		return nil
	}
	return r.backend
}

// LatestFrame returns the most recently completed frame.
func (r *TestRenderer) LatestFrame() TestFrame {
	if r == nil {
		return TestFrame{}
	}
	r.mu.RLock()
	frame := r.latest
	r.mu.RUnlock()
	return frame
}

// WaitForFrame waits for a render newer than the current frame.
func (r *TestRenderer) WaitForFrame(timeout time.Duration) (TestFrame, error) {
	if r == nil {
		return TestFrame{}, errors.New("test renderer is nil")
	}
	return r.WaitForFrameAfter(r.LatestFrame().Stats.Frame, timeout)
}

// WaitForFrameAfter waits for a render newer than frame. Capture the current
// frame number before triggering asynchronous work to avoid missing a fast
// render between posting the work and beginning the wait.
func (r *TestRenderer) WaitForFrameAfter(frame int64, timeout time.Duration) (TestFrame, error) {
	if r == nil {
		return TestFrame{}, errors.New("test renderer is nil")
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return r.waitAfter(ctx, frame)
}

// Flush requests a render and waits for its completed frame.
func (r *TestRenderer) Flush(timeout time.Duration) (TestFrame, error) {
	if r == nil || r.app == nil {
		return TestFrame{}, errors.New("test renderer is not initialized")
	}
	baseline := r.LatestFrame().Stats.Frame
	r.app.Invalidate()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return r.waitAfter(ctx, baseline)
}

// WaitForVisualIdle waits until no new frames arrive for idleFor.
func (r *TestRenderer) WaitForVisualIdle(idleFor, timeout time.Duration) (TestFrame, error) {
	if r == nil {
		return TestFrame{}, errors.New("test renderer is nil")
	}
	if idleFor <= 0 {
		return r.LatestFrame(), nil
	}

	timeoutTimer := time.NewTimer(timeout)
	defer timeoutTimer.Stop()
	idleTimer := time.NewTimer(idleFor)
	defer idleTimer.Stop()
	seen := r.LatestFrame().Stats.Frame

	for {
		select {
		case <-r.frameCh:
			latest := r.LatestFrame().Stats.Frame
			if latest <= seen {
				continue
			}
			seen = latest
			if !idleTimer.Stop() {
				select {
				case <-idleTimer.C:
				default:
				}
			}
			idleTimer.Reset(idleFor)
		case <-idleTimer.C:
			latest := r.LatestFrame()
			if latest.Stats.Frame > seen {
				seen = latest.Stats.Frame
				idleTimer.Reset(idleFor)
				continue
			}
			return latest, nil
		case <-timeoutTimer.C:
			return TestFrame{}, ErrTimeout
		case <-r.done:
			return TestFrame{}, r.stoppedError()
		}
	}
}

// Capture returns the current simulated screen.
func (r *TestRenderer) Capture() string {
	if r == nil || r.backend == nil {
		return ""
	}
	return r.backend.Capture()
}

// CaptureRegion returns a rectangular region of the simulated screen.
func (r *TestRenderer) CaptureRegion(x, y, width, height int) string {
	if r == nil || r.backend == nil {
		return ""
	}
	return r.backend.CaptureRegion(x, y, width, height)
}

// CaptureCell returns the content and style at one screen cell.
func (r *TestRenderer) CaptureCell(x, y int) (rune, []rune, backend.Style) {
	if r == nil || r.backend == nil {
		return ' ', nil, backend.DefaultStyle()
	}
	return r.backend.CaptureCell(x, y)
}

// Stats returns the render summary collected by this renderer.
func (r *TestRenderer) Stats() runtime.RenderSummary {
	if r == nil || r.sampler == nil {
		return runtime.RenderSummary{}
	}
	return r.sampler.Summary()
}

// Close stops the app and waits for the renderer to release its backend.
func (r *TestRenderer) Close() error {
	if r == nil {
		return nil
	}
	r.close.Do(func() {
		r.cancel()
		<-r.done
	})
	r.mu.RLock()
	err := r.runErr
	r.mu.RUnlock()
	if errors.Is(err, context.Canceled) {
		return nil
	}
	return err
}

func (r *TestRenderer) observe(stats runtime.RenderStats) {
	r.sampler.ObserveRender(stats)
	frame := TestFrame{Stats: stats, Screen: r.backend.Capture()}
	r.mu.Lock()
	r.latest = frame
	r.mu.Unlock()
	select {
	case r.frameCh <- struct{}{}:
	default:
	}
}

func (r *TestRenderer) waitAfter(ctx context.Context, baseline int64) (TestFrame, error) {
	for {
		latest := r.LatestFrame()
		if latest.Stats.Frame > baseline {
			return latest, nil
		}
		select {
		case <-r.frameCh:
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				return TestFrame{}, ErrTimeout
			}
			return TestFrame{}, ctx.Err()
		case <-r.done:
			return TestFrame{}, r.stoppedError()
		}
	}
}

func (r *TestRenderer) stoppedError() error {
	r.mu.RLock()
	err := r.runErr
	r.mu.RUnlock()
	if err == nil {
		return errors.New("test renderer stopped")
	}
	return err
}
