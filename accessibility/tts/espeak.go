package tts

import (
	"context"
	"fmt"
	"os/exec"
	"sync"

	"github.com/odvcencio/fluffyui/accessibility"
)

type espeakSpeaker struct {
	mu   sync.Mutex
	cmd  *exec.Cmd
	rate float64
	bin  string
}

func newEspeakSpeaker(rate float64) (accessibility.Speaker, error) {
	bin := "espeak-ng"
	if _, err := exec.LookPath(bin); err != nil {
		bin = "espeak"
		if _, err := exec.LookPath(bin); err != nil {
			return nil, fmt.Errorf("espeak-ng or espeak not found")
		}
	}
	return &espeakSpeaker{rate: rate, bin: bin}, nil
}

func (e *espeakSpeaker) Speak(ctx context.Context, text string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	_ = e.stopLocked()
	var args []string
	if e.rate > 0 && e.rate != 1.0 {
		wpm := int(175 * e.rate)
		args = append(args, "-s", fmt.Sprintf("%d", wpm))
	}
	args = append(args, text)
	e.cmd = exec.CommandContext(ctx, e.bin, args...)
	return e.cmd.Run()
}

func (e *espeakSpeaker) Stop() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.stopLocked()
}

func (e *espeakSpeaker) stopLocked() error {
	if e.cmd != nil && e.cmd.Process != nil {
		_ = e.cmd.Process.Kill()
		_, _ = e.cmd.Process.Wait()
		e.cmd = nil
	}
	return nil
}

func (e *espeakSpeaker) Close() error {
	return e.Stop()
}
