package widgets

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/odvcencio/fluffy-ui/accessibility"
	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/runtime"
	"github.com/odvcencio/fluffy-ui/terminal"
)

// SearchWidget provides a search input overlay for the chat view.
type SearchWidget struct {
	FocusableBase

	query        string
	matchCount   int
	currentMatch int
	label        string

	// Callbacks
	onSearch func(query string)
	onClose  func()
	onNext   func()
	onPrev   func()

	// Styles
	bgStyle     backend.Style
	borderStyle backend.Style
	textStyle   backend.Style
	matchStyle  backend.Style
}

// NewSearchWidget creates a new search widget.
func NewSearchWidget() *SearchWidget {
	w := &SearchWidget{
		label:       "Search",
		bgStyle:     backend.DefaultStyle(),
		borderStyle: backend.DefaultStyle(),
		textStyle:   backend.DefaultStyle(),
		matchStyle:  backend.DefaultStyle().Foreground(backend.ColorYellow),
	}
	w.Base.Role = accessibility.RoleTextbox
	w.syncA11y()
	return w
}

// SetOnSearch sets the search callback.
func (s *SearchWidget) SetOnSearch(fn func(query string)) {
	s.onSearch = fn
}

// SetOnClose sets the close callback.
func (s *SearchWidget) SetOnClose(fn func()) {
	s.onClose = fn
}

// SetOnNavigate sets callbacks for navigating search matches.
func (s *SearchWidget) SetOnNavigate(next, prev func()) {
	s.onNext = next
	s.onPrev = prev
}

// SetStyles configures appearance.
func (s *SearchWidget) SetStyles(bg, border, text, match backend.Style) {
	s.bgStyle = bg
	s.borderStyle = border
	s.textStyle = text
	s.matchStyle = match
}

// SetMatchInfo updates the match count display.
func (s *SearchWidget) SetMatchInfo(current, total int) {
	s.currentMatch = current
	s.matchCount = total
	s.syncA11y()
}

// SetLabel updates the accessibility label.
func (s *SearchWidget) SetLabel(label string) {
	s.label = label
	s.syncA11y()
}

// Query returns the current search query.
func (s *SearchWidget) Query() string {
	return s.query
}

// Measure returns the preferred size (fixed height bar).
func (s *SearchWidget) Measure(constraints runtime.Constraints) runtime.Size {
	return runtime.Size{
		Width:  constraints.MaxWidth,
		Height: 1,
	}
}

// Layout positions at the bottom of the screen.
func (s *SearchWidget) Layout(bounds runtime.Rect) {
	newBounds := runtime.Rect{
		X:      bounds.X,
		Y:      bounds.Y + bounds.Height - 1,
		Width:  bounds.Width,
		Height: 1,
	}
	s.Base.Layout(newBounds)
}

// Render draws the search bar.
func (s *SearchWidget) Render(ctx runtime.RenderContext) {
	b := s.bounds
	buf := ctx.Buffer
	s.syncA11y()

	// Fill background
	for x := b.X; x < b.X+b.Width; x++ {
		buf.Set(x, b.Y, ' ', s.bgStyle)
	}

	// Draw "/ " prefix
	buf.SetString(b.X, b.Y, "/ ", s.borderStyle)

	// Draw query
	queryX := b.X + 2
	maxQuery := b.Width - 20
	query := s.query
	if len(query) > maxQuery {
		query = query[len(query)-maxQuery:]
	}
	buf.SetString(queryX, b.Y, query, s.textStyle)

	// Draw cursor
	cursorX := queryX + len(query)
	if cursorX < b.X+b.Width-15 && s.focused {
		buf.Set(cursorX, b.Y, '█', s.textStyle)
	}

	// Draw match count on the right
	if s.matchCount > 0 {
		matchInfo := strconv.Itoa(s.currentMatch+1) + "/" + strconv.Itoa(s.matchCount)
		infoX := b.X + b.Width - len(matchInfo) - 2
		buf.SetString(infoX, b.Y, matchInfo, s.matchStyle)
	} else if s.query != "" {
		noMatch := "No matches"
		infoX := b.X + b.Width - len(noMatch) - 2
		buf.SetString(infoX, b.Y, noMatch, s.matchStyle)
	}
}

// HandleMessage processes keyboard input.
func (s *SearchWidget) HandleMessage(msg runtime.Message) runtime.HandleResult {
	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		return runtime.Unhandled()
	}

	switch key.Key {
	case terminal.KeyEscape:
		s.query = ""
		s.syncA11y()
		if s.onSearch != nil {
			s.onSearch("")
		}
		return runtime.WithCommand(runtime.PopOverlay{})

	case terminal.KeyEnter:
		// Close search bar but keep highlighting
		return runtime.WithCommand(runtime.PopOverlay{})

	case terminal.KeyUp:
		if s.onPrev != nil {
			s.onPrev()
		}
		return runtime.Handled()

	case terminal.KeyDown:
		if s.onNext != nil {
			s.onNext()
		}
		return runtime.Handled()

	case terminal.KeyBackspace:
		if len(s.query) > 0 {
			s.query = s.query[:len(s.query)-1]
			s.syncA11y()
			if s.onSearch != nil {
				s.onSearch(s.query)
			}
		}
		return runtime.Handled()

	case terminal.KeyRune:
		s.query += string(key.Rune)
		s.syncA11y()
		if s.onSearch != nil {
			s.onSearch(s.query)
		}
		return runtime.Handled()
	}

	return runtime.Unhandled()
}

func (s *SearchWidget) syncA11y() {
	if s == nil {
		return
	}
	if s.Base.Role == "" {
		s.Base.Role = accessibility.RoleTextbox
	}
	label := strings.TrimSpace(s.label)
	if label == "" {
		label = "Search"
	}
	s.Base.Label = label
	s.Base.Value = &accessibility.ValueInfo{Text: s.query}
	if s.matchCount > 0 {
		s.Base.Description = fmt.Sprintf("%d of %d", s.currentMatch+1, s.matchCount)
	} else if s.query != "" {
		s.Base.Description = "no matches"
	} else {
		s.Base.Description = ""
	}
}
