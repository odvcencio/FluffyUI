package fur

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// Handler implements slog.Handler with fur formatting.
type Handler struct {
	opts    HandlerOpts
	console *Console
	attrs   []slog.Attr
	groups  []string
}

// HandlerOpts configures the slog handler.
type HandlerOpts struct {
	Level      slog.Leveler
	TimeFormat string
	ShowSource bool
	ShowTime   bool
	Pretty     bool
}

// NewHandler creates a new slog handler.
func NewHandler(opts HandlerOpts) *Handler {
	if opts.Level == nil {
		opts.Level = slog.LevelInfo
	}
	if opts.TimeFormat == "" {
		opts.TimeFormat = "15:04:05"
	}
	return &Handler{opts: opts, console: Default()}
}

// Enabled reports whether the handler handles the given level.
func (h *Handler) Enabled(_ context.Context, level slog.Level) bool {
	minLevel := slog.LevelInfo
	if h != nil && h.opts.Level != nil {
		minLevel = h.opts.Level.Level()
	}
	return level >= minLevel
}

// Handle logs the record.
func (h *Handler) Handle(_ context.Context, r slog.Record) error {
	if h == nil {
		return nil
	}
	if !h.Enabled(context.Background(), r.Level) {
		return nil
	}
	c := h.console
	if c == nil {
		c = Default()
	}
	prefixText, prefixSpans := h.prefix(r)
	prefixWidth := stringWidth(prefixText)
	messageLines := wrapLines(splitTextLines(r.Message, DefaultStyle()), max(10, c.Width()-prefixWidth))

	var lines []Line
	for i, line := range messageLines {
		var combined Line
		if i == 0 {
			combined = append(combined, prefixSpans...)
		} else {
			appendSpan(&combined, Span{Text: strings.Repeat(" ", prefixWidth), Style: DefaultStyle()})
		}
		combined = append(combined, line...)
		lines = append(lines, combined)
	}

	attrsText := strings.TrimSpace(strings.Join(h.formatAttrs(r), " "))
	if attrsText != "" {
		if h.opts.Pretty {
			attrLines := wrapLines(splitTextLines(attrsText, DefaultStyle()), max(10, c.Width()-prefixWidth))
			for _, line := range attrLines {
				var combined Line
				appendSpan(&combined, Span{Text: strings.Repeat(" ", prefixWidth), Style: DefaultStyle()})
				combined = append(combined, line...)
				lines = append(lines, combined)
			}
		} else if len(lines) > 0 {
			appendSpan(&lines[0], Span{Text: " " + attrsText, Style: DefaultStyle()})
		}
	}

	c.writeLines(lines, true)
	return nil
}

// WithAttrs returns a new handler with additional attributes.
func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if h == nil {
		return h
	}
	clone := *h
	clone.attrs = append(append([]slog.Attr{}, h.attrs...), attrs...)
	return &clone
}

// WithGroup returns a new handler with a group name.
func (h *Handler) WithGroup(name string) slog.Handler {
	if h == nil {
		return h
	}
	if name == "" {
		return h
	}
	clone := *h
	clone.groups = append(append([]string{}, h.groups...), name)
	return &clone
}

func (h *Handler) prefix(r slog.Record) (string, Line) {
	var text strings.Builder
	var spans Line
	if h.opts.ShowTime {
		timestamp := r.Time
		if timestamp.IsZero() {
			timestamp = time.Now()
		}
		stamp := timestamp.Format(h.opts.TimeFormat)
		text.WriteString(stamp)
		text.WriteByte(' ')
		appendSpan(&spans, Span{Text: stamp, Style: h.mutedStyle()})
		appendSpan(&spans, Span{Text: " ", Style: DefaultStyle()})
	}
	level := strings.ToUpper(r.Level.String())
	text.WriteString(level)
	text.WriteByte(' ')
	appendSpan(&spans, Span{Text: level, Style: h.levelStyle(r.Level)})
	appendSpan(&spans, Span{Text: " ", Style: DefaultStyle()})

	if h.opts.ShowSource {
		if src := recordSource(r); src != "" {
			text.WriteString(src)
			text.WriteByte(' ')
			appendSpan(&spans, Span{Text: src, Style: h.mutedStyle()})
			appendSpan(&spans, Span{Text: " ", Style: DefaultStyle()})
		}
	}
	return text.String(), spans
}

func (h *Handler) formatAttrs(r slog.Record) []string {
	var attrs []slog.Attr
	attrs = append(attrs, h.attrs...)
	r.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, a)
		return true
	})
	prefix := strings.Join(h.groups, ".")
	if prefix != "" {
		prefix += "."
	}
	var out []string
	for _, attr := range attrs {
		flattenAttr(prefix, attr, &out)
	}
	return out
}

func flattenAttr(prefix string, attr slog.Attr, out *[]string) {
	attr.Value = attr.Value.Resolve()
	if attr.Value.Kind() == slog.KindGroup {
		groupPrefix := prefix + attr.Key + "."
		for _, child := range attr.Value.Group() {
			flattenAttr(groupPrefix, child, out)
		}
		return
	}
	key := prefix + attr.Key
	value := formatValue(attr.Value)
	*out = append(*out, fmt.Sprintf("%s=%s", key, value))
}

func formatValue(value slog.Value) string {
	switch value.Kind() {
	case slog.KindString:
		return strconv.Quote(value.String())
	case slog.KindInt64:
		return fmt.Sprintf("%d", value.Int64())
	case slog.KindUint64:
		return fmt.Sprintf("%d", value.Uint64())
	case slog.KindFloat64:
		return fmt.Sprintf("%g", value.Float64())
	case slog.KindBool:
		return fmt.Sprintf("%t", value.Bool())
	case slog.KindDuration:
		return value.Duration().String()
	case slog.KindTime:
		return value.Time().Format(time.RFC3339)
	case slog.KindAny:
		return fmt.Sprintf("%v", value.Any())
	default:
		return fmt.Sprintf("%v", value.Any())
	}
}

func recordSource(r slog.Record) string {
	if r.PC == 0 {
		return ""
	}
	fn := runtime.FuncForPC(r.PC)
	if fn == nil {
		return ""
	}
	file, line := fn.FileLine(r.PC)
	if file == "" {
		return ""
	}
	return fmt.Sprintf("%s:%d", shortFile(file), line)
}

func (h *Handler) levelStyle(level slog.Level) Style {
	if h != nil && h.console != nil && h.console.theme != nil {
		switch {
		case level >= slog.LevelError:
			return FromCompositor(h.console.theme.Error)
		case level >= slog.LevelWarn:
			return FromCompositor(h.console.theme.Warning)
		case level >= slog.LevelInfo:
			return FromCompositor(h.console.theme.Info)
		default:
			return FromCompositor(h.console.theme.TextMuted)
		}
	}
	switch {
	case level >= slog.LevelError:
		return DefaultStyle().Foreground(ColorRed)
	case level >= slog.LevelWarn:
		return DefaultStyle().Foreground(ColorYellow)
	case level >= slog.LevelInfo:
		return DefaultStyle().Foreground(ColorCyan)
	default:
		return Dim
	}
}

func (h *Handler) mutedStyle() Style {
	if h != nil && h.console != nil && h.console.theme != nil {
		return FromCompositor(h.console.theme.TextMuted)
	}
	return Dim
}
