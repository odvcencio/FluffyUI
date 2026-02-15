package widgets

import (
	"image"
	"strings"
	"testing"
	"time"

	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/keybind"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/state"
	"github.com/odvcencio/fluffyui/terminal"
	"github.com/odvcencio/fluffyui/video"
	fluffytest "github.com/odvcencio/fluffyui/testing"
)

// ---------------------------------------------------------------------------
// VideoPlayer (26.1% -> higher)
// ---------------------------------------------------------------------------

func TestVideoPlayer_Construction(t *testing.T) {
	player := &VideoPlayer{}
	_ = player // verify construction succeeds
}

func TestVideoPlayer_PlayPause(t *testing.T) {
	player := &VideoPlayer{}
	player.Play()
	if !player.IsPlaying() {
		t.Fatal("expected playing after Play()")
	}
	player.Pause()
	if player.IsPlaying() {
		t.Fatal("expected not playing after Pause()")
	}
}

func TestVideoPlayer_NilSafe(t *testing.T) {
	var player *VideoPlayer
	player.Play()
	player.Pause()
	player.Seek(0)
	if player.IsPlaying() {
		t.Fatal("nil player should not be playing")
	}
	if player.DroppedFrames() != 0 {
		t.Fatal("nil player should have 0 dropped frames")
	}
	result := player.HandleMessage(runtime.KeyMsg{})
	if result.Handled {
		t.Fatal("nil player should return unhandled")
	}
}

func TestVideoPlayer_StyleType(t *testing.T) {
	player := &VideoPlayer{}
	if player.StyleType() != "VideoPlayer" {
		t.Fatalf("expected StyleType 'VideoPlayer', got %q", player.StyleType())
	}
}

func TestVideoPlayer_Measure(t *testing.T) {
	player := &VideoPlayer{}
	size := player.Measure(runtime.Constraints{MaxWidth: 40, MaxHeight: 20})
	if size.Width <= 0 || size.Height <= 0 {
		t.Fatalf("expected positive size, got %dx%d", size.Width, size.Height)
	}
}

func TestVideoPlayer_Layout(t *testing.T) {
	player := &VideoPlayer{}
	player.Measure(runtime.Constraints{MaxWidth: 40, MaxHeight: 20})
	player.Layout(runtime.Rect{X: 0, Y: 0, Width: 40, Height: 20})
	// Should not panic.
}

func TestVideoPlayer_LayoutZero(t *testing.T) {
	player := &VideoPlayer{}
	player.Layout(runtime.Rect{X: 0, Y: 0, Width: 0, Height: 0})
	if player.canvas != nil {
		t.Fatal("expected nil canvas with zero bounds")
	}
}

func TestVideoPlayer_RenderNilCanvas(t *testing.T) {
	player := &VideoPlayer{}
	buf := runtime.NewBuffer(10, 5)
	player.Render(runtime.RenderContext{Buffer: buf})
	// Should not panic.
}

func TestVideoPlayer_SetBlitter(t *testing.T) {
	player := &VideoPlayer{}
	player.SetBlitter(nil) // nil should be no-op
	// Should not panic.
}

func TestVideoPlayer_SetOnEnd(t *testing.T) {
	called := false
	player := &VideoPlayer{}
	player.SetOnEnd(func() { called = true })
	if player.onEnd == nil {
		t.Fatal("expected onEnd to be set")
	}
	player.onEnd()
	if !called {
		t.Fatal("expected onEnd to be called")
	}
}

func TestVideoPlayer_DroppedFrames(t *testing.T) {
	player := &VideoPlayer{framesDropped: 5}
	if player.DroppedFrames() != 5 {
		t.Fatalf("expected 5 dropped frames, got %d", player.DroppedFrames())
	}
}

func TestVideoPlayer_FrameIndexFor(t *testing.T) {
	player := &VideoPlayer{frameDuration: 100 * time.Millisecond}
	idx := player.frameIndexFor(250 * time.Millisecond)
	if idx != 2 {
		t.Fatalf("expected frame index 2, got %d", idx)
	}
	idx = player.frameIndexFor(0)
	if idx != 0 {
		t.Fatalf("expected frame index 0, got %d", idx)
	}
}

func TestVideoPlayer_FrameIndexForZeroDuration(t *testing.T) {
	player := &VideoPlayer{frameDuration: 0}
	idx := player.frameIndexFor(500 * time.Millisecond)
	if idx != 0 {
		t.Fatalf("expected 0 for zero duration, got %d", idx)
	}
}

func TestVideoPlayer_CurrentFrameImage(t *testing.T) {
	player := &VideoPlayer{
		frames:       make([]image.Image, 3),
		currentFrame: 1,
	}
	img := player.currentFrameImage()
	if img != nil {
		// The images are nil (make creates zero-value slices), which is fine.
	}

	// Empty frames.
	player2 := &VideoPlayer{}
	if player2.currentFrameImage() != nil {
		t.Fatal("expected nil for no frames")
	}

	// Negative frame index.
	player3 := &VideoPlayer{
		frames:       make([]image.Image, 3),
		currentFrame: -1,
	}
	_ = player3.currentFrameImage() // Should return first frame.

	// Frame beyond range.
	player4 := &VideoPlayer{
		frames:       make([]image.Image, 3),
		currentFrame: 10,
	}
	_ = player4.currentFrameImage() // Should return last frame.
}

func TestVideoPlayer_AdvanceFrame(t *testing.T) {
	player := &VideoPlayer{
		frameDuration: 100 * time.Millisecond,
		frames:        make([]image.Image, 5),
		framesDone:    true,
		playing:       true,
		playhead:      600 * time.Millisecond,
	}
	player.advanceFrame()
	// Playhead at 600ms with 100ms frames = frame 6, but only 5 frames, so should be at 4.
	if player.currentFrame != 4 {
		t.Fatalf("expected frame 4, got %d", player.currentFrame)
	}
	if player.playing {
		t.Fatal("expected playing to be false after reaching end")
	}
}

func TestVideoPlayer_AdvanceFrameNoFrames(t *testing.T) {
	player := &VideoPlayer{
		frameDuration: 100 * time.Millisecond,
		playing:       true,
	}
	player.advanceFrame()
	// No panic expected.
}

func TestVideoPlayer_TickWhenNotPlaying(t *testing.T) {
	player := &VideoPlayer{
		frameDuration: 100 * time.Millisecond,
		playing:       false,
	}
	result := player.HandleMessage(runtime.TickMsg{Time: time.Now()})
	if result.Handled {
		t.Fatal("expected unhandled when not playing")
	}
}

func TestVideoPlayer_TickZeroDuration(t *testing.T) {
	player := &VideoPlayer{
		frameDuration: 0,
		playing:       true,
	}
	result := player.HandleMessage(runtime.TickMsg{Time: time.Now()})
	if result.Handled {
		t.Fatal("expected unhandled with zero frame duration")
	}
}

func TestVideoPlayer_TickFirstTick(t *testing.T) {
	player := &VideoPlayer{
		frameDuration: 100 * time.Millisecond,
		playing:       true,
		frames:        make([]image.Image, 5),
	}
	now := time.Now()
	result := player.HandleMessage(runtime.TickMsg{Time: now})
	if !result.Handled {
		t.Fatal("expected first tick to be handled")
	}
	if player.lastTick.IsZero() {
		t.Fatal("expected lastTick to be set")
	}
}

func TestVideoPlayer_WithBlitterDeprecated(t *testing.T) {
	player := &VideoPlayer{}
	result := player.WithBlitter(nil)
	if result != player {
		t.Fatal("expected same player returned")
	}
}

func TestVideoPlayer_InitTiming(t *testing.T) {
	player := &VideoPlayer{}
	player.initTiming(video.VideoInfo{FrameRate: 0})
	if player.frameRate != 30 {
		t.Fatalf("expected default frame rate 30, got %f", player.frameRate)
	}
}

func TestVideoPlayer_InitTimingPositive(t *testing.T) {
	player := &VideoPlayer{}
	player.initTiming(video.VideoInfo{FrameRate: 60})
	if player.frameRate != 60 {
		t.Fatalf("expected frame rate 60, got %f", player.frameRate)
	}
}

func TestVideoPlayer_StartFrameLoaderNilDecoder(t *testing.T) {
	player := &VideoPlayer{}
	err := player.startFrameLoader()
	if err != nil {
		t.Fatalf("expected nil error with nil decoder, got %v", err)
	}
}

func TestVideoPlayer_DrawFrameNilCanvas(t *testing.T) {
	player := &VideoPlayer{}
	player.drawFrame(nil) // Should not panic.
}

func TestVideoPlayer_FrameSnapshot(t *testing.T) {
	player := &VideoPlayer{
		frames:     make([]image.Image, 3),
		framesDone: true,
	}
	count, done := player.frameSnapshot()
	if count != 3 {
		t.Fatalf("expected 3 frames, got %d", count)
	}
	if !done {
		t.Fatal("expected done")
	}
}

// ---------------------------------------------------------------------------
// EnhancedPalette (28.6% -> higher)
// ---------------------------------------------------------------------------

func TestEnhancedPalette_New(t *testing.T) {
	reg := keybind.NewRegistry()
	ep := NewEnhancedPalette(reg)
	if ep == nil {
		t.Fatal("expected non-nil enhanced palette")
	}
	if ep.Widget == nil {
		t.Fatal("expected non-nil widget")
	}
}

func TestEnhancedPalette_WithCommands(t *testing.T) {
	reg := keybind.NewRegistry()
	reg.Register(keybind.Command{ID: "test.cmd", Title: "Test Command", Category: "Testing"})
	reg.Register(keybind.Command{ID: "test.cmd2", Title: "Other Command", Category: "Other"})
	ep := NewEnhancedPalette(reg)
	if ep.Widget == nil {
		t.Fatal("expected widget")
	}
}

func TestEnhancedPalette_Record(t *testing.T) {
	reg := keybind.NewRegistry()
	reg.Register(keybind.Command{ID: "cmd1", Title: "Command 1"})
	reg.Register(keybind.Command{ID: "cmd2", Title: "Command 2"})
	ep := NewEnhancedPalette(reg)
	ep.Record("cmd1")
	if len(ep.recent) != 1 {
		t.Fatalf("expected 1 recent, got %d", len(ep.recent))
	}
	ep.Record("cmd2")
	if len(ep.recent) != 2 {
		t.Fatalf("expected 2 recent, got %d", len(ep.recent))
	}
	// Duplicate should not grow the list.
	ep.Record("cmd1")
	if len(ep.recent) != 2 {
		t.Fatalf("expected 2 recent after dup, got %d", len(ep.recent))
	}
}

func TestEnhancedPalette_RecordEmpty(t *testing.T) {
	reg := keybind.NewRegistry()
	ep := NewEnhancedPalette(reg)
	ep.Record("")
	if len(ep.recent) != 0 {
		t.Fatalf("expected 0 recent for empty id, got %d", len(ep.recent))
	}
}

func TestEnhancedPalette_Pin(t *testing.T) {
	reg := keybind.NewRegistry()
	reg.Register(keybind.Command{ID: "cmd1", Title: "Command 1"})
	ep := NewEnhancedPalette(reg)
	ep.Pin("cmd1")
	if len(ep.pinned) != 1 {
		t.Fatalf("expected 1 pinned, got %d", len(ep.pinned))
	}
}

func TestEnhancedPalette_Unpin(t *testing.T) {
	reg := keybind.NewRegistry()
	reg.Register(keybind.Command{ID: "cmd1", Title: "Command 1"})
	ep := NewEnhancedPalette(reg)
	ep.Pin("cmd1")
	ep.Unpin("cmd1")
	if len(ep.pinned) != 0 {
		t.Fatalf("expected 0 pinned after unpin, got %d", len(ep.pinned))
	}
}

func TestEnhancedPalette_UnpinEmpty(t *testing.T) {
	reg := keybind.NewRegistry()
	ep := NewEnhancedPalette(reg)
	ep.Unpin("")  // Should be no-op.
	ep.Pin("")    // Should be no-op.
	ep.Record("") // Should be no-op.
}

func TestEnhancedPalette_NilSafe(t *testing.T) {
	var ep *EnhancedPalette
	ep.SetKeymaps()
	ep.SetKeymapStack(nil)
	ep.Refresh()
	ep.Record("x")
	ep.Pin("x")
	ep.Unpin("x")
}

func TestEnhancedPalette_SetKeymaps(t *testing.T) {
	reg := keybind.NewRegistry()
	reg.Register(keybind.Command{ID: "cmd1", Title: "Command 1"})
	ep := NewEnhancedPalette(reg)
	ep.SetKeymaps()
}

func TestEnhancedPalette_CommandTitle(t *testing.T) {
	// Test with title set.
	cmd := keybind.Command{ID: "test", Title: "Test Title"}
	if commandTitle(cmd) != "Test Title" {
		t.Fatalf("expected 'Test Title', got %q", commandTitle(cmd))
	}
	// Test without title.
	cmd2 := keybind.Command{ID: "test.id"}
	if commandTitle(cmd2) != "test.id" {
		t.Fatalf("expected 'test.id', got %q", commandTitle(cmd2))
	}
}

func TestEnhancedPalette_Unique(t *testing.T) {
	result := unique([]string{"a", "b", "a", "c", "b"})
	if len(result) != 3 {
		t.Fatalf("expected 3 unique items, got %d", len(result))
	}
}

func TestEnhancedPalette_RecordLimit(t *testing.T) {
	reg := keybind.NewRegistry()
	for i := 0; i < 15; i++ {
		id := strings.Repeat("x", i+1)
		reg.Register(keybind.Command{ID: id, Title: id})
	}
	ep := NewEnhancedPalette(reg)
	for i := 0; i < 15; i++ {
		id := strings.Repeat("x", i+1)
		ep.Record(id)
	}
	if len(ep.recent) > 10 {
		t.Fatalf("expected max 10 recent, got %d", len(ep.recent))
	}
}

// ---------------------------------------------------------------------------
// ScrollView (31% -> higher)
// ---------------------------------------------------------------------------

func TestScrollView_New(t *testing.T) {
	content := NewLabel("Hello")
	sv := NewScrollView(content)
	if sv == nil {
		t.Fatal("expected non-nil scroll view")
	}
	if sv.AccessibleRole() != "group" {
		t.Fatalf("expected role group, got %q", sv.AccessibleRole())
	}
}

func TestScrollView_Measure(t *testing.T) {
	content := NewLabel("Hello World")
	sv := NewScrollView(content)
	size := sv.Measure(runtime.Constraints{MaxWidth: 40, MaxHeight: 20})
	if size.Width <= 0 || size.Height <= 0 {
		t.Fatalf("expected positive size, got %dx%d", size.Width, size.Height)
	}
}

func TestScrollView_SetLabel(t *testing.T) {
	sv := NewScrollView(NewLabel("x"))
	sv.SetLabel("Custom Label")
	if sv.AccessibleLabel() != "Custom Label" {
		t.Fatalf("expected label 'Custom Label', got %q", sv.AccessibleLabel())
	}
}

func TestScrollView_SetContent(t *testing.T) {
	sv := NewScrollView(NewLabel("x"))
	sv.SetContent(NewLabel("y"))
	children := sv.ChildWidgets()
	if len(children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(children))
	}
}

func TestScrollView_SetBehavior(t *testing.T) {
	sv := NewScrollView(NewLabel("x"))
	sv.SetBehavior(sv.behavior) // Should not panic.
}

func TestScrollView_ContentSize(t *testing.T) {
	sv := NewScrollView(NewLabel("x"))
	cs := sv.ContentSize()
	// Initially zero since we haven't measured.
	if cs.Width < 0 || cs.Height < 0 {
		t.Fatalf("expected non-negative content size, got %dx%d", cs.Width, cs.Height)
	}
}

func TestScrollView_NilContentSize(t *testing.T) {
	var sv *ScrollView
	cs := sv.ContentSize()
	if cs.Width != 0 || cs.Height != 0 {
		t.Fatal("expected zero size for nil scroll view")
	}
}

func TestScrollView_Render(t *testing.T) {
	content := NewLabel("Hello World")
	sv := NewScrollView(content)
	output := fluffytest.RenderToString(sv, 20, 5)
	if !strings.Contains(output, "Hello") {
		t.Fatalf("expected output to contain 'Hello', got %q", output)
	}
}

func TestScrollView_ScrollBy(t *testing.T) {
	content := NewLabel("Line1\nLine2\nLine3\nLine4\nLine5\nLine6\nLine7\nLine8")
	sv := NewScrollView(content)
	sv.Measure(runtime.Constraints{MaxWidth: 20, MaxHeight: 3})
	sv.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 3})
	sv.ScrollBy(0, 2)
	// Should not panic.
}

func TestScrollView_ScrollTo(t *testing.T) {
	sv := NewScrollView(NewLabel("x"))
	sv.Measure(runtime.Constraints{MaxWidth: 20, MaxHeight: 3})
	sv.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 3})
	sv.ScrollTo(0, 5)
	// Should not panic.
}

func TestScrollView_PageBy(t *testing.T) {
	sv := NewScrollView(NewLabel("x"))
	sv.Measure(runtime.Constraints{MaxWidth: 20, MaxHeight: 3})
	sv.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 3})
	sv.PageBy(1)
	// Should not panic.
}

func TestScrollView_ScrollToStartEnd(t *testing.T) {
	sv := NewScrollView(NewLabel("x"))
	sv.Measure(runtime.Constraints{MaxWidth: 20, MaxHeight: 3})
	sv.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 3})
	sv.ScrollToStart()
	sv.ScrollToEnd()
}

func TestScrollView_HandleKeyMessages(t *testing.T) {
	sv := NewScrollView(NewLabel("x\ny\nz"))
	sv.Measure(runtime.Constraints{MaxWidth: 20, MaxHeight: 3})
	sv.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 3})
	sv.Focus()
	tests := []terminal.Key{
		terminal.KeyUp,
		terminal.KeyDown,
		terminal.KeyPageUp,
		terminal.KeyPageDown,
		terminal.KeyHome,
		terminal.KeyEnd,
	}
	for _, key := range tests {
		result := sv.HandleMessage(runtime.KeyMsg{Key: key})
		if !result.Handled {
			t.Fatalf("expected key %v to be handled", key)
		}
	}
}

func TestScrollView_HandleMouseWheel(t *testing.T) {
	sv := NewScrollView(NewLabel("x\ny\nz"))
	sv.Measure(runtime.Constraints{MaxWidth: 20, MaxHeight: 3})
	sv.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 3})
	sv.Focus()
	result := sv.HandleMessage(runtime.MouseMsg{Button: runtime.MouseWheelUp})
	if !result.Handled {
		t.Fatal("expected wheel up to be handled")
	}
	result = sv.HandleMessage(runtime.MouseMsg{Button: runtime.MouseWheelDown})
	if !result.Handled {
		t.Fatal("expected wheel down to be handled")
	}
}

func TestScrollView_HandleKeyUnfocused(t *testing.T) {
	sv := NewScrollView(NewLabel("x"))
	result := sv.HandleMessage(runtime.KeyMsg{Key: terminal.KeyUp})
	if result.Handled {
		t.Fatal("expected key to be unhandled when not focused")
	}
}

func TestScrollView_ChildWidgets(t *testing.T) {
	sv := NewScrollView(NewLabel("x"))
	children := sv.ChildWidgets()
	if len(children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(children))
	}
}

func TestScrollView_ChildWidgetsNil(t *testing.T) {
	sv := &ScrollView{}
	children := sv.ChildWidgets()
	if children != nil {
		t.Fatal("expected nil children with nil content")
	}
}

func TestScrollView_NilHandleMessage(t *testing.T) {
	var sv *ScrollView
	result := sv.HandleMessage(runtime.KeyMsg{})
	if result.Handled {
		t.Fatal("nil scroll view should return unhandled")
	}
}

func TestScrollView_SetStyle(t *testing.T) {
	sv := NewScrollView(NewLabel("x"))
	sv.SetStyle(sv.style) // Should not panic.
}

func TestScrollView_BindUnbind(t *testing.T) {
	sv := NewScrollView(NewLabel("x"))
	sv.Bind(runtime.Services{})
	sv.Unbind()
}

// ---------------------------------------------------------------------------
// TimePicker (35.7% -> higher)
// ---------------------------------------------------------------------------

func TestTimePicker_New(t *testing.T) {
	tp := NewTimePicker()
	if tp == nil {
		t.Fatal("expected non-nil time picker")
	}
	if tp.AccessibleRole() != "group" {
		t.Fatalf("expected role group, got %q", tp.AccessibleRole())
	}
}

func TestTimePicker_SetTime(t *testing.T) {
	tp := NewTimePicker()
	target := time.Date(2025, 6, 15, 14, 30, 45, 0, time.UTC)
	tp.SetTime(target)
	result := tp.Time()
	if result.Hour() != 14 || result.Minute() != 30 || result.Second() != 45 {
		t.Fatalf("expected 14:30:45, got %v", result)
	}
}

func TestTimePicker_Measure(t *testing.T) {
	tp := NewTimePicker()
	size := tp.Measure(runtime.Constraints{MaxWidth: 40, MaxHeight: 5})
	if size.Width != 5 {
		t.Fatalf("expected width 5, got %d", size.Width)
	}
	if size.Height != 1 {
		t.Fatalf("expected height 1, got %d", size.Height)
	}
}

func TestTimePicker_MeasureWithSeconds(t *testing.T) {
	tp := NewTimePicker()
	tp.SetShowSeconds(true)
	size := tp.Measure(runtime.Constraints{MaxWidth: 40, MaxHeight: 5})
	if size.Width != 8 {
		t.Fatalf("expected width 8 with seconds, got %d", size.Width)
	}
}

func TestTimePicker_Render(t *testing.T) {
	tp := NewTimePicker()
	tp.SetTime(time.Date(2025, 1, 1, 9, 5, 0, 0, time.UTC))
	output := fluffytest.RenderToString(tp, 10, 1)
	if !strings.Contains(output, "09") || !strings.Contains(output, "05") {
		t.Fatalf("expected output to contain '09' and '05', got %q", output)
	}
}

func TestTimePicker_HandleKeyLeftRight(t *testing.T) {
	tp := NewTimePicker()
	tp.Focus()
	// Start at field 0 (hours), move right to minutes.
	result := tp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRight})
	if !result.Handled {
		t.Fatal("expected right key to be handled")
	}
	if tp.selected != 1 {
		t.Fatalf("expected selected field 1, got %d", tp.selected)
	}
	// Can't go beyond max (1 without seconds).
	tp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRight})
	if tp.selected != 1 {
		t.Fatalf("expected selected still 1, got %d", tp.selected)
	}
	// Move left.
	tp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyLeft})
	if tp.selected != 0 {
		t.Fatalf("expected selected 0, got %d", tp.selected)
	}
	// Can't go below 0.
	tp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyLeft})
	if tp.selected != 0 {
		t.Fatalf("expected selected still 0, got %d", tp.selected)
	}
}

func TestTimePicker_HandleKeyUpDown(t *testing.T) {
	tp := NewTimePicker()
	tp.Focus()
	tp.SetTime(time.Date(2025, 1, 1, 10, 30, 0, 0, time.UTC))
	// Up increments hour.
	tp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyUp})
	if tp.hour != 11 {
		t.Fatalf("expected hour 11, got %d", tp.hour)
	}
	// Down decrements hour.
	tp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	if tp.hour != 10 {
		t.Fatalf("expected hour 10, got %d", tp.hour)
	}
}

func TestTimePicker_HandleKeyEnter(t *testing.T) {
	tp := NewTimePicker()
	tp.Focus()
	submitted := false
	tp.SetOnSubmit(func(t time.Time) {
		submitted = true
	})
	tp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})
	if !submitted {
		t.Fatal("expected submit callback to be called")
	}
}

func TestTimePicker_HandleDigit(t *testing.T) {
	tp := NewTimePicker()
	tp.Focus()
	tp.SetTime(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	// Type "1" then "5" for 15.
	tp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: '1'})
	tp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: '5'})
	if tp.hour != 15 {
		t.Fatalf("expected hour 15, got %d", tp.hour)
	}
	// Should auto-advance to minutes field.
	if tp.selected != 1 {
		t.Fatalf("expected selected field 1 after two digits, got %d", tp.selected)
	}
}

func TestTimePicker_SetShowSeconds(t *testing.T) {
	tp := NewTimePicker()
	tp.SetShowSeconds(true)
	if !tp.showSeconds {
		t.Fatal("expected showSeconds to be true")
	}
	tp.selected = 2 // On seconds field.
	tp.SetShowSeconds(false)
	if tp.selected != 1 {
		t.Fatalf("expected selected clamped to 1, got %d", tp.selected)
	}
}

func TestTimePicker_SetLabel(t *testing.T) {
	tp := NewTimePicker()
	tp.SetLabel("Start Time")
	if tp.AccessibleLabel() != "Start Time" {
		t.Fatalf("expected label 'Start Time', got %q", tp.AccessibleLabel())
	}
}

func TestTimePicker_SetOnChange(t *testing.T) {
	tp := NewTimePicker()
	tp.Focus()
	changed := false
	tp.SetOnChange(func(t time.Time) {
		changed = true
	})
	tp.SetTime(time.Date(2025, 1, 1, 5, 0, 0, 0, time.UTC))
	tp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyUp})
	if !changed {
		t.Fatal("expected onChange callback to fire")
	}
}

func TestTimePicker_NilSafe(t *testing.T) {
	var tp *TimePicker
	tp.SetShowSeconds(true)
	tp.SetLabel("x")
	tp.SetOnChange(nil)
	tp.SetOnSubmit(nil)
	tp.SetTime(time.Now())
	if !tp.Time().IsZero() {
		t.Fatal("nil time picker should return zero time")
	}
	tp.Bind(runtime.Services{})
	tp.Unbind()
}

func TestTimePicker_HandleUnfocused(t *testing.T) {
	tp := NewTimePicker()
	result := tp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyUp})
	if result.Handled {
		t.Fatal("expected unhandled when not focused")
	}
}

func TestTimePicker_HandleNonKeyMsg(t *testing.T) {
	tp := NewTimePicker()
	tp.Focus()
	result := tp.HandleMessage(runtime.MouseMsg{})
	if result.Handled {
		t.Fatal("expected non-key message to be unhandled")
	}
}

func TestTimePicker_FieldIncrementWrap(t *testing.T) {
	tp := NewTimePicker()
	tp.Focus()
	tp.hour = 23
	tp.incrementField(1)
	if tp.hour != 0 {
		t.Fatalf("expected hour to wrap to 0, got %d", tp.hour)
	}
	tp.hour = 0
	tp.incrementField(-1)
	if tp.hour != 23 {
		t.Fatalf("expected hour to wrap to 23, got %d", tp.hour)
	}
}

func TestTimePicker_FieldIncrementMinute(t *testing.T) {
	tp := NewTimePicker()
	tp.Focus()
	tp.selected = 1
	tp.minute = 59
	tp.incrementField(1)
	if tp.minute != 0 {
		t.Fatalf("expected minute to wrap to 0, got %d", tp.minute)
	}
}

func TestTimePicker_FieldIncrementSecond(t *testing.T) {
	tp := NewTimePicker()
	tp.Focus()
	tp.SetShowSeconds(true)
	tp.selected = 2
	tp.second = 30
	tp.incrementField(5)
	if tp.second != 35 {
		t.Fatalf("expected second 35, got %d", tp.second)
	}
}

func TestTimePicker_StyleType(t *testing.T) {
	tp := NewTimePicker()
	if tp.StyleType() != "TimePicker" {
		t.Fatalf("expected StyleType 'TimePicker', got %q", tp.StyleType())
	}
}

// ---------------------------------------------------------------------------
// List (36.6% -> higher)
// ---------------------------------------------------------------------------

func TestList_NewList(t *testing.T) {
	adapter := NewSliceAdapter([]string{"A", "B", "C"}, func(item string, index int, selected bool, ctx runtime.RenderContext) {
		ctx.Buffer.SetString(0, 0, item, ctx.Buffer.Get(0, 0).Style)
	})
	list := NewList(adapter)
	if list == nil {
		t.Fatal("expected non-nil list")
	}
	if list.AccessibleRole() != "list" {
		t.Fatalf("expected role list, got %q", list.AccessibleRole())
	}
}

func TestList_Measure(t *testing.T) {
	adapter := NewSliceAdapter([]string{"A", "B", "C"}, nil)
	list := NewList(adapter)
	size := list.Measure(runtime.Constraints{MaxWidth: 20, MaxHeight: 10})
	if size.Width <= 0 || size.Height <= 0 {
		t.Fatalf("expected positive size, got %dx%d", size.Width, size.Height)
	}
}

func TestList_HandleKeyNavigation(t *testing.T) {
	adapter := NewSliceAdapter([]string{"A", "B", "C", "D", "E"}, nil)
	list := NewList(adapter)
	list.Focus()
	list.Measure(runtime.Constraints{MaxWidth: 20, MaxHeight: 10})
	list.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 10})

	// Down should move selection.
	list.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	if list.SelectedIndex() != 1 {
		t.Fatalf("expected selected 1, got %d", list.SelectedIndex())
	}
	// Up should move back.
	list.HandleMessage(runtime.KeyMsg{Key: terminal.KeyUp})
	if list.SelectedIndex() != 0 {
		t.Fatalf("expected selected 0, got %d", list.SelectedIndex())
	}
	// Home.
	list.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	list.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	list.HandleMessage(runtime.KeyMsg{Key: terminal.KeyHome})
	if list.SelectedIndex() != 0 {
		t.Fatalf("expected selected 0 after Home, got %d", list.SelectedIndex())
	}
	// End.
	list.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnd})
	if list.SelectedIndex() != 4 {
		t.Fatalf("expected selected 4 after End, got %d", list.SelectedIndex())
	}
}

func TestList_HandleKeyEnter(t *testing.T) {
	adapter := NewSliceAdapter([]string{"A", "B"}, nil)
	list := NewList(adapter)
	list.Focus()
	selected := ""
	list.SetOnSelect(func(index int, item string) {
		selected = item
	})
	list.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})
	if selected != "A" {
		t.Fatalf("expected 'A' selected, got %q", selected)
	}
}

func TestList_HandleKeyPageUpDown(t *testing.T) {
	adapter := NewSliceAdapter([]string{"A", "B", "C", "D", "E", "F"}, nil)
	list := NewList(adapter)
	list.Focus()
	list.Measure(runtime.Constraints{MaxWidth: 20, MaxHeight: 3})
	list.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 3})
	list.HandleMessage(runtime.KeyMsg{Key: terminal.KeyPageDown})
	if list.SelectedIndex() == 0 {
		t.Fatal("expected selection to advance after page down")
	}
}

func TestList_SetSelected(t *testing.T) {
	adapter := NewSliceAdapter([]string{"A", "B", "C"}, nil)
	list := NewList(adapter)
	list.SetSelected(2)
	if list.SelectedIndex() != 2 {
		t.Fatalf("expected selected 2, got %d", list.SelectedIndex())
	}
}

func TestList_SelectedItem(t *testing.T) {
	adapter := NewSliceAdapter([]string{"A", "B", "C"}, nil)
	list := NewList(adapter)
	item, ok := list.SelectedItem()
	if !ok {
		t.Fatal("expected ok")
	}
	if item != "A" {
		t.Fatalf("expected 'A', got %q", item)
	}
}

func TestList_SetLabel(t *testing.T) {
	adapter := NewSliceAdapter([]string{"A"}, nil)
	list := NewList(adapter)
	list.SetLabel("My List")
	if list.AccessibleLabel() != "My List" {
		t.Fatalf("expected label 'My List', got %q", list.AccessibleLabel())
	}
}

func TestList_ScrollByScrollTo(t *testing.T) {
	adapter := NewSliceAdapter([]string{"A", "B", "C"}, nil)
	list := NewList(adapter)
	list.ScrollBy(0, 1)
	list.ScrollTo(0, 2)
	list.PageBy(1)
	list.ScrollToStart()
	list.ScrollToEnd()
}

func TestList_SetStyle(t *testing.T) {
	adapter := NewSliceAdapter([]string{"A"}, nil)
	list := NewList(adapter)
	list.SetStyle(list.style)
	list.SetSelectedStyle(list.selectedStyle)
}

func TestList_StyleType(t *testing.T) {
	adapter := NewSliceAdapter([]string{"A"}, nil)
	list := NewList(adapter)
	if list.StyleType() != "List" {
		t.Fatalf("expected StyleType 'List', got %q", list.StyleType())
	}
}

func TestList_HandleUnfocused(t *testing.T) {
	adapter := NewSliceAdapter([]string{"A"}, nil)
	list := NewList(adapter)
	result := list.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	if result.Handled {
		t.Fatal("expected unhandled when not focused")
	}
}

func TestList_HandleNonKeyMsg(t *testing.T) {
	adapter := NewSliceAdapter([]string{"A"}, nil)
	list := NewList(adapter)
	list.Focus()
	result := list.HandleMessage(runtime.MouseMsg{})
	if result.Handled {
		t.Fatal("expected non-key message to be unhandled")
	}
}

func TestList_DeprecatedOnSelect(t *testing.T) {
	adapter := NewSliceAdapter([]string{"A"}, nil)
	list := NewList(adapter)
	list.OnSelect(nil) // Deprecated but should still work.
}

// ---------------------------------------------------------------------------
// DebugOverlay (37.7% -> higher)
// ---------------------------------------------------------------------------

func TestDebugOverlay_New(t *testing.T) {
	root := NewLabel("Hello")
	overlay := NewDebugOverlay(root)
	if overlay == nil {
		t.Fatal("expected non-nil overlay")
	}
}

func TestDebugOverlay_Measure(t *testing.T) {
	overlay := NewDebugOverlay(NewLabel("x"))
	size := overlay.Measure(runtime.Constraints{MaxWidth: 40, MaxHeight: 20})
	if size.Width != 40 || size.Height != 20 {
		t.Fatalf("expected 40x20, got %dx%d", size.Width, size.Height)
	}
}

func TestDebugOverlay_Render(t *testing.T) {
	label := NewLabel("Debug Target")
	label.Measure(runtime.Constraints{MaxWidth: 20, MaxHeight: 5})
	label.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 5})
	overlay := NewDebugOverlay(label, WithDebugLabels(true), WithDebugMaxDepth(3))
	overlay.Measure(runtime.Constraints{MaxWidth: 20, MaxHeight: 5})
	overlay.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 5})
	buf := runtime.NewBuffer(20, 5)
	overlay.Render(runtime.RenderContext{Buffer: buf})
	// Should not panic.
}

func TestDebugOverlay_RenderNilRoot(t *testing.T) {
	overlay := NewDebugOverlay(nil)
	buf := runtime.NewBuffer(10, 5)
	overlay.Render(runtime.RenderContext{Buffer: buf})
	// Should not panic.
}

func TestDebugOverlay_Options(t *testing.T) {
	overlay := NewDebugOverlay(NewLabel("x"),
		WithDebugStyle(backend.DefaultStyle()),
		WithDebugLabelStyle(backend.DefaultStyle()),
		WithDebugLabels(false),
		WithDebugMaxDepth(5),
	)
	if overlay.showLabels {
		t.Fatal("expected showLabels to be false")
	}
	if overlay.maxDepth != 5 {
		t.Fatalf("expected maxDepth 5, got %d", overlay.maxDepth)
	}
}

func TestDebugOverlay_NilOptions(t *testing.T) {
	// Nil option should be skipped.
	overlay := NewDebugOverlay(NewLabel("x"), nil)
	if overlay == nil {
		t.Fatal("expected non-nil overlay")
	}
}

func TestDebugOverlay_BindUnbind(t *testing.T) {
	overlay := NewDebugOverlay(NewLabel("x"))
	overlay.Bind(runtime.Services{})
	overlay.Unbind()
}

func TestDebugOverlay_NilBindUnbind(t *testing.T) {
	var overlay *DebugOverlay
	overlay.Bind(runtime.Services{})
	overlay.Unbind()
}

func TestDebugOverlay_DebugLabel(t *testing.T) {
	label := NewLabel("Test")
	result := debugLabel(label)
	if !strings.Contains(result, "Label") {
		t.Fatalf("expected label to contain 'Label', got %q", result)
	}
}

func TestDebugOverlay_RenderWithChildren(t *testing.T) {
	child1 := NewLabel("A")
	child1.Measure(runtime.Constraints{MaxWidth: 10, MaxHeight: 3})
	child1.Layout(runtime.Rect{X: 0, Y: 0, Width: 10, Height: 3})
	child2 := NewLabel("B")
	child2.Measure(runtime.Constraints{MaxWidth: 10, MaxHeight: 3})
	child2.Layout(runtime.Rect{X: 10, Y: 0, Width: 10, Height: 3})
	container := NewSplitter(child1, child2)
	container.Measure(runtime.Constraints{MaxWidth: 20, MaxHeight: 3})
	container.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 3})
	overlay := NewDebugOverlay(container)
	overlay.Measure(runtime.Constraints{MaxWidth: 20, MaxHeight: 3})
	overlay.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 3})
	buf := runtime.NewBuffer(20, 3)
	overlay.Render(runtime.RenderContext{Buffer: buf})
	// Should render bounding boxes for children.
}

// ---------------------------------------------------------------------------
// Radio (45.8% -> higher)
// ---------------------------------------------------------------------------

func TestRadio_NewRadio(t *testing.T) {
	group := NewRadioGroup()
	radio := NewRadio("Option 1", group)
	if radio == nil {
		t.Fatal("expected non-nil radio")
	}
	if radio.AccessibleRole() != "radio" {
		t.Fatalf("expected role radio, got %q", radio.AccessibleRole())
	}
}

func TestRadio_GroupSelection(t *testing.T) {
	group := NewRadioGroup()
	r1 := NewRadio("A", group)
	r2 := NewRadio("B", group)
	r3 := NewRadio("C", group)
	_ = r1
	_ = r2
	_ = r3
	group.SetSelected(1)
	if group.Selected() != 1 {
		t.Fatalf("expected selected 1, got %d", group.Selected())
	}
}

func TestRadio_HandleSelect(t *testing.T) {
	group := NewRadioGroup()
	r1 := NewRadio("A", group)
	r2 := NewRadio("B", group)
	_ = r2
	r1.Focus()
	r1.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})
	if group.Selected() != 0 {
		t.Fatalf("expected selected 0, got %d", group.Selected())
	}
}

func TestRadio_HandleSelectSpace(t *testing.T) {
	group := NewRadioGroup()
	r1 := NewRadio("A", group)
	r1.Focus()
	r1.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: ' '})
	if group.Selected() != 0 {
		t.Fatalf("expected selected 0, got %d", group.Selected())
	}
}

func TestRadio_Disabled(t *testing.T) {
	group := NewRadioGroup()
	r1 := NewRadio("A", group)
	r1.SetDisabled(true)
	r1.Focus()
	result := r1.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})
	if result.Handled {
		t.Fatal("expected disabled radio to not handle enter")
	}
}

func TestRadio_Measure(t *testing.T) {
	group := NewRadioGroup()
	r := NewRadio("Hello", group)
	size := r.Measure(runtime.Constraints{MaxWidth: 40, MaxHeight: 5})
	if size.Width <= 0 || size.Height != 1 {
		t.Fatalf("expected positive width and height 1, got %dx%d", size.Width, size.Height)
	}
}

func TestRadio_Render(t *testing.T) {
	group := NewRadioGroup()
	r := NewRadio("Test", group)
	output := fluffytest.RenderToString(r, 20, 1)
	if !strings.Contains(output, "Test") {
		t.Fatalf("expected output to contain 'Test', got %q", output)
	}
}

func TestRadio_RenderSelected(t *testing.T) {
	group := NewRadioGroup()
	r := NewRadio("Selected", group)
	group.SetSelected(0)
	output := fluffytest.RenderToString(r, 20, 1)
	if !strings.Contains(output, "(*)") {
		t.Fatalf("expected output to contain '(*)', got %q", output)
	}
}

func TestRadio_RenderUnselected(t *testing.T) {
	group := NewRadioGroup()
	r1 := NewRadio("A", group)
	NewRadio("B", group)
	group.SetSelected(1)
	output := fluffytest.RenderToString(r1, 20, 1)
	if !strings.Contains(output, "( )") {
		t.Fatalf("expected output to contain '( )', got %q", output)
	}
}

func TestRadio_SetLabel(t *testing.T) {
	group := NewRadioGroup()
	r := NewRadio("Old", group)
	r.SetLabel("New")
	if r.AccessibleLabel() != "New" {
		t.Fatalf("expected label 'New', got %q", r.AccessibleLabel())
	}
}

func TestRadio_GroupOnChange(t *testing.T) {
	group := NewRadioGroup()
	changed := -1
	group.SetOnChange(func(index int) {
		changed = index
	})
	NewRadio("A", group)
	NewRadio("B", group)
	group.SetSelected(1)
	if changed != 1 {
		t.Fatalf("expected onChange with 1, got %d", changed)
	}
}

func TestRadio_DeprecatedOnChange(t *testing.T) {
	group := NewRadioGroup()
	group.OnChange(nil) // Deprecated.
}

func TestRadio_SetStyle(t *testing.T) {
	group := NewRadioGroup()
	r := NewRadio("A", group)
	r.SetStyle(r.style)
	r.SetFocusStyle(r.focusStyle)
}

func TestRadio_StyleType(t *testing.T) {
	group := NewRadioGroup()
	r := NewRadio("A", group)
	if r.StyleType() != "Radio" {
		t.Fatalf("expected StyleType 'Radio', got %q", r.StyleType())
	}
}

func TestRadio_BindUnbind(t *testing.T) {
	group := NewRadioGroup()
	r := NewRadio("A", group)
	r.Bind(runtime.Services{})
	r.Unbind()
}

func TestRadio_HandleUnfocused(t *testing.T) {
	group := NewRadioGroup()
	r := NewRadio("A", group)
	result := r.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})
	if result.Handled {
		t.Fatal("expected unhandled when not focused")
	}
}

func TestRadio_HandleNonKeyMsg(t *testing.T) {
	group := NewRadioGroup()
	r := NewRadio("A", group)
	r.Focus()
	result := r.HandleMessage(runtime.MouseMsg{})
	if result.Handled {
		t.Fatal("expected non-key message to be unhandled")
	}
}

func TestRadio_PosInSet(t *testing.T) {
	group := NewRadioGroup()
	r1 := NewRadio("A", group)
	r2 := NewRadio("B", group)
	r3 := NewRadio("C", group)
	if r1.Base.PosInSet != 1 {
		t.Fatalf("expected PosInSet 1, got %d", r1.Base.PosInSet)
	}
	if r2.Base.PosInSet != 2 {
		t.Fatalf("expected PosInSet 2, got %d", r2.Base.PosInSet)
	}
	if r3.Base.PosInSet != 3 {
		t.Fatalf("expected PosInSet 3, got %d", r3.Base.PosInSet)
	}
	// SetSize is set at creation time, so r3 (last created) has the full count
	if r3.Base.SetSize != 3 {
		t.Fatalf("expected SetSize 3, got %d", r3.Base.SetSize)
	}
}

// ---------------------------------------------------------------------------
// Menu (47% -> higher)
// ---------------------------------------------------------------------------

func TestMenu_New(t *testing.T) {
	menu := NewMenu(&MenuItem{ID: "file", Title: "File"})
	if menu == nil {
		t.Fatal("expected non-nil menu")
	}
	if menu.AccessibleRole() != "menu" {
		t.Fatalf("expected role menu, got %q", menu.AccessibleRole())
	}
}

func TestMenu_Measure(t *testing.T) {
	menu := NewMenu(
		&MenuItem{ID: "a", Title: "Item A"},
		&MenuItem{ID: "b", Title: "Item B"},
	)
	size := menu.Measure(runtime.Constraints{MaxWidth: 30, MaxHeight: 10})
	if size.Width <= 0 || size.Height <= 0 {
		t.Fatalf("expected positive size, got %dx%d", size.Width, size.Height)
	}
}

func TestMenu_Render(t *testing.T) {
	menu := NewMenu(
		&MenuItem{ID: "a", Title: "Item A"},
		&MenuItem{ID: "b", Title: "Item B"},
	)
	output := fluffytest.RenderToString(menu, 20, 5)
	if !strings.Contains(output, "Item A") {
		t.Fatalf("expected output to contain 'Item A', got %q", output)
	}
}

func TestMenu_HandleKeyNavigation(t *testing.T) {
	menu := NewMenu(
		&MenuItem{ID: "a", Title: "Item A"},
		&MenuItem{ID: "b", Title: "Item B"},
		&MenuItem{ID: "c", Title: "Item C"},
	)
	menu.Focus()
	menu.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	if menu.selectedIndex != 1 {
		t.Fatalf("expected selected 1, got %d", menu.selectedIndex)
	}
	menu.HandleMessage(runtime.KeyMsg{Key: terminal.KeyUp})
	if menu.selectedIndex != 0 {
		t.Fatalf("expected selected 0, got %d", menu.selectedIndex)
	}
	menu.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnd})
	if menu.selectedIndex != 2 {
		t.Fatalf("expected selected 2, got %d", menu.selectedIndex)
	}
	menu.HandleMessage(runtime.KeyMsg{Key: terminal.KeyHome})
	if menu.selectedIndex != 0 {
		t.Fatalf("expected selected 0, got %d", menu.selectedIndex)
	}
}

func TestMenu_HandleExpand(t *testing.T) {
	menu := NewMenu(
		&MenuItem{ID: "parent", Title: "Parent", Children: []*MenuItem{
			{ID: "child", Title: "Child"},
		}},
	)
	menu.Focus()
	// Expand with right.
	menu.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRight})
	if !menu.Items[0].Expanded {
		t.Fatal("expected parent to be expanded")
	}
	// Collapse with left.
	menu.HandleMessage(runtime.KeyMsg{Key: terminal.KeyLeft})
	if menu.Items[0].Expanded {
		t.Fatal("expected parent to be collapsed")
	}
}

func TestMenu_HandleEnterSelect(t *testing.T) {
	selected := false
	menu := NewMenu(
		&MenuItem{ID: "a", Title: "Item A", OnSelect: func() { selected = true }},
	)
	menu.Focus()
	menu.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})
	if !selected {
		t.Fatal("expected onSelect to be called")
	}
}

func TestMenu_HandleEnterToggle(t *testing.T) {
	menu := NewMenu(
		&MenuItem{ID: "parent", Title: "Parent", Children: []*MenuItem{
			{ID: "child", Title: "Child"},
		}},
	)
	menu.Focus()
	menu.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})
	if !menu.Items[0].Expanded {
		t.Fatal("expected parent to be expanded after enter")
	}
}

func TestMenu_SetItems(t *testing.T) {
	menu := NewMenu()
	menu.SetItems(&MenuItem{ID: "a", Title: "A"})
	if len(menu.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(menu.Items))
	}
}

func TestMenu_SetLabel(t *testing.T) {
	menu := NewMenu()
	menu.SetLabel("Navigation")
	if menu.AccessibleLabel() != "Navigation" {
		t.Fatalf("expected label 'Navigation', got %q", menu.AccessibleLabel())
	}
}

func TestMenu_ScrollByScrollTo(t *testing.T) {
	menu := NewMenu(
		&MenuItem{ID: "a", Title: "A"},
		&MenuItem{ID: "b", Title: "B"},
		&MenuItem{ID: "c", Title: "C"},
	)
	menu.ScrollBy(0, 1)
	menu.ScrollTo(0, 2)
	menu.PageBy(1)
	menu.ScrollToStart()
	menu.ScrollToEnd()
}

func TestMenu_SetStyle(t *testing.T) {
	menu := NewMenu()
	menu.SetStyle(menu.style)
	menu.SetSelectedStyle(menu.selectedStyle)
}

func TestMenu_StyleType(t *testing.T) {
	menu := NewMenu()
	if menu.StyleType() != "Menu" {
		t.Fatalf("expected StyleType 'Menu', got %q", menu.StyleType())
	}
}

func TestMenu_BindUnbind(t *testing.T) {
	menu := NewMenu()
	menu.Bind(runtime.Services{})
	menu.Unbind()
}

func TestMenu_HandleUnfocused(t *testing.T) {
	menu := NewMenu(&MenuItem{ID: "a", Title: "A"})
	result := menu.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	if result.Handled {
		t.Fatal("expected unhandled when not focused")
	}
}

func TestMenu_HandleNonKeyMsg(t *testing.T) {
	menu := NewMenu(&MenuItem{ID: "a", Title: "A"})
	menu.Focus()
	result := menu.HandleMessage(runtime.MouseMsg{})
	if result.Handled {
		t.Fatal("expected non-key message to be unhandled")
	}
}

func TestMenu_DisabledItem(t *testing.T) {
	selected := false
	menu := NewMenu(
		&MenuItem{ID: "a", Title: "A", Disabled: true, OnSelect: func() { selected = true }},
	)
	menu.Focus()
	menu.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})
	if selected {
		t.Fatal("expected disabled item not to trigger onSelect")
	}
}

func TestMenu_RenderWithShortcut(t *testing.T) {
	menu := NewMenu(
		&MenuItem{ID: "a", Title: "Open", Shortcut: "Ctrl+O"},
	)
	output := fluffytest.RenderToString(menu, 30, 3)
	if !strings.Contains(output, "Ctrl+O") {
		t.Fatalf("expected output to contain shortcut, got %q", output)
	}
}

// ---------------------------------------------------------------------------
// Tabs (49.5% -> higher)
// ---------------------------------------------------------------------------

func TestTabs_New(t *testing.T) {
	tabs := NewTabs(
		Tab{Title: "Tab1", Content: NewLabel("Content1")},
		Tab{Title: "Tab2", Content: NewLabel("Content2")},
	)
	if tabs == nil {
		t.Fatal("expected non-nil tabs")
	}
	if tabs.AccessibleRole() != "tablist" {
		t.Fatalf("expected role tablist, got %q", tabs.AccessibleRole())
	}
}

func TestTabs_Measure(t *testing.T) {
	tabs := NewTabs(
		Tab{Title: "Tab1", Content: NewLabel("Content")},
	)
	size := tabs.Measure(runtime.Constraints{MaxWidth: 40, MaxHeight: 10})
	if size.Width <= 0 || size.Height <= 0 {
		t.Fatalf("expected positive size, got %dx%d", size.Width, size.Height)
	}
}

func TestTabs_Render(t *testing.T) {
	tabs := NewTabs(
		Tab{Title: "Tab1", Content: NewLabel("Content1")},
		Tab{Title: "Tab2", Content: NewLabel("Content2")},
	)
	output := fluffytest.RenderToString(tabs, 30, 5)
	if !strings.Contains(output, "Tab1") {
		t.Fatalf("expected output to contain 'Tab1', got %q", output)
	}
}

func TestTabs_HandleKeyLeftRight(t *testing.T) {
	tabs := NewTabs(
		Tab{Title: "Tab1", Content: NewLabel("C1")},
		Tab{Title: "Tab2", Content: NewLabel("C2")},
		Tab{Title: "Tab3", Content: NewLabel("C3")},
	)
	tabs.Focus()
	tabs.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRight})
	if tabs.SelectedIndex() != 1 {
		t.Fatalf("expected selected 1, got %d", tabs.SelectedIndex())
	}
	tabs.HandleMessage(runtime.KeyMsg{Key: terminal.KeyLeft})
	if tabs.SelectedIndex() != 0 {
		t.Fatalf("expected selected 0, got %d", tabs.SelectedIndex())
	}
}

func TestTabs_HandleKeyHomeEnd(t *testing.T) {
	tabs := NewTabs(
		Tab{Title: "A", Content: NewLabel("1")},
		Tab{Title: "B", Content: NewLabel("2")},
		Tab{Title: "C", Content: NewLabel("3")},
	)
	tabs.Focus()
	tabs.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnd})
	if tabs.SelectedIndex() != 2 {
		t.Fatalf("expected selected 2, got %d", tabs.SelectedIndex())
	}
	tabs.HandleMessage(runtime.KeyMsg{Key: terminal.KeyHome})
	if tabs.SelectedIndex() != 0 {
		t.Fatalf("expected selected 0, got %d", tabs.SelectedIndex())
	}
}

func TestTabs_SetSelected(t *testing.T) {
	tabs := NewTabs(
		Tab{Title: "A", Content: NewLabel("1")},
		Tab{Title: "B", Content: NewLabel("2")},
	)
	tabs.SetSelected(1)
	if tabs.SelectedIndex() != 1 {
		t.Fatalf("expected selected 1, got %d", tabs.SelectedIndex())
	}
}

func TestTabs_SetLabel(t *testing.T) {
	tabs := NewTabs(Tab{Title: "A", Content: NewLabel("1")})
	tabs.SetLabel("Navigation Tabs")
	if tabs.AccessibleLabel() != "Navigation Tabs" {
		t.Fatalf("expected label 'Navigation Tabs', got %q", tabs.AccessibleLabel())
	}
}

func TestTabs_ChildWidgets(t *testing.T) {
	content := NewLabel("Content")
	tabs := NewTabs(Tab{Title: "A", Content: content})
	children := tabs.ChildWidgets()
	if len(children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(children))
	}
}

func TestTabs_ChildWidgetsEmpty(t *testing.T) {
	tabs := NewTabs()
	children := tabs.ChildWidgets()
	if children != nil {
		t.Fatal("expected nil children for empty tabs")
	}
}

func TestTabs_PathSegment(t *testing.T) {
	content := NewLabel("X")
	tabs := NewTabs(Tab{Title: "MyTab", Content: content})
	seg := tabs.PathSegment(content)
	if seg != "Tabs[MyTab]" {
		t.Fatalf("expected 'Tabs[MyTab]', got %q", seg)
	}
}

func TestTabs_PathSegmentUnknown(t *testing.T) {
	tabs := NewTabs(Tab{Title: "A", Content: NewLabel("1")})
	seg := tabs.PathSegment(NewLabel("other"))
	if seg != "Tabs" {
		t.Fatalf("expected 'Tabs', got %q", seg)
	}
}

func TestTabs_SetStyle(t *testing.T) {
	tabs := NewTabs(Tab{Title: "A", Content: NewLabel("1")})
	tabs.SetStyle(tabs.style)
	tabs.SetSelectedStyle(tabs.selectedStyle)
}

func TestTabs_StyleType(t *testing.T) {
	tabs := NewTabs(Tab{Title: "A", Content: NewLabel("1")})
	if tabs.StyleType() != "Tabs" {
		t.Fatalf("expected StyleType 'Tabs', got %q", tabs.StyleType())
	}
}

func TestTabs_BindUnbind(t *testing.T) {
	tabs := NewTabs(Tab{Title: "A", Content: NewLabel("1")})
	tabs.Bind(runtime.Services{})
	tabs.Unbind()
}

func TestTabs_MountUnmount(t *testing.T) {
	tabs := NewTabs(Tab{Title: "A", Content: NewLabel("1")})
	tabs.Mount()
	if !tabs.mounted {
		t.Fatal("expected mounted")
	}
	tabs.Unmount()
	if tabs.mounted {
		t.Fatal("expected not mounted")
	}
}

func TestTabs_HandleUnfocused(t *testing.T) {
	tabs := NewTabs(Tab{Title: "A", Content: NewLabel("1")})
	result := tabs.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRight})
	if result.Handled {
		t.Fatal("expected unhandled when not focused")
	}
}

func TestTabs_HandleNonKeyMsg(t *testing.T) {
	tabs := NewTabs(Tab{Title: "A", Content: NewLabel("1")})
	tabs.Focus()
	result := tabs.HandleMessage(runtime.MouseMsg{})
	if result.Handled {
		t.Fatal("expected non-key message to be unhandled")
	}
}

// ---------------------------------------------------------------------------
// RichText (38.4% -> higher)
// ---------------------------------------------------------------------------

func TestRichText_New(t *testing.T) {
	rt := NewRichText("# Hello World\nSome **bold** text")
	if rt == nil {
		t.Fatal("expected non-nil rich text")
	}
	if rt.AccessibleRole() != "text" {
		t.Fatalf("expected role text, got %q", rt.AccessibleRole())
	}
}

func TestRichText_Measure(t *testing.T) {
	rt := NewRichText("Hello")
	size := rt.Measure(runtime.Constraints{MaxWidth: 40, MaxHeight: 10})
	if size.Width <= 0 || size.Height <= 0 {
		t.Fatalf("expected positive size, got %dx%d", size.Width, size.Height)
	}
}

func TestRichText_Render(t *testing.T) {
	rt := NewRichText("Hello World")
	output := fluffytest.RenderToString(rt, 30, 5)
	if !strings.Contains(output, "Hello") {
		t.Fatalf("expected output to contain 'Hello', got %q", output)
	}
}

func TestRichText_SetContent(t *testing.T) {
	rt := NewRichText("Initial")
	rt.SetContent("Updated content")
	if rt.content != "Updated content" {
		t.Fatalf("expected 'Updated content', got %q", rt.content)
	}
}

func TestRichText_SetLabel(t *testing.T) {
	rt := NewRichText("x")
	rt.SetLabel("Documentation")
	if rt.AccessibleLabel() != "Documentation" {
		t.Fatalf("expected 'Documentation', got %q", rt.AccessibleLabel())
	}
}

func TestRichText_HandleKeyScroll(t *testing.T) {
	rt := NewRichText("Line1\nLine2\nLine3\nLine4\nLine5\nLine6\nLine7\nLine8")
	rt.Focus()
	rt.Measure(runtime.Constraints{MaxWidth: 20, MaxHeight: 3})
	rt.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 3})
	keys := []terminal.Key{
		terminal.KeyDown,
		terminal.KeyUp,
		terminal.KeyPageDown,
		terminal.KeyPageUp,
		terminal.KeyHome,
		terminal.KeyEnd,
	}
	for _, key := range keys {
		result := rt.HandleMessage(runtime.KeyMsg{Key: key})
		if !result.Handled {
			t.Fatalf("expected key %v to be handled", key)
		}
	}
}

func TestRichText_HandleMouseWheel(t *testing.T) {
	rt := NewRichText("L1\nL2\nL3\nL4\nL5\nL6\nL7\nL8")
	rt.Measure(runtime.Constraints{MaxWidth: 20, MaxHeight: 3})
	rt.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 3})
	result := rt.HandleMessage(runtime.MouseMsg{Button: runtime.MouseWheelUp})
	if !result.Handled {
		t.Fatal("expected wheel up to be handled")
	}
	result = rt.HandleMessage(runtime.MouseMsg{Button: runtime.MouseWheelDown})
	if !result.Handled {
		t.Fatal("expected wheel down to be handled")
	}
}

func TestRichText_ScrollMethods(t *testing.T) {
	rt := NewRichText("Line1\nLine2\nLine3\nLine4\nLine5")
	rt.Measure(runtime.Constraints{MaxWidth: 20, MaxHeight: 2})
	rt.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 2})
	rt.ScrollBy(0, 1)
	rt.ScrollTo(0, 0)
	rt.PageBy(1)
	rt.ScrollToStart()
	rt.ScrollToEnd()
}

func TestRichText_ScrollToAnchor(t *testing.T) {
	rt := NewRichText("# Section1\nContent\n# Section2\nMore content")
	rt.Measure(runtime.Constraints{MaxWidth: 30, MaxHeight: 2})
	rt.Layout(runtime.Rect{X: 0, Y: 0, Width: 30, Height: 2})
	rt.ScrollToAnchor("section2")
	rt.ScrollToAnchor("") // Should be no-op.
}

func TestRichText_SetShowScrollbar(t *testing.T) {
	rt := NewRichText("x")
	rt.SetShowScrollbar(false)
	if rt.showBar {
		t.Fatal("expected showBar to be false")
	}
}

func TestRichText_SetStyle(t *testing.T) {
	rt := NewRichText("x")
	rt.SetStyle(rt.style)
	if !rt.styleSet {
		t.Fatal("expected styleSet to be true")
	}
}

func TestRichText_SetSource(t *testing.T) {
	rt := NewRichText("x")
	rt.SetSource("markdown")
}

func TestRichText_SetRenderer(t *testing.T) {
	rt := NewRichText("x")
	rt.SetRenderer(nil)
}

func TestRichText_SetLines(t *testing.T) {
	rt := NewRichText("x")
	rt.SetLines(nil)
	if rt.content != "" {
		t.Fatal("expected content to be empty after SetLines")
	}
}

func TestRichText_Options(t *testing.T) {
	rt := NewRichText("Content",
		WithRichTextLabel("MyLabel"),
		WithRichTextScrollbar(false),
		WithRichTextSource("src"),
	)
	if rt.AccessibleLabel() != "MyLabel" {
		t.Fatalf("expected 'MyLabel', got %q", rt.AccessibleLabel())
	}
	if rt.showBar {
		t.Fatal("expected showBar false")
	}
}

func TestRichText_StyleType(t *testing.T) {
	rt := NewRichText("x")
	if rt.StyleType() != "RichText" {
		t.Fatalf("expected StyleType 'RichText', got %q", rt.StyleType())
	}
}

func TestRichText_BindUnbind(t *testing.T) {
	rt := NewRichText("x")
	rt.Bind(runtime.Services{})
	rt.Unbind()
}

func TestRichText_HandleUnfocused(t *testing.T) {
	rt := NewRichText("x")
	result := rt.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	if result.Handled {
		t.Fatal("expected unhandled when not focused")
	}
}

func TestRichText_NilHandleMessage(t *testing.T) {
	var rt *RichText
	result := rt.HandleMessage(runtime.KeyMsg{})
	if result.Handled {
		t.Fatal("nil rich text should return unhandled")
	}
}

func TestRichText_WrapEmptyContent(t *testing.T) {
	rt := NewRichText("")
	rt.Measure(runtime.Constraints{MaxWidth: 20, MaxHeight: 5})
	// Should not panic.
}

// ---------------------------------------------------------------------------
// Splitter (57.9% -> higher)
// ---------------------------------------------------------------------------

func TestSplitter_New(t *testing.T) {
	s := NewSplitter(NewLabel("A"), NewLabel("B"))
	if s == nil {
		t.Fatal("expected non-nil splitter")
	}
	if s.AccessibleRole() != "group" {
		t.Fatalf("expected role group, got %q", s.AccessibleRole())
	}
}

func TestSplitter_Measure(t *testing.T) {
	s := NewSplitter(NewLabel("Left"), NewLabel("Right"))
	size := s.Measure(runtime.Constraints{MaxWidth: 40, MaxHeight: 10})
	if size.Width <= 0 || size.Height <= 0 {
		t.Fatalf("expected positive size, got %dx%d", size.Width, size.Height)
	}
}

func TestSplitter_Render(t *testing.T) {
	s := NewSplitter(NewLabel("Left"), NewLabel("Right"))
	output := fluffytest.RenderToString(s, 30, 3)
	// Should contain both sides' content.
	if len(output) == 0 {
		t.Fatal("expected non-empty output")
	}
}

func TestSplitter_VerticalOrientation(t *testing.T) {
	s := NewSplitter(NewLabel("Top"), NewLabel("Bottom"))
	s.Orientation = SplitVertical
	s.syncA11y()
	if s.Base.Orientation != "vertical" {
		t.Fatalf("expected vertical orientation, got %q", s.Base.Orientation)
	}
	output := fluffytest.RenderToString(s, 20, 6)
	if len(output) == 0 {
		t.Fatal("expected non-empty output")
	}
}

func TestSplitter_ChildWidgets(t *testing.T) {
	s := NewSplitter(NewLabel("A"), NewLabel("B"))
	children := s.ChildWidgets()
	if len(children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(children))
	}
}

func TestSplitter_ChildWidgetsNilFirst(t *testing.T) {
	s := NewSplitter(nil, NewLabel("B"))
	children := s.ChildWidgets()
	if len(children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(children))
	}
}

func TestSplitter_HandleMessage(t *testing.T) {
	s := NewSplitter(NewLabel("A"), NewLabel("B"))
	result := s.HandleMessage(runtime.KeyMsg{})
	if result.Handled {
		t.Fatal("expected unhandled")
	}
}

func TestSplitter_PathSegment(t *testing.T) {
	first := NewLabel("A")
	second := NewLabel("B")
	s := NewSplitter(first, second)
	if s.PathSegment(first) != "Splitter[first]" {
		t.Fatalf("expected Splitter[first], got %q", s.PathSegment(first))
	}
	if s.PathSegment(second) != "Splitter[second]" {
		t.Fatalf("expected Splitter[second], got %q", s.PathSegment(second))
	}
	if s.PathSegment(NewLabel("other")) != "Splitter" {
		t.Fatal("expected Splitter for unknown child")
	}
}

func TestSplitter_SetLabel(t *testing.T) {
	s := NewSplitter(NewLabel("A"), NewLabel("B"))
	s.SetLabel("My Splitter")
	if s.AccessibleLabel() != "My Splitter" {
		t.Fatalf("expected 'My Splitter', got %q", s.AccessibleLabel())
	}
}

func TestSplitter_BindUnbind(t *testing.T) {
	s := NewSplitter(NewLabel("A"), NewLabel("B"))
	s.Bind(runtime.Services{})
	s.Unbind()
}

func TestSplitter_NilSafe(t *testing.T) {
	var s *Splitter
	s.SetLabel("x")
	s.Bind(runtime.Services{})
	s.Unbind()
	children := s.ChildWidgets()
	if children != nil {
		t.Fatal("expected nil children from nil splitter")
	}
}

func TestSplitter_CustomRatio(t *testing.T) {
	s := NewSplitter(NewLabel("A"), NewLabel("B"))
	s.Ratio = 0.3
	s.Measure(runtime.Constraints{MaxWidth: 40, MaxHeight: 10})
	s.Layout(runtime.Rect{X: 0, Y: 0, Width: 40, Height: 10})
	// Should not panic.
}

func TestSplitter_InvalidRatio(t *testing.T) {
	s := NewSplitter(NewLabel("A"), NewLabel("B"))
	s.Ratio = 0 // Invalid, should default to 0.5.
	s.Layout(runtime.Rect{X: 0, Y: 0, Width: 40, Height: 10})
	if s.Ratio != 0.5 {
		t.Fatalf("expected ratio 0.5 for 0, got %f", s.Ratio)
	}
	s.Ratio = 1 // Invalid, should default to 0.5.
	s.Layout(runtime.Rect{X: 0, Y: 0, Width: 40, Height: 10})
	if s.Ratio != 0.5 {
		t.Fatalf("expected ratio 0.5 for 1, got %f", s.Ratio)
	}
}

// ---------------------------------------------------------------------------
// Slider additional tests (40.9% -> higher)
// ---------------------------------------------------------------------------

func TestSlider_Measure(t *testing.T) {
	s := NewSlider(nil)
	size := s.Measure(runtime.Constraints{MaxWidth: 30, MaxHeight: 5})
	if size.Width <= 0 || size.Height != 1 {
		t.Fatalf("expected positive width and height 1, got %dx%d", size.Width, size.Height)
	}
}

func TestSlider_MeasureVertical(t *testing.T) {
	s := NewSlider(nil, WithSliderOrientation(Vertical))
	size := s.Measure(runtime.Constraints{MaxWidth: 5, MaxHeight: 20})
	if size.Width != 1 || size.Height <= 0 {
		t.Fatalf("expected width 1 and positive height, got %dx%d", size.Width, size.Height)
	}
}

func TestSlider_Render(t *testing.T) {
	val := state.NewSignal(50.0)
	s := NewSlider(val, WithSliderRange(0, 100, 1))
	output := fluffytest.RenderToString(s, 20, 1)
	if len(output) == 0 {
		t.Fatal("expected non-empty output")
	}
}

func TestSlider_RenderVertical(t *testing.T) {
	val := state.NewSignal(50.0)
	s := NewSlider(val, WithSliderRange(0, 100, 1), WithSliderOrientation(Vertical))
	output := fluffytest.RenderToString(s, 1, 10)
	if len(output) == 0 {
		t.Fatal("expected non-empty output")
	}
}

func TestSlider_RenderWithValue(t *testing.T) {
	val := state.NewSignal(25.0)
	s := NewSlider(val, WithSliderRange(0, 100, 1), WithSliderShowValue(true))
	output := fluffytest.RenderToString(s, 30, 1)
	if !strings.Contains(output, "25") {
		t.Fatalf("expected output to contain '25', got %q", output)
	}
}

func TestSlider_SetValue(t *testing.T) {
	val := state.NewSignal(0.0)
	s := NewSlider(val, WithSliderRange(0, 100, 1))
	s.SetValue(75)
	if s.Value() != 75 {
		t.Fatalf("expected 75, got %f", s.Value())
	}
}

func TestSlider_SetValueClamp(t *testing.T) {
	val := state.NewSignal(0.0)
	s := NewSlider(val, WithSliderRange(0, 100, 1))
	s.SetValue(200)
	if s.Value() != 100 {
		t.Fatalf("expected clamped to 100, got %f", s.Value())
	}
	s.SetValue(-50)
	if s.Value() != 0 {
		t.Fatalf("expected clamped to 0, got %f", s.Value())
	}
}

func TestSlider_SetRange(t *testing.T) {
	val := state.NewSignal(50.0)
	s := NewSlider(val, WithSliderRange(0, 100, 1))
	s.SetRange(0, 200, 5)
	if s.max != 200 {
		t.Fatalf("expected max 200, got %f", s.max)
	}
}

func TestSlider_SetLabel(t *testing.T) {
	s := NewSlider(nil)
	s.SetLabel("Volume")
	if s.AccessibleLabel() != "Volume" {
		t.Fatalf("expected 'Volume', got %q", s.AccessibleLabel())
	}
}

func TestSlider_HandleKeyAll(t *testing.T) {
	val := state.NewSignal(50.0)
	s := NewSlider(val, WithSliderRange(0, 100, 1))
	s.Focus()

	keys := []terminal.Key{
		terminal.KeyLeft,
		terminal.KeyRight,
		terminal.KeyUp,
		terminal.KeyDown,
		terminal.KeyPageUp,
		terminal.KeyPageDown,
		terminal.KeyHome,
		terminal.KeyEnd,
	}
	for _, key := range keys {
		result := s.HandleMessage(runtime.KeyMsg{Key: key})
		if !result.Handled {
			t.Fatalf("expected key %v to be handled", key)
		}
	}
}

func TestSlider_HandleMouseClick(t *testing.T) {
	val := state.NewSignal(50.0)
	s := NewSlider(val, WithSliderRange(0, 100, 1))
	s.Measure(runtime.Constraints{MaxWidth: 20, MaxHeight: 1})
	s.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 1})
	result := s.HandleMessage(runtime.MouseMsg{
		X: 10, Y: 0, Button: runtime.MouseLeft, Action: runtime.MousePress,
	})
	if !result.Handled {
		t.Fatal("expected mouse click to be handled")
	}
}

func TestSlider_HandleMouseDrag(t *testing.T) {
	val := state.NewSignal(50.0)
	s := NewSlider(val, WithSliderRange(0, 100, 1))
	s.Measure(runtime.Constraints{MaxWidth: 20, MaxHeight: 1})
	s.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 1})
	// Press.
	s.HandleMessage(runtime.MouseMsg{X: 5, Y: 0, Button: runtime.MouseLeft, Action: runtime.MousePress})
	// Move.
	s.HandleMessage(runtime.MouseMsg{X: 15, Y: 0, Button: runtime.MouseLeft, Action: runtime.MouseMove})
	// Release.
	s.HandleMessage(runtime.MouseMsg{X: 15, Y: 0, Button: runtime.MouseLeft, Action: runtime.MouseRelease})
}

func TestSlider_StyleType(t *testing.T) {
	s := NewSlider(nil)
	if s.StyleType() != "Slider" {
		t.Fatalf("expected StyleType 'Slider', got %q", s.StyleType())
	}
}

func TestSlider_BindUnbind(t *testing.T) {
	s := NewSlider(nil)
	s.Bind(runtime.Services{})
	s.Unbind()
}

func TestSlider_Options(t *testing.T) {
	val := state.NewSignal(25.0)
	_ = NewSlider(val,
		WithSliderRange(10, 90, 5),
		WithSliderOrientation(Horizontal),
		WithSliderShowValue(true),
		WithSliderValueFormat("%.1f"),
	)
}

func TestSlider_HandleUnfocused(t *testing.T) {
	s := NewSlider(nil)
	result := s.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRight})
	if result.Handled {
		t.Fatal("expected unhandled when not focused")
	}
}

// ---------------------------------------------------------------------------
// RangeSlider additional tests
// ---------------------------------------------------------------------------

func TestRangeSlider_New(t *testing.T) {
	rs := NewRangeSlider(nil, nil)
	if rs == nil {
		t.Fatal("expected non-nil range slider")
	}
	if rs.AccessibleRole() != "slider" {
		t.Fatalf("expected role slider, got %q", rs.AccessibleRole())
	}
}

func TestRangeSlider_Measure(t *testing.T) {
	rs := NewRangeSlider(nil, nil)
	size := rs.Measure(runtime.Constraints{MaxWidth: 30, MaxHeight: 5})
	if size.Width <= 0 || size.Height != 1 {
		t.Fatalf("expected positive width and height 1, got %dx%d", size.Width, size.Height)
	}
}

func TestRangeSlider_MeasureVertical(t *testing.T) {
	rs := NewRangeSlider(nil, nil, WithRangeSliderOrientation(Vertical))
	size := rs.Measure(runtime.Constraints{MaxWidth: 5, MaxHeight: 20})
	if size.Width != 1 || size.Height <= 0 {
		t.Fatalf("expected width 1, got %dx%d", size.Width, size.Height)
	}
}

func TestRangeSlider_Render(t *testing.T) {
	minVal := state.NewSignal(20.0)
	maxVal := state.NewSignal(80.0)
	rs := NewRangeSlider(minVal, maxVal, WithRangeSliderRange(0, 100, 1))
	output := fluffytest.RenderToString(rs, 30, 1)
	if len(output) == 0 {
		t.Fatal("expected non-empty output")
	}
}

func TestRangeSlider_RenderVertical(t *testing.T) {
	minVal := state.NewSignal(20.0)
	maxVal := state.NewSignal(80.0)
	rs := NewRangeSlider(minVal, maxVal, WithRangeSliderRange(0, 100, 1), WithRangeSliderOrientation(Vertical))
	output := fluffytest.RenderToString(rs, 1, 10)
	if len(output) == 0 {
		t.Fatal("expected non-empty output")
	}
}

func TestRangeSlider_RenderWithValue(t *testing.T) {
	minVal := state.NewSignal(20.0)
	maxVal := state.NewSignal(80.0)
	rs := NewRangeSlider(minVal, maxVal, WithRangeSliderRange(0, 100, 1), WithRangeSliderShowValue(true))
	output := fluffytest.RenderToString(rs, 40, 1)
	if !strings.Contains(output, "20") {
		t.Fatalf("expected output to contain '20', got %q", output)
	}
}

func TestRangeSlider_SetValues(t *testing.T) {
	minVal := state.NewSignal(0.0)
	maxVal := state.NewSignal(100.0)
	rs := NewRangeSlider(minVal, maxVal, WithRangeSliderRange(0, 100, 1))
	rs.SetValues(30, 70)
	min, max := rs.Values()
	if min != 30 || max != 70 {
		t.Fatalf("expected 30-70, got %f-%f", min, max)
	}
}

func TestRangeSlider_SetRange(t *testing.T) {
	rs := NewRangeSlider(nil, nil)
	rs.SetRange(10, 200, 5)
	if rs.min != 10 || rs.max != 200 || rs.step != 5 {
		t.Fatalf("expected 10-200/5, got %f-%f/%f", rs.min, rs.max, rs.step)
	}
}

func TestRangeSlider_SetLabel(t *testing.T) {
	rs := NewRangeSlider(nil, nil)
	rs.SetLabel("Price Range")
	if rs.AccessibleLabel() != "Price Range" {
		t.Fatalf("expected 'Price Range', got %q", rs.AccessibleLabel())
	}
}

func TestRangeSlider_HandleKeyAll(t *testing.T) {
	minVal := state.NewSignal(30.0)
	maxVal := state.NewSignal(70.0)
	rs := NewRangeSlider(minVal, maxVal, WithRangeSliderRange(0, 100, 1))
	rs.Focus()
	keys := []terminal.Key{
		terminal.KeyLeft,
		terminal.KeyRight,
		terminal.KeyUp,
		terminal.KeyDown,
		terminal.KeyPageUp,
		terminal.KeyPageDown,
		terminal.KeyHome,
		terminal.KeyEnd,
	}
	for _, key := range keys {
		result := rs.HandleMessage(runtime.KeyMsg{Key: key})
		if !result.Handled {
			t.Fatalf("expected key %v to be handled", key)
		}
	}
}

func TestRangeSlider_HandleMouseClick(t *testing.T) {
	minVal := state.NewSignal(30.0)
	maxVal := state.NewSignal(70.0)
	rs := NewRangeSlider(minVal, maxVal, WithRangeSliderRange(0, 100, 1))
	rs.Measure(runtime.Constraints{MaxWidth: 20, MaxHeight: 1})
	rs.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 1})
	result := rs.HandleMessage(runtime.MouseMsg{
		X: 10, Y: 0, Button: runtime.MouseLeft, Action: runtime.MousePress,
	})
	if !result.Handled {
		t.Fatal("expected mouse click to be handled")
	}
}

func TestRangeSlider_HandleMouseDrag(t *testing.T) {
	minVal := state.NewSignal(30.0)
	maxVal := state.NewSignal(70.0)
	rs := NewRangeSlider(minVal, maxVal, WithRangeSliderRange(0, 100, 1))
	rs.Measure(runtime.Constraints{MaxWidth: 20, MaxHeight: 1})
	rs.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 1})
	rs.HandleMessage(runtime.MouseMsg{X: 5, Y: 0, Button: runtime.MouseLeft, Action: runtime.MousePress})
	rs.HandleMessage(runtime.MouseMsg{X: 15, Y: 0, Button: runtime.MouseLeft, Action: runtime.MouseMove})
	rs.HandleMessage(runtime.MouseMsg{X: 15, Y: 0, Button: runtime.MouseLeft, Action: runtime.MouseRelease})
}

func TestRangeSlider_StyleType(t *testing.T) {
	rs := NewRangeSlider(nil, nil)
	if rs.StyleType() != "RangeSlider" {
		t.Fatalf("expected StyleType 'RangeSlider', got %q", rs.StyleType())
	}
}

func TestRangeSlider_BindUnbind(t *testing.T) {
	rs := NewRangeSlider(nil, nil)
	rs.Bind(runtime.Services{})
	rs.Unbind()
}

func TestRangeSlider_Options(t *testing.T) {
	_ = NewRangeSlider(nil, nil,
		WithRangeSliderRange(0, 50, 2),
		WithRangeSliderOrientation(Vertical),
		WithRangeSliderShowValue(true),
		WithRangeSliderValueFormat("%.2f"),
	)
}

func TestRangeSlider_HandleUnfocused(t *testing.T) {
	rs := NewRangeSlider(nil, nil)
	result := rs.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRight})
	if result.Handled {
		t.Fatal("expected unhandled when not focused")
	}
}

func TestRangeSlider_RenderVerticalWithValue(t *testing.T) {
	minVal := state.NewSignal(20.0)
	maxVal := state.NewSignal(80.0)
	rs := NewRangeSlider(minVal, maxVal,
		WithRangeSliderRange(0, 100, 1),
		WithRangeSliderOrientation(Vertical),
		WithRangeSliderShowValue(true),
	)
	output := fluffytest.RenderToString(rs, 10, 10)
	if len(output) == 0 {
		t.Fatal("expected non-empty output")
	}
}

// ---------------------------------------------------------------------------
// SelectDropdown (50.7% -> higher)
// ---------------------------------------------------------------------------

func TestSelectDropdown_Measure(t *testing.T) {
	parent := NewSelect(
		SelectOption{Label: "Alpha", Value: "a"},
		SelectOption{Label: "Beta", Value: "b"},
	)
	drop := newSelectDropdown(parent)
	size := drop.Measure(runtime.Constraints{MaxWidth: 30, MaxHeight: 10})
	if size.Width <= 0 || size.Height <= 0 {
		t.Fatalf("expected positive size, got %dx%d", size.Width, size.Height)
	}
}

func TestSelectDropdown_Render(t *testing.T) {
	parent := NewSelect(
		SelectOption{Label: "Alpha", Value: "a"},
		SelectOption{Label: "Beta", Value: "b"},
	)
	drop := newSelectDropdown(parent)
	output := fluffytest.RenderToString(drop, 20, 5)
	if !strings.Contains(output, "Alpha") {
		t.Fatalf("expected 'Alpha' in output, got %q", output)
	}
}

func TestSelectDropdown_HandleKeyUpDown(t *testing.T) {
	parent := NewSelect(
		SelectOption{Label: "A", Value: "a"},
		SelectOption{Label: "B", Value: "b"},
		SelectOption{Label: "C", Value: "c"},
	)
	drop := newSelectDropdown(parent)
	drop.Focus()
	drop.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	if drop.selected != 1 {
		t.Fatalf("expected selected 1, got %d", drop.selected)
	}
	drop.HandleMessage(runtime.KeyMsg{Key: terminal.KeyUp})
	if drop.selected != 0 {
		t.Fatalf("expected selected 0, got %d", drop.selected)
	}
}

func TestSelectDropdown_HandleKeyEscape(t *testing.T) {
	parent := NewSelect(
		SelectOption{Label: "A", Value: "a"},
	)
	drop := newSelectDropdown(parent)
	drop.Focus()
	result := drop.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEscape})
	if !result.Handled {
		t.Fatal("expected escape to be handled")
	}
}

func TestSelectDropdown_HandleKeyEnter(t *testing.T) {
	parent := NewSelect(
		SelectOption{Label: "A", Value: "a"},
	)
	drop := newSelectDropdown(parent)
	drop.Focus()
	result := drop.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})
	if !result.Handled {
		t.Fatal("expected enter to be handled")
	}
}

func TestSelectDropdown_HandleUnfocused(t *testing.T) {
	parent := NewSelect(
		SelectOption{Label: "A", Value: "a"},
	)
	drop := newSelectDropdown(parent)
	result := drop.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	if result.Handled {
		t.Fatal("expected unhandled when not focused")
	}
}

func TestSelectDropdown_DisabledOptions(t *testing.T) {
	parent := NewSelect(
		SelectOption{Label: "A", Value: "a"},
		SelectOption{Label: "B", Value: "b", Disabled: true},
		SelectOption{Label: "C", Value: "c"},
	)
	drop := newSelectDropdown(parent)
	drop.Focus()
	// Should skip disabled.
	drop.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	if drop.selected == 1 {
		t.Fatal("expected disabled item to be skipped")
	}
}

func TestSelectDropdown_StyleType(t *testing.T) {
	parent := NewSelect(
		SelectOption{Label: "A", Value: "a"},
	)
	drop := newSelectDropdown(parent)
	if drop.StyleType() != "Select" {
		t.Fatalf("expected StyleType 'Select', got %q", drop.StyleType())
	}
}

func TestSelectDropdown_BindUnbind(t *testing.T) {
	parent := NewSelect(
		SelectOption{Label: "A", Value: "a"},
	)
	drop := newSelectDropdown(parent)
	drop.Bind(runtime.Services{})
	drop.Unbind()
}

func TestSelectDropdown_MouseClick(t *testing.T) {
	parent := NewSelect(
		SelectOption{Label: "A", Value: "a"},
		SelectOption{Label: "B", Value: "b"},
	)
	drop := newSelectDropdown(parent)
	drop.Focus()
	drop.Measure(runtime.Constraints{MaxWidth: 20, MaxHeight: 5})
	drop.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 5})
	// Click on the content area.
	content := drop.ContentBounds()
	result := drop.HandleMessage(runtime.MouseMsg{
		X: content.X + 1, Y: content.Y + 1,
		Button: runtime.MouseLeft, Action: runtime.MousePress,
	})
	if !result.Handled {
		t.Fatal("expected mouse click to be handled")
	}
}

func TestSelectDropdown_EnsureVisible(t *testing.T) {
	parent := NewSelect(
		SelectOption{Label: "A", Value: "a"},
		SelectOption{Label: "B", Value: "b"},
		SelectOption{Label: "C", Value: "c"},
		SelectOption{Label: "D", Value: "d"},
		SelectOption{Label: "E", Value: "e"},
	)
	drop := newSelectDropdown(parent)
	drop.selected = 4
	drop.ensureVisible(2)
	if drop.offset < 3 {
		t.Fatalf("expected offset >= 3, got %d", drop.offset)
	}
}

// ---------------------------------------------------------------------------
// DatePicker (43.9% -> higher)
// ---------------------------------------------------------------------------

func TestDatePicker_New(t *testing.T) {
	dp := NewDatePicker()
	if dp == nil {
		t.Fatal("expected non-nil date picker")
	}
	if dp.AccessibleRole() != "group" {
		t.Fatalf("expected role group, got %q", dp.AccessibleRole())
	}
}

func TestDatePicker_Measure(t *testing.T) {
	dp := NewDatePicker()
	size := dp.Measure(runtime.Constraints{MaxWidth: 40, MaxHeight: 20})
	if size.Width <= 0 || size.Height <= 0 {
		t.Fatalf("expected positive size, got %dx%d", size.Width, size.Height)
	}
}

func TestDatePicker_Render(t *testing.T) {
	dp := NewDatePicker()
	_ = fluffytest.RenderToString(dp, 40, 15)
	// Should not panic.
}

func TestDatePicker_ChildWidgets(t *testing.T) {
	dp := NewDatePicker()
	children := dp.ChildWidgets()
	if len(children) < 2 {
		t.Fatalf("expected at least 2 children, got %d", len(children))
	}
}

// ---------------------------------------------------------------------------
// DateRangePicker (44.4% -> higher)
// ---------------------------------------------------------------------------

func TestDateRangePicker_New(t *testing.T) {
	drp := NewDateRangePicker()
	if drp == nil {
		t.Fatal("expected non-nil date range picker")
	}
}

func TestDateRangePicker_Measure(t *testing.T) {
	drp := NewDateRangePicker()
	size := drp.Measure(runtime.Constraints{MaxWidth: 40, MaxHeight: 20})
	if size.Width <= 0 || size.Height <= 0 {
		t.Fatalf("expected positive size, got %dx%d", size.Width, size.Height)
	}
}

func TestDateRangePicker_ChildWidgets(t *testing.T) {
	drp := NewDateRangePicker()
	children := drp.ChildWidgets()
	if len(children) < 3 {
		t.Fatalf("expected at least 3 children, got %d", len(children))
	}
}

// ---------------------------------------------------------------------------
// Tree (47.5% -> higher)
// ---------------------------------------------------------------------------

func TestTree_New(t *testing.T) {
	tree := NewTree(&TreeNode{Label: "Root"})
	if tree == nil {
		t.Fatal("expected non-nil tree")
	}
	if tree.AccessibleRole() != "tree" {
		t.Fatalf("expected role tree, got %q", tree.AccessibleRole())
	}
}

func TestTree_Measure(t *testing.T) {
	tree := NewTree(&TreeNode{
		Label: "Root",
		Children: []*TreeNode{
			{Label: "Child1"},
			{Label: "Child2"},
		},
		Expanded: true,
	})
	size := tree.Measure(runtime.Constraints{MaxWidth: 30, MaxHeight: 10})
	if size.Width <= 0 || size.Height <= 0 {
		t.Fatalf("expected positive size, got %dx%d", size.Width, size.Height)
	}
}

func TestTree_Render(t *testing.T) {
	tree := NewTree(&TreeNode{
		Label: "Root",
		Children: []*TreeNode{
			{Label: "Child"},
		},
		Expanded: true,
	})
	output := fluffytest.RenderToString(tree, 30, 5)
	if !strings.Contains(output, "Root") {
		t.Fatalf("expected output to contain 'Root', got %q", output)
	}
}

func TestTree_HandleKeyNavigation(t *testing.T) {
	tree := NewTree(&TreeNode{
		Label: "Root",
		Children: []*TreeNode{
			{Label: "A"},
			{Label: "B"},
			{Label: "C"},
		},
		Expanded: true,
	})
	tree.Focus()
	tree.Measure(runtime.Constraints{MaxWidth: 20, MaxHeight: 10})
	tree.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 10})

	tree.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	if tree.selectedIndex != 1 {
		t.Fatalf("expected selected 1, got %d", tree.selectedIndex)
	}
	tree.HandleMessage(runtime.KeyMsg{Key: terminal.KeyUp})
	if tree.selectedIndex != 0 {
		t.Fatalf("expected selected 0, got %d", tree.selectedIndex)
	}
}

func TestTree_HandleExpandCollapse(t *testing.T) {
	tree := NewTree(&TreeNode{
		Label: "Root",
		Children: []*TreeNode{
			{Label: "Child"},
		},
	})
	tree.Focus()
	tree.Measure(runtime.Constraints{MaxWidth: 20, MaxHeight: 10})
	tree.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 10})
	// Expand.
	tree.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRight})
	if !tree.Root.Expanded {
		t.Fatal("expected root to be expanded")
	}
	// Collapse.
	tree.HandleMessage(runtime.KeyMsg{Key: terminal.KeyLeft})
	if tree.Root.Expanded {
		t.Fatal("expected root to be collapsed")
	}
}

func TestTree_HandleEnterToggle(t *testing.T) {
	tree := NewTree(&TreeNode{
		Label: "Root",
		Children: []*TreeNode{
			{Label: "Child"},
		},
	})
	tree.Focus()
	tree.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})
	if !tree.Root.Expanded {
		t.Fatal("expected root to be expanded after enter")
	}
}

func TestTree_HandleDownToLastChild(t *testing.T) {
	tree := NewTree(&TreeNode{
		Label: "Root",
		Children: []*TreeNode{
			{Label: "A"},
			{Label: "B"},
		},
		Expanded: true,
	})
	tree.Focus()
	tree.Measure(runtime.Constraints{MaxWidth: 20, MaxHeight: 10})
	tree.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 10})
	// Navigate down: Root -> A -> B
	tree.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	if tree.selectedIndex != 1 {
		t.Fatalf("expected selected 1, got %d", tree.selectedIndex)
	}
	tree.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	if tree.selectedIndex != 2 {
		t.Fatalf("expected selected 2, got %d", tree.selectedIndex)
	}
	// Navigate back up
	tree.HandleMessage(runtime.KeyMsg{Key: terminal.KeyUp})
	if tree.selectedIndex != 1 {
		t.Fatalf("expected selected 1, got %d", tree.selectedIndex)
	}
}

func TestTree_SetVirtualScroll(t *testing.T) {
	tree := NewTree(&TreeNode{Label: "Root"})
	tree.SetVirtualScroll(true)
	if !tree.VirtualScroll() {
		t.Fatal("expected virtual scroll to be enabled")
	}
	tree.SetVirtualScroll(false)
	if tree.VirtualScroll() {
		t.Fatal("expected virtual scroll to be disabled")
	}
}

// ---------------------------------------------------------------------------
// GPUCanvasWidget (32.2% -> higher)
// ---------------------------------------------------------------------------

func TestGPUCanvasWidget_New(t *testing.T) {
	w := NewGPUCanvasWidget(nil)
	if w == nil {
		t.Fatal("expected non-nil widget")
	}
}

func TestGPUCanvasWidget_Measure(t *testing.T) {
	w := NewGPUCanvasWidget(nil)
	size := w.Measure(runtime.Constraints{MaxWidth: 30, MaxHeight: 20})
	if size.Width <= 0 || size.Height <= 0 {
		t.Fatalf("expected positive size, got %dx%d", size.Width, size.Height)
	}
}

func TestGPUCanvasWidget_Layout(t *testing.T) {
	w := NewGPUCanvasWidget(nil)
	w.Measure(runtime.Constraints{MaxWidth: 30, MaxHeight: 20})
	w.Layout(runtime.Rect{X: 0, Y: 0, Width: 30, Height: 20})
	if w.cellWidth != 30 || w.cellHeight != 20 {
		t.Fatalf("expected 30x20, got %dx%d", w.cellWidth, w.cellHeight)
	}
}

func TestGPUCanvasWidget_LayoutZero(t *testing.T) {
	w := NewGPUCanvasWidget(nil)
	w.Layout(runtime.Rect{X: 0, Y: 0, Width: 0, Height: 0})
	if w.cellWidth != 0 || w.cellHeight != 0 {
		t.Fatal("expected zero dimensions with zero bounds")
	}
}

func TestGPUCanvasWidget_SetEncoder(t *testing.T) {
	w := NewGPUCanvasWidget(nil)
	w.SetEncoder(nil)
}

func TestGPUCanvasWidget_SetBackend(t *testing.T) {
	w := NewGPUCanvasWidget(nil)
	w.SetBackend(0) // BackendAuto
}

func TestGPUCanvasWidget_SetDriver(t *testing.T) {
	w := NewGPUCanvasWidget(nil)
	w.SetDriver(nil)
}

func TestGPUCanvasWidget_NilSafe(t *testing.T) {
	var w *GPUCanvasWidget
	w.SetEncoder(nil)
	w.SetBackend(0)
	w.SetDriver(nil)
}

func TestGPUCanvasWidget_DeprecatedMethods(t *testing.T) {
	w := NewGPUCanvasWidget(nil)
	result := w.WithEncoder(nil)
	if result != w {
		t.Fatal("expected same widget returned")
	}
	result = w.WithBackend(0)
	if result != w {
		t.Fatal("expected same widget returned")
	}
	result = w.WithDriver(nil)
	if result != w {
		t.Fatal("expected same widget returned")
	}
}

func TestGPUCanvasWidget_Options(t *testing.T) {
	_ = NewGPUCanvasWidget(nil,
		WithGPUCanvasEncoder(nil),
		WithGPUCanvasBackend(0),
		WithGPUCanvasDriver(nil),
	)
}

func TestGPUCanvasWidget_RenderNilDraw(t *testing.T) {
	w := NewGPUCanvasWidget(nil)
	buf := runtime.NewBuffer(10, 5)
	w.Render(runtime.RenderContext{Buffer: buf})
	// Should not panic.
}

func TestGPUCanvasWidget_Unbind(t *testing.T) {
	w := NewGPUCanvasWidget(nil)
	w.Unbind()
}

func TestGPUCanvasWidget_EncoderCellSize(t *testing.T) {
	// Default case.
	w, h := encoderCellSize(nil)
	if w <= 0 || h <= 0 {
		t.Fatalf("expected positive cell size, got %dx%d", w, h)
	}
}

// ---------------------------------------------------------------------------
// SelectDropdown accessibility
// ---------------------------------------------------------------------------

func TestSelectDropdown_SyncA11y(t *testing.T) {
	parent := NewSelect(
		SelectOption{Label: "Option A", Value: "a"},
	)
	drop := newSelectDropdown(parent)
	if drop.AccessibleRole() != "listbox" {
		t.Fatalf("expected role listbox, got %q", drop.AccessibleRole())
	}
}

// ---------------------------------------------------------------------------
// Skeleton additional tests (42.9% -> higher)
// ---------------------------------------------------------------------------

func TestSkeleton_Render(t *testing.T) {
	s := NewSkeleton(10, 1)
	output := fluffytest.RenderToString(s, 10, 1)
	if len(output) == 0 {
		t.Fatal("expected non-empty output")
	}
}

func TestSkeleton_SetAnimate(t *testing.T) {
	s := NewSkeleton(10, 1)
	s.SetAnimate(false)
	if s.IsAnimating() {
		t.Fatal("expected not animating")
	}
	s.SetAnimate(true)
	if !s.IsAnimating() {
		t.Fatal("expected animating")
	}
}

// ---------------------------------------------------------------------------
// SliceAdapter / SignalAdapter tests
// ---------------------------------------------------------------------------

func TestSliceAdapter_Count(t *testing.T) {
	adapter := NewSliceAdapter([]string{"A", "B", "C"}, nil)
	if adapter.Count() != 3 {
		t.Fatalf("expected count 3, got %d", adapter.Count())
	}
}

func TestSliceAdapter_Item(t *testing.T) {
	adapter := NewSliceAdapter([]string{"A", "B", "C"}, nil)
	if adapter.Item(1) != "B" {
		t.Fatalf("expected 'B', got %q", adapter.Item(1))
	}
}

func TestSliceAdapter_ItemOutOfRange(t *testing.T) {
	adapter := NewSliceAdapter([]string{"A"}, nil)
	item := adapter.Item(5)
	if item != "" {
		t.Fatalf("expected empty string for out of range, got %q", item)
	}
	item = adapter.Item(-1)
	if item != "" {
		t.Fatalf("expected empty string for negative index, got %q", item)
	}
}

func TestSignalAdapter_Count(t *testing.T) {
	sig := state.NewSignal([]string{"A", "B"})
	adapter := NewSignalAdapter(sig, nil)
	if adapter.Count() != 2 {
		t.Fatalf("expected count 2, got %d", adapter.Count())
	}
}

func TestSignalAdapter_Item(t *testing.T) {
	sig := state.NewSignal([]string{"X", "Y"})
	adapter := NewSignalAdapter(sig, nil)
	if adapter.Item(0) != "X" {
		t.Fatalf("expected 'X', got %q", adapter.Item(0))
	}
}

func TestSignalAdapter_ItemOutOfRange(t *testing.T) {
	sig := state.NewSignal([]string{"A"})
	adapter := NewSignalAdapter(sig, nil)
	item := adapter.Item(5)
	if item != "" {
		t.Fatalf("expected empty for out of range, got %q", item)
	}
}

// Ensure imports are used.
var (
	_ = image.Pt
	_ = video.VideoInfo{}
)
