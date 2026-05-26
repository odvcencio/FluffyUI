package agent

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"m31labs.dev/fluffyui/backend/sim"
	"m31labs.dev/fluffyui/runtime"
	"m31labs.dev/fluffyui/widgets"
)

func TestServerServeAndHandleRequests(t *testing.T) {
	be := sim.New(40, 10)
	app := runtime.NewApp(runtime.AppConfig{
		Backend:           be,
		Root:              widgets.NewLabel("Hello"),
		Update:            runtime.DefaultUpdate,
		FocusRegistration: runtime.FocusRegistrationAuto,
		TickRate:          time.Second / 60,
	})
	runAppForTest(t, app)

	socketPath := filepath.Join(t.TempDir(), "agent.sock")
	srv, err := NewServer(ServerOptions{
		Addr:            "unix:" + socketPath,
		App:             app,
		Token:           "secret",
		AllowText:       false,
		SnapshotTimeout: time.Second,
	})
	if err != nil {
		t.Fatalf("new server: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- srv.Serve(ctx)
	}()

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(socketPath); err == nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)

	send := func(t *testing.T, req any) response {
		t.Helper()
		data, err := json.Marshal(req)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		if _, err := conn.Write(append(data, '\n')); err != nil {
			t.Fatalf("write: %v", err)
		}
		line, err := reader.ReadBytes('\n')
		if err != nil {
			t.Fatalf("read: %v", err)
		}
		var resp response
		if err := json.Unmarshal(line, &resp); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		return resp
	}

	resp := send(t, request{ID: 1})
	if resp.Error != "missing_type" {
		t.Fatalf("expected missing_type, got %+v", resp)
	}

	resp = send(t, request{ID: 2, Type: "ping"})
	if resp.Error != "unauthorized" {
		t.Fatalf("expected unauthorized, got %+v", resp)
	}

	resp = send(t, request{ID: 3, Type: "hello", Token: "nope"})
	if resp.Error != "unauthorized" {
		t.Fatalf("expected unauthorized, got %+v", resp)
	}

	resp = send(t, request{ID: 4, Type: "hello", Token: "secret"})
	if !resp.OK || resp.Capabilities == nil {
		t.Fatalf("expected capabilities, got %+v", resp)
	}

	resp = send(t, request{ID: 5, Type: "snapshot", IncludeText: true})
	if resp.Error != "text_disabled" {
		t.Fatalf("expected text_disabled, got %+v", resp)
	}

	resp = send(t, request{ID: 6, Type: "snapshot"})
	if !resp.OK || resp.Snapshot == nil {
		t.Fatalf("expected snapshot, got %+v", resp)
	}

	resp = send(t, request{ID: 7, Type: "key", Key: "g g"})
	if resp.Error != "invalid_key" {
		t.Fatalf("expected invalid_key, got %+v", resp)
	}
	resp = send(t, request{ID: 7, Type: "key", Key: "enter"})
	if !resp.OK {
		t.Fatalf("expected key ok, got %+v", resp)
	}

	resp = send(t, request{ID: 8, Type: "text", Text: ""})
	if resp.Error != "missing_text" {
		t.Fatalf("expected missing_text, got %+v", resp)
	}
	resp = send(t, request{ID: 8, Type: "text", Text: "hi"})
	if !resp.OK {
		t.Fatalf("expected text ok, got %+v", resp)
	}

	resp = send(t, request{ID: 9, Type: "mouse", Button: "bad", Action: "press"})
	if resp.Error != "invalid_mouse_button" {
		t.Fatalf("expected invalid_mouse_button, got %+v", resp)
	}

	resp = send(t, request{ID: 10, Type: "mouse", Button: "left", Action: "bad"})
	if resp.Error != "invalid_mouse_action" {
		t.Fatalf("expected invalid_mouse_action, got %+v", resp)
	}

	resp = send(t, request{ID: 11, Type: "mouse", Button: "left", Action: "press"})
	if !resp.OK {
		t.Fatalf("expected mouse ok, got %+v", resp)
	}

	resp = send(t, request{ID: 12, Type: "paste", Text: "data"})
	if !resp.OK {
		t.Fatalf("expected paste ok, got %+v", resp)
	}

	resp = send(t, request{ID: 13, Type: "resize", Width: 0, Height: -1})
	if resp.Error != "invalid_resize" {
		t.Fatalf("expected invalid_resize, got %+v", resp)
	}

	resp = send(t, request{ID: 14, Type: "resize", Width: 20, Height: 5})
	if !resp.OK {
		t.Fatalf("expected resize ok, got %+v", resp)
	}

	resp = send(t, request{ID: 15, Type: "ping"})
	if !resp.OK {
		t.Fatalf("expected ping ok, got %+v", resp)
	}

	resp = send(t, request{ID: 16, Type: "unknown"})
	if resp.Error != "unknown_type" {
		t.Fatalf("expected unknown_type, got %+v", resp)
	}

	if _, err := conn.Write([]byte("{bad_json}\n")); err != nil {
		t.Fatalf("write bad json: %v", err)
	}
	line, err := reader.ReadBytes('\n')
	if err != nil {
		t.Fatalf("read bad json response: %v", err)
	}
	var badResp response
	if err := json.Unmarshal(line, &badResp); err != nil {
		t.Fatalf("unmarshal bad response: %v", err)
	}
	if badResp.Error != "bad_json" {
		t.Fatalf("expected bad_json, got %+v", badResp)
	}

	cancel()
	_ = srv.Close()

	err = <-done
	if err != nil && !errors.Is(err, context.Canceled) {
		t.Fatalf("serve error: %v", err)
	}

	if _, err := os.Stat(socketPath); !os.IsNotExist(err) {
		t.Fatalf("expected socket to be removed, err=%v", err)
	}
}
