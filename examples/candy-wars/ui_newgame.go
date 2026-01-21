package main

import (
	"fmt"

	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/runtime"
	"github.com/odvcencio/fluffy-ui/terminal"
	"github.com/odvcencio/fluffy-ui/widgets"
)

// AppView switches between the new game screen and the main game view.
type AppView struct {
	widgets.Component
	game        *Game
	gameView    *GameView
	newGameView *NewGameDialog
	showNewGame bool
	gameStarted bool

	style backend.Style
}

func NewAppView() *AppView {
	game := NewGame()
	view := NewGameView(game)

	app := &AppView{
		game:        game,
		gameView:    view,
		showNewGame: true,
		style:       backend.DefaultStyle(),
	}

	app.newGameView = NewNewGameDialog(func(diff Difficulty) {
		app.showNewGame = false
		app.gameStarted = true
		app.game.StartNewRunWithDifficulty(diff)
		app.gameView.lastPrices = nil
		app.gameView.lastHour = 0
		app.gameView.lastHeat = 0
		app.gameView.lastDay = 0
		app.gameView.lastLoc = -1
		app.gameView.refresh()
	}, func() {
		if app.gameStarted {
			app.showNewGame = false
		}
	})

	app.gameView.onRequestNewGame = func() {
		app.showNewGame = true
		app.Invalidate()
	}

	return app
}

func (a *AppView) Measure(c runtime.Constraints) runtime.Size {
	return c.MaxSize()
}

func (a *AppView) Layout(bounds runtime.Rect) {
	a.Component.Layout(bounds)
	if a.gameView != nil {
		a.gameView.Layout(bounds)
	}
	if a.newGameView != nil {
		a.newGameView.Layout(bounds)
	}
}

func (a *AppView) Render(ctx runtime.RenderContext) {
	if a.showNewGame {
		ctx.Clear(a.style)
		a.newGameView.Render(ctx)
		return
	}
	if a.gameView != nil {
		a.gameView.Render(ctx)
	}
}

func (a *AppView) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if a.showNewGame {
		return a.newGameView.HandleMessage(msg)
	}
	if a.gameView != nil {
		return a.gameView.HandleMessage(msg)
	}
	return runtime.Unhandled()
}

func (a *AppView) ChildWidgets() []runtime.Widget {
	widgets := make([]runtime.Widget, 0, 2)
	if a.gameView != nil {
		widgets = append(widgets, a.gameView)
	}
	if a.newGameView != nil {
		widgets = append(widgets, a.newGameView)
	}
	return widgets
}

// NewGameDialog renders difficulty selection.
type NewGameDialog struct {
	*ModalDialog
	options    []Difficulty
	radios     []*widgets.Radio
	group      *widgets.RadioGroup
	onStart    func(diff Difficulty)
	onCancel   func()
	style      backend.Style
	dimStyle   backend.Style
	labelStyle backend.Style
}

func NewNewGameDialog(onStart func(diff Difficulty), onCancel func()) *NewGameDialog {
	options := []Difficulty{DifficultyNormal, DifficultyNightmare, DifficultyHell}
	group := widgets.NewRadioGroup()
	radios := make([]*widgets.Radio, 0, len(options))

	for _, diff := range options {
		mods := DifficultySettings[diff]
		label := fmt.Sprintf("%s - %s", mods.Name, mods.Tagline)
		radios = append(radios, widgets.NewRadio(label, group))
	}
	group.SetSelected(0)

	height := 12 + len(options)
	if height < 14 {
		height = 14
	}

	d := &NewGameDialog{
		ModalDialog: NewModalDialog("New Game", 64, height),
		options:     options,
		radios:      radios,
		group:       group,
		onStart:     onStart,
		onCancel:    onCancel,
		style:       backend.DefaultStyle(),
		dimStyle:    backend.DefaultStyle().Dim(true),
		labelStyle:  backend.DefaultStyle().Bold(true),
	}

	d.WithActions(
		DialogAction{Label: "Start", Key: 'S', OnSelect: d.startSelected},
		DialogAction{Label: "Cancel", Key: 'C', OnSelect: d.cancel},
	)

	return d
}

func (d *NewGameDialog) startSelected() {
	if d.onStart == nil {
		return
	}
	idx := d.group.Selected()
	if idx < 0 || idx >= len(d.options) {
		idx = 0
	}
	d.onStart(d.options[idx])
}

func (d *NewGameDialog) cancel() {
	if d.onCancel != nil {
		d.onCancel()
	}
}

func (d *NewGameDialog) Layout(bounds runtime.Rect) {
	rect := d.CenteredBounds(bounds)
	d.ModalDialog.Layout(rect)

	content := runtime.Rect{
		X:      rect.X + 2,
		Y:      rect.Y + 2,
		Width:  rect.Width - 4,
		Height: rect.Height - 4,
	}

	y := content.Y + 1
	for _, radio := range d.radios {
		radio.Layout(runtime.Rect{X: content.X, Y: y, Width: content.Width, Height: 1})
		y += 2
	}
}

func (d *NewGameDialog) Render(ctx runtime.RenderContext) {
	d.ModalDialog.Render(ctx)

	bounds := d.Bounds()
	contentX := bounds.X + 2
	contentY := bounds.Y + 2

	ctx.Buffer.SetString(contentX, contentY, "Choose your difficulty:", d.labelStyle)

	for i, radio := range d.radios {
		radio.Render(ctx)
		mods := DifficultySettings[d.options[i]]
		lineY := radio.Bounds().Y + 1
		ctx.Buffer.SetString(contentX+4, lineY, truncPad(mods.Description, bounds.Width-8), d.dimStyle)
	}
}

func (d *NewGameDialog) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if key, ok := msg.(runtime.KeyMsg); ok {
		switch key.Key {
		case terminal.KeyEscape:
			d.cancel()
			return runtime.Handled()
		case terminal.KeyUp:
			idx := d.group.Selected() - 1
			if idx < 0 {
				idx = 0
			}
			d.group.SetSelected(idx)
			return runtime.Handled()
		case terminal.KeyDown:
			idx := d.group.Selected() + 1
			if idx >= len(d.options) {
				idx = len(d.options) - 1
			}
			d.group.SetSelected(idx)
			return runtime.Handled()
		case terminal.KeyEnter:
			d.startSelected()
			return runtime.Handled()
		}
		switch key.Rune {
		case '1', '2', '3':
			idx := int(key.Rune - '1')
			if idx >= 0 && idx < len(d.options) {
				d.group.SetSelected(idx)
				return runtime.Handled()
			}
		case 's', 'S':
			d.startSelected()
			return runtime.Handled()
		case 'c', 'C':
			d.cancel()
			return runtime.Handled()
		}
	}

	for _, radio := range d.radios {
		if result := radio.HandleMessage(msg); result.Handled {
			return result
		}
	}

	return d.ModalDialog.HandleMessage(msg)
}

func (d *NewGameDialog) ChildWidgets() []runtime.Widget {
	widgets := make([]runtime.Widget, 0, len(d.radios))
	for _, radio := range d.radios {
		widgets = append(widgets, radio)
	}
	return widgets
}
