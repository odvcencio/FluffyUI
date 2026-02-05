//go:build !js

package tcell

import (
	"bytes"
	"sync"

	"github.com/gdamore/tcell/v3"
)

type rawTty struct {
	inner   tcell.Tty
	writeMu sync.Mutex

	inlineMode bool
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

// NotifyResize in v3 takes a channel instead of a callback function
func (t *rawTty) NotifyResize(ch chan<- bool) {
	t.inner.NotifyResize(ch)
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

	output := p
	if t.inlineMode {
		output = stripAltScreenEscapes(output)
		if len(output) == 0 {
			return len(p), nil
		}
	}

	n, err := t.inner.Write(output)
	if err != nil {
		return n, err
	}
	if len(output) != len(p) {
		// We consumed all bytes from caller even when alternate-screen escapes
		// were filtered out.
		return len(p), nil
	}
	return n, nil
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

func (t *rawTty) SetInlineMode(enabled bool) {
	if t == nil {
		return
	}
	t.writeMu.Lock()
	t.inlineMode = enabled
	t.writeMu.Unlock()
}

var (
	enterAltScreen = []byte("\x1b[?1049h")
	exitAltScreen  = []byte("\x1b[?1049l")
)

func stripAltScreenEscapes(p []byte) []byte {
	if len(p) == 0 {
		return p
	}
	out := bytes.ReplaceAll(p, enterAltScreen, nil)
	out = bytes.ReplaceAll(out, exitAltScreen, nil)
	return out
}
