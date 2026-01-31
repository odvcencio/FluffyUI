package forms

import "strings"

// FieldType identifies a form field type.
type FieldType string

const (
	FieldText       FieldType = "text"
	FieldNumber     FieldType = "number"
	FieldEmail      FieldType = "email"
	FieldPassword   FieldType = "password"
	FieldCheckbox   FieldType = "checkbox"
	FieldSelect     FieldType = "select"
	FieldMultiSelect FieldType = "multiselect"
	FieldDate       FieldType = "date"
	FieldTime       FieldType = "time"
)

// FieldSpec describes a form field for builders and renderers.
type FieldSpec struct {
	Name        string
	Label       string
	Type        FieldType
	Placeholder string
	Options     []string
	Initial     any
	Validators  []Validator
}

// Builder constructs forms using a fluent DSL.
type Builder struct {
	fields     []FieldSpec
	validators []FormValidator
}

// NewBuilder creates a new form builder.
func NewBuilder() *Builder {
	return &Builder{}
}

// Field adds a custom field spec.
func (b *Builder) Field(spec FieldSpec) *Builder {
	if b == nil {
		return b
	}
	spec.Name = strings.TrimSpace(spec.Name)
	if spec.Name == "" {
		return b
	}
	b.fields = append(b.fields, spec)
	return b
}

// Text adds a text field.
func (b *Builder) Text(name, label, initial string, validators ...Validator) *Builder {
	return b.Field(FieldSpec{
		Name:       name,
		Label:      label,
		Type:       FieldText,
		Initial:    initial,
		Validators: validators,
	})
}

// Number adds a numeric field (float64).
func (b *Builder) Number(name, label string, initial float64, validators ...Validator) *Builder {
	return b.Field(FieldSpec{
		Name:       name,
		Label:      label,
		Type:       FieldNumber,
		Initial:    initial,
		Validators: validators,
	})
}

// Email adds an email field.
func (b *Builder) Email(name, label, initial string, validators ...Validator) *Builder {
	return b.Field(FieldSpec{
		Name:       name,
		Label:      label,
		Type:       FieldEmail,
		Initial:    initial,
		Validators: validators,
	})
}

// Password adds a password field.
func (b *Builder) Password(name, label, initial string, validators ...Validator) *Builder {
	return b.Field(FieldSpec{
		Name:       name,
		Label:      label,
		Type:       FieldPassword,
		Initial:    initial,
		Validators: validators,
	})
}

// Checkbox adds a boolean field.
func (b *Builder) Checkbox(name, label string, initial bool, validators ...Validator) *Builder {
	return b.Field(FieldSpec{
		Name:       name,
		Label:      label,
		Type:       FieldCheckbox,
		Initial:    initial,
		Validators: validators,
	})
}

// Select adds a single-choice select field.
func (b *Builder) Select(name, label string, options []string, initial string, validators ...Validator) *Builder {
	return b.Field(FieldSpec{
		Name:       name,
		Label:      label,
		Type:       FieldSelect,
		Options:    append([]string{}, options...),
		Initial:    initial,
		Validators: validators,
	})
}

// MultiSelect adds a multi-select field.
func (b *Builder) MultiSelect(name, label string, options []string, initial []string, validators ...Validator) *Builder {
	return b.Field(FieldSpec{
		Name:       name,
		Label:      label,
		Type:       FieldMultiSelect,
		Options:    append([]string{}, options...),
		Initial:    append([]string{}, initial...),
		Validators: validators,
	})
}

// Date adds a date field.
func (b *Builder) Date(name, label string, initial any, validators ...Validator) *Builder {
	return b.Field(FieldSpec{
		Name:       name,
		Label:      label,
		Type:       FieldDate,
		Initial:    initial,
		Validators: validators,
	})
}

// Time adds a time field.
func (b *Builder) Time(name, label string, initial any, validators ...Validator) *Builder {
	return b.Field(FieldSpec{
		Name:       name,
		Label:      label,
		Type:       FieldTime,
		Initial:    initial,
		Validators: validators,
	})
}

// Validator adds a form-level validator.
func (b *Builder) Validator(validator FormValidator) *Builder {
	if b == nil || validator == nil {
		return b
	}
	b.validators = append(b.validators, validator)
	return b
}

// Build constructs a Form and returns the field specs used.
func (b *Builder) Build() (*Form, []FieldSpec) {
	if b == nil {
		return NewForm(), nil
	}
	fields := make([]FieldSpec, 0, len(b.fields))
	formFields := make([]Field, 0, len(b.fields))
	for _, spec := range b.fields {
		name := strings.TrimSpace(spec.Name)
		if name == "" {
			continue
		}
		fields = append(fields, spec)
		formFields = append(formFields, NewField(name, spec.Initial, spec.Validators...))
	}
	form := NewForm(formFields...)
	for _, validator := range b.validators {
		form.AddValidator(validator)
	}
	return form, fields
}
