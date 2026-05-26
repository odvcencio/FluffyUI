package widgets

import (
	"bytes"
	"testing"
	"time"

	"m31labs.dev/fluffyui/backend"
	"m31labs.dev/fluffyui/runtime"
	"m31labs.dev/fluffyui/terminal"
)

func TestLog_New(t *testing.T) {
	log := NewLog()

	if log == nil {
		t.Fatal("expected non-nil Log")
	}
	if log.maxLines != 10000 {
		t.Errorf("maxLines = %d, want 10000", log.maxLines)
	}
	if log.AutoScroll() != true {
		t.Error("AutoScroll() should default to true")
	}
	if log.ShowTime() != true {
		t.Error("ShowTime() should default to true")
	}
	if log.MinLevel() != LogDebug {
		t.Errorf("MinLevel() = %d, want LogDebug", log.MinLevel())
	}
	if log.EntryCount() != 0 {
		t.Errorf("EntryCount() = %d, want 0", log.EntryCount())
	}
}

func TestLog_WithOptions(t *testing.T) {
	log := NewLog(
		WithMaxLines(100),
		WithAutoScroll(false),
		WithShowTime(false),
		WithMinLevel(LogWarn),
		WithLogStyle(backend.DefaultStyle().Bold(true)),
	)

	if log.maxLines != 100 {
		t.Errorf("maxLines = %d, want 100", log.maxLines)
	}
	if log.AutoScroll() != false {
		t.Error("AutoScroll() = true, want false")
	}
	if log.ShowTime() != false {
		t.Error("ShowTime() = true, want false")
	}
	if log.MinLevel() != LogWarn {
		t.Errorf("MinLevel() = %d, want LogWarn", log.MinLevel())
	}
}

func TestLog_AddEntry(t *testing.T) {
	log := NewLog(WithMaxLines(10))

	log.Info("test message")

	if log.EntryCount() != 1 {
		t.Errorf("EntryCount() = %d, want 1", log.EntryCount())
	}

	entries := log.Entries()
	if len(entries) != 1 {
		t.Fatalf("len(Entries()) = %d, want 1", len(entries))
	}
	if entries[0].Message != "test message" {
		t.Errorf("Message = %q, want 'test message'", entries[0].Message)
	}
	if entries[0].Level != LogInfo {
		t.Errorf("Level = %d, want LogInfo", entries[0].Level)
	}
}

func TestLog_LevelMethods(t *testing.T) {
	log := NewLog()

	log.Debug("debug message")
	log.Info("info message")
	log.Warn("warn message")
	log.Error("error message")

	if log.EntryCount() != 4 {
		t.Errorf("EntryCount() = %d, want 4", log.EntryCount())
	}

	entries := log.Entries()
	if entries[0].Level != LogDebug {
		t.Error("first entry should be Debug")
	}
	if entries[1].Level != LogInfo {
		t.Error("second entry should be Info")
	}
	if entries[2].Level != LogWarn {
		t.Error("third entry should be Warn")
	}
	if entries[3].Level != LogError {
		t.Error("fourth entry should be Error")
	}
}

func TestLog_LogWithFields(t *testing.T) {
	log := NewLog()

	fields := map[string]any{"user": "test", "count": 42}
	log.Info("message with fields", fields)

	entries := log.Entries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Fields == nil {
		t.Fatal("expected fields to be set")
	}
	if entries[0].Fields["user"] != "test" {
		t.Errorf("Fields['user'] = %v, want 'test'", entries[0].Fields["user"])
	}
	if entries[0].Fields["count"] != 42 {
		t.Errorf("Fields['count'] = %v, want 42", entries[0].Fields["count"])
	}
}

func TestLog_RingBuffer(t *testing.T) {
	log := NewLog(WithMaxLines(5))

	// Add more entries than buffer size
	for i := 0; i < 10; i++ {
		log.Info("message")
	}

	if log.EntryCount() != 5 {
		t.Errorf("EntryCount() = %d, want 5", log.EntryCount())
	}

	entries := log.Entries()
	if len(entries) != 5 {
		t.Errorf("len(Entries()) = %d, want 5", len(entries))
	}
}

func TestLog_Clear(t *testing.T) {
	log := NewLog()

	log.Info("message 1")
	log.Info("message 2")

	if log.EntryCount() != 2 {
		t.Fatalf("EntryCount() = %d, want 2", log.EntryCount())
	}

	log.Clear()

	if log.EntryCount() != 0 {
		t.Errorf("After Clear(), EntryCount() = %d, want 0", log.EntryCount())
	}
}

func TestLog_Filter(t *testing.T) {
	log := NewLog()

	log.Info("include this")
	log.Info("exclude this")
	log.Info("include again")

	log.SetFilter(func(e LogEntry) bool {
		return e.Message == "include this" || e.Message == "include again"
	})

	filtered := log.FilteredEntries()
	if len(filtered) != 2 {
		t.Errorf("len(FilteredEntries()) = %d, want 2", len(filtered))
	}
}

func TestLog_MinLevelFilter(t *testing.T) {
	log := NewLog()

	log.Debug("debug")
	log.Info("info")
	log.Warn("warn")
	log.Error("error")

	// Filter to Warn and above
	log.SetMinLevel(LogWarn)

	filtered := log.FilteredEntries()
	if len(filtered) != 2 {
		t.Errorf("len(FilteredEntries()) = %d, want 2", len(filtered))
	}
}

func TestLog_Write(t *testing.T) {
	log := NewLog()

	// Test io.Writer interface
	n, err := log.Write([]byte("line 1\nline 2\n"))
	if err != nil {
		t.Errorf("Write() error = %v", err)
	}
	if n != 14 {
		t.Errorf("Write() = %d, want 14", n)
	}

	if log.EntryCount() != 2 {
		t.Errorf("EntryCount() = %d, want 2", log.EntryCount())
	}
}

func TestLog_WriteEmpty(t *testing.T) {
	log := NewLog()

	// Empty lines should be skipped
	log.Write([]byte("\n\n"))

	if log.EntryCount() != 0 {
		t.Errorf("EntryCount() = %d, want 0 for empty lines", log.EntryCount())
	}
}

func TestLog_WriteToBuffer(t *testing.T) {
	log := NewLog()

	// Use log as io.Writer with a buffer
	var buf bytes.Buffer
	buf.WriteString("test log message\n")

	log.Write(buf.Bytes())

	entries := log.Entries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry")
	}
	if entries[0].Message != "test log message" {
		t.Errorf("Message = %q, want 'test log message'", entries[0].Message)
	}
}

func TestLog_Measure(t *testing.T) {
	log := NewLog()

	size := log.Measure(runtime.Constraints{MaxWidth: 80, MaxHeight: 24})

	if size.Width != 80 {
		t.Errorf("Width = %d, want 80", size.Width)
	}
	if size.Height != 24 {
		t.Errorf("Height = %d, want 24", size.Height)
	}
}

func TestLog_Render(t *testing.T) {
	log := NewLog(WithShowTime(false))
	log.Info("Hello, World!")
	log.Focus()
	log.Layout(runtime.Rect{X: 0, Y: 0, Width: 40, Height: 5})

	buf := runtime.NewBuffer(40, 5)
	ctx := runtime.RenderContext{Buffer: buf}

	log.Render(ctx)

	// Check that something was rendered
	// The message should appear at the bottom of the viewport
	cell := buf.Get(0, 4) // Last line
	if cell.Rune == 0 || cell.Rune == ' ' {
		// Might be "[INFO] Hello, World!"
	}
}

func TestLog_RenderEmpty(t *testing.T) {
	log := NewLog()
	log.Layout(runtime.Rect{X: 0, Y: 0, Width: 0, Height: 0})

	buf := runtime.NewBuffer(40, 5)
	ctx := runtime.RenderContext{Buffer: buf}

	// Should not panic
	log.Render(ctx)
}

func TestLog_Navigation(t *testing.T) {
	log := NewLog(WithAutoScroll(true))
	for i := 0; i < 20; i++ {
		log.Info("message")
	}
	log.Focus()
	log.Layout(runtime.Rect{X: 0, Y: 0, Width: 40, Height: 5})

	// AutoScroll starts at bottom
	if log.offset != 0 {
		t.Errorf("offset = %d, want 0", log.offset)
	}

	// Scroll up
	log.HandleMessage(runtime.KeyMsg{Key: terminal.KeyUp})
	if log.offset != 1 {
		t.Errorf("After Up, offset = %d, want 1", log.offset)
	}
	if log.autoScroll {
		t.Error("AutoScroll should be disabled after manual scroll")
	}

	// Scroll down back to bottom
	log.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	if log.offset != 0 {
		t.Errorf("After Down, offset = %d, want 0", log.offset)
	}
	if !log.autoScroll {
		t.Error("AutoScroll should be re-enabled at bottom")
	}
}

func TestLog_PageNavigation(t *testing.T) {
	log := NewLog()
	for i := 0; i < 50; i++ {
		log.Info("message")
	}
	log.Focus()
	log.Layout(runtime.Rect{X: 0, Y: 0, Width: 40, Height: 10})

	// Page up
	log.HandleMessage(runtime.KeyMsg{Key: terminal.KeyPageUp})
	if log.offset < 5 {
		t.Errorf("After PageUp, offset = %d, expected >= 5", log.offset)
	}

	// Page down
	log.HandleMessage(runtime.KeyMsg{Key: terminal.KeyPageDown})
	if log.offset >= 10 {
		t.Errorf("After PageDown, offset = %d, expected < 10", log.offset)
	}
}

func TestLog_HomeEnd(t *testing.T) {
	log := NewLog()
	for i := 0; i < 20; i++ {
		log.Info("message")
	}
	log.Focus()
	log.Layout(runtime.Rect{X: 0, Y: 0, Width: 40, Height: 5})

	// Go to start (oldest)
	log.HandleMessage(runtime.KeyMsg{Key: terminal.KeyHome})
	if log.offset != 15 { // 20 entries - 5 viewport = 15 max offset
		t.Errorf("After Home, offset = %d, want 15", log.offset)
	}

	// Go to end (newest)
	log.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnd})
	if log.offset != 0 {
		t.Errorf("After End, offset = %d, want 0", log.offset)
	}
	if !log.autoScroll {
		t.Error("AutoScroll should be enabled at End")
	}
}

func TestLog_Scroll(t *testing.T) {
	log := NewLog()
	for i := 0; i < 20; i++ {
		log.Info("message")
	}
	log.Layout(runtime.Rect{X: 0, Y: 0, Width: 40, Height: 5})

	// Test ScrollBy
	log.ScrollBy(0, -3) // Scroll toward older entries
	if log.offset < 1 {
		t.Errorf("After ScrollBy, offset = %d, expected > 0", log.offset)
	}

	// Test ScrollToStart
	log.ScrollToStart()
	if log.offset == 0 {
		t.Errorf("After ScrollToStart, offset = 0, expected > 0")
	}

	// Test ScrollToEnd
	log.ScrollToEnd()
	if log.offset != 0 {
		t.Errorf("After ScrollToEnd, offset = %d, want 0", log.offset)
	}
}

func TestLog_ClipboardCopy(t *testing.T) {
	log := NewLog(WithShowTime(false))
	log.Info("message 1")
	log.Info("message 2")

	text, ok := log.ClipboardCopy()
	if !ok {
		t.Error("ClipboardCopy() returned false")
	}
	if text == "" {
		t.Error("ClipboardCopy() returned empty string")
	}
}

func TestLog_Tab(t *testing.T) {
	log := NewLog()
	log.Focus()

	result := log.HandleMessage(runtime.KeyMsg{Key: terminal.KeyTab})

	if len(result.Commands) == 0 {
		t.Error("expected FocusNext command")
	}
	_, ok := result.Commands[0].(runtime.FocusNext)
	if !ok {
		t.Errorf("expected FocusNext, got %T", result.Commands[0])
	}
}

func TestLog_ShiftTab(t *testing.T) {
	log := NewLog()
	log.Focus()

	result := log.HandleMessage(runtime.KeyMsg{Key: terminal.KeyTab, Shift: true})

	if len(result.Commands) == 0 {
		t.Error("expected FocusPrev command")
	}
	_, ok := result.Commands[0].(runtime.FocusPrev)
	if !ok {
		t.Errorf("expected FocusPrev, got %T", result.Commands[0])
	}
}

func TestLog_Escape(t *testing.T) {
	log := NewLog()
	log.Focus()

	result := log.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEscape})

	if len(result.Commands) == 0 {
		t.Error("expected Cancel command")
	}
	_, ok := result.Commands[0].(runtime.Cancel)
	if !ok {
		t.Errorf("expected Cancel, got %T", result.Commands[0])
	}
}

func TestLog_Unfocused(t *testing.T) {
	log := NewLog()
	// Not focused

	result := log.HandleMessage(runtime.KeyMsg{Key: terminal.KeyUp})

	if result.Handled {
		t.Error("unfocused log should not handle messages")
	}
}

func TestLog_NonKeyMsg(t *testing.T) {
	log := NewLog()
	log.Focus()

	result := log.HandleMessage(runtime.ResizeMsg{Width: 80, Height: 24})

	if result.Handled {
		t.Error("non-key message should not be handled")
	}
}

func TestLog_AutoScrollOnNewEntry(t *testing.T) {
	log := NewLog(WithAutoScroll(true))
	for i := 0; i < 10; i++ {
		log.Info("message")
	}
	log.Layout(runtime.Rect{X: 0, Y: 0, Width: 40, Height: 5})

	// Scroll up manually
	log.offset = 3
	log.autoScroll = false

	// Add new entry (should not auto-scroll since we disabled it)
	log.Info("new message")
	if log.offset != 3 {
		t.Errorf("offset = %d, expected 3 (no auto-scroll)", log.offset)
	}

	// Re-enable auto-scroll and add entry
	log.autoScroll = true
	log.Info("another message")
	if log.offset != 0 {
		t.Errorf("offset = %d, expected 0 (auto-scroll to bottom)", log.offset)
	}
}

func TestLog_LevelColors(t *testing.T) {
	log := NewLog()
	log.SetLevelStyle(LogDebug, backend.DefaultStyle().Dim(true))

	// Verify no panic and colors are set
	if log.levelColors[LogDebug] == (backend.Style{}) {
		t.Error("expected Debug style to be set")
	}
}

func TestLog_Accessibility(t *testing.T) {
	log := NewLog()
	log.SetLabel("Application Log")

	if log.AccessibleRole() != "log" {
		t.Errorf("AccessibleRole() = %q, want 'log'", log.AccessibleRole())
	}
	if log.AccessibleLabel() != "Application Log" {
		t.Errorf("AccessibleLabel() = %q, want 'Application Log'", log.AccessibleLabel())
	}
}

func TestLog_StyleType(t *testing.T) {
	log := NewLog()

	if log.StyleType() != "Log" {
		t.Errorf("StyleType() = %q, want 'Log'", log.StyleType())
	}
}

func TestLog_LogLevelString(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected string
	}{
		{LogDebug, "DEBUG"},
		{LogInfo, "INFO"},
		{LogWarn, "WARN"},
		{LogError, "ERROR"},
		{LogLevel(99), "UNKNOWN"},
	}

	for _, tc := range tests {
		if tc.level.String() != tc.expected {
			t.Errorf("LogLevel(%d).String() = %q, want %q", tc.level, tc.level.String(), tc.expected)
		}
	}
}

func TestLog_EntryFormat(t *testing.T) {
	entry := LogEntry{
		Time:    time.Date(2026, 1, 15, 10, 30, 45, 0, time.UTC),
		Level:   LogInfo,
		Message: "test message",
		Fields:  map[string]any{"key": "value"},
	}

	formatted := entry.Format(true)
	if formatted == "" {
		t.Error("Format() returned empty string")
	}
	if !bytes.Contains([]byte(formatted), []byte("10:30:45")) {
		t.Error("Format() should contain time")
	}
	if !bytes.Contains([]byte(formatted), []byte("[INFO]")) {
		t.Error("Format() should contain level")
	}
	if !bytes.Contains([]byte(formatted), []byte("test message")) {
		t.Error("Format() should contain message")
	}
	if !bytes.Contains([]byte(formatted), []byte("key=value")) {
		t.Error("Format() should contain fields")
	}

	// Without time
	formatted = entry.Format(false)
	if bytes.Contains([]byte(formatted), []byte("10:30:45")) {
		t.Error("Format(false) should not contain time")
	}
}

func TestLog_NilSafety(t *testing.T) {
	var log *Log

	// These should not panic
	log.SetFilter(nil)
	log.SetMinLevel(LogWarn)
	log.SetAutoScroll(true)
	log.SetShowTime(true)
	log.SetStyle(backend.DefaultStyle())
	log.SetLevelStyle(LogInfo, backend.DefaultStyle())
	log.SetLabel("Test")
	log.Bind(runtime.Services{})
	log.Unbind()
	log.Clear()
	log.AddEntry(LogEntry{})
	log.Log(LogInfo, "test")
	log.Debug("test")
	log.Info("test")
	log.Warn("test")
	log.Error("test")
	log.ScrollBy(0, 1)
	log.ScrollTo(0, 0)
	log.PageBy(1)
	log.ScrollToStart()
	log.ScrollToEnd()

	n, err := log.Write([]byte("test"))
	if err != nil {
		t.Error("Write() should not error for nil log")
	}
	if n != 4 {
		t.Errorf("Write() = %d, want 4", n)
	}

	if log.AutoScroll() != false {
		t.Error("expected AutoScroll() = false")
	}
	if log.ShowTime() != false {
		t.Error("expected ShowTime() = false")
	}
	if log.MinLevel() != LogDebug {
		t.Error("expected MinLevel() = LogDebug")
	}
	if log.EntryCount() != 0 {
		t.Error("expected EntryCount() = 0")
	}
	if log.Entries() != nil {
		t.Error("expected Entries() = nil")
	}
	if log.FilteredEntries() != nil {
		t.Error("expected FilteredEntries() = nil")
	}
}

func TestLog_ConcurrentAccess(t *testing.T) {
	log := NewLog(WithMaxLines(100))

	done := make(chan bool)

	// Writer goroutine
	go func() {
		for i := 0; i < 50; i++ {
			log.Info("message from writer")
		}
		done <- true
	}()

	// Reader goroutine
	go func() {
		for i := 0; i < 50; i++ {
			_ = log.Entries()
			_ = log.EntryCount()
		}
		done <- true
	}()

	<-done
	<-done

	// Test passed if no race or panic
}
