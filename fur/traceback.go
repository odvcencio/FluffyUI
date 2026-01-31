package fur

import (
	"errors"
	"fmt"
	"go/build"
	"os"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
)

// Traceback renders an error with stack trace and source context.
func Traceback(err error) Renderable {
	return TracebackWith(err, TracebackOpts{})
}

// TracebackOpts configures traceback rendering.
type TracebackOpts struct {
	Context    int
	MaxFrames  int
	HideStdlib bool
}

// TracebackWith renders an error with stack trace and options.
func TracebackWith(err error, opts TracebackOpts) Renderable {
	if opts.Context <= 0 {
		opts.Context = 3
	}
	if opts.MaxFrames <= 0 {
		opts.MaxFrames = 20
	}
	return tracebackRenderable{err: err, opts: opts}
}

// Wrap captures stack trace at call site.
func Wrap(err error) error {
	if err == nil {
		return nil
	}
	if hasStack(err) {
		return err
	}
	return &traceError{err: err, stack: captureStack(3)}
}

// WrapMsg wraps with additional message.
func WrapMsg(err error, msg string) error {
	if err == nil {
		return nil
	}
	return &traceError{err: err, msg: msg, stack: captureStack(3)}
}

// InstallPanicHandler enables richer panic output.
func InstallPanicHandler() {
	debug.SetTraceback("all")
}

type tracebackRenderable struct {
	err  error
	opts TracebackOpts
}

func (t tracebackRenderable) Render(width int) []Line {
	if t.err == nil {
		return nil
	}
	if width <= 0 {
		width = 80
	}
	var lines []Line
	errLines := wrapLines(splitTextLines(t.err.Error(), DefaultStyle().Foreground(ColorRed)), width-2)
	lines = append(lines, renderBox("Error", errLines, width, DefaultStyle().Foreground(ColorRed))...)
	lines = append(lines, Line{})
	lines = append(lines, Line{{Text: "Traceback (most recent call last):", Style: Dim}})
	lines = append(lines, Line{})

	frames := collectFrames(t.err, t.opts)
	for i, frame := range frames {
		title := fmt.Sprintf("%s:%d in %s", shortFile(frame.File), frame.Line, shortFunc(frame.Function))
		content := frameContext(frame, t.opts.Context)
		lines = append(lines, renderBox(title, content, width, Dim)...)
		if i < len(frames)-1 {
			lines = append(lines, Line{})
		}
	}
	return lines
}

type traceError struct {
	err   error
	msg   string
	stack []uintptr
}

func (e *traceError) Error() string {
	if e == nil {
		return ""
	}
	if e.msg == "" {
		return e.err.Error()
	}
	return e.msg + ": " + e.err.Error()
}

func (e *traceError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.err
}

func (e *traceError) Stack() []uintptr {
	if e == nil {
		return nil
	}
	return e.stack
}

type stackTracer interface {
	Stack() []uintptr
}

func hasStack(err error) bool {
	var tracer stackTracer
	return errors.As(err, &tracer)
}

func captureStack(skip int) []uintptr {
	pcs := make([]uintptr, 64)
	n := runtime.Callers(skip, pcs)
	return pcs[:n]
}

func collectFrames(err error, opts TracebackOpts) []runtime.Frame {
	stack := stackFromError(err)
	if len(stack) == 0 {
		stack = captureStack(4)
	}
	frames := runtime.CallersFrames(stack)
	var out []runtime.Frame
	for {
		frame, more := frames.Next()
		if !isInternalFrame(frame) && (!opts.HideStdlib || !isStdlibFrame(frame)) {
			out = append(out, frame)
		}
		if !more {
			break
		}
	}
	if len(out) > opts.MaxFrames {
		out = out[len(out)-opts.MaxFrames:]
	}
	for i := 0; i < len(out)/2; i++ {
		out[i], out[len(out)-1-i] = out[len(out)-1-i], out[i]
	}
	return out
}

func stackFromError(err error) []uintptr {
	var tracer stackTracer
	if errors.As(err, &tracer) {
		return tracer.Stack()
	}
	return nil
}

func isInternalFrame(frame runtime.Frame) bool {
	return strings.Contains(frame.Function, "/fur.") || strings.Contains(frame.Function, ".fur.")
}

func isStdlibFrame(frame runtime.Frame) bool {
	root := build.Default.GOROOT
	if root != "" {
		if strings.HasPrefix(frame.File, root+string(os.PathSeparator)) {
			return true
		}
	}
	return strings.Contains(frame.File, string(os.PathSeparator)+"runtime"+string(os.PathSeparator))
}

func shortFunc(name string) string {
	if name == "" {
		return ""
	}
	parts := strings.Split(name, "/")
	short := parts[len(parts)-1]
	short = strings.TrimPrefix(short, "(")
	return short
}

func frameContext(frame runtime.Frame, context int) []Line {
	if context <= 0 {
		context = 1
	}
	file := frame.File
	data, err := os.ReadFile(file)
	if err != nil {
		return splitTextLines("(source unavailable)", Dim)
	}
	source := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	line := frame.Line
	if line <= 0 || line > len(source) {
		return splitTextLines("(source unavailable)", Dim)
	}
	start := max(1, line-context)
	end := min(len(source), line+context)
	lineWidth := len(strconv.Itoa(end))
	var lines []Line
	for i := start; i <= end; i++ {
		prefix := "  "
		if i == line {
			prefix = "→ "
		}
		number := fmt.Sprintf("%*d", lineWidth, i)
		text := strings.TrimRight(source[i-1], "\r")
		formatted := fmt.Sprintf("%s%s │ %s", prefix, number, text)
		lines = append(lines, splitTextLines(formatted, DefaultStyle())...)
	}
	return lines
}
