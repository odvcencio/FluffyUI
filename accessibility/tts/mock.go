package tts

import (
	"context"
	"sync"

	"m31labs.dev/fluffyui/accessibility"
)

// Mock records spoken text for testing. It implements accessibility.Speaker.
type Mock struct {
	mu      sync.Mutex
	history []string
}

// NewMock creates a mock speaker.
func NewMock() *Mock {
	return &Mock{}
}

// Speak records the text.
func (m *Mock) Speak(_ context.Context, text string) error {
	m.mu.Lock()
	m.history = append(m.history, text)
	m.mu.Unlock()
	return nil
}

// Stop is a no-op.
func (m *Mock) Stop() error { return nil }

// Close is a no-op.
func (m *Mock) Close() error { return nil }

// History returns a copy of all spoken text.
func (m *Mock) History() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]string, len(m.history))
	copy(out, m.history)
	return out
}

// Verify that Mock implements Speaker at compile time.
var _ accessibility.Speaker = (*Mock)(nil)
