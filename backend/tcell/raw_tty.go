package tcell

import (
	"sync"

	"github.com/gdamore/tcell/v2"
)

type rawTty struct {
	inner   tcell.Tty
	writeMu sync.Mutex
}

func (t *rawTty) Start() error {
	return t.inner.Start()
}

func (t *rawTty) Stop() error {
	return t.inner.Stop()
}

func (t *rawTty) Drain() error {
	return t.inner.Drain()
}

func (t *rawTty) NotifyResize(cb func()) {
	t.inner.NotifyResize(cb)
}

func (t *rawTty) WindowSize() (tcell.WindowSize, error) {
	return t.inner.WindowSize()
}

func (t *rawTty) Read(p []byte) (int, error) {
	return t.inner.Read(p)
}

func (t *rawTty) Write(p []byte) (int, error) {
	t.writeMu.Lock()
	defer t.writeMu.Unlock()
	return t.inner.Write(p)
}

func (t *rawTty) Close() error {
	return t.inner.Close()
}

func (t *rawTty) WriteRaw(p []byte) error {
	if t == nil {
		return nil
	}
	t.writeMu.Lock()
	defer t.writeMu.Unlock()
	_, err := t.inner.Write(p)
	return err
}
