package main

import (
	"fmt"

	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/terminal"
)

type blackMarketKind int

const (
	blackMarketWeapon blackMarketKind = iota
	blackMarketArmor
	blackMarketCandy
	blackMarketItem
)

type blackMarketEntry struct {
	Key      rune
	Kind     blackMarketKind
	Name     string
	Price    int
	Owned    bool
	Equipped bool
	Detail   string
}

func (v *GameView) openBlackMarketDialog() {
	if v.game.GameOver.Get() || v.game.InCombat() {
		return
	}
	if !v.game.blackMarketUnlocked() {
		v.game.Message.Set("Black Market is locked. Reach Day 15 or trade $1000.")
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
	v.showStash = false
	v.showStats = false
	v.tradeInput.Blur()
	v.bankInput.Blur()
	v.loanInput.Blur()
	v.craftInput.Blur()
	v.stashInput.Blur()

	v.showBlackMarket = true
	v.blackMarketIndex = 0
	v.blackMarketInput.Clear()
	v.blackMarketInput.Focus()
	v.Invalidate()
}

func (v *GameView) blackMarketDialogRect(bounds runtime.Rect, rows int) runtime.Rect {
	dialogW := 72
	dialogH := rows + 8
	if dialogH < 12 {
		dialogH = 12
	}
	x := bounds.X + (bounds.Width-dialogW)/2
	y := bounds.Y + (bounds.Height-dialogH)/2
	return runtime.Rect{X: x, Y: y, Width: dialogW, Height: dialogH}
}

func (v *GameView) blackMarketInputRect(bounds runtime.Rect) runtime.Rect {
	entries := v.blackMarketEntries()
	rect := v.blackMarketDialogRect(bounds, len(entries))
	label := "Qty: "
	hint := "[A-Z] Select  [Enter] Buy  [Esc] Close"
	inputX := rect.X + 2 + len(label)
	hintX := rect.X + rect.Width - 2 - len(hint)
	inputW := hintX - inputX - 1
	if inputW < 4 {
		inputW = 4
	}
	return runtime.Rect{X: inputX, Y: rect.Y + rect.Height - 3, Width: inputW, Height: 1}
}

func (v *GameView) renderBlackMarketDialog(ctx runtime.RenderContext) {
	entries := v.blackMarketEntries()
	bounds := v.Bounds()
	rect := v.blackMarketDialogRect(bounds, len(entries))
	ctx.Buffer.Fill(rect, ' ', v.style)
	ctx.Buffer.DrawBox(rect, v.accentStyle)

	ctx.Buffer.SetString(rect.X+2, rect.Y, " BLACK MARKET ", v.accentStyle)
	ctx.Buffer.SetString(rect.X+2, rect.Y+2, "Premium gear and contraband candy.", v.dimStyle)

	v.ensureBlackMarketSelection(entries)
	row := rect.Y + 4
	for i, entry := range entries {
		style := v.style
		if i == v.blackMarketIndex {
			style = v.accentStyle
		}
		ctx.Buffer.SetString(rect.X+2, row+i, truncPad(entryLine(entry), rect.Width-4), style)
	}

	label := "Qty: "
	hint := "[A-Z] Select  [Enter] Buy  [Esc] Close"
	ctx.Buffer.SetString(rect.X+2, rect.Y+rect.Height-3, label, v.style)
	v.blackMarketInput.Layout(v.blackMarketInputRect(bounds))
	v.blackMarketInput.Render(ctx)
	hintX := rect.X + rect.Width - 2 - len(hint)
	ctx.Buffer.SetString(hintX, rect.Y+rect.Height-3, hint, v.dimStyle)
}

func (v *GameView) handleBlackMarketInput(msg runtime.Message) runtime.HandleResult {
	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		return v.blackMarketInput.HandleMessage(msg)
	}
	switch key.Key {
	case terminal.KeyEscape:
		v.showBlackMarket = false
		v.blackMarketInput.Blur()
		v.Invalidate()
		return runtime.Handled()
	case terminal.KeyEnter:
		entries := v.blackMarketEntries()
		if len(entries) == 0 {
			return runtime.Handled()
		}
		v.ensureBlackMarketSelection(entries)
		entry := entries[v.blackMarketIndex]
		used := v.buyBlackMarketEntry(entry)
		if used {
			v.blackMarketInput.Clear()
		}
		v.refresh()
		return runtime.Handled()
	}

	if key.Rune != 0 {
		r := key.Rune
		if r >= 'A' && r <= 'Z' {
			r = r - 'A' + 'a'
		}
		if r >= 'a' && r <= 'z' {
			entries := v.blackMarketEntries()
			for i, entry := range entries {
				if entry.Key == r {
					v.blackMarketIndex = i
					v.Invalidate()
					return runtime.Handled()
				}
			}
		}
	}

	return v.blackMarketInput.HandleMessage(msg)
}

func (v *GameView) blackMarketEntries() []blackMarketEntry {
	entries := make([]blackMarketEntry, 0, 8)
	key := 'a'
	for _, weapon := range Weapons {
		if !weapon.RequiresBlackMarket {
			continue
		}
		entries = append(entries, blackMarketEntry{
			Key:      key,
			Kind:     blackMarketWeapon,
			Name:     weapon.Name,
			Price:    blackMarketWeaponPrice(weapon),
			Owned:    v.game.ownedWeapons[weapon.Name],
			Equipped: v.game.equippedWeapon == weapon.Name,
			Detail:   fmt.Sprintf("(%s) +%d ATK", weaponClassLabel(weapon.Class), weapon.AtkBonus),
		})
		key++
	}
	for _, armor := range Armors {
		if !armor.RequiresBlackMarket {
			continue
		}
		entries = append(entries, blackMarketEntry{
			Key:      key,
			Kind:     blackMarketArmor,
			Name:     armor.Name,
			Price:    blackMarketArmorPrice(armor),
			Owned:    v.game.ownedArmors[armor.Name],
			Equipped: v.game.equippedArmor == armor.Name,
			Detail:   fmt.Sprintf("+%d DEF SPD %+d", armor.DefBonus, armor.SPDMod),
		})
		key++
	}
	for _, item := range BlackMarketItems {
		entries = append(entries, blackMarketEntry{
			Key:    key,
			Kind:   blackMarketItem,
			Name:   item.Name,
			Price:  item.Price,
			Detail: item.Description,
		})
		key++
	}
	entries = append(entries, blackMarketEntry{
		Key:    key,
		Kind:   blackMarketCandy,
		Name:   "Rare Import",
		Price:  blackMarketRarePrice,
		Detail: "Contraband candy",
	})
	return entries
}

func (v *GameView) ensureBlackMarketSelection(entries []blackMarketEntry) {
	if len(entries) == 0 {
		v.blackMarketIndex = 0
		return
	}
	if v.blackMarketIndex < 0 || v.blackMarketIndex >= len(entries) {
		v.blackMarketIndex = 0
	}
}

func entryLine(entry blackMarketEntry) string {
	status := ""
	switch entry.Kind {
	case blackMarketCandy:
		status = fmt.Sprintf("$%d each", entry.Price)
	default:
		status = gearStatus(entry.Owned, entry.Equipped, false, entry.Price)
	}
	return fmt.Sprintf("[%c] %-22s %s | %s", entry.Key, entry.Name, entry.Detail, status)
}

func (v *GameView) buyBlackMarketEntry(entry blackMarketEntry) bool {
	switch entry.Kind {
	case blackMarketWeapon:
		return v.game.BuyBlackMarketWeapon(entry.Name)
	case blackMarketArmor:
		return v.game.BuyBlackMarketArmor(entry.Name)
	case blackMarketCandy:
		qty := parseInputAmount(v.blackMarketInput.Text())
		if qty <= 0 {
			qty = 1
		}
		return v.game.BuyBlackMarketRareImport(qty)
	case blackMarketItem:
		qty := parseInputAmount(v.blackMarketInput.Text())
		if qty <= 0 {
			qty = 1
		}
		return v.game.BuyBlackMarketItem(entry.Name, qty)
	default:
		return false
	}
}
