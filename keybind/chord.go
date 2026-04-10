package keybind

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/odvcencio/fluffyui/terminal"
)

// Chord represents a multi-key sequence (e.g., Ctrl+K followed by Ctrl+C).
type Chord []terminal.KeyEvent

// ChordState indicates the current state of chord recognition.
type ChordState int

const (
	// ChordIdle means no chord is in progress.
	ChordIdle ChordState = iota
	// ChordPending means a partial chord prefix has been recognized.
	ChordPending
	// ChordCompleted means a chord was fully matched and its action was invoked.
	ChordCompleted
)

// defaultChordTimeout is the maximum time between chord key presses.
const defaultChordTimeout = 1 * time.Second

// chordBinding stores a parsed chord sequence and its action.
type chordBinding struct {
	sequence []KeyPress
	action   func()
}

// ChordHandler manages multi-key chord recognition. It maintains a set of
// chord bindings and tracks pending key presses, applying a configurable
// timeout between keys.
type ChordHandler struct {
	mu       sync.Mutex
	bindings map[string]*chordBinding // chord string -> binding
	pending  []KeyPress
	timeout  time.Duration
	timer    *time.Timer
}

// NewChordHandler creates a ChordHandler with default settings.
func NewChordHandler() *ChordHandler {
	return &ChordHandler{
		bindings: make(map[string]*chordBinding),
		timeout:  defaultChordTimeout,
	}
}

// Bind registers a chord sequence to an action. The chord format uses
// space-separated key descriptions, e.g. "ctrl+k ctrl+c".
func (h *ChordHandler) Bind(chord string, action func()) {
	if h == nil || chord == "" || action == nil {
		return
	}
	key, err := ParseKeySequence(chord)
	if err != nil || len(key.Sequence) == 0 {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.bindings == nil {
		h.bindings = make(map[string]*chordBinding)
	}
	h.bindings[chord] = &chordBinding{
		sequence: key.Sequence,
		action:   action,
	}
}

// Unbind removes a chord binding.
func (h *ChordHandler) Unbind(chord string) {
	if h == nil || chord == "" {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.bindings, chord)
}

// HandleKey processes a key event. It returns the chord state:
//   - ChordIdle if the key does not match any chord or prefix
//   - ChordPending if the key extends a partial chord match
//   - ChordCompleted if the key completed a chord (action was invoked)
func (h *ChordHandler) HandleKey(key terminal.KeyEvent) ChordState {
	if h == nil {
		return ChordIdle
	}
	h.mu.Lock()
	defer h.mu.Unlock()

	h.stopTimer()

	press := KeyPressFromTerminal(key)
	// Normalize: if it's a ctrl key (e.g. KeyCtrlC), set the Ctrl flag.
	if isCtrlKey(press.Key) {
		press.Ctrl = true
	}

	seq := append(h.pending, press)

	// Check if seq is a prefix of any longer binding. If so, prefer waiting
	// for the longer chord even if there's an exact match for the current
	// sequence (VS Code behavior: ctrl+k waits even if ctrl+k is also bound,
	// because ctrl+k ctrl+c might follow).
	if h.isPrefix(seq) {
		h.pending = seq
		h.startTimer()
		return ChordPending
	}

	// Check for exact match.
	for _, binding := range h.bindings {
		if matchesSequence(binding.sequence, seq) {
			h.pending = nil
			action := binding.action
			// Unlock to allow re-entrant calls from the action.
			h.mu.Unlock()
			action()
			h.mu.Lock()
			return ChordCompleted
		}
	}

	// No match with accumulated sequence. Try just this key alone
	// (the pending sequence was a dead end).
	if len(h.pending) > 0 {
		solo := []KeyPress{press}

		for _, binding := range h.bindings {
			if matchesSequence(binding.sequence, solo) {
				h.pending = nil
				action := binding.action
				h.mu.Unlock()
				action()
				h.mu.Lock()
				return ChordCompleted
			}
		}

		if h.isPrefix(solo) {
			h.pending = solo
			h.startTimer()
			return ChordPending
		}
	}

	h.pending = nil
	return ChordIdle
}

// Reset clears any pending chord sequence.
func (h *ChordHandler) Reset() {
	if h == nil {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	h.stopTimer()
	h.pending = nil
}

// SetTimeout sets the maximum time between chord key presses.
func (h *ChordHandler) SetTimeout(d time.Duration) {
	if h == nil {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	h.timeout = d
}

// PendingKeys returns the current pending key sequence as a display string,
// or an empty string if no chord is in progress.
func (h *ChordHandler) PendingKeys() string {
	if h == nil {
		return ""
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if len(h.pending) == 0 {
		return ""
	}
	parts := make([]string, len(h.pending))
	for i, press := range h.pending {
		parts[i] = FormatKeyPress(press)
	}
	return strings.Join(parts, " ")
}

// isPrefix reports whether seq is a prefix of any registered chord.
// Caller must hold h.mu.
func (h *ChordHandler) isPrefix(seq []KeyPress) bool {
	for _, binding := range h.bindings {
		if len(seq) < len(binding.sequence) && prefixMatches(binding.sequence, seq) {
			return true
		}
	}
	return false
}

// startTimer starts the chord timeout. Caller must hold h.mu.
func (h *ChordHandler) startTimer() {
	h.timer = time.AfterFunc(h.timeout, func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		h.pending = nil
	})
}

// stopTimer stops any running chord timeout. Caller must hold h.mu.
func (h *ChordHandler) stopTimer() {
	if h.timer != nil {
		h.timer.Stop()
		h.timer = nil
	}
}

// matchesSequence reports whether a binding sequence matches the given key presses exactly.
func matchesSequence(binding, seq []KeyPress) bool {
	if len(binding) != len(seq) {
		return false
	}
	for i := range binding {
		if !binding[i].Equal(seq[i]) {
			return false
		}
	}
	return true
}

// prefixMatches reports whether seq is a prefix of binding.
func prefixMatches(binding, seq []KeyPress) bool {
	if len(seq) > len(binding) {
		return false
	}
	for i := range seq {
		if !binding[i].Equal(seq[i]) {
			return false
		}
	}
	return true
}

// ParseKey parses a key description like "ctrl+k", "shift+enter", "alt+a", "F5"
// into a terminal.KeyEvent.
func ParseKey(desc string) (terminal.KeyEvent, error) {
	if desc == "" {
		return terminal.KeyEvent{}, fmt.Errorf("empty key description")
	}
	press, err := parseKeyToken(desc)
	if err != nil {
		return terminal.KeyEvent{}, err
	}
	if press.Ctrl && press.Key == terminal.KeyCtrlK && press.Rune == 0 {
		press.Key = terminal.KeyRune
		press.Rune = 'k'
	}
	return terminal.KeyEvent{
		Key:   press.Key,
		Rune:  press.Rune,
		Alt:   press.Alt,
		Ctrl:  press.Ctrl,
		Shift: press.Shift,
	}, nil
}

// FormatKey formats a terminal.KeyEvent as a human-readable string like "Ctrl+K".
func FormatKey(key terminal.KeyEvent) string {
	press := KeyPressFromTerminal(key)
	if isCtrlKey(press.Key) {
		press.Ctrl = true
	}
	return FormatKeyPress(press)
}

// FormatChord formats a chord description string (e.g. "ctrl+k ctrl+c") into
// a display string (e.g. "Ctrl+K Ctrl+C").
func FormatChord(chord string) string {
	key, err := ParseKeySequence(chord)
	if err != nil || len(key.Sequence) == 0 {
		return chord
	}
	return FormatKeySequence(key)
}
