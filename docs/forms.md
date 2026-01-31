# Forms

The `forms` package provides field state, validation, and form-level
coordination. It is UI-agnostic so it can be wired to any widget.

## Fields and validators

```go
name := forms.NewField("name", "", forms.Required("Name is required"))
email := forms.NewField("email", "", forms.Email("Invalid email"))

form := forms.NewForm(name, email)
```

## Updating values

Wire input widgets to form fields:

```go
input.SetOnChange(func(text string) {
    form.Set("name", text)
})
```

## Submit and validation

```go
form.OnSubmit(func(values forms.Values) {
    fmt.Println("submit:", values)
})

form.Submit()
errors := form.Validate()
```

## Cross-field validation

```go
form.AddValidator(forms.FieldsMatch("password", "confirm", "Passwords must match"))
```

## Builder DSL

Use the fluent builder when defining forms declaratively:

```go
builder := forms.NewBuilder().
    Text("name", "Name", "", forms.Required("Name required")).
    Email("email", "Email", "", forms.Email("Invalid email")).
    Checkbox("tos", "Terms", false)

form, specs := builder.Build()
```

The returned `specs` slice can be used to drive custom form renderers.

See `examples/settings-form` for a full example.
