package widgets

import (
	"errors"
	"sort"
	"strings"
	"sync"

	"github.com/odvcencio/fluffyui/runtime"
)

// WidgetPlugin describes a third-party widget plugin.
type WidgetPlugin struct {
	ID          string
	Name        string
	Version     string
	Description string
	Categories  []string
	New         func() runtime.Widget
}

// Validate returns an error if the plugin metadata is incomplete.
func (p WidgetPlugin) Validate() error {
	if strings.TrimSpace(p.ID) == "" {
		return errors.New("widget plugin id is required")
	}
	if p.New == nil {
		return errors.New("widget plugin requires New factory")
	}
	return nil
}

type widgetPluginRegistry struct {
	mu      sync.RWMutex
	plugins map[string]WidgetPlugin
}

var defaultPluginRegistry = &widgetPluginRegistry{plugins: map[string]WidgetPlugin{}}

// RegisterWidgetPlugin registers a plugin globally.
func RegisterWidgetPlugin(plugin WidgetPlugin) error {
	return defaultPluginRegistry.register(plugin)
}

// MustRegisterWidgetPlugin registers a plugin or panics.
func MustRegisterWidgetPlugin(plugin WidgetPlugin) {
	if err := RegisterWidgetPlugin(plugin); err != nil {
		panic(err)
	}
}

// WidgetPlugins returns all registered plugins sorted by ID.
func WidgetPlugins() []WidgetPlugin {
	return defaultPluginRegistry.list()
}

// WidgetPluginByID fetches a plugin by ID.
func WidgetPluginByID(id string) (WidgetPlugin, bool) {
	return defaultPluginRegistry.get(id)
}

func (r *widgetPluginRegistry) register(plugin WidgetPlugin) error {
	if r == nil {
		return errors.New("plugin registry is nil")
	}
	if err := plugin.Validate(); err != nil {
		return err
	}
	id := strings.TrimSpace(plugin.ID)
	if id == "" {
		return errors.New("widget plugin id is required")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.plugins == nil {
		r.plugins = map[string]WidgetPlugin{}
	}
	if _, exists := r.plugins[id]; exists {
		return errors.New("widget plugin already registered: " + id)
	}
	r.plugins[id] = plugin
	return nil
}

func (r *widgetPluginRegistry) list() []WidgetPlugin {
	if r == nil {
		return nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	if len(r.plugins) == 0 {
		return nil
	}
	ids := make([]string, 0, len(r.plugins))
	for id := range r.plugins {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	out := make([]WidgetPlugin, 0, len(ids))
	for _, id := range ids {
		out = append(out, r.plugins[id])
	}
	return out
}

func (r *widgetPluginRegistry) get(id string) (WidgetPlugin, bool) {
	if r == nil {
		return WidgetPlugin{}, false
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.plugins[id]
	return p, ok
}

