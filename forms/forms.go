// Package forms provides form state management and validation.
package forms

import (
	"reflect"
	"strings"

	"github.com/odvcencio/fluffy-ui/state"
)

// Values is a map of form field values.
type Values map[string]any

// ValidationError describes a validation failure.
type ValidationError struct {
	Field   string
	Message string
}

// Field represents a form field.
type Field interface {
	Name() string
	Value() any
	SetValue(any)

	Dirty() bool
	Touched() bool
	Valid() bool
	Errors() []string

	Validate() []ValidationError
	Reset()
}

// FieldBase provides common field state management.
type FieldBase struct {
	name       string
	initial    any
	validators []Validator
	dirty      *state.Signal[bool]
	touched    *state.Signal[bool]
	errors     *state.Signal[[]string]
}

// NewFieldBase constructs a field base.
func NewFieldBase(name string, initial any, validators ...Validator) FieldBase {
	return FieldBase{
		name:       strings.TrimSpace(name),
		initial:    initial,
		validators: validators,
		dirty:      state.NewSignal(false),
		touched:    state.NewSignal(false),
		errors:     state.NewSignal([]string{}),
	}
}

// Name returns the field name.
func (f *FieldBase) Name() string {
	if f == nil {
		return ""
	}
	return f.name
}

// Dirty reports whether the field value differs from its initial value.
func (f *FieldBase) Dirty() bool {
	if f == nil || f.dirty == nil {
		return false
	}
	return f.dirty.Get()
}

// Touched reports whether the field has been modified.
func (f *FieldBase) Touched() bool {
	if f == nil || f.touched == nil {
		return false
	}
	return f.touched.Get()
}

// Errors returns the current validation errors.
func (f *FieldBase) Errors() []string {
	if f == nil || f.errors == nil {
		return nil
	}
	return f.errors.Get()
}

// Valid reports whether the field has no validation errors.
func (f *FieldBase) Valid() bool {
	return len(f.Errors()) == 0
}

// SetInitial updates the initial value used for dirty tracking.
func (f *FieldBase) SetInitial(value any) {
	if f == nil {
		return
	}
	f.initial = value
}

// ResetState clears dirty/touched/errors.
func (f *FieldBase) ResetState() {
	if f == nil {
		return
	}
	if f.dirty != nil {
		f.dirty.Set(false)
	}
	if f.touched != nil {
		f.touched.Set(false)
	}
	if f.errors != nil {
		f.errors.Set([]string{})
	}
}

// SetValidators updates the validators.
func (f *FieldBase) SetValidators(validators ...Validator) {
	if f == nil {
		return
	}
	f.validators = validators
}

// MarkTouched marks the field as touched.
func (f *FieldBase) MarkTouched() {
	if f == nil || f.touched == nil {
		return
	}
	f.touched.Set(true)
}

// UpdateDirty updates dirty state based on the provided value.
func (f *FieldBase) UpdateDirty(value any) {
	if f == nil || f.dirty == nil {
		return
	}
	f.dirty.Set(!reflect.DeepEqual(f.initial, value))
}

// SetErrors updates the error list.
func (f *FieldBase) SetErrors(messages []string) {
	if f == nil || f.errors == nil {
		return
	}
	if messages == nil {
		messages = []string{}
	}
	f.errors.Set(messages)
}

// ValidateValue validates a value using the field validators.
func (f *FieldBase) ValidateValue(value any) []ValidationError {
	if f == nil {
		return nil
	}
	var out []ValidationError
	for _, validator := range f.validators {
		if validator == nil {
			continue
		}
		if err := validator.Validate(value); err != nil {
			if err.Field == "" {
				err.Field = f.name
			}
			out = append(out, *err)
		}
	}
	return out
}

// SimpleField is a basic field implementation storing its own value.
type SimpleField struct {
	FieldBase
	value any
}

// NewField creates a new simple field.
func NewField(name string, initial any, validators ...Validator) *SimpleField {
	base := NewFieldBase(name, initial, validators...)
	return &SimpleField{
		FieldBase: base,
		value:     initial,
	}
}

// Value returns the current value.
func (f *SimpleField) Value() any {
	if f == nil {
		return nil
	}
	return f.value
}

// SetValue updates the value and dirty state.
func (f *SimpleField) SetValue(value any) {
	if f == nil {
		return
	}
	f.value = value
	f.UpdateDirty(value)
	f.MarkTouched()
	f.updateErrors()
}

// Validate runs validation on the current value.
func (f *SimpleField) Validate() []ValidationError {
	if f == nil {
		return nil
	}
	errs := f.ValidateValue(f.value)
	f.setErrorsFromValidation(errs)
	return errs
}

// Reset restores the initial value.
func (f *SimpleField) Reset() {
	if f == nil {
		return
	}
	f.value = f.initial
	f.ResetState()
}

func (f *SimpleField) updateErrors() {
	errs := f.ValidateValue(f.value)
	f.setErrorsFromValidation(errs)
}

func (f *SimpleField) setErrorsFromValidation(errs []ValidationError) {
	if f == nil {
		return
	}
	if len(errs) == 0 {
		f.SetErrors([]string{})
		return
	}
	messages := make([]string, 0, len(errs))
	for _, err := range errs {
		messages = append(messages, err.Message)
	}
	f.SetErrors(messages)
}

// Form manages a collection of fields.
type Form struct {
	fields     map[string]Field
	order      []string
	validators []FormValidator
	onSubmit   func(values Values)
	onCancel   func()

	dirty      *state.Signal[bool]
	valid      *state.Signal[bool]
	submitting *state.Signal[bool]
}

// NewForm constructs an empty form.
func NewForm(fields ...Field) *Form {
	form := &Form{
		fields:     make(map[string]Field),
		dirty:      state.NewSignal(false),
		valid:      state.NewSignal(true),
		submitting: state.NewSignal(false),
	}
	for _, field := range fields {
		form.AddField(field)
	}
	return form
}

// AddField registers a field.
func (f *Form) AddField(field Field) {
	if f == nil || field == nil || strings.TrimSpace(field.Name()) == "" {
		return
	}
	name := field.Name()
	if f.fields == nil {
		f.fields = make(map[string]Field)
	}
	if _, exists := f.fields[name]; !exists {
		f.order = append(f.order, name)
	}
	f.fields[name] = field
	f.updateDirty()
}

// AddValidator registers a cross-field validator.
func (f *Form) AddValidator(validator FormValidator) {
	if f == nil || validator == nil {
		return
	}
	f.validators = append(f.validators, validator)
}

// OnSubmit sets the submit callback.
func (f *Form) OnSubmit(fn func(values Values)) {
	if f == nil {
		return
	}
	f.onSubmit = fn
}

// OnCancel sets the cancel callback.
func (f *Form) OnCancel(fn func()) {
	if f == nil {
		return
	}
	f.onCancel = fn
}

// Field returns a field by name.
func (f *Form) Field(name string) Field {
	if f == nil {
		return nil
	}
	return f.fields[name]
}

// Get retrieves a field value.
func (f *Form) Get(name string) any {
	if f == nil {
		return nil
	}
	field := f.Field(name)
	if field == nil {
		return nil
	}
	return field.Value()
}

// Set updates a field value.
func (f *Form) Set(name string, value any) {
	if f == nil {
		return
	}
	field := f.Field(name)
	if field == nil {
		return
	}
	field.SetValue(value)
	f.updateDirty()
}

// Values returns a snapshot of form values.
func (f *Form) Values() Values {
	if f == nil {
		return nil
	}
	values := make(Values, len(f.fields))
	for name, field := range f.fields {
		values[name] = field.Value()
	}
	return values
}

// Submit validates and invokes the submit callback.
func (f *Form) Submit() {
	if f == nil {
		return
	}
	f.submitting.Set(true)
	errors := f.Validate()
	if len(errors) == 0 && f.onSubmit != nil {
		f.onSubmit(f.Values())
	}
	f.submitting.Set(false)
}

// Cancel invokes the cancel callback.
func (f *Form) Cancel() {
	if f == nil {
		return
	}
	if f.onCancel != nil {
		f.onCancel()
	}
}

// Reset resets all fields to their initial values.
func (f *Form) Reset() {
	if f == nil {
		return
	}
	for _, name := range f.order {
		if field := f.fields[name]; field != nil {
			field.Reset()
		}
	}
	f.updateDirty()
	f.valid.Set(true)
}

// Validate validates all fields and form validators.
func (f *Form) Validate() []ValidationError {
	if f == nil {
		return nil
	}
	var errors []ValidationError
	for _, name := range f.order {
		field := f.fields[name]
		if field == nil {
			continue
		}
		errors = append(errors, field.Validate()...)
	}
	values := f.Values()
	for _, validator := range f.validators {
		if validator == nil {
			continue
		}
		errors = append(errors, validator.Validate(values)...)
	}
	f.valid.Set(len(errors) == 0)
	return errors
}

// DirtySignal returns the form dirty signal.
func (f *Form) DirtySignal() *state.Signal[bool] {
	if f == nil {
		return nil
	}
	return f.dirty
}

// ValidSignal returns the form valid signal.
func (f *Form) ValidSignal() *state.Signal[bool] {
	if f == nil {
		return nil
	}
	return f.valid
}

// SubmittingSignal returns the form submitting signal.
func (f *Form) SubmittingSignal() *state.Signal[bool] {
	if f == nil {
		return nil
	}
	return f.submitting
}

func (f *Form) updateDirty() {
	if f == nil || f.dirty == nil {
		return
	}
	dirty := false
	for _, field := range f.fields {
		if field != nil && field.Dirty() {
			dirty = true
			break
		}
	}
	f.dirty.Set(dirty)
}
