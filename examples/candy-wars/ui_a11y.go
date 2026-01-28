package main

import (
	"fmt"
	"strings"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/widgets"
)

type A11yNode struct {
	widgets.Base
	children []runtime.Widget
}

func NewA11yNode(role accessibility.Role, label string) *A11yNode {
	node := &A11yNode{}
	node.Base.Role = role
	node.Base.Label = label
	return node
}

func (n *A11yNode) Measure(constraints runtime.Constraints) runtime.Size {
	return runtime.Size{}
}

func (n *A11yNode) Layout(bounds runtime.Rect) {
	n.Base.Layout(bounds)
}

func (n *A11yNode) Render(ctx runtime.RenderContext) {}

func (n *A11yNode) HandleMessage(msg runtime.Message) runtime.HandleResult {
	return runtime.Unhandled()
}

func (n *A11yNode) ChildWidgets() []runtime.Widget {
	return n.children
}

func (n *A11yNode) SetChildren(children ...runtime.Widget) {
	n.children = children
}

func a11yDialog(label, description string, children ...runtime.Widget) *A11yNode {
	node := NewA11yNode(accessibility.RoleDialog, label)
	node.Base.Description = description
	node.children = children
	return node
}

func a11yGroup(label string, children ...runtime.Widget) *A11yNode {
	node := NewA11yNode(accessibility.RoleGroup, label)
	node.children = children
	return node
}

func a11yList(label string, items []runtime.Widget) *A11yNode {
	node := NewA11yNode(accessibility.RoleList, label)
	node.children = items
	return node
}

func a11yListItem(label string, selected, disabled bool) *A11yNode {
	node := NewA11yNode(accessibility.RoleListItem, label)
	node.Base.State = accessibility.StateSet{
		Selected: selected,
		Disabled: disabled,
	}
	return node
}

func a11yText(line string) *A11yNode {
	return NewA11yNode(accessibility.RoleText, line)
}

func a11yLine(line string, width int) string {
	if width > 0 {
		line = truncPad(line, width)
	}
	return strings.TrimRight(line, " ")
}

func (v *GameView) dialogA11yWidgets() []runtime.Widget {
	var out []runtime.Widget
	if v.showTrade {
		out = append(out, v.tradeDialogA11y())
	}
	if v.showBank {
		out = append(out, v.bankDialogA11y())
	}
	if v.showLoan {
		out = append(out, v.loanDialogA11y())
	}
	if v.showCraft {
		out = append(out, v.craftDialogA11y())
	}
	if v.showUpgrades {
		out = append(out, v.upgradesDialogA11y())
	}
	if v.showIntel {
		out = append(out, v.intelDialogA11y())
	}
	if v.showGear {
		out = append(out, v.gearDialogA11y())
	}
	if v.showBlackMarket {
		out = append(out, v.blackMarketDialogA11y())
	}
	if v.showStats {
		out = append(out, v.statsDialogA11y())
	}
	if v.showStash {
		out = append(out, v.stashDialogA11y())
	}
	if v.game.InCombat() {
		out = append(out, v.combatDialogA11y())
	}
	if v.showItems {
		out = append(out, v.itemsDialogA11y())
	}
	if v.game.GameOver.Get() {
		out = append(out, v.gameOverDialogA11y())
	}
	return out
}

func (v *GameView) tradeDialogA11y() runtime.Widget {
	action := "Buy"
	if !v.tradeIsBuy {
		action = "Sell"
	}
	title := fmt.Sprintf("%s %s", action, v.tradeCandy)

	price := v.game.Prices.Get()[v.tradeCandy]
	if v.tradeIsBuy {
		price = v.game.buyPrice(v.tradeCandy)
	} else {
		price = v.game.sellPrice(v.tradeCandy)
	}

	width := v.tradeDialogRect(v.Bounds()).Width - 4
	priceLine := a11yLine(fmt.Sprintf("Price: $%d each", price), width)
	stockLine := ""
	if v.tradeIsBuy {
		stockLine = fmt.Sprintf("Stock: %d", v.game.availableStock(v.tradeCandy))
	} else {
		inv := v.game.Inventory.Get()
		stockLine = fmt.Sprintf("You own: %d", inv[v.tradeCandy])
	}
	stockLine = a11yLine(stockLine, width)
	cashLine := a11yLine(fmt.Sprintf("Cash: $%d", v.game.Cash.Get()), width)

	lines := []runtime.Widget{
		a11yText(priceLine),
		a11yText(stockLine),
		a11yText(cashLine),
		a11yGroup("Qty", v.tradeInput),
		a11yText(a11yLine("[Enter] [Esc]", width)),
	}
	return a11yDialog(title, "", lines...)
}

func (v *GameView) bankDialogA11y() runtime.Widget {
	width := v.bankDialogRect(v.Bounds()).Width - 4
	cashLine := a11yLine(fmt.Sprintf("Cash: $%d", v.game.Cash.Get()), width)
	bankLine := a11yLine(fmt.Sprintf("Bank: $%d / $%d", v.game.Bank.Get(), v.game.BankLimit), width)
	mode := "Deposit"
	if v.bankAction == bankWithdraw {
		mode = "Withdraw"
	}
	modeLine := a11yLine(fmt.Sprintf("Mode: %s", mode), width)
	lines := []runtime.Widget{
		a11yText(cashLine),
		a11yText(bankLine),
		a11yText(modeLine),
		a11yGroup("Amount", v.bankInput),
		a11yText(a11yLine("[D]eposit  [W]ithdraw  [Enter] Apply  [Esc] Close", width)),
	}
	return a11yDialog("Bank", "", lines...)
}

func (v *GameView) loanDialogA11y() runtime.Widget {
	width := v.loanDialogRect(v.Bounds()).Width - 4
	items := make([]runtime.Widget, 0, len(LoanTiers))
	for i, tier := range LoanTiers {
		disabled := !v.game.loanTierUnlocked(i)
		line := ""
		if disabled {
			line = fmt.Sprintf("[%d] %s (LOCKED)", i+1, tier.Name)
		} else {
			line = fmt.Sprintf("[%d] %s $%d @ %d%%/day (Heat +%d)", i+1, tier.Name, tier.Amount, tier.InterestPercent, tier.HeatPenalty)
		}
		items = append(items, a11yListItem(a11yLine(line, width), false, disabled))
	}

	totalLoans := v.game.TotalLoanDebt()
	loansLine := "Loans: none"
	if len(v.game.Loans) > 0 {
		parts := make([]string, 0, len(v.game.Loans))
		for _, loan := range v.game.Loans {
			tier := LoanTiers[loan.Tier]
			parts = append(parts, fmt.Sprintf("%s $%d", tier.Name, loan.Balance))
		}
		loansLine = "Loans: " + strings.Join(parts, ", ")
	}

	lines := []runtime.Widget{
		a11yList("Loan Tiers", items),
		a11yText(a11yLine(fmt.Sprintf("Active loans: $%d", totalLoans), width)),
		a11yText(a11yLine(loansLine, width)),
		a11yGroup("Repay", v.loanInput),
		a11yText(a11yLine("[1-3] Take  [R]epay  [Enter] Apply  [Esc] Close", width)),
	}
	return a11yDialog("Loan Shark", "", lines...)
}

func (v *GameView) craftDialogA11y() runtime.Widget {
	width := v.craftDialogRect(v.Bounds()).Width - 4
	invLine := strings.TrimRight(v.craftInventoryLine(width), " ")

	items := make([]runtime.Widget, 0, len(CraftedItems))
	for i, item := range CraftedItems {
		line := fmt.Sprintf("[%d] %s: %s | %s", i+1, item.Name, formatIngredients(item.Ingredients), item.Description)
		items = append(items, a11yListItem(a11yLine(line, width), i == v.craftIndex, false))
	}

	lines := []runtime.Widget{
		a11yText(invLine),
		a11yList("Recipes", items),
		a11yGroup("Qty", v.craftInput),
		a11yText(a11yLine("[1-7] Select  [Enter] Craft  [Esc] Close", width)),
	}
	return a11yDialog("Crafting", "", lines...)
}

func (v *GameView) itemsDialogA11y() runtime.Widget {
	items := v.usableItemNames(v.itemsInCombat)
	width := v.itemsDialogRect(v.Bounds(), len(items)).Width - 4

	title := "Items"
	if v.itemsInCombat {
		title = "Items (Combat)"
	}

	if len(items) == 0 {
		lines := []runtime.Widget{
			a11yText(a11yLine("No usable items.", width)),
			a11yText(a11yLine("[Press any key]", width)),
		}
		return a11yDialog(title, "", lines...)
	}

	inv := v.game.Inventory.Get()
	listItems := make([]runtime.Widget, 0, len(items))
	for i, name := range items {
		item, _ := itemDefByName(name)
		count := inv[name]
		line := fmt.Sprintf("[%d] %s x%d - %s", i+1, item.Name, count, item.Description)
		listItems = append(listItems, a11yListItem(a11yLine(line, width), false, false))
	}

	lines := []runtime.Widget{
		a11yList("Items", listItems),
		a11yText(a11yLine("[1-9] Use  [Esc] Close", width)),
	}
	return a11yDialog(title, "", lines...)
}

func (v *GameView) upgradesDialogA11y() runtime.Widget {
	width := v.upgradesDialogRect(v.Bounds()).Width - 4
	statsLine := fmt.Sprintf("Stats: ATK %d  DEF %d  SPD %d", v.game.playerATK(), v.game.playerDEF(), v.game.playerSPD())

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

	combatLines := v.upgradeCombatLines()
	combatItems := make([]runtime.Widget, 0, len(combatLines))
	for _, line := range combatLines {
		combatItems = append(combatItems, a11yListItem(a11yLine(line, width), false, !combatAvailable))
	}
	combatList := a11yList("Combat Upgrades", combatItems)
	combatGroup := a11yGroup(combatHeader, combatList)
	if !combatAvailable {
		combatGroup.Base.State.Disabled = true
	}

	tradeLines := v.upgradeTradeLines()
	tradeItems := make([]runtime.Widget, 0, len(tradeLines))
	for _, line := range tradeLines {
		tradeItems = append(tradeItems, a11yListItem(a11yLine(line, width), false, !tradeAvailable))
	}
	tradeList := a11yList("Trade Upgrades", tradeItems)
	tradeGroup := a11yGroup(tradeHeader, tradeList)
	if !tradeAvailable {
		tradeGroup.Base.State.Disabled = true
	}

	lines := []runtime.Widget{
		a11yText(a11yLine(statsLine, width)),
		combatGroup,
		tradeGroup,
		a11yText(a11yLine("[1-9,0] Buy  [Esc] Close", width)),
	}
	return a11yDialog("Upgrades", "", lines...)
}

func (v *GameView) intelDialogA11y() runtime.Widget {
	loc := v.game.Location.Get()
	adjacent := LocationAdjacency[loc]
	width := v.intelDialogRect(v.Bounds(), len(adjacent)).Width - 4

	lines := []runtime.Widget{
		a11yText(a11yLine("Adjacent prices (current hour):", width)),
	}

	if len(adjacent) == 0 {
		lines = append(lines,
			a11yText(a11yLine("No adjacent locations.", width)),
			a11yText(a11yLine("[Press any key]", width)),
		)
		return a11yDialog("Informant Intel", "", lines...)
	}

	items := make([]runtime.Widget, 0, len(adjacent))
	for _, adj := range adjacent {
		prices := v.game.PricesForLocation(adj)
		line := fmt.Sprintf("%-10s %s", Locations[adj].Name+":", formatIntelPrices(v.game, adj, prices))
		items = append(items, a11yListItem(a11yLine(line, width), false, false))
	}
	lines = append(lines, a11yList("Locations", items))
	lines = append(lines, a11yText(a11yLine("[Press any key]", width)))
	return a11yDialog("Informant Intel", "", lines...)
}

func (v *GameView) gearDialogA11y() runtime.Widget {
	entries := v.gearEntries()
	width := v.gearDialogRect(v.Bounds(), len(entries)).Width - 4

	items := make([]runtime.Widget, 0, len(entries))
	for _, entry := range entries {
		line := a11yLine(entry.Label, width)
		items = append(items, a11yListItem(line, entry.Equipped, entry.Locked))
	}

	lines := []runtime.Widget{
		a11yText(a11yLine("Weapon Triangle: Blunt > Quick > Sharp > Blunt", width)),
		a11yList("Gear", items),
		a11yText(a11yLine("[A-Z] Buy/Equip  [Esc] Close", width)),
	}
	return a11yDialog("Gear Shop", "", lines...)
}

func (v *GameView) blackMarketDialogA11y() runtime.Widget {
	entries := v.blackMarketEntries()
	v.ensureBlackMarketSelection(entries)
	width := v.blackMarketDialogRect(v.Bounds(), len(entries)).Width - 4

	items := make([]runtime.Widget, 0, len(entries))
	for i, entry := range entries {
		line := a11yLine(entryLine(entry), width)
		items = append(items, a11yListItem(line, i == v.blackMarketIndex, false))
	}

	lines := []runtime.Widget{
		a11yText(a11yLine("Premium gear and contraband candy.", width)),
		a11yList("Market", items),
		a11yGroup("Qty", v.blackMarketInput),
		a11yText(a11yLine("[A-Z] Select  [Enter] Buy  [Esc] Close", width)),
	}
	return a11yDialog("Black Market", "", lines...)
}

func (v *GameView) statsDialogA11y() runtime.Widget {
	lines := v.statsLines()
	width := v.statsDialogRect(v.Bounds(), len(lines)).Width - 4

	title := "Career Stats"
	switch v.statsTab {
	case statsAchievements:
		title = "Achievements"
	case statsHistory:
		title = "Run History"
	}

	items := make([]runtime.Widget, 0, len(lines))
	for _, line := range lines {
		line = a11yLine(line, width)
		if strings.TrimSpace(line) == "" {
			continue
		}
		items = append(items, a11yListItem(line, false, false))
	}

	children := []runtime.Widget{
		a11yList("Stats", items),
		a11yText(a11yLine("[1] Career  [2] Achievements  [3] History  [Esc] Close", width)),
	}
	return a11yDialog(title, "", children...)
}

func (v *GameView) stashDialogA11y() runtime.Widget {
	width := v.stashDialogRect(v.Bounds()).Width - 4

	modeLabel := "Deposit"
	if v.stashMode == stashWithdraw {
		modeLabel = "Withdraw"
	}
	countLine := fmt.Sprintf("Backpack: %d/%d  Stash: %d/%d  Mode: %s",
		v.game.InventoryCount(), v.game.Capacity,
		v.game.StashCount(), v.game.stashCapacity,
		modeLabel,
	)

	lines := []runtime.Widget{
		a11yText(a11yLine(countLine, width)),
	}

	items := v.stashItemNames()
	v.ensureStashSelection(items)
	if len(items) == 0 {
		lines = append(lines, a11yText(a11yLine("No items available.", width)))
	} else {
		listItems := make([]runtime.Widget, 0, len(items))
		for i, name := range items {
			count := v.stashItemCount(name)
			line := fmt.Sprintf("[%d] %s x%d", i+1, name, count)
			listItems = append(listItems, a11yListItem(a11yLine(line, width), name == v.stashItem, false))
		}
		lines = append(lines, a11yList("Items", listItems))
	}

	lines = append(lines,
		a11yGroup("Qty", v.stashInput),
		a11yText(a11yLine("[D]eposit [W]ithdraw [Enter] Move [Esc] Close", width)),
	)

	return a11yDialog("Secret Stash", "", lines...)
}

func (v *GameView) combatDialogA11y() runtime.Widget {
	if !v.game.InCombat() {
		return nil
	}
	rect := v.combatDialogRect(v.Bounds())
	width := rect.Width - 4

	enemy := v.game.Combat.Enemy
	title := fmt.Sprintf("Combat: %s", enemy.Name)

	lines := []runtime.Widget{
		a11yText(a11yLine(fmt.Sprintf("You HP: %d", v.game.HP.Get()), width)),
		a11yText(a11yLine(fmt.Sprintf("%s HP: %d", shortName(enemy.Name), v.game.Combat.EnemyHP), width)),
	}

	logs := v.game.Combat.Log
	maxLines := rect.Height - 6
	if maxLines < 0 {
		maxLines = 0
	}
	if len(logs) > maxLines {
		logs = logs[len(logs)-maxLines:]
	}
	if len(logs) > 0 {
		logItems := make([]runtime.Widget, 0, len(logs))
		for _, line := range logs {
			line = a11yLine(line, width)
			if strings.TrimSpace(line) == "" {
				continue
			}
			logItems = append(logItems, a11yListItem(line, false, false))
		}
		if len(logItems) > 0 {
			lines = append(lines, a11yList("Combat Log", logItems))
		}
	}

	lines = append(lines, a11yText(a11yLine("[A]ttack  [D]efend  [I]tem  [F]lee", width)))
	return a11yDialog(title, "", lines...)
}

func (v *GameView) gameOverDialogA11y() runtime.Widget {
	dialogW := 50
	width := dialogW - 4
	msg := v.game.GameOverMsg.Get()
	msgLines := splitLines(msg, width)

	items := make([]runtime.Widget, 0, len(msgLines))
	for _, line := range msgLines {
		line = a11yLine(line, width)
		if strings.TrimSpace(line) == "" {
			continue
		}
		items = append(items, a11yListItem(line, false, false))
	}

	children := []runtime.Widget{}
	if len(items) > 0 {
		children = append(children, a11yList("Message", items))
	}
	children = append(children, a11yText(a11yLine("[R]estart  [Q]uit", width)))
	return a11yDialog("Game Over", "", children...)
}
