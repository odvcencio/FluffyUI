package widgets

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"m31labs.dev/fluffyui/accessibility"
	"m31labs.dev/fluffyui/backend"
	"m31labs.dev/fluffyui/runtime"
	"m31labs.dev/fluffyui/scroll"
	"m31labs.dev/fluffyui/terminal"
)

// LogLevel represents the severity of a log entry.
type LogLevel int

const (
	// LogDebug is for debug-level messages.
	LogDebug LogLevel = iota
	// LogInfo is for informational messages.
	LogInfo
	// LogWarn is for warning messages.
	LogWarn
	// LogError is for error messages.
	LogError
)

// String returns a string representation of the log level.
func (l LogLevel) String() string {
	switch l {
	case LogDebug:
		return "DEBUG"
	case LogInfo:
		return "INFO"
	case LogWarn:
		return "WARN"
	case LogError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// LogEntry represents a single log message.
type LogEntry struct {
	Time    time.Time
	Level   LogLevel
	Message string
	Fields  map[string]any
}

// Format returns a formatted string representation of the entry.
func (e LogEntry) Format(showTime bool) string {
	var sb strings.Builder

	if showTime {
		sb.WriteString(e.Time.Format("15:04:05"))
		sb.WriteString(" ")
	}

	sb.WriteString("[")
	sb.WriteString(e.Level.String())
	sb.WriteString("] ")
	sb.WriteString(e.Message)

	if len(e.Fields) > 0 {
		sb.WriteString(" {")
		first := true
		for k, v := range e.Fields {
			if !first {
				sb.WriteString(", ")
			}
			first = false
			sb.WriteString(k)
			sb.WriteString("=")
			sb.WriteString(fmt.Sprint(v))
		}
		sb.WriteString("}")
	}

	return sb.String()
}

// Log is an auto-scrolling log viewer widget with filtering and level-based coloring.
type Log struct {
	FocusableBase

	maxLines    int                      // Ring buffer size
	autoScroll  bool                     // Scroll to bottom on new entries
	filter      func(LogEntry) bool      // Custom filter function
	levelColors map[LogLevel]backend.Style // Level-specific styles

	entries    []LogEntry // Ring buffer of entries
	head       int        // Index of oldest entry
	count      int        // Number of entries in buffer
	offset     int        // Scroll offset (0 = bottom in auto-scroll mode)
	showTime   bool       // Show timestamps
	minLevel   LogLevel   // Minimum level to display
	label         string
	style         backend.Style
	selectedStyle backend.Style
	services      runtime.Services

	mu sync.Mutex
}

// LogOption is a functional option for Log.
type LogOption func(*Log)

// WithMaxLines sets the maximum number of lines to keep.
func WithMaxLines(max int) LogOption {
	return func(l *Log) {
		if max < 1 {
			max = 1
		}
		l.maxLines = max
	}
}

// WithAutoScroll enables or disables auto-scrolling.
func WithAutoScroll(auto bool) LogOption {
	return func(l *Log) {
		l.autoScroll = auto
	}
}

// WithLogFilter sets a custom filter function.
func WithLogFilter(filter func(LogEntry) bool) LogOption {
	return func(l *Log) {
		l.filter = filter
	}
}

// WithLevelColors sets custom level-based colors.
func WithLevelColors(colors map[LogLevel]backend.Style) LogOption {
	return func(l *Log) {
		l.levelColors = colors
	}
}

// WithShowTime enables or disables timestamp display.
func WithShowTime(show bool) LogOption {
	return func(l *Log) {
		l.showTime = show
	}
}

// WithMinLevel sets the minimum level to display.
func WithMinLevel(level LogLevel) LogOption {
	return func(l *Log) {
		l.minLevel = level
	}
}

// WithLogStyle sets the base style.
func WithLogStyle(s backend.Style) LogOption {
	return func(l *Log) {
		l.style = s
	}
}

// NewLog creates a new log viewer widget.
func NewLog(opts ...LogOption) *Log {
	l := &Log{
		maxLines:   10000,
		autoScroll: true,
		showTime:   true,
		label:      "Log Viewer",
		style:      backend.DefaultStyle(),
		selectedStyle: backend.DefaultStyle().Reverse(true),
		levelColors: map[LogLevel]backend.Style{
			LogDebug: backend.DefaultStyle().Dim(true),
			LogInfo:  backend.DefaultStyle(),
			LogWarn:  backend.DefaultStyle().Foreground(backend.ColorYellow),
			LogError: backend.DefaultStyle().Foreground(backend.ColorRed),
		},
	}
	for _, opt := range opts {
		opt(l)
	}
	l.entries = make([]LogEntry, l.maxLines)
	l.Base.Role = accessibility.RoleLog
	l.Base.Live = accessibility.LivePolite
	l.Base.Relevant = accessibility.RelevantAdditions
	l.syncA11y()
	return l
}

// Bind attaches app services.
func (l *Log) Bind(services runtime.Services) {
	if l == nil {
		return
	}
	l.services = services
}

// Unbind releases app services.
func (l *Log) Unbind() {
	if l == nil {
		return
	}
	l.services = runtime.Services{}
}

// Write implements io.Writer, parsing lines as log entries.
func (l *Log) Write(p []byte) (n int, err error) {
	if l == nil {
		return len(p), nil
	}
	lines := strings.Split(string(p), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		l.Log(LogInfo, line)
	}
	return len(p), nil
}

// Log adds a new entry to the log.
func (l *Log) Log(level LogLevel, message string, fields ...map[string]any) {
	if l == nil {
		return
	}

	var fieldMap map[string]any
	if len(fields) > 0 {
		fieldMap = fields[0]
	}

	entry := LogEntry{
		Time:    time.Now(),
		Level:   level,
		Message: message,
		Fields:  fieldMap,
	}

	l.AddEntry(entry)
}

// Debug adds a debug-level entry.
func (l *Log) Debug(message string, fields ...map[string]any) {
	l.Log(LogDebug, message, fields...)
}

// Info adds an info-level entry.
func (l *Log) Info(message string, fields ...map[string]any) {
	l.Log(LogInfo, message, fields...)
}

// Warn adds a warning-level entry.
func (l *Log) Warn(message string, fields ...map[string]any) {
	l.Log(LogWarn, message, fields...)
}

// Error adds an error-level entry.
func (l *Log) Error(message string, fields ...map[string]any) {
	l.Log(LogError, message, fields...)
}

// AddEntry adds a log entry to the ring buffer.
func (l *Log) AddEntry(entry LogEntry) {
	if l == nil {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// Calculate the index for the new entry
	idx := (l.head + l.count) % l.maxLines
	l.entries[idx] = entry

	if l.count < l.maxLines {
		l.count++
	} else {
		// Buffer is full, advance head
		l.head = (l.head + 1) % l.maxLines
	}

	// Auto-scroll to bottom
	if l.autoScroll {
		l.offset = 0
	}

	// Announce new entries with priority based on level.
	msg := strings.TrimSpace(entry.Message)
	announcer := l.services.Announcer()
	level := entry.Level

	l.Invalidate()

	if msg != "" && announcer != nil {
		priority := accessibility.PriorityPolite
		if level >= LogError {
			priority = accessibility.PriorityAssertive
		}
		announcer.Announce(msg, priority)
	}
}

// Clear removes all entries from the log.
func (l *Log) Clear() {
	if l == nil {
		return
	}

	l.mu.Lock()
	l.head = 0
	l.count = 0
	l.offset = 0
	l.mu.Unlock()
	l.syncA11y()
}

// EntryCount returns the number of entries in the log.
func (l *Log) EntryCount() int {
	if l == nil {
		return 0
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	return l.count
}

// Entries returns a copy of all entries.
func (l *Log) Entries() []LogEntry {
	if l == nil {
		return nil
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	result := make([]LogEntry, l.count)
	for i := 0; i < l.count; i++ {
		idx := (l.head + i) % l.maxLines
		result[i] = l.entries[idx]
	}
	return result
}

// FilteredEntries returns entries that pass the current filter.
func (l *Log) FilteredEntries() []LogEntry {
	if l == nil {
		return nil
	}

	all := l.Entries()
	if l.filter == nil && l.minLevel == LogDebug {
		return all
	}

	var result []LogEntry
	for _, entry := range all {
		if entry.Level < l.minLevel {
			continue
		}
		if l.filter != nil && !l.filter(entry) {
			continue
		}
		result = append(result, entry)
	}
	return result
}

// SetFilter sets the filter function.
func (l *Log) SetFilter(filter func(LogEntry) bool) {
	if l == nil {
		return
	}
	l.filter = filter
	l.syncA11y()
}

// SetMinLevel sets the minimum level to display.
func (l *Log) SetMinLevel(level LogLevel) {
	if l == nil {
		return
	}
	l.minLevel = level
	l.syncA11y()
}

// MinLevel returns the minimum level.
func (l *Log) MinLevel() LogLevel {
	if l == nil {
		return LogDebug
	}
	return l.minLevel
}

// SetAutoScroll enables or disables auto-scrolling.
func (l *Log) SetAutoScroll(auto bool) {
	if l == nil {
		return
	}
	l.autoScroll = auto
}

// AutoScroll returns whether auto-scrolling is enabled.
func (l *Log) AutoScroll() bool {
	if l == nil {
		return false
	}
	return l.autoScroll
}

// SetShowTime enables or disables timestamp display.
func (l *Log) SetShowTime(show bool) {
	if l == nil {
		return
	}
	l.showTime = show
}

// ShowTime returns whether timestamps are shown.
func (l *Log) ShowTime() bool {
	if l == nil {
		return false
	}
	return l.showTime
}

// SetStyle sets the base style.
func (l *Log) SetStyle(style backend.Style) {
	if l == nil {
		return
	}
	l.style = style
}

// SetLevelStyle sets the style for a specific level.
func (l *Log) SetLevelStyle(level LogLevel, style backend.Style) {
	if l == nil {
		return
	}
	if l.levelColors == nil {
		l.levelColors = make(map[LogLevel]backend.Style)
	}
	l.levelColors[level] = style
}

// SetLabel sets the accessibility label.
func (l *Log) SetLabel(label string) {
	if l == nil {
		return
	}
	l.label = label
	l.syncA11y()
}

// StyleType returns the selector type name.
func (l *Log) StyleType() string {
	return "Log"
}

// Measure returns the desired size.
func (l *Log) Measure(constraints runtime.Constraints) runtime.Size {
	return l.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		height := contentConstraints.MaxHeight
		if height <= 0 {
			height = contentConstraints.MinHeight
		}
		return contentConstraints.Constrain(runtime.Size{Width: contentConstraints.MaxWidth, Height: height})
	})
}

// Layout stores the assigned bounds.
func (l *Log) Layout(bounds runtime.Rect) {
	if l == nil {
		return
	}
	l.FocusableBase.Layout(bounds)
}

// Render draws the log viewer.
func (l *Log) Render(ctx runtime.RenderContext) {
	if l == nil {
		return
	}
	l.syncA11y()
	outer := l.bounds
	content := l.ContentBounds()
	if outer.Width <= 0 || outer.Height <= 0 {
		return
	}

	baseStyle := mergeBackendStyles(resolveBaseStyle(ctx, l, backend.DefaultStyle(), false), l.style)
	ctx.Buffer.Fill(outer, ' ', baseStyle)

	if content.Width <= 0 || content.Height <= 0 {
		return
	}

	entries := l.FilteredEntries()
	if len(entries) == 0 {
		return
	}

	// Calculate visible range
	// offset is from the bottom (0 = showing latest entries)
	startIdx := len(entries) - content.Height - l.offset
	if startIdx < 0 {
		startIdx = 0
	}
	endIdx := startIdx + content.Height
	if endIdx > len(entries) {
		endIdx = len(entries)
	}

	for i := startIdx; i < endIdx; i++ {
		entry := entries[i]
		rowY := content.Y + (i - startIdx)

		// Get level-specific style
		style := baseStyle
		if levelStyle, ok := l.levelColors[entry.Level]; ok {
			style = mergeBackendStyles(baseStyle, levelStyle)
		}

		// Format and render the line
		line := entry.Format(l.showTime)
		line = truncateString(line, content.Width)
		writePadded(ctx.Buffer, content.X, rowY, content.Width, line, style)
	}
}

// HandleMessage handles keyboard input.
func (l *Log) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if l == nil || !l.focused {
		return runtime.Unhandled()
	}

	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		return runtime.Unhandled()
	}

	entries := l.FilteredEntries()
	viewportHeight := l.ContentBounds().Height
	if viewportHeight < 1 {
		viewportHeight = 1
	}
	maxOffset := len(entries) - viewportHeight
	if maxOffset < 0 {
		maxOffset = 0
	}

	switch key.Key {
	case terminal.KeyUp:
		l.offset++
		if l.offset > maxOffset {
			l.offset = maxOffset
		}
		l.autoScroll = false
		return runtime.Handled()

	case terminal.KeyDown:
		l.offset--
		if l.offset <= 0 {
			l.offset = 0
			l.autoScroll = true
		}
		return runtime.Handled()

	case terminal.KeyPageUp:
		l.offset += viewportHeight
		if l.offset > maxOffset {
			l.offset = maxOffset
		}
		l.autoScroll = false
		return runtime.Handled()

	case terminal.KeyPageDown:
		l.offset -= viewportHeight
		if l.offset <= 0 {
			l.offset = 0
			l.autoScroll = true
		}
		return runtime.Handled()

	case terminal.KeyHome:
		l.offset = maxOffset
		l.autoScroll = false
		return runtime.Handled()

	case terminal.KeyEnd:
		l.offset = 0
		l.autoScroll = true
		return runtime.Handled()

	case terminal.KeyCtrlC:
		if l.copyToClipboard() {
			return runtime.Handled()
		}

	case terminal.KeyTab:
		if key.Shift {
			return runtime.WithCommand(runtime.FocusPrev{})
		}
		return runtime.WithCommand(runtime.FocusNext{})

	case terminal.KeyEscape:
		return runtime.WithCommand(runtime.Cancel{})
	}

	return runtime.Unhandled()
}

// ScrollBy scrolls by delta rows.
func (l *Log) ScrollBy(dx, dy int) {
	if l == nil || dy == 0 {
		return
	}
	entries := l.FilteredEntries()
	viewportHeight := l.ContentBounds().Height
	if viewportHeight < 1 {
		viewportHeight = 1
	}
	maxOffset := len(entries) - viewportHeight
	if maxOffset < 0 {
		maxOffset = 0
	}

	l.offset -= dy // Negative because offset is from bottom
	if l.offset < 0 {
		l.offset = 0
		l.autoScroll = true
	}
	if l.offset > maxOffset {
		l.offset = maxOffset
	}
	l.Invalidate()
}

// ScrollTo scrolls to an absolute row.
func (l *Log) ScrollTo(x, y int) {
	if l == nil {
		return
	}
	entries := l.FilteredEntries()
	viewportHeight := l.ContentBounds().Height
	if viewportHeight < 1 {
		viewportHeight = 1
	}
	maxOffset := len(entries) - viewportHeight
	if maxOffset < 0 {
		maxOffset = 0
	}

	l.offset = len(entries) - y - viewportHeight
	if l.offset < 0 {
		l.offset = 0
		l.autoScroll = true
	}
	if l.offset > maxOffset {
		l.offset = maxOffset
	}
	l.Invalidate()
}

// PageBy scrolls by pages.
func (l *Log) PageBy(pages int) {
	if l == nil {
		return
	}
	viewportHeight := l.ContentBounds().Height
	if viewportHeight < 1 {
		viewportHeight = 1
	}
	l.ScrollBy(0, pages*viewportHeight)
}

// ScrollToStart scrolls to the oldest entries.
func (l *Log) ScrollToStart() {
	if l == nil {
		return
	}
	entries := l.FilteredEntries()
	viewportHeight := l.ContentBounds().Height
	if viewportHeight < 1 {
		viewportHeight = 1
	}
	l.offset = len(entries) - viewportHeight
	if l.offset < 0 {
		l.offset = 0
	}
	l.autoScroll = false
	l.Invalidate()
}

// ScrollToEnd scrolls to the newest entries.
func (l *Log) ScrollToEnd() {
	if l == nil {
		return
	}
	l.offset = 0
	l.autoScroll = true
	l.Invalidate()
}

// ClipboardCopy returns visible log entries for copying.
func (l *Log) ClipboardCopy() (string, bool) {
	if l == nil {
		return "", false
	}
	entries := l.FilteredEntries()
	if len(entries) == 0 {
		return "", false
	}

	var sb strings.Builder
	for _, entry := range entries {
		sb.WriteString(entry.Format(l.showTime))
		sb.WriteString("\n")
	}
	return sb.String(), true
}

func (l *Log) copyToClipboard() bool {
	cb := l.services.Clipboard()
	if cb == nil || !cb.Available() {
		return false
	}
	text, ok := l.ClipboardCopy()
	if !ok {
		return false
	}
	_ = cb.Write(text)
	return true
}

func (l *Log) syncA11y() {
	if l == nil {
		return
	}
	if l.Base.Role == "" {
		l.Base.Role = accessibility.RoleLog
	}
	label := strings.TrimSpace(l.label)
	if label == "" {
		label = "Log Viewer"
	}
	l.Base.Label = label

	entries := l.FilteredEntries()
	l.Base.Description = fmt.Sprintf("%d entries", len(entries))

	if len(entries) > 0 {
		latest := entries[len(entries)-1]
		l.Base.Value = &accessibility.ValueInfo{Text: latest.Message}
	} else {
		l.Base.Value = nil
	}
}

var _ scroll.Controller = (*Log)(nil)
var _ runtime.Widget = (*Log)(nil)
var _ runtime.Focusable = (*Log)(nil)
var _ runtime.Bindable = (*Log)(nil)
var _ runtime.Unbindable = (*Log)(nil)
