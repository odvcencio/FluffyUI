package main

import (
	"fmt"

	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/terminal"
	"github.com/odvcencio/fluffyui/widgets"
)

type saveLoadMode int

const (
	saveMode saveLoadMode = iota
	loadMode
)

// SaveLoadDialog renders a save/load slot table.
type SaveLoadDialog struct {
	*ModalDialog
	mode     saveLoadMode
	table    *widgets.Table
	slots    []SaveSlotInfo
	selected int

	onConfirm func(slot int)
	style     backend.Style
}

func NewSaveLoadDialog(mode saveLoadMode, onConfirm func(slot int)) *SaveLoadDialog {
	table := widgets.NewTable(
		widgets.TableColumn{Title: "Slot", Width: 6},
		widgets.TableColumn{Title: "Name", Width: 16},
		widgets.TableColumn{Title: "Day", Width: 6},
		widgets.TableColumn{Title: "Worth", Width: 8},
		widgets.TableColumn{Title: "Saved", Width: 16},
	)
	width := 70
	height := 8 + maxSaveSlots
	if height < 12 {
		height = 12
	}

	title := "Save Game"
	if mode == loadMode {
		title = "Load Game"
	}

	d := &SaveLoadDialog{
		ModalDialog: NewModalDialog(title, width, height),
		mode:        mode,
		table:       table,
		onConfirm:   onConfirm,
		style:       backend.DefaultStyle(),
	}

	actionLabel := "Save"
	actionKey := 'S'
	if mode == loadMode {
		actionLabel = "Load"
		actionKey = 'L'
	}

	d.WithActions(
		DialogAction{Label: actionLabel, Key: actionKey, OnSelect: d.confirmSelection},
		DialogAction{Label: "Cancel", Key: 'C', OnSelect: d.dismiss},
	)

	return d
}

func (d *SaveLoadDialog) UpdateSlots(slots []SaveSlotInfo) {
	d.slots = slots
	rows := make([][]string, len(slots))
	for i, slot := range slots {
		name := slot.Name
		day := "-"
		worth := "-"
		saved := "-"
		if slot.Empty {
			name = "<empty>"
		} else {
			day = fmt.Sprintf("Day %d", slot.Day)
			worth = fmt.Sprintf("$%d", slot.NetWorth)
			saved = slot.SavedAt.Format("01/02 15:04")
		}
		rows[i] = []string{
			fmt.Sprintf("%d", slot.Slot),
			name,
			day,
			worth,
			saved,
		}
	}
	d.table.SetRows(rows)
}

func (d *SaveLoadDialog) confirmSelection() {
	if d.onConfirm == nil {
		return
	}
	if d.selected < 0 || d.selected >= len(d.slots) {
		return
	}
	d.onConfirm(d.slots[d.selected].Slot)
}

func (d *SaveLoadDialog) dismiss() {
	// Trigger dismiss via the underlying dialog
	d.Dialog.Focus()
	d.Dialog.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEscape})
}

func (d *SaveLoadDialog) Layout(bounds runtime.Rect) {
	rect := d.CenteredBounds(bounds)
	d.ModalDialog.Layout(rect)
	content := runtime.Rect{
		X:      rect.X + 2,
		Y:      rect.Y + 2,
		Width:  rect.Width - 4,
		Height: rect.Height - 4,
	}
	d.table.Layout(content)
}

func (d *SaveLoadDialog) Render(ctx runtime.RenderContext) {
	d.ModalDialog.Render(ctx)

	ctx.Buffer.SetString(d.Bounds().X+2, d.Bounds().Y+1, "Select a slot:", d.style)
	d.table.Render(ctx)
}

func (d *SaveLoadDialog) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if key, ok := msg.(runtime.KeyMsg); ok {
		switch key.Key {
		case terminal.KeyEscape:
			d.dismiss()
			return runtime.Handled()
		case terminal.KeyEnter:
			d.confirmSelection()
			return runtime.Handled()
		}
		switch key.Rune {
		case 'c', 'C':
			d.dismiss()
			return runtime.Handled()
		case 's', 'S':
			if d.mode == saveMode {
				d.confirmSelection()
				return runtime.Handled()
			}
		case 'l', 'L':
			if d.mode == loadMode {
				d.confirmSelection()
				return runtime.Handled()
			}
		case '1', '2', '3', '4', '5':
			idx := int(key.Rune - '1')
			if idx >= 0 && idx < len(d.slots) {
				d.selected = idx
				d.table.SetSelected(idx)
				d.table.Focus()
			}
		}
	}

	d.table.Focus()
	result := d.table.HandleMessage(msg)
	d.selected = d.table.SelectedIndex()
	return result
}

func (d *SaveLoadDialog) ChildWidgets() []runtime.Widget {
	return []runtime.Widget{d.table}
}
