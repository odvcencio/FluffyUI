package runtime

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
)

var (
	layoutDebugOnce sync.Once
	layoutDebugMu   sync.RWMutex

	layoutDebugEnabled bool
	layoutDebugWriter  io.Writer = os.Stderr

	layoutWarningMu   sync.Mutex
	layoutWarningSeen = map[string]struct{}{}
)

// LayoutDebugEnabled reports whether layout diagnostics are enabled.
// It is enabled by default when FLUFFYUI_LAYOUT_DEBUG is a truthy value.
func LayoutDebugEnabled() bool {
	layoutDebugOnce.Do(initLayoutDebug)
	layoutDebugMu.RLock()
	defer layoutDebugMu.RUnlock()
	return layoutDebugEnabled
}

// SetLayoutDebug enables or disables layout diagnostics at runtime.
func SetLayoutDebug(enabled bool) {
	layoutDebugOnce.Do(initLayoutDebug)
	layoutDebugMu.Lock()
	layoutDebugEnabled = enabled
	layoutDebugMu.Unlock()
}

// SetLayoutDebugWriter overrides the destination for layout diagnostics.
// Passing nil restores os.Stderr.
func SetLayoutDebugWriter(w io.Writer) {
	layoutDebugOnce.Do(initLayoutDebug)
	layoutDebugMu.Lock()
	if w == nil {
		w = os.Stderr
	}
	layoutDebugWriter = w
	layoutDebugMu.Unlock()
}

// WarnInvalidConstraints emits a debug warning when constraints are impossible
// to satisfy (min > max in either dimension).
func WarnInvalidConstraints(scope string, c Constraints) {
	if c.MinWidth <= c.MaxWidth && c.MinHeight <= c.MaxHeight {
		return
	}
	emitLayoutWarning(scope, "unsatisfiable constraints", c, Size{Width: -1, Height: -1})
}

// WarnZeroMeasure emits a debug warning for zero-size measurements when
// constraints would allow visible output.
func WarnZeroMeasure(scope string, c Constraints, measured Size) {
	WarnInvalidConstraints(scope, c)
	if measured.Width != 0 || measured.Height != 0 {
		return
	}
	if c.MaxWidth <= 0 || c.MaxHeight <= 0 {
		return
	}
	emitLayoutWarning(scope, "zero-size measurement", c, measured)
}

func initLayoutDebug() {
	layoutDebugEnabled = parseLayoutDebug(os.Getenv("FLUFFYUI_LAYOUT_DEBUG"))
}

func parseLayoutDebug(raw string) bool {
	raw = strings.TrimSpace(strings.ToLower(raw))
	switch raw {
	case "1", "true", "yes", "on", "debug":
		return true
	case "0", "false", "no", "off", "":
		return false
	default:
		enabled, err := strconv.ParseBool(raw)
		return err == nil && enabled
	}
}

func emitLayoutWarning(scope string, issue string, c Constraints, measured Size) {
	layoutDebugOnce.Do(initLayoutDebug)
	layoutDebugMu.RLock()
	enabled := layoutDebugEnabled
	writer := layoutDebugWriter
	layoutDebugMu.RUnlock()
	if !enabled {
		return
	}
	if writer == nil {
		writer = os.Stderr
	}
	if scope == "" {
		scope = "unknown"
	}

	key := fmt.Sprintf("%s|%s|%d|%d|%d|%d|%d|%d",
		scope,
		issue,
		c.MinWidth,
		c.MaxWidth,
		c.MinHeight,
		c.MaxHeight,
		measured.Width,
		measured.Height,
	)
	layoutWarningMu.Lock()
	if _, exists := layoutWarningSeen[key]; exists {
		layoutWarningMu.Unlock()
		return
	}
	layoutWarningSeen[key] = struct{}{}
	layoutWarningMu.Unlock()

	if measured.Width >= 0 && measured.Height >= 0 {
		_, _ = fmt.Fprintf(
			writer,
			"fluffyui layout debug: %s in %s (constraints min=%dx%d max=%dx%d, measured=%dx%d)\n",
			issue,
			scope,
			c.MinWidth,
			c.MinHeight,
			c.MaxWidth,
			c.MaxHeight,
			measured.Width,
			measured.Height,
		)
		return
	}

	_, _ = fmt.Fprintf(
		writer,
		"fluffyui layout debug: %s in %s (constraints min=%dx%d max=%dx%d)\n",
		issue,
		scope,
		c.MinWidth,
		c.MinHeight,
		c.MaxWidth,
		c.MaxHeight,
	)
}

func resetLayoutWarningsForTest() {
	layoutWarningMu.Lock()
	layoutWarningSeen = map[string]struct{}{}
	layoutWarningMu.Unlock()
}
