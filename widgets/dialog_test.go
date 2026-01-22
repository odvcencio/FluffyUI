package widgets

import (
	"testing"
	"time"

	"github.com/odvcencio/fluffy-ui/runtime"
	"github.com/odvcencio/fluffy-ui/terminal"
)

func TestDialog_BasicCreation(t *testing.T) {
	clicked := false
	dialog := NewDialog("Title", "Body text",
		DialogButton{Label: "OK", OnClick: func() { clicked = true }},
	)

	if dialog.Title != "Title" {
		t.Errorf("Title = %q, want %q", dialog.Title, "Title")
	}
	if dialog.Body != "Body text" {
		t.Errorf("Body = %q, want %q", dialog.Body, "Body text")
	}
	if !dialog.dismissable {
		t.Error("Dialog should be dismissable by default")
	}
	if len(dialog.Buttons) != 1 {
		t.Errorf("Buttons count = %d, want 1", len(dialog.Buttons))
	}

	// Test button click
	dialog.Buttons[0].OnClick()
	if !clicked {
		t.Error("Button OnClick was not called")
	}
}

func TestDialog_KeyboardShortcuts(t *testing.T) {
	yesClicked := false
	noClicked := false

	dialog := NewDialog("Confirm", "Are you sure?",
		DialogButton{Label: "Yes", Key: 'Y', OnClick: func() { yesClicked = true }},
		DialogButton{Label: "No", Key: 'N', OnClick: func() { noClicked = true }},
	)
	dialog.Focus()

	// Test lowercase 'y' triggers Yes
	dialog.HandleMessage(runtime.KeyMsg{Rune: 'y'})
	if !yesClicked {
		t.Error("Lowercase 'y' should trigger Yes button")
	}

	// Test uppercase 'N' triggers No
	dialog.HandleMessage(runtime.KeyMsg{Rune: 'N'})
	if !noClicked {
		t.Error("Uppercase 'N' should trigger No button")
	}
}

func TestDialog_OnDismiss(t *testing.T) {
	dismissed := false
	dialog := NewDialog("Alert", "Message").
		OnDismiss(func() { dismissed = true })
	dialog.Focus()

	dialog.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEscape})
	if !dismissed {
		t.Error("OnDismiss should be called when Escape is pressed")
	}
}

func TestDialog_WithDismissable(t *testing.T) {
	dismissed := false
	dialog := NewDialog("Alert", "Message").
		WithDismissable(false).
		OnDismiss(func() { dismissed = true })
	dialog.Focus()

	dialog.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEscape})
	if dismissed {
		t.Error("OnDismiss should not be called when dismissable is false")
	}
}

func TestDialog_TimerProgress(t *testing.T) {
	dialog := NewDialog("Alert", "Message").
		WithAutoDismiss(100 * time.Millisecond)

	// Progress should start near 0
	progress := dialog.TimerProgress(dialog.startTime)
	if progress != 0 {
		t.Errorf("Initial progress = %f, want 0", progress)
	}

	// Progress at half duration
	halfTime := dialog.startTime.Add(50 * time.Millisecond)
	progress = dialog.TimerProgress(halfTime)
	if progress < 0.4 || progress > 0.6 {
		t.Errorf("Half progress = %f, want ~0.5", progress)
	}

	// Progress should clamp at 1.0
	endTime := dialog.startTime.Add(200 * time.Millisecond)
	progress = dialog.TimerProgress(endTime)
	if progress != 1.0 {
		t.Errorf("End progress = %f, want 1.0", progress)
	}
}

func TestDialog_ShouldDismiss(t *testing.T) {
	dialog := NewDialog("Alert", "Message").
		WithAutoDismiss(50 * time.Millisecond)

	// Should not dismiss immediately
	if dialog.ShouldDismiss(dialog.startTime) {
		t.Error("Should not dismiss at start")
	}

	// Should dismiss after duration
	endTime := dialog.startTime.Add(100 * time.Millisecond)
	if !dialog.ShouldDismiss(endTime) {
		t.Error("Should dismiss after duration elapsed")
	}

	// Should not dismiss when paused
	dialog.PauseTimer()
	if dialog.ShouldDismiss(endTime) {
		t.Error("Should not dismiss when paused")
	}

	// Should dismiss after resume
	dialog.ResumeTimer()
	if !dialog.ShouldDismiss(endTime) {
		t.Error("Should dismiss after resume")
	}
}

func TestDialog_CenteredBounds(t *testing.T) {
	dialog := NewDialog("Title", "Body")

	parent := runtime.Rect{X: 0, Y: 0, Width: 80, Height: 24}
	centered := dialog.CenteredBounds(parent)

	// Check it's centered horizontally
	expectedX := (parent.Width - centered.Width) / 2
	if centered.X != expectedX {
		t.Errorf("X = %d, want %d", centered.X, expectedX)
	}

	// Check it's centered vertically
	expectedY := (parent.Height - centered.Height) / 2
	if centered.Y != expectedY {
		t.Errorf("Y = %d, want %d", centered.Y, expectedY)
	}
}

func TestDialog_ArrowNavigation(t *testing.T) {
	dialog := NewDialog("Test", "Body",
		DialogButton{Label: "A"},
		DialogButton{Label: "B"},
		DialogButton{Label: "C"},
	)
	dialog.Focus()

	// Initial selection should be 0
	if dialog.selected != 0 {
		t.Errorf("Initial selected = %d, want 0", dialog.selected)
	}

	// Right arrow should move to next button
	dialog.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRight})
	if dialog.selected != 1 {
		t.Errorf("After right, selected = %d, want 1", dialog.selected)
	}

	// Left arrow should move to previous button
	dialog.HandleMessage(runtime.KeyMsg{Key: terminal.KeyLeft})
	if dialog.selected != 0 {
		t.Errorf("After left, selected = %d, want 0", dialog.selected)
	}

	// Should not go below 0
	dialog.HandleMessage(runtime.KeyMsg{Key: terminal.KeyLeft})
	if dialog.selected != 0 {
		t.Errorf("Should clamp at 0, got %d", dialog.selected)
	}

	// Move to end
	dialog.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRight})
	dialog.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRight})
	dialog.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRight})
	if dialog.selected != 2 {
		t.Errorf("Should clamp at 2, got %d", dialog.selected)
	}
}

func TestDialog_EnterActivatesSelected(t *testing.T) {
	clicked := make([]bool, 3)
	dialog := NewDialog("Test", "Body",
		DialogButton{Label: "A", OnClick: func() { clicked[0] = true }},
		DialogButton{Label: "B", OnClick: func() { clicked[1] = true }},
		DialogButton{Label: "C", OnClick: func() { clicked[2] = true }},
	)
	dialog.Focus()

	// Move to B and press Enter
	dialog.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRight})
	dialog.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})

	if clicked[0] || !clicked[1] || clicked[2] {
		t.Errorf("Only B should be clicked, got %v", clicked)
	}
}

func TestDialog_WithContent(t *testing.T) {
	content := NewText("Custom content")
	dialog := NewDialog("Title", "").WithContent(content)

	if dialog.Content != content {
		t.Error("Content should be set")
	}

	children := dialog.ChildWidgets()
	if len(children) != 1 || children[0] != content {
		t.Error("ChildWidgets should return content")
	}
}

func TestDialog_ChildWidgetsNil(t *testing.T) {
	dialog := NewDialog("Title", "Body")

	children := dialog.ChildWidgets()
	if children != nil {
		t.Error("ChildWidgets should return nil when no content")
	}
}

func TestDialog_Measure(t *testing.T) {
	dialog := NewDialog("Short", "This is a longer body text")
	size := dialog.Measure(runtime.Loose(100, 100))

	// Width should accommodate the longer body
	if size.Width < len("This is a longer body text")+4 {
		t.Errorf("Width = %d, should accommodate body text", size.Width)
	}
}

func TestDialog_MeasureWithAutoDismiss(t *testing.T) {
	dialog := NewDialog("Title", "Body")
	sizeWithout := dialog.Measure(runtime.Loose(100, 100))

	dialogWithTimer := NewDialog("Title", "Body").
		WithAutoDismiss(5 * time.Second)
	sizeWith := dialogWithTimer.Measure(runtime.Loose(100, 100))

	// Height should be 1 more with timer bar
	if sizeWith.Height != sizeWithout.Height+1 {
		t.Errorf("Height with timer = %d, want %d", sizeWith.Height, sizeWithout.Height+1)
	}
}
