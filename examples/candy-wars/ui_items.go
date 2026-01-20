package main

import (
	"fmt"
	"strings"

	"github.com/odvcencio/fluffy-ui/runtime"
	"github.com/odvcencio/fluffy-ui/terminal"
)

func (v *GameView) openCraftDialog() {
	if v.game.GameOver.Get() || v.game.InCombat() {
		return
	}
	if !v.game.craftingUnlocked() {
		v.game.Message.Set("Crafting is locked. Reach Day 7 or earn $500 profit.")
		v.refresh()
		return
	}
	v.showTrade = false
	v.showBank = false
	v.showLoan = false
	v.showItems = false
	v.showUpgrades = false
	v.showIntel = false
	v.showGear = false
	v.showBlackMarket = false
	v.showStats = false
	v.showStash = false
	v.tradeInput.Blur()
	v.bankInput.Blur()
	v.loanInput.Blur()
	v.stashInput.Blur()
	v.blackMarketInput.Blur()

	v.showCraft = true
	v.craftInput.Clear()
	v.craftInput.Focus()
	v.Invalidate()
}

func (v *GameView) craftDialogRect(bounds runtime.Rect) runtime.Rect {
	dialogW := 66
	dialogH := len(CraftedItems) + 7
	if dialogH < 10 {
		dialogH = 10
	}
	x := bounds.X + (bounds.Width-dialogW)/2
	y := bounds.Y + (bounds.Height-dialogH)/2
	return runtime.Rect{X: x, Y: y, Width: dialogW, Height: dialogH}
}

func (v *GameView) craftInputRect(bounds runtime.Rect) runtime.Rect {
	rect := v.craftDialogRect(bounds)
	label := "Qty: "
	hint := "[1-7] Select  [Enter] Craft  [Esc] Close"
	inputX := rect.X + 2 + len(label)
	hintX := rect.X + rect.Width - 2 - len(hint)
	inputW := hintX - inputX - 1
	if inputW < 4 {
		inputW = 4
	}
	return runtime.Rect{X: inputX, Y: rect.Y + rect.Height - 3, Width: inputW, Height: 1}
}

func (v *GameView) renderCraftDialog(ctx runtime.RenderContext) {
	bounds := v.Bounds()
	rect := v.craftDialogRect(bounds)
	ctx.Buffer.Fill(rect, ' ', v.style)
	ctx.Buffer.DrawBox(rect, v.accentStyle)

	ctx.Buffer.SetString(rect.X+2, rect.Y, " CRAFTING ", v.accentStyle)

	invLine := v.craftInventoryLine(rect.Width - 4)
	ctx.Buffer.SetString(rect.X+2, rect.Y+2, invLine, v.dimStyle)

	listStart := rect.Y + 3
	for i, item := range CraftedItems {
		line := fmt.Sprintf("[%d] %s: %s | %s", i+1, item.Name, formatIngredients(item.Ingredients), item.Description)
		line = truncPad(line, rect.Width-4)
		style := v.style
		if i == v.craftIndex {
			style = v.accentStyle
		}
		ctx.Buffer.SetString(rect.X+2, listStart+i, line, style)
	}

	label := "Qty: "
	hint := "[1-7] Select  [Enter] Craft  [Esc] Close"
	ctx.Buffer.SetString(rect.X+2, rect.Y+rect.Height-3, label, v.style)
	v.craftInput.Layout(v.craftInputRect(bounds))
	v.craftInput.Render(ctx)
	hintX := rect.X + rect.Width - 2 - len(hint)
	ctx.Buffer.SetString(hintX, rect.Y+rect.Height-3, hint, v.dimStyle)
}

func (v *GameView) craftInventoryLine(width int) string {
	inv := v.game.Inventory.Get()
	parts := make([]string, 0, len(CandyTypes))
	for _, candy := range CandyTypes {
		if qty := inv[candy.Name]; qty > 0 {
			parts = append(parts, fmt.Sprintf("%s:%d", shortName(candy.Name), qty))
		}
	}
	line := "Ingredients: none"
	if len(parts) > 0 {
		line = "Ingredients: " + strings.Join(parts, " ")
	}
	return truncPad(line, width)
}

func (v *GameView) handleCraftInput(msg runtime.Message) runtime.HandleResult {
	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		return v.craftInput.HandleMessage(msg)
	}

	switch key.Key {
	case terminal.KeyEscape:
		v.showCraft = false
		v.craftInput.Blur()
		v.Invalidate()
		return runtime.Handled()
	case terminal.KeyEnter:
		qty := parseInputAmount(v.craftInput.Text())
		if qty <= 0 {
			qty = 1
		}
		v.game.CraftItem(v.craftIndex, qty)
		v.craftInput.Clear()
		v.refresh()
		return runtime.Handled()
	}

	if key.Rune >= '1' && key.Rune <= '9' {
		idx := int(key.Rune - '1')
		if idx >= 0 && idx < len(CraftedItems) {
			v.craftIndex = idx
			v.Invalidate()
			return runtime.Handled()
		}
	}

	return v.craftInput.HandleMessage(msg)
}

func (v *GameView) openItemsDialog(inCombat bool) {
	if v.game.GameOver.Get() {
		return
	}
	v.showTrade = false
	v.showBank = false
	v.showLoan = false
	v.showCraft = false
	v.showUpgrades = false
	v.showIntel = false
	v.showGear = false
	v.showBlackMarket = false
	v.showStats = false
	v.showStash = false
	v.tradeInput.Blur()
	v.bankInput.Blur()
	v.loanInput.Blur()
	v.craftInput.Blur()
	v.stashInput.Blur()
	v.blackMarketInput.Blur()

	v.itemsInCombat = inCombat
	v.showItems = true
	v.Invalidate()
}

func (v *GameView) itemsDialogRect(bounds runtime.Rect, items int) runtime.Rect {
	dialogW := 58
	dialogH := items + 6
	if dialogH < 8 {
		dialogH = 8
	}
	x := bounds.X + (bounds.Width-dialogW)/2
	y := bounds.Y + (bounds.Height-dialogH)/2
	return runtime.Rect{X: x, Y: y, Width: dialogW, Height: dialogH}
}

func (v *GameView) renderItemsDialog(ctx runtime.RenderContext) {
	items := v.usableItemNames(v.itemsInCombat)
	bounds := v.Bounds()
	rect := v.itemsDialogRect(bounds, len(items))
	ctx.Buffer.Fill(rect, ' ', v.style)
	ctx.Buffer.DrawBox(rect, v.accentStyle)

	title := " ITEMS "
	if v.itemsInCombat {
		title = " ITEMS (COMBAT) "
	}
	ctx.Buffer.SetString(rect.X+2, rect.Y, title, v.accentStyle)

	if len(items) == 0 {
		ctx.Buffer.SetString(rect.X+2, rect.Y+2, "No usable items.", v.style)
		ctx.Buffer.SetString(rect.X+2, rect.Y+rect.Height-2, "[Press any key]", v.dimStyle)
		return
	}

	inv := v.game.Inventory.Get()
	for i, name := range items {
		item, _ := itemDefByName(name)
		count := inv[name]
		line := fmt.Sprintf("[%d] %s x%d - %s", i+1, item.Name, count, item.Description)
		line = truncPad(line, rect.Width-4)
		ctx.Buffer.SetString(rect.X+2, rect.Y+2+i, line, v.style)
	}
	ctx.Buffer.SetString(rect.X+2, rect.Y+rect.Height-2, "[1-9] Use  [Esc] Close", v.dimStyle)
}

func (v *GameView) handleItemsInput(msg runtime.Message) runtime.HandleResult {
	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		return runtime.Handled()
	}

	if key.Key == terminal.KeyEscape {
		v.showItems = false
		v.Invalidate()
		return runtime.Handled()
	}

	items := v.usableItemNames(v.itemsInCombat)
	if len(items) == 0 {
		v.showItems = false
		v.Invalidate()
		return runtime.Handled()
	}

	if key.Rune >= '1' && key.Rune <= '9' {
		idx := int(key.Rune - '1')
		if idx >= 0 && idx < len(items) {
			name := items[idx]
			used := false
			if v.itemsInCombat {
				used = v.game.CombatUseItem(name)
			} else {
				used = v.game.UseItem(name, false)
			}
			if used {
				v.showItems = false
			}
			v.refresh()
			return runtime.Handled()
		}
	}

	return runtime.Handled()
}

func (v *GameView) usableItemNames(inCombat bool) []string {
	inv := v.game.Inventory.Get()
	names := make([]string, 0, len(CraftedItems)+len(BlackMarketItems))
	for _, item := range allItemDefs() {
		if inv[item.Name] <= 0 {
			continue
		}
		if inCombat && !item.UseInCombat {
			continue
		}
		if !inCombat && !item.UseOutOfCombat {
			continue
		}
		names = append(names, item.Name)
	}
	return names
}

func formatIngredients(ingredients map[string]int) string {
	parts := make([]string, 0, len(ingredients))
	for _, candy := range CandyTypes {
		if qty, ok := ingredients[candy.Name]; ok && qty > 0 {
			parts = append(parts, fmt.Sprintf("%dx %s", qty, candy.Name))
		}
	}
	if len(parts) == 0 {
		return "None"
	}
	return strings.Join(parts, ", ")
}
