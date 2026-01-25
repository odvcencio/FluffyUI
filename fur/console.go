package fur

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/odvcencio/fluffy-ui/compositor"
	"golang.org/x/term"
)

// Console provides styled terminal output.
type Console struct {
	out     io.Writer
	width   int
	theme   *Theme
	markup  *MarkupParser
	noColor bool
	mu      sync.Mutex
}

// Option configures a Console.
type Option func(*Console)

// Default returns the default console (stdout).
func Default() *Console {
	return defaultConsole
}

// New creates a console with options.
func New(opts ...Option) *Console {
	c := &Console{
		out:    os.Stdout,
		width:  0,
		theme:  DefaultTheme(),
		markup: DefaultMarkupParser(),
	}
	for _, opt := range opts {
		if opt != nil {
			opt(c)
		}
	}
	return c
}

var defaultConsole = New()

// WithOutput sets the output writer.
func WithOutput(w io.Writer) Option {
	return func(c *Console) {
		if w != nil {
			c.out = w
		}
	}
}

// WithWidth sets the console width.
func WithWidth(w int) Option {
	return func(c *Console) {
		c.width = w
	}
}

// WithTheme sets the console theme.
func WithTheme(t *Theme) Option {
	return func(c *Console) {
		c.theme = t
	}
}

// WithNoColor disables ANSI styling output.
func WithNoColor() Option {
	return func(c *Console) {
		c.noColor = true
	}
}

// Print outputs text with word wrapping.
func (c *Console) Print(a ...any) {
	c.printMarkup(fmt.Sprint(a...), false)
}

// Println outputs text with newline.
func (c *Console) Println(a ...any) {
	c.printMarkup(fmt.Sprint(a...), true)
}

// Printf outputs formatted text.
func (c *Console) Printf(format string, a ...any) {
	c.printMarkup(fmt.Sprintf(format, a...), false)
}

// Log outputs with timestamp and caller location.
func (c *Console) Log(a ...any) {
	c.logMessage(fmt.Sprint(a...))
}

// Logf outputs formatted log message.
func (c *Console) Logf(format string, a ...any) {
	c.logMessage(fmt.Sprintf(format, a...))
}

// Rule prints a horizontal divider.
func (c *Console) Rule(title ...string) {
	c.Render(Rule(title...))
}

// Render outputs any Renderable.
func (c *Console) Render(r Renderable) {
	if c == nil || r == nil {
		return
	}
	lines := r.Render(c.Width())
	c.writeLines(lines, true)
}

// Clear clears the terminal screen.
func (c *Console) Clear() {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	_, _ = io.WriteString(c.out, compositor.ANSIClearScreen+compositor.ANSICursorHome)
}

// Width returns terminal width.
func (c *Console) Width() int {
	if c == nil {
		return 80
	}
	if c.width > 0 {
		return c.width
	}
	if detected := detectWidth(c.out); detected > 0 {
		return detected
	}
	return 80
}

func (c *Console) printMarkup(text string, newline bool) {
	if c == nil {
		return
	}
	parser := c.markup
	if parser == nil {
		parser = DefaultMarkupParser()
	}
	lines := parser.Parse(text)
	lines = wrapLines(lines, c.Width())
	c.writeLines(lines, newline)
}

func (c *Console) logMessage(msg string) {
	if c == nil {
		return
	}
	stamp := time.Now().Format("15:04:05")
	source := callerLocation()
	prefix := stamp
	if source != "" {
		prefix += " " + source
	}
	prefix += " "
	prefixWidth := stringWidth(prefix)
	messageLines := wrapLines(splitTextLines(msg, DefaultStyle()), max(10, c.Width()-prefixWidth))
	muted := c.themeMutedStyle()
	var lines []Line
	for i, line := range messageLines {
		var combined Line
		if i == 0 {
			appendSpan(&combined, Span{Text: stamp, Style: muted})
			if source != "" {
				appendSpan(&combined, Span{Text: " ", Style: DefaultStyle()})
				appendSpan(&combined, Span{Text: source, Style: muted})
			}
			appendSpan(&combined, Span{Text: " ", Style: DefaultStyle()})
		} else {
			appendSpan(&combined, Span{Text: strings.Repeat(" ", prefixWidth), Style: DefaultStyle()})
		}
		combined = append(combined, line...)
		lines = append(lines, combined)
	}
	c.writeLines(lines, true)
}

func (c *Console) themeMutedStyle() Style {
	if c != nil && c.theme != nil {
		return FromCompositor(c.theme.TextMuted)
	}
	return Dim
}

func detectWidth(w io.Writer) int {
	if file, ok := w.(*os.File); ok {
		fd := int(file.Fd())
		if term.IsTerminal(fd) {
			if width, _, err := term.GetSize(fd); err == nil && width > 0 {
				return width
			}
		}
	}
	if env := os.Getenv("COLUMNS"); env != "" {
		if value, err := strconv.Atoi(env); err == nil && value > 0 {
			return value
		}
	}
	return 0
}

func (c *Console) writeLines(lines []Line, newlineAfterLast bool) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	var buf strings.Builder
	current := DefaultStyle()
	styleSet := false
	for i, line := range lines {
		for _, span := range line {
			if c.noColor {
				buf.WriteString(span.Text)
				continue
			}
			if !styleSet || !current.Equal(span.Style) {
				buf.WriteString(compositor.StyleToANSI(span.Style.ToCompositor()))
				current = span.Style
				styleSet = true
			}
			buf.WriteString(span.Text)
		}
		if !c.noColor && styleSet {
			buf.WriteString(compositor.ANSIReset)
			current = DefaultStyle()
			styleSet = false
		}
		if i < len(lines)-1 || newlineAfterLast {
			buf.WriteByte('\n')
		}
	}
	if len(lines) == 0 && newlineAfterLast {
		buf.WriteByte('\n')
	}
	_, _ = io.WriteString(c.out, buf.String())
}

func callerLocation() string {
	pcs := make([]uintptr, 16)
	n := runtime.Callers(3, pcs)
	frames := runtime.CallersFrames(pcs[:n])
	for {
		frame, more := frames.Next()
		if !strings.Contains(frame.Function, "/fur.") && !strings.Contains(frame.Function, ".fur.") && !strings.HasPrefix(frame.Function, "runtime.") {
			file := shortFile(frame.File)
			return fmt.Sprintf("%s:%d", file, frame.Line)
		}
		if !more {
			break
		}
	}
	return ""
}

func shortFile(path string) string {
	clean := filepath.ToSlash(path)
	parts := strings.Split(clean, "/")
	if len(parts) >= 2 {
		return parts[len(parts)-2] + "/" + parts[len(parts)-1]
	}
	return filepath.Base(path)
}
