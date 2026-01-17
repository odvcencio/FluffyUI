package keybind

// KeymapStack manages a stack of keymaps.
type KeymapStack struct {
	stack []*Keymap
}

// Push adds a keymap to the stack.
func (s *KeymapStack) Push(keymap *Keymap) {
	if s == nil || keymap == nil {
		return
	}
	s.stack = append(s.stack, keymap)
}

// Pop removes the top keymap.
func (s *KeymapStack) Pop() *Keymap {
	if s == nil || len(s.stack) == 0 {
		return nil
	}
	last := s.stack[len(s.stack)-1]
	s.stack = s.stack[:len(s.stack)-1]
	return last
}

// Current returns the top keymap.
func (s *KeymapStack) Current() *Keymap {
	if s == nil || len(s.stack) == 0 {
		return nil
	}
	return s.stack[len(s.stack)-1]
}

// All returns a copy of the keymap stack.
func (s *KeymapStack) All() []*Keymap {
	if s == nil || len(s.stack) == 0 {
		return nil
	}
	out := make([]*Keymap, len(s.stack))
	copy(out, s.stack)
	return out
}

// Match searches the stack top-down for a match.
func (s *KeymapStack) Match(seq []KeyPress, ctx Context) keyMatch {
	if s == nil || len(seq) == 0 {
		return keyMatch{}
	}
	for i := len(s.stack) - 1; i >= 0; i-- {
		match := s.stack[i].Match(seq, ctx)
		if match.Binding != nil {
			return match
		}
		if match.Prefix {
			return match
		}
	}
	return keyMatch{}
}
