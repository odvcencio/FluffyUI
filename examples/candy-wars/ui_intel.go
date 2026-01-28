package main

import (
	"fmt"
	"strings"

	"github.com/odvcencio/fluffyui/runtime"
)

func (v *GameView) openIntelDialog() {
	if v.game.GameOver.Get() || v.game.InCombat() {
		return
	}
	if !v.game.hasInformant {
		v.game.Message.Set("No informant network yet.")
		v.refresh()
		return
	}
	v.showTrade = false
	v.showBank = false
	v.showLoan = false
	v.showCraft = false
	v.showItems = false
	v.showUpgrades = false
	v.showStash = false
	v.showGear = false
	v.showBlackMarket = false
	v.showStats = false
	v.tradeInput.Blur()
	v.bankInput.Blur()
	v.loanInput.Blur()
	v.craftInput.Blur()
	v.stashInput.Blur()
	v.blackMarketInput.Blur()
	v.showIntel = true
	v.Invalidate()
}

func (v *GameView) intelDialogRect(bounds runtime.Rect, rows int) runtime.Rect {
	dialogW := 70
	dialogH := rows + 6
	if dialogH < 10 {
		dialogH = 10
	}
	x := bounds.X + (bounds.Width-dialogW)/2
	y := bounds.Y + (bounds.Height-dialogH)/2
	return runtime.Rect{X: x, Y: y, Width: dialogW, Height: dialogH}
}

func (v *GameView) renderIntelDialog(ctx runtime.RenderContext) {
	loc := v.game.Location.Get()
	adjacent := LocationAdjacency[loc]
	bounds := v.Bounds()
	rect := v.intelDialogRect(bounds, len(adjacent))
	ctx.Buffer.Fill(rect, ' ', v.style)
	ctx.Buffer.DrawBox(rect, v.accentStyle)

	ctx.Buffer.SetString(rect.X+2, rect.Y, " INFORMANT INTEL ", v.accentStyle)
	ctx.Buffer.SetString(rect.X+2, rect.Y+2, "Adjacent prices (current hour):", v.dimStyle)

	if len(adjacent) == 0 {
		ctx.Buffer.SetString(rect.X+2, rect.Y+4, "No adjacent locations.", v.style)
		ctx.Buffer.SetString(rect.X+2, rect.Y+rect.Height-2, "[Press any key]", v.dimStyle)
		return
	}

	for i, adj := range adjacent {
		prices := v.game.PricesForLocation(adj)
		line := fmt.Sprintf("%-10s %s", Locations[adj].Name+":", formatIntelPrices(v.game, adj, prices))
		ctx.Buffer.SetString(rect.X+2, rect.Y+4+i, truncPad(line, rect.Width-4), v.style)
	}

	ctx.Buffer.SetString(rect.X+2, rect.Y+rect.Height-2, "[Press any key]", v.dimStyle)
}

func (v *GameView) handleIntelInput(msg runtime.Message) runtime.HandleResult {
	if _, ok := msg.(runtime.KeyMsg); ok {
		v.showIntel = false
		v.Invalidate()
	}
	return runtime.Handled()
}

func formatIntelPrices(game *Game, loc int, prices MarketPrices) string {
	parts := make([]string, 0, len(CandyTypes))
	for _, candy := range CandyTypes {
		if game.isShortage(loc, candy.Name) {
			parts = append(parts, fmt.Sprintf("%s--", candy.Emoji))
			continue
		}
		price := prices[candy.Name]
		parts = append(parts, fmt.Sprintf("%s%d", candy.Emoji, price))
	}
	return strings.Join(parts, " ")
}
