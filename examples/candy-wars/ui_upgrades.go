package main

import (
	"fmt"

	"github.com/odvcencio/fluffy-ui/runtime"
	"github.com/odvcencio/fluffy-ui/terminal"
)

func (v *GameView) openUpgradesDialog() {
	if v.game.GameOver.Get() || v.game.InCombat() {
		return
	}
	v.showTrade = false
	v.showBank = false
	v.showLoan = false
	v.showCraft = false
	v.showItems = false
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
	v.showUpgrades = true
	v.Invalidate()
}

func (v *GameView) upgradesDialogRect(bounds runtime.Rect) runtime.Rect {
	dialogW := 70
	dialogH := 18
	x := bounds.X + (bounds.Width-dialogW)/2
	y := bounds.Y + (bounds.Height-dialogH)/2
	return runtime.Rect{X: x, Y: y, Width: dialogW, Height: dialogH}
}

func (v *GameView) renderUpgradesDialog(ctx runtime.RenderContext) {
	bounds := v.Bounds()
	rect := v.upgradesDialogRect(bounds)
	ctx.Buffer.Fill(rect, ' ', v.style)
	ctx.Buffer.DrawBox(rect, v.accentStyle)

	ctx.Buffer.SetString(rect.X+2, rect.Y, " UPGRADES ", v.accentStyle)

	statsLine := fmt.Sprintf("Stats: ATK %d  DEF %d  SPD %d", v.game.playerATK(), v.game.playerDEF(), v.game.playerSPD())
	ctx.Buffer.SetString(rect.X+2, rect.Y+2, truncPad(statsLine, rect.Width-4), v.dimStyle)

	combatAvailable := v.game.canAccessCombatUpgrades()
	tradeAvailable := v.game.canAccessTradeUpgrades()

	combatHeader := "Combat (Gymnasium)"
	if !combatAvailable {
		combatHeader += " - Visit Gym to buy"
	}
	tradeHeader := "Trader (Cafeteria)"
	if !tradeAvailable {
		tradeHeader += " - Visit Cafeteria to buy"
	}

	headerStyle := v.accentStyle
	if !combatAvailable {
		headerStyle = v.dimStyle
	}
	ctx.Buffer.SetString(rect.X+2, rect.Y+4, truncPad(combatHeader, rect.Width-4), headerStyle)

	lines := v.upgradeCombatLines()
	for i, line := range lines {
		style := v.style
		if !combatAvailable {
			style = v.dimStyle
		}
		ctx.Buffer.SetString(rect.X+2, rect.Y+5+i, truncPad(line, rect.Width-4), style)
	}

	headerStyle = v.accentStyle
	if !tradeAvailable {
		headerStyle = v.dimStyle
	}
	ctx.Buffer.SetString(rect.X+2, rect.Y+10, truncPad(tradeHeader, rect.Width-4), headerStyle)

	lines = v.upgradeTradeLines()
	for i, line := range lines {
		style := v.style
		if !tradeAvailable {
			style = v.dimStyle
		}
		ctx.Buffer.SetString(rect.X+2, rect.Y+11+i, truncPad(line, rect.Width-4), style)
	}

	ctx.Buffer.SetString(rect.X+2, rect.Y+rect.Height-2, "[1-9,0] Buy  [Esc] Close", v.dimStyle)
}

func (v *GameView) upgradeCombatLines() []string {
	lines := make([]string, 0, 5)
	lines = append(lines, fmt.Sprintf("[1] Workout Sessions (+2 ATK) $%d (%d/%d)", statUpgradeCost, v.game.workoutCount, statUpgradeMax))
	lines = append(lines, fmt.Sprintf("[2] Thick Skin (+2 DEF) $%d (%d/%d)", statUpgradeCost, v.game.thickSkinCount, statUpgradeMax))
	lines = append(lines, fmt.Sprintf("[3] Track Practice (+2 SPD) $%d (%d/%d)", statUpgradeCost, v.game.trackPracticeCount, statUpgradeMax))

	muscleStatus := fmt.Sprintf("$%d", muscleCost)
	if v.game.hasMuscle {
		muscleStatus = "Owned"
	}
	lines = append(lines, fmt.Sprintf("[4] Hire Muscle (+10 ATK) %s", muscleStatus))

	intimidationStatus := fmt.Sprintf("$%d", intimidationCost)
	if v.game.hasIntimidation {
		intimidationStatus = "Owned"
	}
	lines = append(lines, fmt.Sprintf("[5] Intimidation (20%% flee) %s", intimidationStatus))
	return lines
}

func (v *GameView) upgradeTradeLines() []string {
	lines := make([]string, 0, 5)
	backpackStatus := "MAX"
	if v.game.backpackTier < len(backpackTierCosts) {
		cost := backpackTierCosts[v.game.backpackTier]
		backpackStatus = fmt.Sprintf("$%d (Tier %d/%d)", cost, v.game.backpackTier+1, len(backpackTierCosts))
	}
	lines = append(lines, fmt.Sprintf("[6] Bigger Backpack (+slots) %s", backpackStatus))

	stashStatus := fmt.Sprintf("$%d", stashCost)
	if v.game.hasStash {
		stashStatus = fmt.Sprintf("Owned (%d slots)", v.game.stashCapacity)
	}
	lines = append(lines, fmt.Sprintf("[7] Secret Stash %s", stashStatus))

	bikeStatus := fmt.Sprintf("$%d", bikeCost)
	if v.game.hasBike {
		bikeStatus = "Owned"
	}
	lines = append(lines, fmt.Sprintf("[8] Bike (Travel 1h) %s", bikeStatus))

	informantStatus := fmt.Sprintf("$%d", informantCost)
	if v.game.hasInformant {
		informantStatus = "Owned"
	}
	lines = append(lines, fmt.Sprintf("[9] Informant Network %s", informantStatus))

	bankStatus := fmt.Sprintf("$%d", bankExpansionCost)
	if v.game.bankExpanded {
		bankStatus = fmt.Sprintf("Owned (Limit $%d)", bankLimitExpanded)
	}
	lines = append(lines, fmt.Sprintf("[0] Bank Expansion %s", bankStatus))
	return lines
}

func (v *GameView) handleUpgradesInput(msg runtime.Message) runtime.HandleResult {
	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		return runtime.Handled()
	}
	if key.Key == terminal.KeyEscape {
		v.showUpgrades = false
		v.Invalidate()
		return runtime.Handled()
	}
	switch key.Rune {
	case '1':
		v.game.BuyWorkout()
	case '2':
		v.game.BuyThickSkin()
	case '3':
		v.game.BuyTrackPractice()
	case '4':
		v.game.BuyHireMuscle()
	case '5':
		v.game.BuyIntimidation()
	case '6':
		v.game.BuyBackpackUpgrade()
	case '7':
		v.game.BuySecretStash()
	case '8':
		v.game.BuyBike()
	case '9':
		v.game.BuyInformant()
	case '0':
		v.game.BuyBankExpansion()
	default:
		return runtime.Handled()
	}
	v.refresh()
	return runtime.Handled()
}
