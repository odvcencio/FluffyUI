package keybind

// ModeManager manages keymap modes.
type ModeManager struct {
	modes   map[string]*Keymap
	current string
	stack   []string
}

// NewModeManager creates an empty mode manager.
func NewModeManager() *ModeManager {
	return &ModeManager{modes: make(map[string]*Keymap)}
}

// Register associates a keymap with a mode name.
func (m *ModeManager) Register(name string, keymap *Keymap) {
	if m == nil || name == "" {
		return
	}
	if m.modes == nil {
		m.modes = make(map[string]*Keymap)
	}
	m.modes[name] = keymap
	if m.current == "" {
		m.current = name
	}
}

// Current returns the active keymap.
func (m *ModeManager) Current() *Keymap {
	if m == nil || m.current == "" {
		return nil
	}
	return m.modes[m.current]
}

// CurrentName returns the active mode name.
func (m *ModeManager) CurrentName() string {
	if m == nil {
		return ""
	}
	return m.current
}

// Push pushes the current mode onto the stack and switches to mode.
func (m *ModeManager) Push(mode string) {
	if m == nil || mode == "" {
		return
	}
	if m.current != "" {
		m.stack = append(m.stack, m.current)
	}
	m.current = mode
}

// Pop restores the previous mode.
func (m *ModeManager) Pop() {
	if m == nil || len(m.stack) == 0 {
		return
	}
	last := m.stack[len(m.stack)-1]
	m.stack = m.stack[:len(m.stack)-1]
	m.current = last
}

// Set switches to a mode without modifying the stack.
func (m *ModeManager) Set(mode string) {
	if m == nil || mode == "" {
		return
	}
	m.current = mode
}
