package widgets

import (
	"testing"

	"github.com/odvcencio/fluffyui/runtime"
)

func TestWidgetPluginRegistry(t *testing.T) {
	registry := &widgetPluginRegistry{plugins: map[string]WidgetPlugin{}}
	plugin := WidgetPlugin{
		ID:          "demo.widget",
		Name:        "Demo Widget",
		Version:     "1.0.0",
		Description: "Test plugin",
		New: func() runtime.Widget {
			return NewLabel("Demo")
		},
	}

	if err := registry.register(plugin); err != nil {
		t.Fatalf("unexpected register error: %v", err)
	}

	if _, ok := registry.get("demo.widget"); !ok {
		t.Fatalf("expected plugin to be registered")
	}

	if err := registry.register(plugin); err == nil {
		t.Fatalf("expected duplicate registration to error")
	}

	list := registry.list()
	if len(list) != 1 {
		t.Fatalf("expected 1 plugin, got %d", len(list))
	}
}

