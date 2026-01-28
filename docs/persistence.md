# Persistence

FluffyUI can capture and restore widget state via `runtime.Persistable`.

## Implement Persistable

```go
type MyForm struct {
    widgets.Component
    value *state.Signal[string]
}

func (f *MyForm) MarshalState() ([]byte, error) {
    return json.Marshal(map[string]any{
        "value": f.value.Get(),
    })
}

func (f *MyForm) UnmarshalState(data []byte) error {
    var payload struct{ Value string }
    if err := json.Unmarshal(data, &payload); err != nil {
        return err
    }
    f.value.Set(payload.Value)
    return nil
}
```

## Capture and restore

```go
snapshot, _ := runtime.CaptureState(root)
_ = runtime.SaveSnapshot("state.json", snapshot)

loaded, _ := runtime.LoadSnapshot("state.json")
_ = runtime.ApplyState(root, loaded)
```

Widgets are keyed by `runtime.Keyed` or `Base.ID`. Assign stable IDs to
persist state across runs.
