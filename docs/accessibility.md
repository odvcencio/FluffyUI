# Accessibility

FluffyUI treats accessibility as a first-class constraint. Widgets can expose
roles, labels, and state, and the runtime can announce focus changes.

## Accessible interface

Widgets that implement `accessibility.Accessible` provide metadata for screen
readers and logs:

```go
type Accessible interface {
    AccessibleRole() Role
    AccessibleLabel() string
    AccessibleDescription() string
    AccessibleState() StateSet
    AccessibleValue() *ValueInfo
}
```

Most input widgets embed `accessibility.Base`, which supplies these fields.

## Roles and state

Use roles like `RoleButton`, `RoleCheckbox`, or `RoleTextbox` to describe
semantics. The `StateSet` captures selection, checked, disabled, and other
status flags.

## Announcer

`accessibility.Announcer` is a central place to publish changes. The default
`SimpleAnnouncer` keeps a history in memory and can notify listeners.

```go
announcer := &accessibility.SimpleAnnouncer{}
announcer.SetOnMessage(func(msg accessibility.Announcement) {
    fmt.Println("announce:", msg.Message)
})
```

The screen announces focus changes automatically when an announcer is set in
`runtime.AppConfig`.

## Focus indicators

Focus styling is configured at the app level:

```go
app := runtime.NewApp(runtime.AppConfig{
    FocusStyle: &accessibility.FocusStyle{
        Indicator: "> ",
        Style:     backend.DefaultStyle().Bold(true),
    },
})
```

Use a short ASCII indicator so it remains visible across terminal fonts.

## Testing announcements

For tests, `accessibility.SimpleAnnouncer` keeps a history you can assert on.
The `testing` helpers expose this directly:

```go
announcer := fluffytest.NewAnnouncer()
app := runtime.NewApp(runtime.AppConfig{
    Announcer: announcer,
    // ...
})

// ... run app ...

fluffytest.AssertAnnounced(t, announcer, "Line 3 of 10")
```
