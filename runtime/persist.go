package runtime

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Persistable captures widget state for persistence.
type Persistable interface {
	MarshalState() ([]byte, error)
	UnmarshalState([]byte) error
}

// PersistSnapshot stores serialized widget state keyed by widget key.
type PersistSnapshot struct {
	Widgets map[string]json.RawMessage `json:"widgets"`
}

// CaptureState walks the widget tree and captures Persistable state keyed by widget Key().
func CaptureState(root Widget) (PersistSnapshot, error) {
	snapshot := PersistSnapshot{Widgets: map[string]json.RawMessage{}}
	var firstErr error
	walkWidgets(root, func(widget Widget) {
		persistable, ok := widget.(Persistable)
		if !ok || persistable == nil {
			return
		}
		key := widgetKey(widget)
		if key == "" {
			return
		}
		data, err := persistable.MarshalState()
		if err != nil {
			if firstErr == nil {
				firstErr = fmt.Errorf("persist %s: %w", key, err)
			}
			return
		}
		snapshot.Widgets[key] = data
	})
	return snapshot, firstErr
}

// ApplyState walks the widget tree and restores Persistable state by widget key.
func ApplyState(root Widget, snapshot PersistSnapshot) error {
	if snapshot.Widgets == nil {
		return nil
	}
	var firstErr error
	walkWidgets(root, func(widget Widget) {
		persistable, ok := widget.(Persistable)
		if !ok || persistable == nil {
			return
		}
		key := widgetKey(widget)
		if key == "" {
			return
		}
		data, ok := snapshot.Widgets[key]
		if !ok {
			return
		}
		if err := persistable.UnmarshalState(data); err != nil {
			if firstErr == nil {
				firstErr = fmt.Errorf("restore %s: %w", key, err)
			}
		}
	})
	return firstErr
}

// SaveSnapshot writes a snapshot to disk as JSON.
func SaveSnapshot(path string, snapshot PersistSnapshot) error {
	payload, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return fmt.Errorf("encode snapshot: %w", err)
	}
	return os.WriteFile(path, payload, 0o600)
}

// LoadSnapshot reads a snapshot from disk.
func LoadSnapshot(path string) (PersistSnapshot, error) {
	payload, err := os.ReadFile(path)
	if err != nil {
		return PersistSnapshot{}, fmt.Errorf("read snapshot: %w", err)
	}
	var snapshot PersistSnapshot
	if err := json.Unmarshal(payload, &snapshot); err != nil {
		return PersistSnapshot{}, fmt.Errorf("decode snapshot: %w", err)
	}
	if snapshot.Widgets == nil {
		snapshot.Widgets = map[string]json.RawMessage{}
	}
	return snapshot, nil
}

func walkWidgets(root Widget, fn func(Widget)) {
	visited := map[Widget]struct{}{}
	var walk func(Widget)
	walk = func(node Widget) {
		if node == nil {
			return
		}
		if _, ok := visited[node]; ok {
			return
		}
		visited[node] = struct{}{}
		fn(node)
		if container, ok := node.(ChildProvider); ok {
			for _, child := range container.ChildWidgets() {
				walk(child)
			}
		}
	}
	walk(root)
}

func widgetKey(widget Widget) string {
	if widget == nil {
		return ""
	}
	if keyed, ok := widget.(Keyed); ok {
		if key := strings.TrimSpace(keyed.Key()); key != "" {
			return key
		}
	}
	if ider, ok := widget.(interface{ ID() string }); ok {
		return strings.TrimSpace(ider.ID())
	}
	return ""
}
