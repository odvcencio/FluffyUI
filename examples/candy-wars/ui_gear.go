package main

import (
	"fmt"

	"github.com/odvcencio/fluffy-ui/runtime"
	"github.com/odvcencio/fluffy-ui/terminal"
)

type gearEntry struct {
	Key      rune
	Label    string
	Weapon   bool
	Name     string
	Owned    bool
	Equipped bool
	Locked   bool
	Price    int
}

func (v *GameView) openGearDialog() {
	if v.game.GameOver.Get() || v.game.InCombat() {
		return
	}
	if !v.game.gearShopAvailable() {
		v.game.Message.Set("Gear shop is only at the Gymnasium.")
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
	v.showStash = false
	v.showBlackMarket = false
	v.showStats = false
	v.tradeInput.Blur()
	v.bankInput.Blur()
	v.loanInput.Blur()
	v.craftInput.Blur()
	v.stashInput.Blur()
	v.blackMarketInput.Blur()
	v.showGear = true
	v.Invalidate()
}

func (v *GameView) gearDialogRect(bounds runtime.Rect, rows int) runtime.Rect {
	dialogW := 72
	dialogH := rows + 8
	if dialogH < 14 {
		dialogH = 14
	}
	x := bounds.X + (bounds.Width-dialogW)/2
	y := bounds.Y + (bounds.Height-dialogH)/2
	return runtime.Rect{X: x, Y: y, Width: dialogW, Height: dialogH}
}

func (v *GameView) renderGearDialog(ctx runtime.RenderContext) {
	entries := v.gearEntries()
	bounds := v.Bounds()
	rect := v.gearDialogRect(bounds, len(entries))
	ctx.Buffer.Fill(rect, ' ', v.style)
	ctx.Buffer.DrawBox(rect, v.accentStyle)

	ctx.Buffer.SetString(rect.X+2, rect.Y, " GEAR SHOP ", v.accentStyle)
	ctx.Buffer.SetString(rect.X+2, rect.Y+2, "Weapon Triangle: Blunt > Quick > Sharp > Blunt", v.dimStyle)

	row := rect.Y + 4
	for i, entry := range entries {
		style := v.style
		if entry.Locked {
			style = v.dimStyle
		}
		ctx.Buffer.SetString(rect.X+2, row+i, truncPad(entry.Label, rect.Width-4), style)
	}

	ctx.Buffer.SetString(rect.X+2, rect.Y+rect.Height-2, "[A-Z] Buy/Equip  [Esc] Close", v.dimStyle)
}

func (v *GameView) handleGearInput(msg runtime.Message) runtime.HandleResult {
	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		return runtime.Handled()
	}
	if key.Key == terminal.KeyEscape {
		v.showGear = false
		v.Invalidate()
		return runtime.Handled()
	}
	entries := v.gearEntries()
	for _, entry := range entries {
		if key.Rune == entry.Key || key.Rune == rune(entry.Key-32) {
			if entry.Locked {
				v.game.Message.Set("That gear is still locked.")
				v.refresh()
				return runtime.Handled()
			}
			if entry.Weapon {
				v.game.BuyOrEquipWeapon(entry.Name)
			} else {
				v.game.BuyOrEquipArmor(entry.Name)
			}
			v.refresh()
			return runtime.Handled()
		}
	}
	return runtime.Handled()
}

func (v *GameView) gearEntries() []gearEntry {
	entries := make([]gearEntry, 0, len(Weapons)+len(Armors))
	key := 'a'

	for _, weapon := range Weapons {
		owned := v.game.ownedWeapons[weapon.Name]
		equipped := v.game.equippedWeapon == weapon.Name
		locked := weapon.RequiresBlackMarket && !v.game.blackMarketUnlocked()
		price := v.game.weaponPrice(weapon)
		status := gearStatus(owned, equipped, locked, price)
		extra := fmt.Sprintf("(%s) +%d ATK", weaponClassLabel(weapon.Class), weapon.AtkBonus)
		if weapon.StunChance > 0 {
			extra += fmt.Sprintf(" Stun%d%%", weapon.StunChance)
		}
		label := fmt.Sprintf("[%c] %-22s %s | %s", key, weapon.Name, extra, status)
		entries = append(entries, gearEntry{
			Key:      key,
			Label:    label,
			Weapon:   true,
			Name:     weapon.Name,
			Owned:    owned,
			Equipped: equipped,
			Locked:   locked,
			Price:    price,
		})
		key++
	}

	for _, armor := range Armors {
		owned := v.game.ownedArmors[armor.Name]
		equipped := v.game.equippedArmor == armor.Name
		locked := armor.RequiresBlackMarket && !v.game.blackMarketUnlocked()
		price := v.game.armorPrice(armor)
		status := gearStatus(owned, equipped, locked, price)
		extra := fmt.Sprintf("+%d DEF SPD %+d", armor.DefBonus, armor.SPDMod)
		label := fmt.Sprintf("[%c] %-22s %s | %s", key, armor.Name, extra, status)
		entries = append(entries, gearEntry{
			Key:      key,
			Label:    label,
			Weapon:   false,
			Name:     armor.Name,
			Owned:    owned,
			Equipped: equipped,
			Locked:   locked,
			Price:    price,
		})
		key++
	}
	return entries
}

func gearStatus(owned bool, equipped bool, locked bool, price int) string {
	if locked {
		return "LOCKED"
	}
	if equipped {
		return "Equipped"
	}
	if owned {
		return "Owned"
	}
	if price <= 0 {
		return "Free"
	}
	return fmt.Sprintf("$%d", price)
}
