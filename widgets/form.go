package widgets

import (
	"fmt"
	"strings"

	"m31labs.dev/fluffyui/accessibility"
	"m31labs.dev/fluffyui/backend"
	"m31labs.dev/fluffyui/runtime"
	"m31labs.dev/fluffyui/terminal"
)

// validatable is a duck-typing interface for widgets that support validation.
// The Input widget already implements this.
type validatable interface {
	Valid() bool
	Errors() []string
}

// textProvider is a duck-typing interface for widgets that expose text content.
type textProvider interface {
	Text() string
}

// FormFieldDef defines a labeled form field.
type FormFieldDef struct {
	Label  string
	Widget runtime.Widget
}

// FormField creates a form field definition.
func FormField(label string, widget runtime.Widget) FormFieldDef {
	return FormFieldDef{Label: label, Widget: widget}
}

// formEntry holds the internal state for a single form field.
type formEntry struct {
	def      FormFieldDef
	labelW   *Label
	errorMsg string // current error (empty = valid)
}

// Form is a container widget that manages a set of labeled input fields
// with validation, error display, and submit handling.
type Form struct {
	Component
	fields      []formEntry
	onSubmit    func(map[string]string)
	onCancel    func()
	submitLabel string
	cancelLabel string
	focusIdx    int
	showErrors  bool
	label       string
	submitBtn   *Button
	cancelBtn   *Button
}

// NewForm creates a new form with the given field definitions.
func NewForm(fields ...FormFieldDef) *Form {
	f := &Form{
		submitLabel: "Submit",
		cancelLabel: "Cancel",
		label:       "Form",
	}
	f.Base.Role = accessibility.RoleForm

	for _, def := range fields {
		entry := formEntry{
			def:    def,
			labelW: NewLabel(def.Label + ":"),
		}
		f.fields = append(f.fields, entry)
	}

	f.submitBtn = NewButton(f.submitLabel, WithOnClick(func() {
		f.submit()
	}))
	f.cancelBtn = NewButton(f.cancelLabel, WithOnClick(func() {
		f.cancel()
	}))

	f.syncA11y()
	return f
}

// SetOnSubmit sets the callback invoked when the form is submitted.
// The callback receives a map of label -> text value for each field.
func (f *Form) SetOnSubmit(fn func(map[string]string)) {
	if f == nil {
		return
	}
	f.onSubmit = fn
}

// SetOnCancel sets the callback invoked when the form is cancelled.
func (f *Form) SetOnCancel(fn func()) {
	if f == nil {
		return
	}
	f.onCancel = fn
}

// SetSubmitLabel sets the label for the submit button.
func (f *Form) SetSubmitLabel(label string) {
	if f == nil {
		return
	}
	f.submitLabel = label
	if f.submitBtn != nil {
		f.submitBtn.SetLabel(label)
	}
}

// SetCancelLabel sets the label for the cancel button.
func (f *Form) SetCancelLabel(label string) {
	if f == nil {
		return
	}
	f.cancelLabel = label
	if f.cancelBtn != nil {
		f.cancelBtn.SetLabel(label)
	}
}

// SetLabel sets the accessibility label for the form.
func (f *Form) SetLabel(label string) {
	if f == nil {
		return
	}
	f.label = label
	f.syncA11y()
}

// FieldCount returns the number of fields in the form.
func (f *Form) FieldCount() int {
	if f == nil {
		return 0
	}
	return len(f.fields)
}

// Validate runs validators on all fields and returns true if all pass.
func (f *Form) Validate() bool {
	if f == nil {
		return true
	}
	allValid := true
	for i := range f.fields {
		entry := &f.fields[i]
		if v, ok := entry.def.Widget.(validatable); ok {
			if !v.Valid() {
				allValid = false
				errs := v.Errors()
				if len(errs) > 0 {
					entry.errorMsg = errs[0]
				} else {
					entry.errorMsg = "Invalid"
				}
			} else {
				entry.errorMsg = ""
			}
		} else {
			entry.errorMsg = ""
		}
	}
	f.showErrors = true
	return allValid
}

// Errors returns a map of field index to error message for fields with errors.
func (f *Form) Errors() map[int]string {
	if f == nil {
		return nil
	}
	errs := make(map[int]string)
	for i, entry := range f.fields {
		if entry.errorMsg != "" {
			errs[i] = entry.errorMsg
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return errs
}

// StyleType returns the selector type name for FSS styling.
func (f *Form) StyleType() string {
	return "Form"
}

// Bind attaches app services to the form and its children.
func (f *Form) Bind(services runtime.Services) {
	if f == nil {
		return
	}
	f.Component.Bind(services)
}

// Unbind releases app services from the form and its children.
func (f *Form) Unbind() {
	if f == nil {
		return
	}
	f.Component.Unbind()
}

// Measure returns the size needed for the form.
func (f *Form) Measure(constraints runtime.Constraints) runtime.Size {
	return f.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		width := 0
		height := 0

		for _, entry := range f.fields {
			// Label line
			labelWidth := textWidth(entry.labelW.text)
			if labelWidth > width {
				width = labelWidth
			}
			height++ // label

			// Widget line
			widgetSize := entry.def.Widget.Measure(contentConstraints)
			if widgetSize.Width > width {
				width = widgetSize.Width
			}
			height++ // widget

			// Error line (if showing errors and has error)
			if f.showErrors && entry.errorMsg != "" {
				errWidth := textWidth(entry.errorMsg)
				if errWidth > width {
					width = errWidth
				}
				height++ // error
			}

			// Blank line between fields
			height++
		}

		// Button row
		buttonWidth := textWidth("["+f.submitLabel+"]") + 1 + textWidth("["+f.cancelLabel+"]")
		if buttonWidth > width {
			width = buttonWidth
		}
		height++ // buttons

		return contentConstraints.Constrain(runtime.Size{
			Width:  width,
			Height: height,
		})
	})
}

// Layout positions the form and all its child widgets.
func (f *Form) Layout(bounds runtime.Rect) {
	f.Base.Layout(bounds)

	content := f.ContentBounds()
	if content.Width <= 0 || content.Height <= 0 {
		return
	}

	y := content.Y
	for i := range f.fields {
		entry := &f.fields[i]

		// Label
		entry.labelW.Layout(runtime.Rect{
			X: content.X, Y: y, Width: content.Width, Height: 1,
		})
		y++

		// Widget
		widgetSize := entry.def.Widget.Measure(runtime.Constraints{
			MinWidth: 0, MaxWidth: content.Width,
			MinHeight: 0, MaxHeight: max(0, content.Y+content.Height-y),
		})
		widgetHeight := widgetSize.Height
		if widgetHeight < 1 {
			widgetHeight = 1
		}
		entry.def.Widget.Layout(runtime.Rect{
			X: content.X, Y: y, Width: content.Width, Height: widgetHeight,
		})
		y += widgetHeight

		// Error line
		if f.showErrors && entry.errorMsg != "" {
			y++ // error takes 1 line
		}

		// Blank line
		y++
	}

	// Submit/Cancel buttons
	submitWidth := textWidth("[" + f.submitLabel + "]")
	cancelWidth := textWidth("[" + f.cancelLabel + "]")
	if y < content.Y+content.Height {
		f.submitBtn.Layout(runtime.Rect{
			X: content.X, Y: y, Width: submitWidth, Height: 1,
		})
		f.cancelBtn.Layout(runtime.Rect{
			X: content.X + submitWidth + 1, Y: y, Width: cancelWidth, Height: 1,
		})
	}
}

// Render draws the form to the buffer.
func (f *Form) Render(ctx runtime.RenderContext) {
	if f == nil {
		return
	}
	content := f.ContentBounds()
	if content.Width <= 0 || content.Height <= 0 {
		return
	}
	f.syncA11y()

	baseStyle := backend.DefaultStyle()

	y := content.Y
	for _, entry := range f.fields {
		if y >= content.Y+content.Height {
			break
		}

		// Render label
		runtime.RenderChild(ctx, entry.labelW)
		y++

		// Render widget
		runtime.RenderChild(ctx, entry.def.Widget)
		widgetBounds := runtime.Rect{}
		if bp, ok := entry.def.Widget.(runtime.BoundsProvider); ok {
			widgetBounds = bp.Bounds()
		}
		widgetHeight := widgetBounds.Height
		if widgetHeight < 1 {
			widgetHeight = 1
		}
		y += widgetHeight

		// Render error message
		if f.showErrors && entry.errorMsg != "" {
			if y < content.Y+content.Height {
				errStyle := baseStyle.Foreground(backend.ColorRed)
				errText := entry.errorMsg
				if textWidth(errText) > content.Width {
					errText = truncateString(errText, content.Width)
				}
				ctx.Buffer.SetString(content.X, y, errText, errStyle)
				y++
			}
		}

		// Blank line
		y++
	}

	// Render buttons
	runtime.RenderChild(ctx, f.submitBtn)
	runtime.RenderChild(ctx, f.cancelBtn)
}

// HandleMessage processes keyboard input for the form.
func (f *Form) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if f == nil {
		return runtime.Unhandled()
	}

	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		// Delegate non-key messages to the focused field
		if f.focusIdx >= 0 && f.focusIdx < len(f.fields) {
			return f.fields[f.focusIdx].def.Widget.HandleMessage(msg)
		}
		return runtime.Unhandled()
	}

	switch key.Key {
	case terminal.KeyTab:
		if key.Shift {
			f.focusPrev()
		} else {
			f.focusNext()
		}
		return runtime.Handled()

	case terminal.KeyEnter:
		// Submit on Enter unless it's a multiline widget
		f.submit()
		return runtime.Handled()

	case terminal.KeyEscape:
		f.cancel()
		return runtime.Handled()
	}

	// Delegate to focused field widget
	if f.focusIdx >= 0 && f.focusIdx < len(f.fields) {
		return f.fields[f.focusIdx].def.Widget.HandleMessage(msg)
	}

	return runtime.Unhandled()
}

// ChildWidgets returns all child widgets for focus traversal and binding.
func (f *Form) ChildWidgets() []runtime.Widget {
	if f == nil {
		return nil
	}
	children := make([]runtime.Widget, 0, len(f.fields)*2+2)
	for _, entry := range f.fields {
		children = append(children, entry.labelW)
		children = append(children, entry.def.Widget)
	}
	children = append(children, f.submitBtn, f.cancelBtn)
	return children
}

// PathSegment returns a debug path segment for the given child.
func (f *Form) PathSegment(child runtime.Widget) string {
	if f == nil {
		return "Form"
	}
	for i, entry := range f.fields {
		if entry.labelW == child {
			return "Form[label-" + entry.def.Label + "]"
		}
		if entry.def.Widget == child {
			return fmt.Sprintf("Form[field-%d]", i)
		}
	}
	if child == f.submitBtn {
		return "Form[submit]"
	}
	if child == f.cancelBtn {
		return "Form[cancel]"
	}
	return "Form"
}

func (f *Form) submit() {
	valid := f.Validate()

	if !valid {
		// Announce errors
		if announcer := f.Services.Announcer(); announcer != nil {
			var errParts []string
			for _, entry := range f.fields {
				if entry.errorMsg != "" {
					errParts = append(errParts, entry.def.Label+": "+entry.errorMsg)
				}
			}
			if len(errParts) > 0 {
				announcer.Announce(strings.Join(errParts, ". "), accessibility.PriorityAssertive)
			}
		}
		return
	}

	if f.onSubmit != nil {
		values := make(map[string]string, len(f.fields))
		for _, entry := range f.fields {
			if tp, ok := entry.def.Widget.(textProvider); ok {
				values[entry.def.Label] = tp.Text()
			} else {
				values[entry.def.Label] = ""
			}
		}
		f.onSubmit(values)
	}
}

func (f *Form) cancel() {
	if f.onCancel != nil {
		f.onCancel()
	}
}

func (f *Form) focusNext() {
	if len(f.fields) == 0 {
		return
	}
	// Blur current
	f.blurCurrent()
	f.focusIdx++
	if f.focusIdx >= len(f.fields) {
		f.focusIdx = 0
	}
	f.focusCurrent()
}

func (f *Form) focusPrev() {
	if len(f.fields) == 0 {
		return
	}
	f.blurCurrent()
	f.focusIdx--
	if f.focusIdx < 0 {
		f.focusIdx = len(f.fields) - 1
	}
	f.focusCurrent()
}

func (f *Form) blurCurrent() {
	if f.focusIdx >= 0 && f.focusIdx < len(f.fields) {
		if focusable, ok := f.fields[f.focusIdx].def.Widget.(runtime.Focusable); ok {
			focusable.Blur()
		}
	}
}

func (f *Form) focusCurrent() {
	if f.focusIdx >= 0 && f.focusIdx < len(f.fields) {
		if focusable, ok := f.fields[f.focusIdx].def.Widget.(runtime.Focusable); ok {
			focusable.Focus()
		}
	}
}

func (f *Form) syncA11y() {
	if f == nil {
		return
	}
	if f.Base.Role == "" {
		f.Base.Role = accessibility.RoleForm
	}
	label := strings.TrimSpace(f.label)
	if label == "" {
		label = "Form"
	}
	f.Base.Label = label
}

var _ runtime.Widget = (*Form)(nil)
var _ runtime.Bindable = (*Form)(nil)
var _ runtime.Unbindable = (*Form)(nil)
var _ runtime.ChildProvider = (*Form)(nil)
