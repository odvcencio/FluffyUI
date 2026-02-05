//go:build linux && !js

package tcell

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v3"
	"golang.org/x/sys/unix"
)

func TestRawTTYInlineModeStripsAltScreenPTY(t *testing.T) {
	master, slave := openLinuxPTY(t)
	defer master.Close()
	defer slave.Close()

	raw := &rawTty{inner: &ptyTty{file: master}}
	raw.SetInlineMode(true)

	payload := []byte("\x1b[?1049hhello-inline\x1b[?1049l\n")
	n, err := raw.Write(payload)
	if err != nil {
		t.Fatalf("raw.Write failed: %v", err)
	}
	if n != len(payload) {
		t.Fatalf("raw.Write wrote %d bytes, want %d", n, len(payload))
	}

	out, err := readPTYOutput(slave, time.Second)
	if err != nil {
		t.Fatalf("failed reading PTY output: %v", err)
	}

	if strings.Contains(out, "\x1b[?1049h") || strings.Contains(out, "\x1b[?1049l") {
		t.Fatalf("alternate-screen escape leaked through PTY: %q", out)
	}
	if !strings.Contains(out, "hello-inline") {
		t.Fatalf("expected payload body in PTY output, got %q", out)
	}
}

func openLinuxPTY(t *testing.T) (*os.File, *os.File) {
	t.Helper()

	master, err := os.OpenFile("/dev/ptmx", os.O_RDWR|unix.O_NOCTTY, 0)
	if err != nil {
		t.Skipf("PTY unavailable (/dev/ptmx open failed): %v", err)
	}

	fd := int(master.Fd())
	if err := unix.IoctlSetPointerInt(fd, unix.TIOCSPTLCK, 0); err != nil {
		_ = master.Close()
		t.Skipf("PTY unavailable (unlock failed): %v", err)
	}

	ptyNum, err := unix.IoctlGetInt(fd, unix.TIOCGPTN)
	if err != nil {
		_ = master.Close()
		t.Skipf("PTY unavailable (query number failed): %v", err)
	}

	slavePath := fmt.Sprintf("/dev/pts/%d", ptyNum)
	slave, err := os.OpenFile(slavePath, os.O_RDWR|unix.O_NOCTTY, 0)
	if err != nil {
		_ = master.Close()
		t.Skipf("PTY unavailable (open slave failed: %s): %v", slavePath, err)
	}

	return master, slave
}

func readPTYOutput(slave *os.File, timeout time.Duration) (string, error) {
	if slave == nil {
		return "", fmt.Errorf("nil PTY slave")
	}
	pollTimeout := int(timeout / time.Millisecond)
	if pollTimeout <= 0 {
		pollTimeout = 1
	}

	fds := []unix.PollFd{
		{
			Fd:     int32(slave.Fd()),
			Events: unix.POLLIN,
		},
	}
	n, err := unix.Poll(fds, pollTimeout)
	if err != nil {
		return "", err
	}
	if n == 0 {
		return "", fmt.Errorf("timed out waiting for PTY output")
	}

	buf := make([]byte, 512)
	count, err := slave.Read(buf)
	if err != nil {
		return "", err
	}
	if count == 0 {
		return "", fmt.Errorf("PTY returned no data")
	}
	return string(buf[:count]), nil
}

type ptyTty struct {
	file *os.File
}

func (t *ptyTty) Start() error {
	return nil
}

func (t *ptyTty) Stop() error {
	return nil
}

func (t *ptyTty) Drain() error {
	return nil
}

func (t *ptyTty) NotifyResize(ch chan<- bool) {}

func (t *ptyTty) WindowSize() (tcell.WindowSize, error) {
	if t == nil || t.file == nil {
		return tcell.WindowSize{}, nil
	}
	ws, err := unix.IoctlGetWinsize(int(t.file.Fd()), unix.TIOCGWINSZ)
	if err != nil {
		return tcell.WindowSize{}, err
	}
	return tcell.WindowSize{
		Width:       int(ws.Col),
		Height:      int(ws.Row),
		PixelWidth:  int(ws.Xpixel),
		PixelHeight: int(ws.Ypixel),
	}, nil
}

func (t *ptyTty) Read(p []byte) (int, error) {
	if t == nil || t.file == nil {
		return 0, os.ErrInvalid
	}
	return t.file.Read(p)
}

func (t *ptyTty) Write(p []byte) (int, error) {
	if t == nil || t.file == nil {
		return 0, os.ErrInvalid
	}
	return t.file.Write(p)
}

func (t *ptyTty) Close() error {
	if t == nil || t.file == nil {
		return nil
	}
	return t.file.Close()
}

var _ tcell.Tty = (*ptyTty)(nil)
