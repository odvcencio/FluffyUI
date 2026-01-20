package main

import (
	"fmt"
	"sort"

	"github.com/odvcencio/fluffy-ui/runtime"
	"github.com/odvcencio/fluffy-ui/terminal"
)

type stashMode int

const (
	stashDeposit stashMode = iota
	stashWithdraw
)

func (v *GameView) openStashDialog() {
	if v.game.GameOver.Get() || v.game.InCombat() {
		return
	}
	if !v.game.hasStash {
		v.game.Message.Set("No stash available yet.")
		v.refresh()
		return
	}
	v.showTrade = false
	v.showBank = false
	v.showLoan = false
	v.showCraft = false
	v.showItems = false
	v.showUpgrades = false
	v.showIntel = false
	v.showGear = false
	v.showBlackMarket = false
	v.showStats = false
	v.showIntel = false
	v.tradeInput.Blur()
	v.bankInput.Blur()
	v.loanInput.Blur()
	v.craftInput.Blur()
	v.blackMarketInput.Blur()
	v.stashMode = stashDeposit
	v.stashInput.Clear()
	v.stashInput.Focus()
	v.showStash = true
	v.Invalidate()
}

func (v *GameView) stashDialogRect(bounds runtime.Rect) runtime.Rect {
	dialogW := 60
	dialogH := 12
	x := bounds.X + (bounds.Width-dialogW)/2
	y := bounds.Y + (bounds.Height-dialogH)/2
	return runtime.Rect{X: x, Y: y, Width: dialogW, Height: dialogH}
}

func (v *GameView) stashInputRect(bounds runtime.Rect) runtime.Rect {
	rect := v.stashDialogRect(bounds)
	label := "Qty: "
	hint := "[D]eposit [W]ithdraw [Enter] Move [Esc] Close"
	inputX := rect.X + 2 + len(label)
	hintX := rect.X + rect.Width - 2 - len(hint)
	inputW := hintX - inputX - 1
	if inputW < 4 {
		inputW = 4
	}
	return runtime.Rect{X: inputX, Y: rect.Y + rect.Height - 3, Width: inputW, Height: 1}
}

func (v *GameView) renderStashDialog(ctx runtime.RenderContext) {
	bounds := v.Bounds()
	rect := v.stashDialogRect(bounds)
	ctx.Buffer.Fill(rect, ' ', v.style)
	ctx.Buffer.DrawBox(rect, v.accentStyle)

	ctx.Buffer.SetString(rect.X+2, rect.Y, " SECRET STASH ", v.accentStyle)

	modeLabel := "Deposit"
	if v.stashMode == stashWithdraw {
		modeLabel = "Withdraw"
	}
	countLine := fmt.Sprintf("Backpack: %d/%d  Stash: %d/%d  Mode: %s",
		v.game.InventoryCount(), v.game.Capacity,
		v.game.StashCount(), v.game.stashCapacity,
		modeLabel,
	)
	ctx.Buffer.SetString(rect.X+2, rect.Y+2, truncPad(countLine, rect.Width-4), v.dimStyle)

	items := v.stashItemNames()
	v.ensureStashSelection(items)
	if len(items) == 0 {
		ctx.Buffer.SetString(rect.X+2, rect.Y+4, "No items available.", v.style)
	} else {
		for i, name := range items {
			count := v.stashItemCount(name)
			line := fmt.Sprintf("[%d] %s x%d", i+1, name, count)
			line = truncPad(line, rect.Width-4)
			style := v.style
			if name == v.stashItem {
				style = v.accentStyle
			}
			ctx.Buffer.SetString(rect.X+2, rect.Y+4+i, line, style)
		}
	}

	label := "Qty: "
	ctx.Buffer.SetString(rect.X+2, rect.Y+rect.Height-3, label, v.style)
	v.stashInput.Layout(v.stashInputRect(bounds))
	v.stashInput.Render(ctx)
	ctx.Buffer.SetString(rect.X+2, rect.Y+rect.Height-2, "[D]eposit [W]ithdraw [Enter] Move [Esc] Close", v.dimStyle)
}

func (v *GameView) handleStashInput(msg runtime.Message) runtime.HandleResult {
	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		return v.stashInput.HandleMessage(msg)
	}
	switch key.Key {
	case terminal.KeyEscape:
		v.showStash = false
		v.stashInput.Blur()
		v.Invalidate()
		return runtime.Handled()
	case terminal.KeyEnter:
		qty := parseInputAmount(v.stashInput.Text())
		if qty <= 0 {
			qty = 1
		}
		if v.stashItem == "" {
			return runtime.Handled()
		}
		if v.stashMode == stashDeposit {
			v.game.StashDeposit(v.stashItem, qty)
		} else {
			v.game.StashWithdraw(v.stashItem, qty)
		}
		v.stashInput.Clear()
		v.refresh()
		return runtime.Handled()
	}

	switch key.Rune {
	case 'd', 'D':
		v.stashMode = stashDeposit
		v.stashInput.Clear()
		v.Invalidate()
		return runtime.Handled()
	case 'w', 'W':
		v.stashMode = stashWithdraw
		v.stashInput.Clear()
		v.Invalidate()
		return runtime.Handled()
	}

	if key.Rune >= '1' && key.Rune <= '9' {
		items := v.stashItemNames()
		idx := int(key.Rune - '1')
		if idx >= 0 && idx < len(items) {
			v.stashItem = items[idx]
			v.Invalidate()
			return runtime.Handled()
		}
	}

	return v.stashInput.HandleMessage(msg)
}

func (v *GameView) stashItemNames() []string {
	var source Inventory
	if v.stashMode == stashWithdraw {
		source = v.game.stash
	} else {
		source = v.game.Inventory.Get()
	}
	names := make([]string, 0, len(source))
	for name, qty := range source {
		if qty > 0 {
			names = append(names, name)
		}
	}
	return orderInventoryNames(names)
}

func (v *GameView) stashItemCount(name string) int {
	if v.stashMode == stashWithdraw {
		return v.game.stash[name]
	}
	return v.game.Inventory.Get()[name]
}

func (v *GameView) ensureStashSelection(items []string) {
	if len(items) == 0 {
		v.stashItem = ""
		return
	}
	for _, name := range items {
		if name == v.stashItem {
			return
		}
	}
	v.stashItem = items[0]
}

func orderInventoryNames(names []string) []string {
	if len(names) == 0 {
		return names
	}
	priority := make(map[string]int, len(CandyTypes))
	for i, candy := range CandyTypes {
		priority[candy.Name] = i
	}
	sort.Slice(names, func(i, j int) bool {
		pi, okI := priority[names[i]]
		pj, okJ := priority[names[j]]
		if okI && okJ {
			return pi < pj
		}
		if okI {
			return true
		}
		if okJ {
			return false
		}
		return names[i] < names[j]
	})
	return names
}
