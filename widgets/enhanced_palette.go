package widgets

import (
	"sort"

	"github.com/odvcencio/fluffy-ui/keybind"
)

// EnhancedPalette wraps a command registry with palette UI.
type EnhancedPalette struct {
	Widget   *PaletteWidget
	registry *keybind.CommandRegistry
	recent   []string
	pinned   []string
	keymaps  []*keybind.Keymap
}

// NewEnhancedPalette creates a palette from a registry.
func NewEnhancedPalette(registry *keybind.CommandRegistry) *EnhancedPalette {
	palette := &EnhancedPalette{
		Widget:   NewPaletteWidget("Commands"),
		registry: registry,
	}
	palette.Refresh()
	return palette
}

// SetKeymaps supplies keymaps for shortcut display.
func (p *EnhancedPalette) SetKeymaps(keymaps ...*keybind.Keymap) {
	if p == nil {
		return
	}
	p.keymaps = keymaps
	p.Refresh()
}

// SetKeymapStack supplies keymaps from a stack.
func (p *EnhancedPalette) SetKeymapStack(stack *keybind.KeymapStack) {
	if p == nil || stack == nil {
		return
	}
	p.keymaps = stack.All()
	p.Refresh()
}

// Refresh rebuilds palette items from the registry.
func (p *EnhancedPalette) Refresh() {
	if p == nil || p.Widget == nil || p.registry == nil {
		return
	}
	shortcuts := keybind.CommandShortcuts(p.keymaps...)
	commands := p.registry.List()
	sort.Slice(commands, func(i, j int) bool {
		return commands[i].ID < commands[j].ID
	})
	items := make([]PaletteItem, 0, len(commands))
	for _, cmd := range commands {
		shortcut := keybind.FormatKeySequences(shortcuts[cmd.ID])
		items = append(items, PaletteItem{
			ID:          cmd.ID,
			Category:    cmd.Category,
			Label:       commandTitle(cmd),
			Description: cmd.Description,
			Shortcut:    shortcut,
		})
	}
	items = append(p.buildPinned(shortcuts), append(p.buildRecent(shortcuts), items...)...)
	p.Widget.SetItems(items)
}

// Record marks a command as recently used.
func (p *EnhancedPalette) Record(id string) {
	if p == nil || id == "" {
		return
	}
	p.recent = append([]string{id}, p.recent...)
	p.recent = unique(p.recent)
	if len(p.recent) > 10 {
		p.recent = p.recent[:10]
	}
	p.Refresh()
}

// Pin adds a command to pinned list.
func (p *EnhancedPalette) Pin(id string) {
	if p == nil || id == "" {
		return
	}
	p.pinned = append(p.pinned, id)
	p.pinned = unique(p.pinned)
	p.Refresh()
}

// Unpin removes a command from pinned list.
func (p *EnhancedPalette) Unpin(id string) {
	if p == nil || id == "" {
		return
	}
	next := make([]string, 0, len(p.pinned))
	for _, item := range p.pinned {
		if item != id {
			next = append(next, item)
		}
	}
	p.pinned = next
	p.Refresh()
}

func (p *EnhancedPalette) buildRecent(shortcuts map[string][]keybind.Key) []PaletteItem {
	if p == nil || len(p.recent) == 0 {
		return nil
	}
	items := make([]PaletteItem, 0, len(p.recent))
	for _, id := range p.recent {
		if cmd, ok := p.registry.Get(id); ok {
			shortcut := keybind.FormatKeySequences(shortcuts[cmd.ID])
			items = append(items, PaletteItem{
				ID:          cmd.ID,
				Category:    "Recent",
				Label:       commandTitle(cmd),
				Description: cmd.Description,
				Shortcut:    shortcut,
			})
		}
	}
	return items
}

func (p *EnhancedPalette) buildPinned(shortcuts map[string][]keybind.Key) []PaletteItem {
	if p == nil || len(p.pinned) == 0 {
		return nil
	}
	items := make([]PaletteItem, 0, len(p.pinned))
	for _, id := range p.pinned {
		if cmd, ok := p.registry.Get(id); ok {
			shortcut := keybind.FormatKeySequences(shortcuts[cmd.ID])
			items = append(items, PaletteItem{
				ID:          cmd.ID,
				Category:    "Pinned",
				Label:       commandTitle(cmd),
				Description: cmd.Description,
				Shortcut:    shortcut,
			})
		}
	}
	return items
}

func commandTitle(cmd keybind.Command) string {
	if cmd.Title != "" {
		return cmd.Title
	}
	return cmd.ID
}

func unique(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	out := make([]string, 0, len(items))
	for _, item := range items {
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	return out
}
