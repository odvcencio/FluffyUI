package main

import (
	"testing"
	"time"
)

func TestModalDialogTimerProgressClamps(t *testing.T) {
	dialog := NewModalDialog("Test", 10, 5)
	dialog.WithAutoDismiss(100 * time.Millisecond)

	// Progress should start near 0
	progress := dialog.TimerProgress(time.Now())
	if progress > 0.1 {
		t.Fatalf("expected initial progress near 0, got %.3f", progress)
	}

	// Progress should clamp to 1 after duration
	time.Sleep(150 * time.Millisecond)
	progress = dialog.TimerProgress(time.Now())
	if progress != 1 {
		t.Fatalf("expected progress to clamp to 1 after duration, got %.3f", progress)
	}
}

func TestModalDialogShouldDismissRespectsPause(t *testing.T) {
	dialog := NewModalDialog("Test", 10, 5)
	dialog.WithAutoDismiss(50 * time.Millisecond)

	// Should not dismiss immediately
	if dialog.ShouldDismiss(time.Now()) {
		t.Fatalf("expected dialog not to dismiss immediately")
	}

	// Wait for duration to elapse
	time.Sleep(100 * time.Millisecond)

	// Should dismiss after duration
	if !dialog.ShouldDismiss(time.Now()) {
		t.Fatalf("expected dialog to dismiss after duration")
	}

	// Test pause behavior with a fresh dialog
	dialog2 := NewModalDialog("Test2", 10, 5)
	dialog2.WithAutoDismiss(50 * time.Millisecond)
	dialog2.PauseTimer()

	time.Sleep(100 * time.Millisecond)

	// Should not dismiss when paused
	if dialog2.ShouldDismiss(time.Now()) {
		t.Fatalf("expected paused dialog not to dismiss")
	}

	// Resume and check
	dialog2.ResumeTimer()
	if !dialog2.ShouldDismiss(time.Now()) {
		t.Fatalf("expected resumed dialog to dismiss")
	}
}

