//go:build linux

package tts

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"

	"github.com/odvcencio/fluffyui/accessibility"
)

type ssipSpeaker struct {
	mu   sync.Mutex
	conn net.Conn
	w    *bufio.Writer
	r    *bufio.Reader
	rate float64
}

func newSSIPSpeaker(rate float64) (accessibility.Speaker, error) {
	sockPath := ssipSocketPath()
	conn, err := net.Dial("unix", sockPath)
	if err != nil {
		return nil, fmt.Errorf("ssip connect %s: %w", sockPath, err)
	}
	s := &ssipSpeaker{
		conn: conn,
		w:    bufio.NewWriter(conn),
		r:    bufio.NewReader(conn),
		rate: rate,
	}
	if err := s.init(); err != nil {
		conn.Close()
		return nil, err
	}
	return s, nil
}

func (s *ssipSpeaker) init() error {
	if err := s.send("SET self CLIENT_NAME fluffyui:tts:main"); err != nil {
		return err
	}
	if _, err := s.readResponse(); err != nil {
		return fmt.Errorf("ssip client name: %w", err)
	}
	if s.rate > 0 && s.rate != 1.0 {
		pct := int((s.rate - 1.0) * 100)
		if err := s.send(fmt.Sprintf("SET self RATE %d", pct)); err != nil {
			return err
		}
		if _, err := s.readResponse(); err != nil {
			return fmt.Errorf("ssip set rate: %w", err)
		}
	}
	return nil
}

func (s *ssipSpeaker) Speak(ctx context.Context, text string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.send("SPEAK"); err != nil {
		return err
	}
	if _, err := s.readResponse(); err != nil {
		return fmt.Errorf("ssip speak cmd: %w", err)
	}

	// Send text lines, then end marker
	for _, line := range strings.Split(text, "\n") {
		if line == "." {
			line = ".."
		}
		if err := s.send(line); err != nil {
			return err
		}
	}
	if err := s.send("."); err != nil {
		return err
	}

	// Read response, consuming event lines (7xx)
	done := make(chan error, 1)
	go func() {
		_, err := s.readResponse()
		done <- err
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		_ = s.cancelLocked()
		return ctx.Err()
	}
}

func (s *ssipSpeaker) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.cancelLocked()
}

func (s *ssipSpeaker) cancelLocked() error {
	if err := s.send("CANCEL self"); err != nil {
		return err
	}
	_, err := s.readResponse()
	return err
}

func (s *ssipSpeaker) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_ = s.send("QUIT")
	return s.conn.Close()
}

func (s *ssipSpeaker) send(line string) error {
	_, err := fmt.Fprintf(s.w, "%s\r\n", line)
	if err != nil {
		return err
	}
	return s.w.Flush()
}

func (s *ssipSpeaker) readResponse() (string, error) {
	var last string
	for {
		line, err := s.r.ReadString('\n')
		if err != nil {
			return "", err
		}
		line = strings.TrimRight(line, "\r\n")
		// 7xx lines are event notifications — skip them
		if len(line) >= 3 && line[0] == '7' {
			continue
		}
		last = line
		// A line starting with 2xx or 3xx (with dash = continuation, space = final)
		if len(line) >= 4 && line[3] == ' ' {
			code := line[:3]
			if code[0] == '2' || code[0] == '3' {
				return last, nil
			}
			return last, fmt.Errorf("ssip error: %s", line)
		}
	}
}

func ssipSocketPath() string {
	if dir := os.Getenv("XDG_RUNTIME_DIR"); dir != "" {
		return dir + "/speech-dispatcher/speechd.sock"
	}
	return fmt.Sprintf("/run/user/%d/speech-dispatcher/speechd.sock", os.Getuid())
}
