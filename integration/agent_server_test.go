package integration

import (
	"bufio"
	"context"
	"encoding/json"
	"net"
	"path/filepath"
	"testing"
	"time"

	"github.com/odvcencio/fluffyui/agent"
	"github.com/odvcencio/fluffyui/backend/sim"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/widgets"
)

type serverRequest struct {
	ID          int    `json:"id,omitempty"`
	Type        string `json:"type"`
	Token       string `json:"token,omitempty"`
	Key         string `json:"key,omitempty"`
	Text        string `json:"text,omitempty"`
	Width       int    `json:"width,omitempty"`
	Height      int    `json:"height,omitempty"`
	IncludeText bool   `json:"include_text,omitempty"`
}

type serverResponse struct {
	ID           int                 `json:"id,omitempty"`
	OK           bool                `json:"ok,omitempty"`
	Error        string              `json:"error,omitempty"`
	Message      string              `json:"message,omitempty"`
	Snapshot     *agent.Snapshot     `json:"snapshot,omitempty"`
	Capabilities *agent.Capabilities `json:"capabilities,omitempty"`
}

func TestAgentServerJSONLIntegration(t *testing.T) {
	be := sim.New(40, 10)
	app := runtime.NewApp(runtime.AppConfig{
		Backend:           be,
		Root:              widgets.NewLabel("Hello"),
		Update:            runtime.DefaultUpdate,
		FocusRegistration: runtime.FocusRegistrationAuto,
		TickRate:          time.Second / 60,
	})
	runApp(t, app)

	socketPath := filepath.Join(t.TempDir(), "agent.sock")
	srv, err := agent.NewServer(agent.ServerOptions{
		Addr:      "unix:" + socketPath,
		App:       app,
		Token:     "secret",
		AllowText: false,
	})
	if err != nil {
		t.Fatalf("new server: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- srv.Serve(ctx)
	}()
	t.Cleanup(func() {
		cancel()
		_ = srv.Close()
		select {
		case <-done:
		case <-time.After(appTimeout):
			t.Fatalf("timeout waiting for server shutdown")
		}
	})

	waitForSocket(t, socketPath)

	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)
	send := func(t *testing.T, req serverRequest) serverResponse {
		t.Helper()
		payload, err := encodeJSONLine(req)
		if err != nil {
			t.Fatalf("encode: %v", err)
		}
		if err := conn.SetDeadline(time.Now().Add(appTimeout)); err != nil {
			t.Fatalf("deadline: %v", err)
		}
		if _, err := conn.Write(payload); err != nil {
			t.Fatalf("write: %v", err)
		}
		line, err := reader.ReadBytes('\n')
		if err != nil {
			t.Fatalf("read: %v", err)
		}
		var resp serverResponse
		if err := json.Unmarshal(line, &resp); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		return resp
	}

	resp := send(t, serverRequest{ID: 1, Type: "ping"})
	if resp.Error != "unauthorized" {
		t.Fatalf("expected unauthorized, got %+v", resp)
	}

	resp = send(t, serverRequest{ID: 2, Type: "hello", Token: "secret"})
	if !resp.OK || resp.Capabilities == nil {
		t.Fatalf("expected capabilities, got %+v", resp)
	}

	resp = send(t, serverRequest{ID: 3, Type: "snapshot", IncludeText: true})
	if resp.Error != "text_disabled" {
		t.Fatalf("expected text_disabled, got %+v", resp)
	}

	resp = send(t, serverRequest{ID: 4, Type: "snapshot"})
	if !resp.OK || resp.Snapshot == nil {
		t.Fatalf("expected snapshot, got %+v", resp)
	}

	resp = send(t, serverRequest{ID: 5, Type: "key", Key: "enter"})
	if !resp.OK {
		t.Fatalf("expected key ok, got %+v", resp)
	}

	resp = send(t, serverRequest{ID: 6, Type: "text", Text: "hi"})
	if !resp.OK {
		t.Fatalf("expected text ok, got %+v", resp)
	}

	resp = send(t, serverRequest{ID: 7, Type: "resize", Width: 50, Height: 15})
	if !resp.OK {
		t.Fatalf("expected resize ok, got %+v", resp)
	}
}
