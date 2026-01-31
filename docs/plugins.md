# Plugin System

FluffyUI exposes a lightweight registry for third-party widgets. Plugins can
register themselves at init time and expose a factory for tooling or catalogs.

## Registering a plugin

```go
func init() {
    widgets.MustRegisterWidgetPlugin(widgets.WidgetPlugin{
        ID:          "acme.spark",
        Name:        "Acme Spark",
        Version:     "1.0.0",
        Description: "Custom sparkline widget",
        Categories:  []string{"charts"},
        New: func() runtime.Widget {
            return acme.NewSpark()
        },
    })
}
```

## Enumerating plugins

```go
for _, plugin := range widgets.WidgetPlugins() {
    fmt.Println(plugin.ID, plugin.Name)
}
```

This registry is intentionally simple and designed to work without Go's
`plugin` package (which is not supported on all platforms).
