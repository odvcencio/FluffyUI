package widgets

import (
	"fmt"
	"strings"
	"time"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/animation"
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/state"
	"github.com/odvcencio/fluffyui/terminal"
)

// Accordion groups collapsible sections.
type Accordion struct {
	FocusableBase
	sections      []*AccordionSection
	allowMultiple bool
	selected      int
	label         string

	style         backend.Style
	headerStyle   backend.Style
	selectedStyle backend.Style
	disabledStyle backend.Style

	services runtime.Services
	subs     state.Subscriptions
}

// AccordionSection is a single collapsible section.
type AccordionSection struct {
	title    *state.Signal[string]
	content  runtime.Widget
	expanded *state.Signal[bool]
	disabled *state.Signal[bool]

	animationDuration time.Duration
	easing            animation.EasingFunc

	visible float64
	target  float64

	headerBounds  runtime.Rect
	contentBounds runtime.Rect
}

// AccordionOption configures an accordion.
type AccordionOption = Option[Accordion]

// AccordionSectionOption configures a section.
type AccordionSectionOption = Option[AccordionSection]

// NewAccordion creates an accordion.
func NewAccordion(sections ...*AccordionSection) *Accordion {
	a := &Accordion{
		sections:      sections,
		allowMultiple: false,
		selected:      0,
		label:         "Accordion",
		style:         backend.DefaultStyle(),
		headerStyle:   backend.DefaultStyle(),
		selectedStyle: backend.DefaultStyle().Reverse(true),
		disabledStyle: backend.DefaultStyle().Dim(true),
	}
	a.Base.Role = accessibility.RoleGroup
	a.syncA11y()
	a.enforceSingleExpanded()
	return a
}

// NewAccordionSection creates a section.
func NewAccordionSection(title string, content runtime.Widget, opts ...AccordionSectionOption) *AccordionSection {
	sec := &AccordionSection{
		title:             state.NewSignal(strings.TrimSpace(title)),
		content:           content,
		expanded:          state.NewSignal(false),
		disabled:          state.NewSignal(false),
		animationDuration: 200 * time.Millisecond,
		easing:            animation.OutCubic,
	}
	for _, opt := range opts {
		opt(sec)
	}
	return sec
}

// Title returns the section title.
func (s *AccordionSection) Title() string {
	if s == nil || s.title == nil {
		return ""
	}
	return s.title.Get()
}

// SetTitle updates the section title.
func (s *AccordionSection) SetTitle(title string) {
	if s == nil || s.title == nil {
		return
	}
	s.title.Set(strings.TrimSpace(title))
}

// Content returns the section content widget.
func (s *AccordionSection) Content() runtime.Widget {
	if s == nil {
		return nil
	}
	return s.content
}

// SetContent updates the section content widget.
func (s *AccordionSection) SetContent(content runtime.Widget) {
	if s == nil {
		return
	}
	s.content = content
}

// Expanded reports whether the section is expanded.
func (s *AccordionSection) Expanded() bool {
	if s == nil || s.expanded == nil {
		return false
	}
	return s.expanded.Get()
}

// SetExpanded updates the expanded state.
func (s *AccordionSection) SetExpanded(expanded bool) {
	if s == nil || s.expanded == nil {
		return
	}
	s.expanded.Set(expanded)
}

// Disabled reports whether the section is disabled.
func (s *AccordionSection) Disabled() bool {
	if s == nil || s.disabled == nil {
		return false
	}
	return s.disabled.Get()
}

// SetDisabled updates the disabled state.
func (s *AccordionSection) SetDisabled(disabled bool) {
	if s == nil || s.disabled == nil {
		return
	}
	s.disabled.Set(disabled)
}

// WithSectionExpanded sets initial expanded state.
func WithSectionExpanded(expanded bool) AccordionSectionOption {
	return func(s *AccordionSection) {
		if s != nil && s.expanded != nil {
			s.expanded.Set(expanded)
		}
	}
}

// WithSectionDisabled sets initial disabled state.
func WithSectionDisabled(disabled bool) AccordionSectionOption {
	return func(s *AccordionSection) {
		if s != nil && s.disabled != nil {
			s.disabled.Set(disabled)
		}
	}
}

// WithSectionAnimation sets animation configuration.
func WithSectionAnimation(duration time.Duration, easing animation.EasingFunc) AccordionSectionOption {
	return func(s *AccordionSection) {
		if s == nil {
			return
		}
		if duration >= 0 {
			s.animationDuration = duration
		}
		if easing != nil {
			s.easing = easing
		}
	}
}

// SetSections updates accordion sections.
func (a *Accordion) SetSections(sections ...*AccordionSection) {
	if a == nil {
		return
	}
	a.sections = sections
	if a.selected >= len(a.sections) {
		a.selected = max(0, len(a.sections)-1)
	}
	a.enforceSingleExpanded()
	a.refreshSubscriptions()
	a.invalidate()
}

// AddSection appends a section.
func (a *Accordion) AddSection(section *AccordionSection) {
	if a == nil || section == nil {
		return
	}
	a.sections = append(a.sections, section)
	a.enforceSingleExpanded()
	a.refreshSubscriptions()
	a.invalidate()
}

// SetAllowMultiple toggles multiple expansion.
func (a *Accordion) SetAllowMultiple(allow bool) {
	if a == nil {
		return
	}
	a.allowMultiple = allow
	a.enforceSingleExpanded()
	a.invalidate()
}

// SetLabel updates the accessibility label.
func (a *Accordion) SetLabel(label string) {
	if a == nil {
		return
	}
	a.label = label
	a.syncA11y()
}

// SetSelected moves selection to the provided index.
func (a *Accordion) SetSelected(index int) {
	if a == nil {
		return
	}
	a.moveSelectionTo(index)
}

// SetStyles updates styles.
func (a *Accordion) SetStyles(base, header, selected, disabled backend.Style) {
	if a == nil {
		return
	}
	a.style = base
	a.headerStyle = header
	a.selectedStyle = selected
	a.disabledStyle = disabled
}

// StyleType returns the selector type name.
func (a *Accordion) StyleType() string {
	return "Accordion"
}

// Bind attaches app services.
func (a *Accordion) Bind(services runtime.Services) {
	if a == nil {
		return
	}
	a.services = services
	a.refreshSubscriptions()
}

// Unbind releases app services.
func (a *Accordion) Unbind() {
	if a == nil {
		return
	}
	a.subs.Clear()
	a.services = runtime.Services{}
}

// Measure returns desired size.
func (a *Accordion) Measure(constraints runtime.Constraints) runtime.Size {
	return a.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		width := contentConstraints.MaxWidth
		if width <= 0 {
			width = contentConstraints.MinWidth
		}
		height := 0
		for _, section := range a.sections {
			height += 1
			if section == nil || section.content == nil {
				continue
			}
			if section.expanded != nil && section.expanded.Get() {
				size := section.content.Measure(runtime.Constraints{
					MinWidth:  width,
					MaxWidth:  width,
					MinHeight: 0,
					MaxHeight: contentConstraints.MaxHeight,
				})
				height += size.Height
			}
		}
		return contentConstraints.Constrain(runtime.Size{Width: width, Height: height})
	})
}

// Layout positions headers and content.
func (a *Accordion) Layout(bounds runtime.Rect) {
	a.FocusableBase.Layout(bounds)
	content := a.ContentBounds()
	y := content.Y
	for i, section := range a.sections {
		if section == nil {
			continue
		}
		headerBounds := runtime.Rect{X: content.X, Y: y, Width: content.Width, Height: 1}
		section.headerBounds = headerBounds
		y++

		contentHeight := 0
		if section.content != nil {
			size := section.content.Measure(runtime.Constraints{
				MinWidth:  content.Width,
				MaxWidth:  content.Width,
				MinHeight: 0,
				MaxHeight: max(0, content.Height-(y-content.Y)),
			})
			contentHeight = size.Height
		}
		target := 0
		if section.expanded != nil && section.expanded.Get() {
			target = contentHeight
		}
		a.syncSectionTarget(section, float64(target))
		visible := int(section.visible)
		if visible < 0 {
			visible = 0
		}
		if visible > contentHeight {
			visible = contentHeight
		}
		section.contentBounds = runtime.Rect{X: content.X, Y: y, Width: content.Width, Height: visible}
		if section.content != nil {
			section.content.Layout(section.contentBounds)
		}
		y += visible
		if i == len(a.sections)-1 {
			break
		}
	}
}

// Render draws headers and visible content.
func (a *Accordion) Render(ctx runtime.RenderContext) {
	if a == nil {
		return
	}
	a.syncA11y()
	outer := a.bounds
	if outer.Width <= 0 || outer.Height <= 0 {
		return
	}
	baseStyle := mergeBackendStyles(resolveBaseStyle(ctx, a, backend.DefaultStyle(), false), a.style)
	headerStyle := mergeBackendStyles(baseStyle, a.headerStyle)
	selectedStyle := mergeBackendStyles(baseStyle, a.selectedStyle)
	disabledStyle := mergeBackendStyles(baseStyle, a.disabledStyle)
	ctx.Buffer.Fill(outer, ' ', baseStyle)

	for idx, section := range a.sections {
		if section == nil {
			continue
		}
		style := headerStyle
		if section.disabled != nil && section.disabled.Get() {
			style = disabledStyle
		} else if a.focused && idx == a.selected {
			style = selectedStyle
		}
		title := ""
		if section.title != nil {
			title = section.title.Get()
		}
		icon := ">"
		if section.expanded != nil && section.expanded.Get() {
			icon = "v"
		}
		label := fmt.Sprintf("%s %s", icon, title)
		writePadded(ctx.Buffer, section.headerBounds.X, section.headerBounds.Y, section.headerBounds.Width, label, style)
		if section.contentBounds.Height > 0 {
			runtime.RenderChild(ctx, section.content)
		}
	}
}

// HandleMessage handles navigation and toggling.
func (a *Accordion) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if a == nil {
		return runtime.Unhandled()
	}
	switch m := msg.(type) {
	case runtime.MouseMsg:
		if m.Action == runtime.MousePress && m.Button == runtime.MouseLeft {
			for i, section := range a.sections {
				if section == nil || section.headerBounds.Height == 0 {
					continue
				}
				if section.headerBounds.Contains(m.X, m.Y) {
					a.selected = i
					a.toggleSection(i)
					return runtime.Handled()
				}
			}
		}
	case runtime.KeyMsg:
		if !a.focused {
			return runtime.Unhandled()
		}
		switch m.Key {
		case terminal.KeyUp:
			a.moveSelection(-1)
			return runtime.Handled()
		case terminal.KeyDown:
			a.moveSelection(1)
			return runtime.Handled()
		case terminal.KeyHome:
			a.moveSelectionTo(0)
			return runtime.Handled()
		case terminal.KeyEnd:
			a.moveSelectionTo(len(a.sections) - 1)
			return runtime.Handled()
		case terminal.KeyLeft:
			a.setExpanded(a.selected, false)
			return runtime.Handled()
		case terminal.KeyRight:
			a.setExpanded(a.selected, true)
			return runtime.Handled()
		case terminal.KeyEnter:
			a.toggleSection(a.selected)
			return runtime.Handled()
		case terminal.KeyRune:
			if m.Rune == ' ' {
				a.toggleSection(a.selected)
				return runtime.Handled()
			}
		}
	}
	return runtime.Unhandled()
}

// ChildWidgets returns section content widgets.
func (a *Accordion) ChildWidgets() []runtime.Widget {
	if a == nil {
		return nil
	}
	children := []runtime.Widget{}
	for _, section := range a.sections {
		if section != nil && section.content != nil {
			children = append(children, section.content)
		}
	}
	return children
}

// PathSegment returns a debug path segment for the given child.
func (a *Accordion) PathSegment(child runtime.Widget) string {
	if a == nil {
		return "Accordion"
	}
	for i, section := range a.sections {
		if section == nil || section.content != child {
			continue
		}
		title := ""
		if section.title != nil {
			title = strings.TrimSpace(section.title.Get())
		}
		if title != "" {
			return fmt.Sprintf("Accordion[%s]", title)
		}
		return fmt.Sprintf("Accordion[%d]", i)
	}
	return "Accordion"
}

// ToggleSection toggles a section by index.
func (a *Accordion) ToggleSection(index int) {
	a.toggleSection(index)
}

func (a *Accordion) toggleSection(index int) {
	if a == nil {
		return
	}
	section := a.sectionAt(index)
	if section == nil || section.disabled != nil && section.disabled.Get() {
		return
	}
	expanded := section.expanded != nil && section.expanded.Get()
	a.setExpanded(index, !expanded)
}

func (a *Accordion) setExpanded(index int, expanded bool) {
	if a == nil {
		return
	}
	section := a.sectionAt(index)
	if section == nil || section.expanded == nil {
		return
	}
	if !a.allowMultiple && expanded {
		for i, s := range a.sections {
			if s == nil || s.expanded == nil {
				continue
			}
			if i != index && s.expanded.Get() {
				s.expanded.Set(false)
			}
		}
	}
	section.expanded.Set(expanded)
	a.invalidate()
}

func (a *Accordion) moveSelection(delta int) {
	if a == nil || len(a.sections) == 0 {
		return
	}
	index := a.selected
	for i := 0; i < len(a.sections); i++ {
		index = (index + delta + len(a.sections)) % len(a.sections)
		if !a.isDisabled(index) {
			a.selected = index
			a.invalidate()
			return
		}
	}
}

func (a *Accordion) moveSelectionTo(index int) {
	if a == nil || len(a.sections) == 0 {
		return
	}
	if index < 0 {
		index = 0
	}
	if index >= len(a.sections) {
		index = len(a.sections) - 1
	}
	if a.isDisabled(index) {
		return
	}
	a.selected = index
	a.invalidate()
}

func (a *Accordion) isDisabled(index int) bool {
	section := a.sectionAt(index)
	return section != nil && section.disabled != nil && section.disabled.Get()
}

func (a *Accordion) sectionAt(index int) *AccordionSection {
	if a == nil || index < 0 || index >= len(a.sections) {
		return nil
	}
	return a.sections[index]
}

func (a *Accordion) enforceSingleExpanded() {
	if a == nil || a.allowMultiple {
		return
	}
	keep := -1
	if current := a.sectionAt(a.selected); current != nil && current.expanded != nil && current.expanded.Get() {
		keep = a.selected
	} else {
		for i, section := range a.sections {
			if section != nil && section.expanded != nil && section.expanded.Get() {
				keep = i
				break
			}
		}
	}
	for i, section := range a.sections {
		if section == nil || section.expanded == nil {
			continue
		}
		if i != keep && section.expanded.Get() {
			section.expanded.Set(false)
		}
	}
}

func (a *Accordion) syncSectionTarget(section *AccordionSection, target float64) {
	if section == nil {
		return
	}
	if section.target == target && section.visible == target {
		return
	}
	if section.target != target {
		section.target = target
		a.animateSection(section, target)
	}
	if section.visible == 0 && target > 0 && a.services.Animator() == nil {
		section.visible = target
	}
}

func (a *Accordion) animateSection(section *AccordionSection, target float64) {
	if a == nil || section == nil {
		return
	}
	if a.services.ReducedMotion() || section.animationDuration <= 0 {
		section.visible = target
		return
	}
	animator := a.services.Animator()
	if animator == nil {
		section.visible = target
		return
	}
	easing := section.easing
	if easing == nil {
		easing = animation.OutCubic
	}
	animator.Animate(section, "height", func() animation.Animatable {
		return animation.Float64(section.visible)
	}, func(value animation.Animatable) {
		section.visible = float64(value.(animation.Float64))
		a.services.Relayout()
	}, animation.Float64(target), animation.TweenConfig{
		Duration: section.animationDuration,
		Easing:   easing,
	})
}

func (a *Accordion) refreshSubscriptions() {
	if a == nil {
		return
	}
	a.subs.Clear()
	a.subs.SetScheduler(a.services.Scheduler())
	for _, section := range a.sections {
		if section == nil {
			continue
		}
		a.subs.Observe(section.title, func() { a.invalidate() })
		a.subs.Observe(section.expanded, func() { a.invalidate() })
		a.subs.Observe(section.disabled, func() { a.invalidate() })
	}
}

func (a *Accordion) syncA11y() {
	if a == nil {
		return
	}
	if a.Base.Role == "" {
		a.Base.Role = accessibility.RoleGroup
	}
	label := strings.TrimSpace(a.label)
	if label == "" {
		label = "Accordion"
	}
	a.Base.Label = label
	if sel := a.sectionAt(a.selected); sel != nil && sel.title != nil {
		a.Base.Value = &accessibility.ValueInfo{Text: sel.title.Get()}
	}
}

func (a *Accordion) invalidate() {
	if a == nil {
		return
	}
	a.Invalidate()
	a.services.Invalidate()
}

var _ runtime.Widget = (*Accordion)(nil)
var _ runtime.Focusable = (*Accordion)(nil)
var _ runtime.ChildProvider = (*Accordion)(nil)
var _ runtime.Bindable = (*Accordion)(nil)
var _ runtime.Unbindable = (*Accordion)(nil)
