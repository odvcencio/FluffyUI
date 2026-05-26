package fluffy

import (
	"fmt"
	"time"

	"m31labs.dev/fluffyui/runtime"
	"m31labs.dev/fluffyui/style"
)

// WatchStyles parses an FSS stylesheet file and starts watching it for
// changes. When the file is modified on disk, the returned stylesheet is
// updated in place so any app or widget holding a reference to it will
// see the new rules on the next render pass.
//
// Returns the parsed stylesheet and a stop function that halts the watcher.
// The caller should pass the stylesheet to WithStylesheet when building
// an app:
//
//	sheet, stop, err := fluffy.WatchStyles("theme.fss")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer stop()
//	app, _ := fluffy.NewApp(fluffy.WithStylesheet(sheet), fluffy.WithRoot(root))
//	app.Run(ctx)
func WatchStyles(path string) (sheet *style.Stylesheet, stop func(), err error) {
	sheet, err = style.ParseFile(path)
	if err != nil {
		return nil, func() {}, fmt.Errorf("fluffy.WatchStyles: %w", err)
	}

	watcher := style.NewFileWatcher(path, sheet)
	if startErr := watcher.Start(); startErr != nil {
		return sheet, func() {}, fmt.Errorf("fluffy.WatchStyles: %w", startErr)
	}

	return sheet, watcher.Stop, nil
}

// WatchStylesForApp starts watching an FSS file and applies changes to the
// given app. Unlike WatchStyles, this function schedules stylesheet updates
// on the app's event loop so that layout and rendering are invalidated
// automatically.
//
// Returns a stop function to halt the watcher.
//
//	stop := fluffy.WatchStylesForApp(app, "theme.fss")
//	defer stop()
func WatchStylesForApp(app *runtime.App, path string) (func(), error) {
	if app == nil {
		return func() {}, fmt.Errorf("fluffy.WatchStylesForApp: app is nil")
	}

	sheet, err := style.ParseFile(path)
	if err != nil {
		return func() {}, fmt.Errorf("fluffy.WatchStylesForApp: %w", err)
	}

	// Apply the initial stylesheet.
	app.SetStylesheet(sheet)

	watcher := style.NewFileWatcher(path, sheet)
	watcher.SetOnChange(func(s *style.Stylesheet) {
		// The FileWatcher already replaced the rules in sheet in-place.
		// We call SetStylesheet to trigger a re-render on the app loop.
		if sched := app.Services().Scheduler(); sched != nil {
			sched.Schedule(func() {
				app.SetStylesheet(s)
			})
		} else {
			app.SetStylesheet(s)
		}
	})

	if startErr := watcher.Start(); startErr != nil {
		return func() {}, fmt.Errorf("fluffy.WatchStylesForApp: %w", startErr)
	}

	return watcher.Stop, nil
}

// WithHotReload returns an AppOption that watches an FSS file and
// automatically applies stylesheet changes when the file is modified.
// The watcher is started in OnReady and stopped in OnQuit.
//
//	app, _ := fluffy.NewApp(
//	    fluffy.WithRoot(root),
//	    fluffy.WithHotReload("theme.fss"),
//	)
func WithHotReload(path string) AppOption {
	return WithHotReloadInterval(path, 500*time.Millisecond)
}

// WithHotReloadInterval is like WithHotReload but with a custom polling interval.
func WithHotReloadInterval(path string, interval time.Duration) AppOption {
	return func(b *appBuilder) {
		if b == nil {
			return
		}

		// Parse the file eagerly so the initial render has the right styles.
		sheet, err := style.ParseFile(path)
		if err != nil {
			// If the file can't be parsed, fall through silently.
			// The app will use the default stylesheet.
			return
		}
		b.cfg.Stylesheet = sheet

		// Wire up start/stop on the app lifecycle.
		prevOnReady := b.cfg.OnReady
		prevOnQuit := b.cfg.OnQuit

		var stopWatcher func()

		b.cfg.OnReady = func(a *runtime.App) {
			watcher := style.NewFileWatcher(path, sheet)
			watcher.SetInterval(interval)
			watcher.SetOnChange(func(s *style.Stylesheet) {
				if sched := a.Services().Scheduler(); sched != nil {
					sched.Schedule(func() {
						a.SetStylesheet(s)
					})
				} else {
					a.SetStylesheet(s)
				}
			})
			if startErr := watcher.Start(); startErr == nil {
				stopWatcher = watcher.Stop
			}
			if prevOnReady != nil {
				prevOnReady(a)
			}
		}

		b.cfg.OnQuit = func(a *runtime.App) {
			if stopWatcher != nil {
				stopWatcher()
			}
			if prevOnQuit != nil {
				prevOnQuit(a)
			}
		}
	}
}
