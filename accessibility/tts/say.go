//go:build darwin

package tts

import (
	"context"
	"fmt"
	"os/exec"
	"sync"

	"m31labs.dev/fluffyui/accessibility"
)

type saySpeaker struct {
	mu   sync.Mutex
	cmd  *exec.Cmd
	rate float64
}

func newSaySpeaker(rate float64) (accessibility.Speaker, error) {
	if _, err := exec.LookPath("say"); err != nil {
		return nil, fmt.Errorf("macOS say not found")
	}
	return &saySpeaker{rate: rate}, nil
}

func (s *saySpeaker) Speak(ctx context.Context, text string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_ = s.stopLocked()
	var args []string
	if s.rate > 0 && s.rate != 1.0 {
		wpm := int(200 * s.rate)
		args = append(args, "-r", fmt.Sprintf("%d", wpm))
	}
	args = append(args, text)
	s.cmd = exec.CommandContext(ctx, "say", args...)
	return s.cmd.Run()
}

func (s *saySpeaker) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.stopLocked()
}

func (s *saySpeaker) stopLocked() error {
	if s.cmd != nil && s.cmd.Process != nil {
		_ = s.cmd.Process.Kill()
		_, _ = s.cmd.Process.Wait()
		s.cmd = nil
	}
	return nil
}

func (s *saySpeaker) Close() error {
	return s.Stop()
}
