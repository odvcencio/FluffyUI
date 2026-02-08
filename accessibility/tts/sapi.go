package tts

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"

	"github.com/odvcencio/fluffyui/accessibility"
)

// sapiSpeaker uses a persistent PowerShell process with System.Speech.
// The process is started once and reused for all utterances, avoiding the
// ~1.4s startup overhead per speech call.
type sapiSpeaker struct {
	mu     sync.Mutex
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	reader *bufio.Reader
	rate   float64
	bin    string // "powershell" or "powershell.exe"
	gen    uint64 // process generation — prevents stale goroutines from killing new processes
}

func newSAPISpeaker(rate float64) (accessibility.Speaker, error) {
	for _, name := range []string{"powershell", "powershell.exe"} {
		if _, err := exec.LookPath(name); err == nil {
			return &sapiSpeaker{rate: rate, bin: name}, nil
		}
	}
	return nil, fmt.Errorf("powershell not found for SAPI")
}

func (s *sapiSpeaker) ensureProcess() error {
	if s.cmd != nil {
		return nil
	}
	rateVal := 0
	if s.rate > 0 && s.rate != 1.0 {
		rateVal = int((s.rate - 1.0) * 5)
	}
	// Persistent PowerShell script: loads System.Speech once, reads lines from stdin,
	// speaks each synchronously, writes "OK" after each completes.
	script := fmt.Sprintf(
		`Add-Type -AssemblyName System.Speech;`+
			`$synth = New-Object System.Speech.Synthesis.SpeechSynthesizer;`+
			`$synth.Rate = %d;`+
			`Write-Output "READY";`+
			`while ($line = [Console]::In.ReadLine()) {`+
			`  if ($line -eq "QUIT") { break };`+
			`  $synth.Speak($line);`+
			`  Write-Output "OK"}`+
			`$synth.Dispose()`,
		rateVal,
	)
	s.cmd = exec.Command(s.bin, "-NoProfile", "-Command", script)
	var err error
	s.stdin, err = s.cmd.StdinPipe()
	if err != nil {
		s.cmd = nil
		return fmt.Errorf("sapi stdin pipe: %w", err)
	}
	stdout, err := s.cmd.StdoutPipe()
	if err != nil {
		s.stdin.Close()
		s.cmd = nil
		return fmt.Errorf("sapi stdout pipe: %w", err)
	}
	s.reader = bufio.NewReader(stdout)
	if err := s.cmd.Start(); err != nil {
		s.cmd = nil
		return fmt.Errorf("sapi start: %w", err)
	}
	// Wait for "READY" signal (System.Speech loaded).
	line, err := s.reader.ReadString('\n')
	if err != nil {
		s.killLocked()
		return fmt.Errorf("sapi ready: %w", err)
	}
	if !strings.HasPrefix(strings.TrimSpace(line), "READY") {
		s.killLocked()
		return fmt.Errorf("sapi: unexpected startup response: %s", line)
	}
	s.gen++
	return nil
}

func (s *sapiSpeaker) Speak(ctx context.Context, text string) error {
	s.mu.Lock()
	if err := s.ensureProcess(); err != nil {
		s.mu.Unlock()
		return err
	}
	gen := s.gen
	stdin := s.stdin
	reader := s.reader
	s.mu.Unlock()

	// Write text to the persistent process.
	escaped := strings.ReplaceAll(text, "\n", " ")
	if _, err := fmt.Fprintln(stdin, escaped); err != nil {
		s.mu.Lock()
		if s.gen == gen {
			s.killLocked()
		}
		s.mu.Unlock()
		return err
	}

	// Wait for "OK" response (blocks until speech completes).
	done := make(chan error, 1)
	go func() {
		_, err := reader.ReadString('\n')
		done <- err
	}()

	select {
	case err := <-done:
		if err != nil {
			s.mu.Lock()
			if s.gen == gen {
				s.killLocked()
			}
			s.mu.Unlock()
		}
		return err
	case <-ctx.Done():
		// Only kill the process if it's still the same generation.
		// A newer generation means a new process was started and we
		// must not kill it.
		s.mu.Lock()
		if s.gen == gen {
			s.killLocked()
		}
		s.mu.Unlock()
		return ctx.Err()
	}
}

func (s *sapiSpeaker) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.killLocked()
}

func (s *sapiSpeaker) killLocked() error {
	if s.cmd != nil && s.cmd.Process != nil {
		_ = s.cmd.Process.Kill()
		_, _ = s.cmd.Process.Wait()
	}
	s.cmd = nil
	s.stdin = nil
	s.reader = nil
	return nil
}

func (s *sapiSpeaker) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.stdin != nil {
		// Graceful shutdown.
		fmt.Fprintln(s.stdin, "QUIT")
	}
	return s.killLocked()
}
