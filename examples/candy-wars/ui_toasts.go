package main

import (
	"fmt"
	"time"

	"github.com/odvcencio/fluffy-ui/toast"
)

// GameToasts manages toast notifications for game events.
type GameToasts struct {
	manager  *toast.ToastManager
	view     *GameView
	active   []*toast.Toast
	onChange func([]*toast.Toast)
}

func NewGameToasts(view *GameView) *GameToasts {
	t := &GameToasts{
		manager: toast.NewToastManager(),
		view:    view,
	}
	t.manager.SetOnChange(t.handleChange)
	return t
}

func (t *GameToasts) handleChange(items []*toast.Toast) {
	t.active = items
	if t.onChange != nil {
		t.onChange(items)
	}
}

// SetOnChange registers a callback for toast updates.
func (t *GameToasts) SetOnChange(fn func([]*toast.Toast)) {
	t.onChange = fn
	if fn != nil {
		fn(t.active)
	}
}

// Manager returns the underlying toast manager.
func (t *GameToasts) Manager() *toast.ToastManager {
	return t.manager
}

// Active returns currently visible toasts.
func (t *GameToasts) Active() []*toast.Toast {
	return t.active
}

// ShowInfo shows an info toast.
func (t *GameToasts) ShowInfo(title, message string) {
	t.manager.Show(toast.ToastInfo, title, message, 3*time.Second)
}

// ShowSuccess shows a success toast.
func (t *GameToasts) ShowSuccess(title, message string) {
	t.manager.Show(toast.ToastSuccess, title, message, 3*time.Second)
}

// ShowWarning shows a warning toast.
func (t *GameToasts) ShowWarning(title, message string) {
	t.manager.Show(toast.ToastWarning, title, message, 4*time.Second)
}

// ShowError shows an error toast.
func (t *GameToasts) ShowError(title, message string) {
	t.manager.Show(toast.ToastError, title, message, 5*time.Second)
}

// ShowPriceAlert shows a price change notification.
func (t *GameToasts) ShowPriceAlert(candy string, oldPrice, newPrice int, location string) {
	if oldPrice == 0 {
		return
	}
	change := newPrice - oldPrice
	pct := float64(change) / float64(oldPrice) * 100

	if pct >= 20 {
		title := fmt.Sprintf("%s +%.0f%%", candy, pct)
		message := fmt.Sprintf("$%d at %s", newPrice, location)
		t.ShowSuccess(title, message)
	} else if pct <= -20 {
		title := fmt.Sprintf("%s %.0f%%", candy, pct)
		message := fmt.Sprintf("$%d at %s", newPrice, location)
		t.ShowWarning(title, message)
	}
}

// ShowTimeWarning shows a time-related warning.
func (t *GameToasts) ShowTimeWarning(hoursLeft int) {
	if hoursLeft == 2 {
		t.ShowWarning("Time Warning", "2 hours until school ends!")
	} else if hoursLeft == 1 {
		t.ShowWarning("Last Hour!", "Only 1 hour left today!")
	}
}

// ShowHeatWarning shows a heat threshold warning.
func (t *GameToasts) ShowHeatWarning(heat int) {
	if heat >= 75 {
		t.ShowError("DANGER", "Teachers are onto you!")
	} else if heat >= 50 {
		t.ShowWarning("Warning", "Getting suspicious...")
	}
}

// ShowDebtReminder shows a debt reminder.
func (t *GameToasts) ShowDebtReminder(debt, daysLeft int) {
	if daysLeft <= 5 && debt > 0 {
		t.ShowWarning("Debt Due", fmt.Sprintf("$%d debt, %d days left!", debt, daysLeft))
	}
}

// ShowAchievement shows an achievement unlock.
func (t *GameToasts) ShowAchievement(name string) {
	t.manager.Show(toast.ToastSuccess, "Achievement Unlocked!", name, 5*time.Second)
}
