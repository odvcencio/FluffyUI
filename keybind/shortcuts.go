package keybind

// CommandShortcuts builds a map of command IDs to key sequences.
func CommandShortcuts(keymaps ...*Keymap) map[string][]Key {
	if len(keymaps) == 0 {
		return nil
	}
	out := make(map[string][]Key)
	seen := make(map[string]map[string]struct{})
	for _, keymap := range keymaps {
		for km := keymap; km != nil; km = km.Parent {
			for _, binding := range km.Bindings {
				if binding.Command == "" || len(binding.Key.Sequence) == 0 {
					continue
				}
				keyStr := FormatKeySequence(binding.Key)
				if keyStr == "" {
					continue
				}
				if _, ok := seen[binding.Command]; !ok {
					seen[binding.Command] = make(map[string]struct{})
				}
				if _, ok := seen[binding.Command][keyStr]; ok {
					continue
				}
				seen[binding.Command][keyStr] = struct{}{}
				out[binding.Command] = append(out[binding.Command], binding.Key)
			}
		}
	}
	return out
}
