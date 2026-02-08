package widgets

import (
	"fmt"
	"strings"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/terminal"
)

// Pagination is a focusable page navigation widget.
// It renders a page indicator like: < 1 2 [3] 4 5 ... 10 >
type Pagination struct {
	FocusableBase

	total    int
	current  int
	pageSize int
	onChange func(page int)
	services runtime.Services

	style    backend.Style
	styleSet bool
}

// NewPagination creates a pagination widget.
func NewPagination(total, pageSize int) *Pagination {
	if total < 0 {
		total = 0
	}
	if pageSize < 1 {
		pageSize = 1
	}
	p := &Pagination{
		total:    total,
		current:  1,
		pageSize: pageSize,
		style:    backend.DefaultStyle(),
	}
	p.Base.Role = accessibility.RoleGroup
	p.Base.Label = "Pagination"
	p.syncA11y()
	return p
}

// SetOnChange sets the page change handler.
func (p *Pagination) SetOnChange(fn func(page int)) {
	if p == nil {
		return
	}
	p.onChange = fn
}

// SetPage sets the current page (1-based).
func (p *Pagination) SetPage(page int) {
	if p == nil {
		return
	}
	totalPages := p.totalPages()
	if page < 1 {
		page = 1
	}
	if totalPages > 0 && page > totalPages {
		page = totalPages
	}
	old := p.current
	p.current = page
	p.syncA11y()
	if old != page {
		if p.onChange != nil {
			p.onChange(page)
		}
		if announcer := p.services.Announcer(); announcer != nil {
			announcer.Announce(fmt.Sprintf("Page %d of %d", page, totalPages), accessibility.PriorityPolite)
		}
	}
}

// Page returns the current page (1-based).
func (p *Pagination) Page() int {
	if p == nil {
		return 0
	}
	return p.current
}

// TotalPages returns the total number of pages.
func (p *Pagination) TotalPages() int {
	if p == nil {
		return 0
	}
	return p.totalPages()
}

func (p *Pagination) totalPages() int {
	if p.total <= 0 {
		return 0
	}
	pages := p.total / p.pageSize
	if p.total%p.pageSize != 0 {
		pages++
	}
	return pages
}

// SetStyle sets the widget style.
func (p *Pagination) SetStyle(style backend.Style) {
	if p == nil {
		return
	}
	p.style = style
	p.styleSet = true
}

// StyleType returns the selector type name.
func (p *Pagination) StyleType() string {
	return "Pagination"
}

// Bind attaches app services.
func (p *Pagination) Bind(services runtime.Services) {
	if p == nil {
		return
	}
	p.services = services
}

// Unbind releases app services.
func (p *Pagination) Unbind() {
	if p == nil {
		return
	}
	p.services = runtime.Services{}
}

// Measure returns the desired size.
func (p *Pagination) Measure(constraints runtime.Constraints) runtime.Size {
	return p.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		text := p.renderText()
		width := textWidth(text)
		if width < 1 {
			width = 1
		}
		return contentConstraints.Constrain(runtime.Size{Width: width, Height: 1})
	})
}

// Render draws the pagination widget.
func (p *Pagination) Render(ctx runtime.RenderContext) {
	if p == nil {
		return
	}
	p.syncA11y()
	bounds := p.ContentBounds()
	if bounds.Width <= 0 || bounds.Height <= 0 {
		return
	}
	style := resolveBaseStyle(ctx, p, p.style, p.styleSet)
	text := p.renderText()
	writePadded(ctx.Buffer, bounds.X, bounds.Y, bounds.Width, truncateString(text, bounds.Width), style)
}

func (p *Pagination) renderText() string {
	totalPages := p.totalPages()
	if totalPages <= 0 {
		return "< >"
	}
	var parts []string
	parts = append(parts, "<")

	// Show up to 5 page numbers around current, with ellipsis
	windowStart := p.current - 2
	if windowStart < 1 {
		windowStart = 1
	}
	windowEnd := windowStart + 4
	if windowEnd > totalPages {
		windowEnd = totalPages
		windowStart = windowEnd - 4
		if windowStart < 1 {
			windowStart = 1
		}
	}

	if windowStart > 1 {
		parts = append(parts, "1")
		if windowStart > 2 {
			parts = append(parts, "...")
		}
	}
	for i := windowStart; i <= windowEnd; i++ {
		if i == p.current {
			parts = append(parts, fmt.Sprintf("[%d]", i))
		} else {
			parts = append(parts, fmt.Sprintf("%d", i))
		}
	}
	if windowEnd < totalPages {
		if windowEnd < totalPages-1 {
			parts = append(parts, "...")
		}
		parts = append(parts, fmt.Sprintf("%d", totalPages))
	}

	parts = append(parts, ">")
	return strings.Join(parts, " ")
}

// HandleMessage processes keyboard input.
func (p *Pagination) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if p == nil || !p.focused {
		return runtime.Unhandled()
	}
	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		return runtime.Unhandled()
	}
	switch key.Key {
	case terminal.KeyLeft:
		if p.current > 1 {
			p.SetPage(p.current - 1)
			return runtime.Handled()
		}
	case terminal.KeyRight:
		if p.current < p.totalPages() {
			p.SetPage(p.current + 1)
			return runtime.Handled()
		}
	case terminal.KeyHome:
		p.SetPage(1)
		return runtime.Handled()
	case terminal.KeyEnd:
		p.SetPage(p.totalPages())
		return runtime.Handled()
	}
	return runtime.Unhandled()
}

func (p *Pagination) syncA11y() {
	if p == nil {
		return
	}
	if p.Base.Role == "" {
		p.Base.Role = accessibility.RoleGroup
	}
	totalPages := p.totalPages()
	p.Base.Label = fmt.Sprintf("Page %d of %d", p.current, totalPages)
	p.Base.PosInSet = p.current
	p.Base.SetSize = totalPages
}

var _ runtime.Widget = (*Pagination)(nil)
var _ runtime.Focusable = (*Pagination)(nil)
var _ runtime.Bindable = (*Pagination)(nil)
var _ runtime.Unbindable = (*Pagination)(nil)
