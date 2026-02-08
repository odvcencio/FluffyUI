//go:build linux

package tts

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// mockSSIPServer simulates a speech-dispatcher SSIP server on a Unix socket.
func mockSSIPServer(t *testing.T, listener net.Listener, spoken chan<- string) {
	t.Helper()
	for {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		go handleSSIPConn(t, conn, spoken)
	}
}

func handleSSIPConn(t *testing.T, conn net.Conn, spoken chan<- string) {
	t.Helper()
	defer conn.Close()
	r := bufio.NewReader(conn)
	w := bufio.NewWriter(conn)

	respond := func(code int, msg string) {
		fmt.Fprintf(w, "%d %s\r\n", code, msg)
		w.Flush()
	}

	inSpeech := false
	var speechText []string

	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return
		}
		line = strings.TrimRight(line, "\r\n")

		if inSpeech {
			if line == "." {
				text := strings.Join(speechText, "\n")
				spoken <- text
				speechText = nil
				inSpeech = false
				respond(200, "OK SPEAKING")
			} else {
				speechText = append(speechText, line)
			}
			continue
		}

		cmd := strings.ToUpper(strings.TrimSpace(line))
		switch {
		case strings.HasPrefix(cmd, "SET"):
			respond(200, "OK SET")
		case cmd == "SPEAK":
			respond(230, "OK RECEIVING DATA")
			inSpeech = true
		case strings.HasPrefix(cmd, "CANCEL"):
			respond(200, "OK CANCELED")
		case cmd == "QUIT":
			respond(200, "OK BYE")
			return
		default:
			respond(200, "OK")
		}
	}
}

func TestSSIPSpeaker(t *testing.T) {
	dir := t.TempDir()
	sockPath := filepath.Join(dir, "speechd.sock")

	listener, err := net.Listen("unix", sockPath)
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	spoken := make(chan string, 10)
	go mockSSIPServer(t, listener, spoken)

	// Override XDG_RUNTIME_DIR so ssipSocketPath() finds our test socket
	spdDir := filepath.Join(dir, "speech-dispatcher")
	os.MkdirAll(spdDir, 0o700)
	os.Symlink(sockPath, filepath.Join(spdDir, "speechd.sock"))
	t.Setenv("XDG_RUNTIME_DIR", dir)

	s, err := newSSIPSpeaker(1.0)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	if err := s.Speak(context.Background(), "hello world"); err != nil {
		t.Fatal(err)
	}

	select {
	case text := <-spoken:
		if text != "hello world" {
			t.Errorf("expected 'hello world', got %q", text)
		}
	default:
		t.Error("no text was spoken")
	}
}

func TestSSIPStop(t *testing.T) {
	dir := t.TempDir()
	sockPath := filepath.Join(dir, "speechd.sock")

	listener, err := net.Listen("unix", sockPath)
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	spoken := make(chan string, 10)
	go mockSSIPServer(t, listener, spoken)

	spdDir := filepath.Join(dir, "speech-dispatcher")
	os.MkdirAll(spdDir, 0o700)
	os.Symlink(sockPath, filepath.Join(spdDir, "speechd.sock"))
	t.Setenv("XDG_RUNTIME_DIR", dir)

	s, err := newSSIPSpeaker(1.0)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	if err := s.Stop(); err != nil {
		t.Fatal(err)
	}
}
