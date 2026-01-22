package style

import (
	"os"
	"time"
)

// WatchFile polls a stylesheet file and invokes onChange when it updates.
// The returned function stops the watcher.
func WatchFile(path string, interval time.Duration, onChange func(*Stylesheet, error)) func() {
	if interval <= 0 {
		interval = 500 * time.Millisecond
	}
	done := make(chan struct{})

	var lastMod time.Time
	if info, err := os.Stat(path); err == nil {
		lastMod = info.ModTime()
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		var lastErr error
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				info, err := os.Stat(path)
				if err != nil {
					if onChange != nil && (lastErr == nil || err.Error() != lastErr.Error()) {
						onChange(nil, err)
					}
					lastErr = err
					continue
				}
				lastErr = nil
				modTime := info.ModTime()
				if modTime.After(lastMod) {
					lastMod = modTime
					sheet, err := ParseFile(path)
					if onChange != nil {
						onChange(sheet, err)
					}
				}
			}
		}
	}()

	return func() { close(done) }
}
