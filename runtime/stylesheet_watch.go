package runtime

import (
	"time"

	"m31labs.dev/fluffyui/style"
)

// WatchStylesheetFile watches a stylesheet file and applies updates on the app loop.
// It returns a stop function to halt polling.
func WatchStylesheetFile(app *App, path string, interval time.Duration) func() {
	if app == nil {
		return func() {}
	}
	scheduler := app.Services().Scheduler()
	return style.WatchFile(path, interval, func(sheet *style.Stylesheet, err error) {
		if err != nil || sheet == nil {
			return
		}
		if scheduler != nil {
			scheduler.Schedule(func() {
				app.SetStylesheet(sheet)
			})
			return
		}
		app.SetStylesheet(sheet)
	})
}
