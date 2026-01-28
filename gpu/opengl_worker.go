//go:build !js

package gpu

import (
	"runtime"
)

type glWorker struct {
	tasks chan func()
	done  chan struct{}
}

func newGLWorker() *glWorker {
	w := &glWorker{
		tasks: make(chan func()),
		done:  make(chan struct{}),
	}
	go func() {
		runtime.LockOSThread()
		for fn := range w.tasks {
			if fn != nil {
				fn()
			}
		}
		close(w.done)
	}()
	return w
}

func (w *glWorker) run(fn func()) {
	if w == nil {
		if fn != nil {
			fn()
		}
		return
	}
	done := make(chan struct{})
	w.tasks <- func() {
		if fn != nil {
			fn()
		}
		close(done)
	}
	<-done
}

func (w *glWorker) stop() {
	if w == nil {
		return
	}
	close(w.tasks)
	<-w.done
}
