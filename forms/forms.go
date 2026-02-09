// Package forms provides form state management and validation.
package forms

import (
	"context"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/odvcencio/fluffyui/state"
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
	name            string
	initial         any
	validators      []Validator
	asyncValidators []AsyncValidator
	dirty           *state.Signal[bool]
	touched         *state.Signal[bool]
	errors          *state.Signal[[]string]
	validating      *state.Signal[bool]
	validationID    atomic.Uint64
	errMu           sync.Mutex
	syncErrors      []string
	asyncErrors     []string
}

// NewFieldBase constructs a field base.
func NewFieldBase(name string, initial any, validators ...Validator) *FieldBase {
	return &FieldBase{
		name:       strings.TrimSpace(name),
		initial:    initial,
		validators: validators,
		dirty:      state.NewSignal(false),
		touched:    state.NewSignal(false),
		errors:     state.NewSignal([]string{}),
		validating: state.NewSignal(false),
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

// Validating reports whether async validation is running.
func (f *FieldBase) Validating() bool {
	if f == nil || f.validating == nil {
		return false
	}
	return f.validating.Get()
}

// ValidatingSignal returns the async validation signal.
func (f *FieldBase) ValidatingSignal() *state.Signal[bool] {
	if f == nil {
		return nil
	}
	return f.validating
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
	if f.validating != nil {
		f.validating.Set(false)
	}
	f.errMu.Lock()
	f.syncErrors = nil
	f.asyncErrors = nil
	f.errMu.Unlock()
}

// SetValidators updates the validators.
func (f *FieldBase) SetValidators(validators ...Validator) {
	if f == nil {
		return
	}
	f.validators = validators
}

// SetAsyncValidators updates async validators.
func (f *FieldBase) SetAsyncValidators(validators ...AsyncValidator) {
	if f == nil {
		return
	}
	if validators == nil {
		f.asyncValidators = nil
		return
	}
	f.asyncValidators = append([]AsyncValidator{}, validators...)
}

// AsyncValidators returns a copy of async validators.
func (f *FieldBase) AsyncValidators() []AsyncValidator {
	if f == nil || len(f.asyncValidators) == 0 {
		return nil
	}
	return append([]AsyncValidator{}, f.asyncValidators...)
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
	f.dirty.Set(!fieldDeepEqual(f.initial, value))
}

// fieldDeepEqual uses fast path for common types before falling back to reflect.DeepEqual.
func fieldDeepEqual(a, b any) bool {
	// Fast path: try direct comparison for common types.
	switch av := a.(type) {
	case int:
		if bv, ok := b.(int); ok {
			return av == bv
		}
		return false
	case string:
		if bv, ok := b.(string); ok {
			return av == bv
		}
		return false
	case bool:
		if bv, ok := b.(bool); ok {
			return av == bv
		}
		return false
	case float64:
		if bv, ok := b.(float64); ok {
			return av == bv
		}
		return false
	case int64:
		if bv, ok := b.(int64); ok {
			return av == bv
		}
		return false
	case int32:
		if bv, ok := b.(int32); ok {
			return av == bv
		}
		return false
	case uint:
		if bv, ok := b.(uint); ok {
			return av == bv
		}
		return false
	case uint64:
		if bv, ok := b.(uint64); ok {
			return av == bv
		}
		return false
	case uint32:
		if bv, ok := b.(uint32); ok {
			return av == bv
		}
		return false
	}
	// Fall back to reflect.DeepEqual for complex types
	return reflect.DeepEqual(a, b)
}

// SetErrors updates the error list.
func (f *FieldBase) SetErrors(messages []string) {
	if f == nil || f.errors == nil {
		return
	}
	if messages == nil {
		messages = []string{}
	}
	f.errMu.Lock()
	f.syncErrors = append([]string{}, messages...)
	f.asyncErrors = nil
	f.errMu.Unlock()
	f.errors.Set(messages)
}

func (f *FieldBase) setSyncErrors(errs []ValidationError) {
	if f == nil || f.errors == nil {
		return
	}
	messages := validationMessages(errs)
	f.errMu.Lock()
	f.syncErrors = messages
	combined := append([]string{}, f.syncErrors...)
	combined = append(combined, f.asyncErrors...)
	f.errMu.Unlock()
	f.errors.Set(combined)
}

func (f *FieldBase) setAsyncErrors(errs []ValidationError) {
	if f == nil || f.errors == nil {
		return
	}
	messages := validationMessages(errs)
	f.errMu.Lock()
	f.asyncErrors = messages
	combined := append([]string{}, f.syncErrors...)
	combined = append(combined, f.asyncErrors...)
	f.errMu.Unlock()
	f.errors.Set(combined)
}

func validationMessages(errs []ValidationError) []string {
	if len(errs) == 0 {
		return []string{}
	}
	messages := make([]string, 0, len(errs))
	for _, err := range errs {
		messages = append(messages, err.Message)
	}
	return messages
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

func (f *FieldBase) bumpValidationID() uint64 {
	if f == nil {
		return 0
	}
	return f.validationID.Add(1)
}

func (f *FieldBase) currentValidationID() uint64 {
	if f == nil {
		return 0
	}
	return f.validationID.Load()
}

func (f *FieldBase) clearAsyncErrors() {
	if f == nil || f.errors == nil {
		return
	}
	f.errMu.Lock()
	f.asyncErrors = nil
	combined := append([]string{}, f.syncErrors...)
	f.errMu.Unlock()
	f.errors.Set(combined)
}

func (f *FieldBase) setValidating(on bool) {
	if f == nil || f.validating == nil {
		return
	}
	f.validating.Set(on)
}

// SimpleField is a basic field implementation storing its own value.
type SimpleField struct {
	*FieldBase
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
	if f == nil || f.FieldBase == nil {
		return
	}
	f.value = value
	f.bumpValidationID()
	f.clearAsyncErrors()
	f.UpdateDirty(value)
	f.MarkTouched()
	f.updateErrors()
}

// Validate runs validation on the current value.
func (f *SimpleField) Validate() []ValidationError {
	if f == nil || f.FieldBase == nil {
		return nil
	}
	errs := f.ValidateValue(f.value)
	f.setErrorsFromValidation(errs)
	return errs
}

// Reset restores the initial value.
func (f *SimpleField) Reset() {
	if f == nil || f.FieldBase == nil {
		return
	}
	f.value = f.initial
	f.bumpValidationID()
	f.clearAsyncErrors()
	f.ResetState()
}

func (f *SimpleField) updateErrors() {
	if f == nil || f.FieldBase == nil {
		return
	}
	errs := f.ValidateValue(f.value)
	f.setErrorsFromValidation(errs)
}

func (f *SimpleField) setErrorsFromValidation(errs []ValidationError) {
	if f == nil || f.FieldBase == nil {
		return
	}
	f.setSyncErrors(errs)
}

// ValidateAsync runs async validators and updates errors when complete.
func (f *SimpleField) ValidateAsync(ctx context.Context) <-chan []ValidationError {
	out := make(chan []ValidationError, 1)
	if f == nil || f.FieldBase == nil {
		out <- nil
		return out
	}
	validators := f.AsyncValidators()
	if len(validators) == 0 {
		out <- nil
		return out
	}
	if ctx == nil {
		ctx = context.Background()
	}
	value := f.value
	validationID := f.currentValidationID()
	f.setValidating(true)

	go func() {
		defer f.setValidating(false)
		var errs []ValidationError
		for _, validator := range validators {
			if validator == nil {
				continue
			}
			if ctx.Err() != nil {
				break
			}
			if err := validator.ValidateAsync(ctx, value); err != nil {
				if err.Field == "" {
					err.Field = f.name
				}
				errs = append(errs, *err)
			}
		}
		if ctx.Err() == nil && f.currentValidationID() == validationID {
			f.setAsyncErrors(errs)
		}
		out <- errs
	}()

	return out
}

// Form manages a collection of fields.
type Form struct {
	fields          map[string]Field
	order           []string
	validators      []FormValidator
	asyncValidators []AsyncFormValidator
	onSubmit        func(values Values)
	onCancel        func()

	dirty      *state.Signal[bool]
	valid      *state.Signal[bool]
	submitting *state.Signal[bool]
	validating *state.Signal[bool]

	dependencies map[string][]string
	dependents   map[string][]string
}

// NewForm constructs an empty form.
func NewForm(fields ...Field) *Form {
	form := &Form{
		fields:       make(map[string]Field),
		dirty:        state.NewSignal(false),
		valid:        state.NewSignal(true),
		submitting:   state.NewSignal(false),
		validating:   state.NewSignal(false),
		dependencies: make(map[string][]string),
		dependents:   make(map[string][]string),
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

// AddAsyncValidator registers an async cross-field validator.
func (f *Form) AddAsyncValidator(validator AsyncFormValidator) {
	if f == nil || validator == nil {
		return
	}
	f.asyncValidators = append(f.asyncValidators, validator)
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
	f.revalidateDependents(name)
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

// SubmitAsync runs validation asynchronously and invokes onSubmit if valid.
func (f *Form) SubmitAsync(ctx context.Context) <-chan []ValidationError {
	out := make(chan []ValidationError, 1)
	if f == nil {
		out <- nil
		return out
	}
	f.submitting.Set(true)
	go func() {
		errs := <-f.ValidateAsync(ctx)
		if len(errs) == 0 && f.onSubmit != nil {
			f.onSubmit(f.Values())
		}
		f.submitting.Set(false)
		out <- errs
	}()
	return out
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

// ValidateAsync validates fields and async validators, returning a channel of errors.
func (f *Form) ValidateAsync(ctx context.Context) <-chan []ValidationError {
	out := make(chan []ValidationError, 1)
	if f == nil {
		out <- nil
		return out
	}
	if ctx == nil {
		ctx = context.Background()
	}

	f.validating.Set(true)
	syncErrors := f.Validate()
	values := f.Values()

	type asyncField interface {
		ValidateAsync(context.Context) <-chan []ValidationError
	}

	go func() {
		var mu sync.Mutex
		errors := append([]ValidationError{}, syncErrors...)
		var wg sync.WaitGroup

		for _, name := range f.order {
			field := f.fields[name]
			if field == nil {
				continue
			}
			if af, ok := field.(asyncField); ok {
				wg.Add(1)
				go func(ch <-chan []ValidationError) {
					defer wg.Done()
					if ch == nil {
						return
					}
					select {
					case <-ctx.Done():
						return
					case errs := <-ch:
						if len(errs) == 0 {
							return
						}
						mu.Lock()
						errors = append(errors, errs...)
						mu.Unlock()
					}
				}(af.ValidateAsync(ctx))
			}
		}

		for _, validator := range f.asyncValidators {
			if validator == nil {
				continue
			}
			wg.Add(1)
			go func(v AsyncFormValidator) {
				defer wg.Done()
				if ctx.Err() != nil {
					return
				}
				errs := v.ValidateAsync(ctx, values)
				if len(errs) == 0 {
					return
				}
				mu.Lock()
				errors = append(errors, errs...)
				mu.Unlock()
			}(validator)
		}

		wg.Wait()
		f.valid.Set(len(errors) == 0)
		f.validating.Set(false)
		out <- errors
	}()

	return out
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

// ValidatingSignal returns the async validating signal.
func (f *Form) ValidatingSignal() *state.Signal[bool] {
	if f == nil {
		return nil
	}
	return f.validating
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

// SetDependencies declares that a field depends on other fields.
func (f *Form) SetDependencies(field string, deps ...string) {
	if f == nil {
		return
	}
	field = strings.TrimSpace(field)
	if field == "" {
		return
	}
	if f.dependencies == nil {
		f.dependencies = make(map[string][]string)
	}

	seen := make(map[string]struct{}, len(deps))
	var out []string
	for _, dep := range deps {
		name := strings.TrimSpace(dep)
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}

	f.dependencies[field] = out
	f.rebuildDependents()
}

func (f *Form) rebuildDependents() {
	if f == nil {
		return
	}
	f.dependents = make(map[string][]string)
	for field, deps := range f.dependencies {
		for _, dep := range deps {
			f.dependents[dep] = append(f.dependents[dep], field)
		}
	}
}

func (f *Form) revalidateDependents(changed string) {
	if f == nil {
		return
	}
	changed = strings.TrimSpace(changed)
	if changed == "" || len(f.dependents) == 0 {
		return
	}
	dependents := f.dependents[changed]
	if len(dependents) == 0 {
		return
	}
	for _, name := range dependents {
		field := f.fields[name]
		if field == nil {
			continue
		}
		field.Validate()
		if af, ok := field.(interface {
			ValidateAsync(context.Context) <-chan []ValidationError
		}); ok {
			af.ValidateAsync(context.Background())
		}
	}
}
