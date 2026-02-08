package tts

import (
	"context"
	"testing"
)

func TestMockSpeaker(t *testing.T) {
	m := NewMock()
	if err := m.Speak(context.Background(), "hello"); err != nil {
		t.Fatal(err)
	}
	if err := m.Speak(context.Background(), "world"); err != nil {
		t.Fatal(err)
	}
	history := m.History()
	if len(history) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(history))
	}
	if history[0] != "hello" {
		t.Errorf("expected 'hello', got %q", history[0])
	}
	if history[1] != "world" {
		t.Errorf("expected 'world', got %q", history[1])
	}
}

func TestMockStop(t *testing.T) {
	m := NewMock()
	if err := m.Stop(); err != nil {
		t.Fatal(err)
	}
}

func TestMockClose(t *testing.T) {
	m := NewMock()
	if err := m.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestAutoDoesNotPanic(t *testing.T) {
	// Auto should never panic even if no TTS is available.
	_, _ = Auto()
}

func TestAutoMockBackend(t *testing.T) {
	s, err := Auto(WithBackend("mock"))
	if err != nil {
		t.Fatal(err)
	}
	if s == nil {
		t.Fatal("expected non-nil speaker for mock backend")
	}
	defer s.Close()
	if err := s.Speak(context.Background(), "test"); err != nil {
		t.Fatal(err)
	}
}

func TestAutoUnknownBackend(t *testing.T) {
	_, err := Auto(WithBackend("nonexistent"))
	if err == nil {
		t.Error("expected error for unknown backend")
	}
}

func TestWithRate(t *testing.T) {
	s, err := Auto(WithBackend("mock"), WithRate(1.5))
	if err != nil {
		t.Fatal(err)
	}
	if s == nil {
		t.Fatal("expected non-nil speaker")
	}
	defer s.Close()
}
