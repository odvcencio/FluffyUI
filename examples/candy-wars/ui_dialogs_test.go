package main

import (
	"math"
	"testing"
	"time"
)

func TestModalDialogTimerProgressClamps(t *testing.T) {
	dialog := NewModalDialog("Test", 10, 5)
	dialog.WithAutoDismiss(10 * time.Second)

	base := time.Date(2024, time.January, 1, 12, 0, 0, 0, time.UTC)
	dialog.startTime = base

	progress := dialog.TimerProgress(base.Add(5 * time.Second))
	if math.Abs(progress-0.5) > 0.0001 {
		t.Fatalf("expected progress to be 0.5, got %.3f", progress)
	}

	if dialog.TimerProgress(base.Add(-1*time.Second)) != 0 {
		t.Fatalf("expected progress to clamp to 0 for negative elapsed")
	}

	if dialog.TimerProgress(base.Add(12*time.Second)) != 1 {
		t.Fatalf("expected progress to clamp to 1 for elapsed beyond duration")
	}
}

func TestModalDialogShouldDismissRespectsPause(t *testing.T) {
	dialog := NewModalDialog("Test", 10, 5)
	dialog.WithAutoDismiss(2 * time.Second)

	base := time.Date(2024, time.January, 1, 12, 0, 0, 0, time.UTC)
	dialog.startTime = base

	if !dialog.ShouldDismiss(base.Add(3 * time.Second)) {
		t.Fatalf("expected dialog to dismiss after duration")
	}

	dialog.PauseTimer()
	if dialog.ShouldDismiss(base.Add(3 * time.Second)) {
		t.Fatalf("expected paused dialog not to dismiss")
	}

	dialog.ResumeTimer()
	if !dialog.ShouldDismiss(base.Add(3 * time.Second)) {
		t.Fatalf("expected resumed dialog to dismiss")
	}
}
