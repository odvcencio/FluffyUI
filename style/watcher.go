package style

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// FileWatcher watches an FSS file for changes and reloads the stylesheet.
// It uses polling (stat-based) for portability across platforms.
type FileWatcher struct {
	path       string
	stylesheet *Stylesheet
	interval   time.Duration // polling interval, default 500ms
	stopCh     chan struct{}
	mu         sync.Mutex
	lastMod    time.Time
	running    bool
	onChange   func(*Stylesheet) // callback when stylesheet reloads
	onError    func(error)       // callback on parse errors
}

// NewFileWatcher creates a FileWatcher for the given FSS file path.
// The watcher will update the provided stylesheet when the file changes.
// It does not start watching until Start is called.
func NewFileWatcher(path string, stylesheet *Stylesheet) *FileWatcher {
	return &FileWatcher{
		path:       path,
		stylesheet: stylesheet,
		interval:   500 * time.Millisecond,
		stopCh:     make(chan struct{}),
	}
}

// SetInterval changes the polling interval. Must be called before Start.
// Values <= 0 are ignored.
func (w *FileWatcher) SetInterval(d time.Duration) {
	if w == nil || d <= 0 {
		return
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	w.interval = d
}

// SetOnChange registers a callback invoked after the stylesheet is
// successfully reloaded from disk. The callback receives the new stylesheet.
func (w *FileWatcher) SetOnChange(fn func(*Stylesheet)) {
	if w == nil {
		return
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	w.onChange = fn
}

// SetOnError registers a callback invoked when a parse error occurs during
// reload. The old stylesheet is preserved; the error is informational.
func (w *FileWatcher) SetOnError(fn func(error)) {
	if w == nil {
		return
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	w.onError = fn
}

// Start begins watching the file in a background goroutine. It records the
// file's current modification time so that only subsequent changes trigger a
// reload. Returns an error if the watcher is already running.
func (w *FileWatcher) Start() error {
	if w == nil {
		return fmt.Errorf("style: FileWatcher is nil")
	}
	w.mu.Lock()
	if w.running {
		w.mu.Unlock()
		return fmt.Errorf("style: FileWatcher already running")
	}
	w.running = true

	// Snapshot current mod time so we only react to future changes.
	if info, err := os.Stat(w.path); err == nil {
		w.lastMod = info.ModTime()
	}

	interval := w.interval
	w.mu.Unlock()

	go w.poll(interval)
	return nil
}

// Stop stops the watcher goroutine. It is safe to call multiple times.
func (w *FileWatcher) Stop() {
	if w == nil {
		return
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	if !w.running {
		return
	}
	w.running = false
	close(w.stopCh)
}

// Reload manually reparses the FSS file and replaces the stylesheet rules.
// On parse error the old stylesheet is preserved and onError is called.
func (w *FileWatcher) Reload() error {
	if w == nil {
		return fmt.Errorf("style: FileWatcher is nil")
	}

	sheet, err := ParseFile(w.path)
	if err != nil {
		w.mu.Lock()
		onErr := w.onError
		w.mu.Unlock()
		if onErr != nil {
			onErr(err)
		}
		return err
	}

	w.replaceStylesheet(sheet)

	// Update last mod time.
	if info, statErr := os.Stat(w.path); statErr == nil {
		w.mu.Lock()
		w.lastMod = info.ModTime()
		w.mu.Unlock()
	}

	w.mu.Lock()
	onCh := w.onChange
	w.mu.Unlock()
	if onCh != nil {
		onCh(w.stylesheet)
	}

	return nil
}

// poll is the background goroutine that checks for file changes.
func (w *FileWatcher) poll(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopCh:
			return
		case <-ticker.C:
			info, err := os.Stat(w.path)
			if err != nil {
				// File may have been temporarily removed; skip this tick.
				continue
			}
			w.mu.Lock()
			lastMod := w.lastMod
			w.mu.Unlock()

			if info.ModTime().After(lastMod) {
				_ = w.Reload()
			}
		}
	}
}

// replaceStylesheet atomically replaces the rules and variables in the
// target stylesheet with those from src. This preserves the pointer
// identity of w.stylesheet so any code holding a reference to it
// sees the updated rules.
func (w *FileWatcher) replaceStylesheet(src *Stylesheet) {
	if w.stylesheet == nil || src == nil {
		return
	}

	// Read from source under its lock.
	src.mu.RLock()
	newRules := make([]Rule, len(src.rules))
	copy(newRules, src.rules)
	newVars := make(map[string]string, len(src.variables))
	for k, v := range src.variables {
		newVars[k] = v
	}
	nextOrder := src.nextOrder
	src.mu.RUnlock()

	// Write to target under its lock.
	w.stylesheet.mu.Lock()
	w.stylesheet.rules = newRules
	w.stylesheet.variables = newVars
	w.stylesheet.nextOrder = nextOrder
	w.stylesheet.indicesStale = true
	w.stylesheet.mu.Unlock()
}
