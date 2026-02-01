package agent

import (
	"bufio"
	"context"
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/odvcencio/fluffyui/backend/sim"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/widgets"
)

func TestServerRequestsProcessGuards(t *testing.T) {
	agt := New(Config{Sim: sim.New(10, 4)})
	opts := DefaultEnhancedServerOptions()
	opts.Addr = "unix:" + filepath.Join(t.TempDir(), "enhanced.sock")
	opts.Agent = agt
	opts.SessionPoolLimits = PoolLimits{GlobalRateLimit: 1, GlobalBurstLimit: 1}

	server, err := NewEnhancedServer(opts)
	if err != nil {
		t.Fatalf("new enhanced server: %v", err)
	}

	expired := NewSession("expired", ModeNormal, SessionLimits{IdleTimeout: time.Millisecond})
	expired.lastSeen.Store(time.Now().Add(-time.Hour))
	resp := server.processRequest(context.Background(), &serverSession{session: expired}, request{Type: "ping"})
	if resp.Error != "session_expired" {
		t.Fatalf("expected session_expired, got %+v", resp)
	}

	rejected := NewSession("rejected", ModeNormal, DefaultSessionLimits())
	rejected.Reject()
	resp = server.processRequest(context.Background(), &serverSession{session: rejected}, request{Type: "ping"})
	if resp.Error != "session_rejected" {
		t.Fatalf("expected session_rejected, got %+v", resp)
	}

	okSession := NewSession("ok", ModeNormal, DefaultSessionLimits())
	okSess := &serverSession{session: okSession, authed: true}
	resp = server.processRequest(context.Background(), okSess, request{Type: "ping"})
	if !resp.OK {
		t.Fatalf("expected ping ok, got %+v", resp)
	}
	resp = server.processRequest(context.Background(), okSess, request{Type: "ping"})
	if resp.Error != "rate_limited" {
		t.Fatalf("expected rate_limited, got %+v", resp)
	}
}

func TestServerRequestsServeAndHandle(t *testing.T) {
	be := sim.New(40, 10)
	app := runtime.NewApp(runtime.AppConfig{
		Backend:           be,
		Root:              widgets.NewLabel("Hello"),
		Update:            runtime.DefaultUpdate,
		FocusRegistration: runtime.FocusRegistrationAuto,
		TickRate:          time.Second / 60,
	})
	runAppForTest(t, app)

	socketPath := filepath.Join(t.TempDir(), "enhanced.sock")
	opts := DefaultEnhancedServerOptions()
	opts.Addr = "unix:" + socketPath
	opts.App = app
	opts.Token = "token"
	opts.AllowText = false
	opts.EnableHealthCheck = false
	opts.SessionLimits = SessionLimits{MaxRequestsPerSec: 1000, BurstLimit: 1000}

	server, err := NewEnhancedServer(opts)
	if err != nil {
		t.Fatalf("new enhanced server: %v", err)
	}
	if err := server.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	t.Cleanup(func() {
		_ = server.Stop()
	})

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

	resp = send(t, request{ID: 3, Type: "hello", Token: "bad"})
	if resp.Error != "unauthorized" {
		t.Fatalf("expected unauthorized, got %+v", resp)
	}

	resp = send(t, request{ID: 4, Type: "hello", Token: "token"})
	if !resp.OK || resp.Capabilities == nil {
		t.Fatalf("expected capabilities, got %+v", resp)
	}

	resp = send(t, request{ID: 5, Type: "health"})
	if !resp.OK || !strings.Contains(resp.Message, "healthy") {
		t.Fatalf("expected health response, got %+v", resp)
	}

	resp = send(t, request{ID: 6, Type: "stats"})
	if !resp.OK || resp.Message == "" {
		t.Fatalf("expected stats response, got %+v", resp)
	}

	resp = send(t, request{ID: 7, Type: "snapshot", IncludeText: true})
	if resp.Error != "text_disabled" {
		t.Fatalf("expected text_disabled, got %+v", resp)
	}

	resp = send(t, request{ID: 8, Type: "snapshot"})
	if !resp.OK || resp.Snapshot == nil {
		t.Fatalf("expected snapshot, got %+v", resp)
	}

	resp = send(t, request{ID: 9, Type: "key", Key: "g g"})
	if resp.Error != "invalid_key" {
		t.Fatalf("expected invalid_key, got %+v", resp)
	}
	resp = send(t, request{ID: 9, Type: "key", Key: "enter"})
	if !resp.OK {
		t.Fatalf("expected key ok, got %+v", resp)
	}

	resp = send(t, request{ID: 10, Type: "text", Text: ""})
	if resp.Error != "missing_text" {
		t.Fatalf("expected missing_text, got %+v", resp)
	}
	resp = send(t, request{ID: 10, Type: "text", Text: "hi"})
	if !resp.OK {
		t.Fatalf("expected text ok, got %+v", resp)
	}

	resp = send(t, request{ID: 11, Type: "mouse", Button: "bad", Action: "press"})
	if resp.Error != "invalid_mouse_button" {
		t.Fatalf("expected invalid_mouse_button, got %+v", resp)
	}

	resp = send(t, request{ID: 12, Type: "mouse", Button: "left", Action: "bad"})
	if resp.Error != "invalid_mouse_action" {
		t.Fatalf("expected invalid_mouse_action, got %+v", resp)
	}

	resp = send(t, request{ID: 13, Type: "mouse", Button: "left", Action: "press"})
	if !resp.OK {
		t.Fatalf("expected mouse ok, got %+v", resp)
	}

	resp = send(t, request{ID: 14, Type: "paste", Text: "data"})
	if !resp.OK {
		t.Fatalf("expected paste ok, got %+v", resp)
	}

	resp = send(t, request{ID: 15, Type: "resize", Width: 0, Height: -1})
	if resp.Error != "invalid_resize" {
		t.Fatalf("expected invalid_resize, got %+v", resp)
	}

	resp = send(t, request{ID: 16, Type: "resize", Width: 20, Height: 5})
	if !resp.OK {
		t.Fatalf("expected resize ok, got %+v", resp)
	}

	resp = send(t, request{ID: 17, Type: "background_task"})
	if !resp.OK || !strings.Contains(resp.Message, "session") {
		t.Fatalf("expected background task message, got %+v", resp)
	}

	resp = send(t, request{ID: 18, Type: "task_status"})
	if !resp.OK || resp.Message == "" {
		t.Fatalf("expected task status, got %+v", resp)
	}

	block := make(chan struct{})
	task, err := server.SubmitBackgroundTask("task", "", "session", func(ctx context.Context, task *BackgroundTask) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-block:
			return nil
		}
	})
	if err != nil {
		t.Fatalf("submit background task: %v", err)
	}

	resp = send(t, request{ID: 19, Type: "task_status", Text: task.ID})
	if !resp.OK || resp.Message == "" {
		t.Fatalf("expected task status, got %+v", resp)
	}

	resp = send(t, request{ID: 20, Type: "task_cancel"})
	if resp.Error != "missing_task_id" {
		t.Fatalf("expected missing_task_id, got %+v", resp)
	}

	resp = send(t, request{ID: 21, Type: "task_cancel", Text: "missing"})
	if resp.Error != "task_not_found" {
		t.Fatalf("expected task_not_found, got %+v", resp)
	}

	resp = send(t, request{ID: 22, Type: "task_cancel", Text: task.ID})
	if !resp.OK {
		t.Fatalf("expected task cancel ok, got %+v", resp)
	}

	resp = send(t, request{ID: 23, Type: "close"})
	if !resp.OK || resp.Message != "closing" {
		t.Fatalf("expected closing, got %+v", resp)
	}

	close(block)

	if err := server.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	if _, err := os.Stat(socketPath); !os.IsNotExist(err) {
		t.Fatalf("expected socket removed, err=%v", err)
	}

	stats := server.Stats()
	if stats.SessionStats.TotalSessions < 0 {
		t.Fatalf("unexpected stats: %+v", stats)
	}
	server.updateHealth()
	if !server.Health().Healthy {
		t.Fatalf("expected healthy status")
	}
}

func TestServerRequestsStartStopGuards(t *testing.T) {
	if err := (*EnhancedServer)(nil).Start(); err == nil {
		t.Fatal("expected error for nil server start")
	}
	if err := (*EnhancedServer)(nil).Stop(); err != nil {
		t.Fatal("expected nil stop error for nil server")
	}

	agt := New(Config{Sim: sim.New(10, 4)})
	opts := DefaultEnhancedServerOptions()
	opts.Addr = "unix:" + filepath.Join(t.TempDir(), "guard.sock")
	opts.Agent = agt

	server, err := NewEnhancedServer(opts)
	if err != nil {
		t.Fatalf("new enhanced server: %v", err)
	}
	if err := server.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	if err := server.Start(); err == nil {
		t.Fatal("expected already running error")
	}
	if err := server.Stop(); err != nil {
		t.Fatalf("stop: %v", err)
	}
	if err := server.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
}

func TestServerRequestsStatsAndHealthHelpers(t *testing.T) {
	server := &EnhancedServer{}
	if server.Health().Healthy {
		t.Fatal("expected nil server health to be unhealthy")
	}
	if stats := server.Stats(); stats.Running {
		t.Fatalf("unexpected stats: %+v", stats)
	}

	server.onQueueFull(&Request{})
	server.onRequestStart(&Request{})
	server.onRequestDone(&Request{}, time.Millisecond, nil)
}
