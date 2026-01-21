package main

import (
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/runtime"
	"github.com/odvcencio/fluffy-ui/terminal"
	"github.com/odvcencio/fluffy-ui/widgets"
)

type GameView struct {
	widgets.Component
	game *Game

	// UI elements
	header       *widgets.Label
	statusPanel  *widgets.Panel
	tabs         *widgets.Tabs
	tradeTab     *TradeTabContent
	inventoryTab *InventoryTabContent
	statsTabView *StatsTabContent
	mapTab       *MapTabContent
	messageLabel *widgets.Label
	statsLabel   *widgets.Label
	inventoryLbl *widgets.Label
	scheduleLbl  *widgets.Label
	statusLbl    *widgets.Label
	heatGauge    *widgets.Progress

	toasts     *GameToasts
	toastStack *widgets.ToastStack

	saveDialog *SaveLoadDialog
	loadDialog *SaveLoadDialog
	showSave   bool
	showLoad   bool

	eventModal *EventModal

	// Trade dialog state
	showTrade  bool
	tradeCandy string
	tradeIsBuy bool
	tradeQty   int
	tradeInput *widgets.Input

	// Bank dialog state
	showBank   bool
	bankAction bankAction
	bankInput  *widgets.Input

	// Loan dialog state
	showLoan  bool
	loanInput *widgets.Input
	loanRepay bool

	// Crafting dialog state
	showCraft  bool
	craftIndex int
	craftInput *widgets.Input

	// Items dialog state
	showItems     bool
	itemsInCombat bool

	// Upgrades dialog state
	showUpgrades bool

	// Informant dialog state
	showIntel bool

	// Gear dialog state
	showGear bool

	// Black market dialog state
	showBlackMarket  bool
	blackMarketIndex int
	blackMarketInput *widgets.Input

	// Stats dialog state
	showStats bool
	statsTab  statsTab

	// Stash dialog state
	showStash  bool
	stashMode  stashMode
	stashItem  string
	stashInput *widgets.Input

	// Layout
	focusIndex   int
	style        backend.Style
	dimStyle     backend.Style
	accentStyle  backend.Style
	successStyle backend.Style
	dangerStyle  backend.Style
	warningStyle backend.Style

	lastPrices MarketPrices
	lastHour   int
	lastHeat   int
	lastDay    int
	lastLoc    int

	onRequestNewGame func()
}

type bankAction int

const (
	bankDeposit bankAction = iota
	bankWithdraw
)

func NewGameView(game *Game) *GameView {
	v := &GameView{
		game:         game,
		style:        backend.DefaultStyle(),
		dimStyle:     backend.DefaultStyle().Dim(true),
		accentStyle:  backend.DefaultStyle().Bold(true),
		successStyle: backend.DefaultStyle().Foreground(backend.ColorGreen).Bold(true),
		dangerStyle:  backend.DefaultStyle().Foreground(backend.ColorRed).Bold(true),
		warningStyle: backend.DefaultStyle().Foreground(backend.ColorYellow),
		lastLoc:      -1,
	}

	v.header = widgets.NewLabel("CANDY WARS - Jefferson Middle School").WithStyle(backend.DefaultStyle().Bold(true).Foreground(backend.ColorYellow))

	v.tradeTab = NewTradeTabContent(game, v)
	v.inventoryTab = NewInventoryTabContent(game, v)
	v.statsTabView = NewStatsTabContent(game, v)
	v.mapTab = NewMapTabContent(game, v)
	v.tabs = widgets.NewTabs(
		widgets.Tab{Title: "Trade", Content: v.tradeTab},
		widgets.Tab{Title: "Inventory", Content: v.inventoryTab},
		widgets.Tab{Title: "Stats", Content: v.statsTabView},
		widgets.Tab{Title: "Map", Content: v.mapTab},
	)

	v.messageLabel = widgets.NewLabel("")
	v.statsLabel = widgets.NewLabel("").WithStyle(v.dimStyle)
	v.inventoryLbl = widgets.NewLabel("")
	v.scheduleLbl = widgets.NewLabel("").WithStyle(v.dimStyle)
	v.statusLbl = widgets.NewLabel("").WithStyle(v.dimStyle)

	v.heatGauge = widgets.NewProgress()
	v.heatGauge.Max = 100
	v.heatGauge.Label = "Heat"

	v.toasts = NewGameToasts(v)
	v.toastStack = widgets.NewToastStack()
	v.toastStack.SetOnDismiss(func(id string) {
		v.toasts.Manager().Dismiss(id)
	})
	v.toasts.SetOnChange(v.toastStack.SetToasts)

	v.tradeInput = widgets.NewInput()
	v.tradeInput.SetPlaceholder("Quantity")
	v.tradeInput.OnChange(func(string) {
		v.Invalidate()
	})

	v.bankInput = widgets.NewInput()
	v.bankInput.SetPlaceholder("Amount")
	v.bankInput.OnChange(func(string) {
		v.Invalidate()
	})
	v.bankAction = bankDeposit

	v.loanInput = widgets.NewInput()
	v.loanInput.SetPlaceholder("Amount")
	v.loanInput.OnChange(func(string) {
		v.Invalidate()
	})
	v.loanRepay = false

	v.craftInput = widgets.NewInput()
	v.craftInput.SetPlaceholder("1")
	v.craftInput.OnChange(func(string) {
		v.Invalidate()
	})
	v.craftIndex = 0

	v.stashInput = widgets.NewInput()
	v.stashInput.SetPlaceholder("1")
	v.stashInput.OnChange(func(string) {
		v.Invalidate()
	})
	v.stashMode = stashDeposit

	v.blackMarketInput = widgets.NewInput()
	v.blackMarketInput.SetPlaceholder("1")
	v.blackMarketInput.OnChange(func(string) {
		v.Invalidate()
	})
	v.blackMarketIndex = 0
	v.statsTab = statsCareer

	v.saveDialog = NewSaveLoadDialog(saveMode, func(slot int) {
		v.handleSaveSlot(slot)
	})
	v.saveDialog.OnDismiss(func() {
		v.showSave = false
		v.Invalidate()
	})
	v.loadDialog = NewSaveLoadDialog(loadMode, func(slot int) {
		v.handleLoadSlot(slot)
	})
	v.loadDialog.OnDismiss(func() {
		v.showLoad = false
		v.Invalidate()
	})

	v.refresh()
	return v
}

func (v *GameView) Mount() {
	v.Observe(v.game.Cash, v.refresh)
	v.Observe(v.game.Debt, v.refresh)
	v.Observe(v.game.Bank, v.refresh)
	v.Observe(v.game.Day, v.refresh)
	v.Observe(v.game.Hour, v.refresh)
	v.Observe(v.game.Prices, v.refresh)
	v.Observe(v.game.Inventory, v.refresh)
	v.Observe(v.game.Message, v.refresh)
	v.Observe(v.game.Heat, v.refresh)
	v.Observe(v.game.HP, v.refresh)
	v.Observe(v.game.Location, v.refresh)
	v.Observe(v.game.ShowEvent, v.refresh)
	v.Observe(v.game.GameOver, v.refresh)
	v.refresh()
}

func (v *GameView) Unmount() {
	v.Subs.Clear()
}

func (v *GameView) refresh() {
	if v.tradeTab != nil {
		v.tradeTab.UpdateMarketTable()
	}
	v.messageLabel.SetText(v.game.Message.Get())
	statsLine := fmt.Sprintf(
		"Life %d  |  Record %d-%d  |  Best $%d",
		v.game.Runs, v.game.Wins, v.game.Losses, v.game.BestWorth,
	)
	if unlocked, total := v.game.AchievementCount(); total > 0 {
		statsLine += fmt.Sprintf("  |  Achv %d/%d", unlocked, total)
	}
	v.statsLabel.SetText(statsLine)
	if v.game.GameOver.Get() && v.showTrade {
		v.showTrade = false
		v.tradeInput.Blur()
	}
	if v.game.GameOver.Get() && v.showBank {
		v.showBank = false
		v.bankInput.Blur()
	}
	if v.game.GameOver.Get() && v.showLoan {
		v.showLoan = false
		v.loanInput.Blur()
	}
	if v.game.GameOver.Get() && v.showCraft {
		v.showCraft = false
		v.craftInput.Blur()
	}
	if v.game.GameOver.Get() && v.showItems {
		v.showItems = false
	}
	if v.game.GameOver.Get() && v.showUpgrades {
		v.showUpgrades = false
	}
	if v.game.GameOver.Get() && v.showIntel {
		v.showIntel = false
	}
	if v.game.GameOver.Get() && v.showGear {
		v.showGear = false
	}
	if v.game.GameOver.Get() && v.showBlackMarket {
		v.showBlackMarket = false
		v.blackMarketInput.Blur()
	}
	if v.game.GameOver.Get() && v.showStats {
		v.showStats = false
	}
	if v.game.GameOver.Get() && v.showStash {
		v.showStash = false
		v.stashInput.Blur()
	}
	if v.game.GameOver.Get() && v.showSave {
		v.showSave = false
	}
	if v.game.GameOver.Get() && v.showLoad {
		v.showLoad = false
	}

	inv := v.game.Inventory.Get()
	invText := fmt.Sprintf("Backpack: %d/%d", v.game.InventoryCount(), v.game.Capacity)
	if len(inv) > 0 {
		invText += " |"
		for name, qty := range inv {
			if qty > 0 {
				invText += fmt.Sprintf(" %s:%d", shortName(name), qty)
			}
		}
	}
	if v.game.tradeBuffUses > 0 {
		invText += fmt.Sprintf(" | Charm:%d", v.game.tradeBuffUses)
	}
	if v.game.hasStash {
		invText += fmt.Sprintf(" | Stash:%d/%d", v.game.StashCount(), v.game.stashCapacity)
	}
	weaponName := shortName(v.game.playerWeapon().Name)
	armorName := shortName(v.game.playerArmor().Name)
	invText += fmt.Sprintf(" | Gear:%s/%s", weaponName, armorName)
	v.inventoryLbl.SetText(invText)
	v.scheduleLbl.SetText(formatScheduleLine(v.game))
	v.statusLbl.SetText(formatStatusLine(v.game))

	heat := v.game.Heat.Get()
	v.heatGauge.Value = float64(heat)

	v.syncToasts()
	v.syncEventModal()

	v.Invalidate()
}

func (v *GameView) updateMarketTable() {
	if v.tradeTab != nil {
		v.tradeTab.UpdateMarketTable()
	}
}

func (v *GameView) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.MaxSize()
}

func (v *GameView) Layout(bounds runtime.Rect) {
	v.Component.Layout(bounds)

	// Header row
	v.header.Layout(runtime.Rect{X: bounds.X, Y: bounds.Y, Width: bounds.Width, Height: 1})

	contentTop := bounds.Y + 2
	helpY := bounds.Y + bounds.Height - 2
	statsY := helpY - 1
	messageY := helpY - 2
	heatY := helpY - 3
	statusY := helpY - 4
	scheduleY := helpY - 5
	inventoryY := helpY - 6

	tabsHeight := inventoryY - contentTop - 1
	if tabsHeight < 1 {
		tabsHeight = 1
	}

	v.tabs.Layout(runtime.Rect{X: bounds.X, Y: contentTop, Width: bounds.Width, Height: tabsHeight})

	v.inventoryLbl.Layout(runtime.Rect{X: bounds.X, Y: inventoryY, Width: bounds.Width, Height: 1})
	v.scheduleLbl.Layout(runtime.Rect{X: bounds.X, Y: scheduleY, Width: bounds.Width, Height: 1})
	v.statusLbl.Layout(runtime.Rect{X: bounds.X, Y: statusY, Width: bounds.Width, Height: 1})
	v.heatGauge.Layout(runtime.Rect{X: bounds.X, Y: heatY, Width: bounds.Width, Height: 1})
	v.messageLabel.Layout(runtime.Rect{X: bounds.X, Y: messageY, Width: bounds.Width, Height: 1})
	v.statsLabel.Layout(runtime.Rect{X: bounds.X, Y: statsY, Width: bounds.Width, Height: 1})

	if v.toastStack != nil {
		v.toastStack.Layout(bounds)
	}

	if v.eventModal != nil {
		v.eventModal.Layout(v.eventModal.CenteredBounds(bounds))
	}

	if v.showSave && v.saveDialog != nil {
		v.saveDialog.Layout(bounds)
	}
	if v.showLoad && v.loadDialog != nil {
		v.loadDialog.Layout(bounds)
	}

	// Trade input (hidden unless trading)
	if v.showTrade {
		inputRect := v.tradeInputRect(bounds)
		v.tradeInput.Layout(inputRect)
	}

	if v.showCraft {
		inputRect := v.craftInputRect(bounds)
		v.craftInput.Layout(inputRect)
	}

	if v.showStash {
		inputRect := v.stashInputRect(bounds)
		v.stashInput.Layout(inputRect)
	}

	if v.showBlackMarket {
		inputRect := v.blackMarketInputRect(bounds)
		v.blackMarketInput.Layout(inputRect)
	}
}

func (v *GameView) Render(ctx runtime.RenderContext) {
	if ctx.Buffer == nil {
		return
	}
	bounds := v.Bounds()
	ctx.Clear(v.style)

	// Header
	v.header.Render(ctx)

	// Status bar with colored segments
	y := bounds.Y + 1
	loc := Locations[v.game.Location.Get()]
	x := bounds.X

	// Day
	dayText := fmt.Sprintf("Day %d/%d", v.game.Day.Get(), v.game.MaxDays)
	ctx.Buffer.SetString(x, y, dayText, v.accentStyle)
	x += len(dayText) + 1

	ctx.Buffer.SetString(x, y, "|", v.dimStyle)
	x += 2

	// Hour
	hourText := fmt.Sprintf("Hour %d/%d", v.game.Hour.Get(), hoursPerDay)
	ctx.Buffer.SetString(x, y, hourText, v.accentStyle)
	x += len(hourText) + 1

	ctx.Buffer.SetString(x, y, "|", v.dimStyle)
	x += 2

	// Cash (green)
	cashText := fmt.Sprintf("Cash: $%d", v.game.Cash.Get())
	ctx.Buffer.SetString(x, y, cashText, v.successStyle)
	x += len(cashText) + 1

	ctx.Buffer.SetString(x, y, "|", v.dimStyle)
	x += 2

	// Debt (red)
	debtText := fmt.Sprintf("Debt: $%d", v.game.TotalDebt())
	ctx.Buffer.SetString(x, y, debtText, v.dangerStyle)
	x += len(debtText) + 1

	ctx.Buffer.SetString(x, y, "|", v.dimStyle)
	x += 2

	// Bank
	bankText := fmt.Sprintf("Bank: $%d", v.game.Bank.Get())
	ctx.Buffer.SetString(x, y, bankText, v.accentStyle)
	x += len(bankText) + 1

	ctx.Buffer.SetString(x, y, "|", v.dimStyle)
	x += 2

	// HP
	hp := v.game.HP.Get()
	hpStyle := v.accentStyle
	if hp <= maxHP/3 {
		hpStyle = v.dangerStyle
	} else if hp <= maxHP/2 {
		hpStyle = v.warningStyle
	}
	hpText := fmt.Sprintf("HP: %d", hp)
	ctx.Buffer.SetString(x, y, hpText, hpStyle)
	x += len(hpText) + 1

	ctx.Buffer.SetString(x, y, "|", v.dimStyle)
	x += 2

	// Location
	locText := fmt.Sprintf("Location: %s", loc.Name)
	ctx.Buffer.SetString(x, y, locText, v.accentStyle)

	// Tabs
	v.tabs.Render(ctx)

	// Inventory
	v.inventoryLbl.Render(ctx)
	v.scheduleLbl.Render(ctx)
	v.statusLbl.Render(ctx)
	v.heatGauge.Render(ctx)

	// Message
	v.messageLabel.Render(ctx)
	v.statsLabel.Render(ctx)

	// Help text
	helpY := bounds.Y + bounds.Height - 2
	help := "[B]uy  [S]ell  [C]raft  [U]se Item  [G]Upgrades  [W]Gear  [R]ival"
	if v.game.hasStash {
		help += "  [H]Stash"
	}
	if v.game.hasInformant {
		help += "  [N]Intel"
	}
	if v.game.blackMarketUnlocked() {
		help += "  [M]Black Market"
	}
	help += "  [T]Stats  [P]ay Debt  [K]Bank  [L]Loan  [1-6]Travel  [E]nd Day  [F5]Save  [F9]Load  [[]/[]] Tabs  [Q]uit"
	if v.game.GameOver.Get() {
		help = "[R]estart  [Q]uit"
	}
	ctx.Buffer.SetString(bounds.X, helpY, help, v.dimStyle)

	// Trade dialog
	if v.showTrade {
		v.renderTradeDialog(ctx)
	}

	// Bank dialog
	if v.showBank {
		v.renderBankDialog(ctx)
	}

	// Loan dialog
	if v.showLoan {
		v.renderLoanDialog(ctx)
	}

	// Craft dialog
	if v.showCraft {
		v.renderCraftDialog(ctx)
	}

	// Upgrades dialog
	if v.showUpgrades {
		v.renderUpgradesDialog(ctx)
	}

	// Informant dialog
	if v.showIntel {
		v.renderIntelDialog(ctx)
	}

	// Gear dialog
	if v.showGear {
		v.renderGearDialog(ctx)
	}

	// Black market dialog
	if v.showBlackMarket {
		v.renderBlackMarketDialog(ctx)
	}

	// Stats dialog
	if v.showStats {
		v.renderStatsDialog(ctx)
	}

	// Stash dialog
	if v.showStash {
		v.renderStashDialog(ctx)
	}

	// Event dialog
	if v.eventModal != nil && !v.game.InCombat() {
		v.eventModal.Render(ctx)
	}

	// Combat dialog
	if v.game.InCombat() {
		v.renderCombatDialog(ctx)
	}

	// Items dialog
	if v.showItems {
		v.renderItemsDialog(ctx)
	}

	// Game over dialog
	if v.game.GameOver.Get() {
		v.renderGameOverDialog(ctx)
	}

	// Save/load dialogs
	if v.showSave && v.saveDialog != nil {
		v.saveDialog.Render(ctx)
	}
	if v.showLoad && v.loadDialog != nil {
		v.loadDialog.Render(ctx)
	}

	// Toasts
	if v.toastStack != nil {
		v.toastStack.SetNow(time.Now())
		v.toastStack.Render(ctx)
	}
}

func (v *GameView) tradeDialogRect(bounds runtime.Rect) runtime.Rect {
	dialogW := 40
	dialogH := 8
	x := bounds.X + (bounds.Width-dialogW)/2
	y := bounds.Y + (bounds.Height-dialogH)/2
	return runtime.Rect{X: x, Y: y, Width: dialogW, Height: dialogH}
}

func (v *GameView) bankDialogRect(bounds runtime.Rect) runtime.Rect {
	dialogW := 44
	dialogH := 9
	x := bounds.X + (bounds.Width-dialogW)/2
	y := bounds.Y + (bounds.Height-dialogH)/2
	return runtime.Rect{X: x, Y: y, Width: dialogW, Height: dialogH}
}

func (v *GameView) loanDialogRect(bounds runtime.Rect) runtime.Rect {
	dialogW := 60
	dialogH := 12
	x := bounds.X + (bounds.Width-dialogW)/2
	y := bounds.Y + (bounds.Height-dialogH)/2
	return runtime.Rect{X: x, Y: y, Width: dialogW, Height: dialogH}
}

func (v *GameView) combatDialogRect(bounds runtime.Rect) runtime.Rect {
	dialogW := 60
	dialogH := 12
	x := bounds.X + (bounds.Width-dialogW)/2
	y := bounds.Y + (bounds.Height-dialogH)/2
	return runtime.Rect{X: x, Y: y, Width: dialogW, Height: dialogH}
}

func (v *GameView) tradeInputRect(bounds runtime.Rect) runtime.Rect {
	dialog := v.tradeDialogRect(bounds)
	label := "Qty: "
	hint := "[Enter] [Esc]"
	inputX := dialog.X + 2 + len(label)
	hintX := dialog.X + dialog.Width - 2 - len(hint)
	inputW := hintX - inputX - 1
	if inputW < 4 {
		inputW = 4
	}
	return runtime.Rect{X: inputX, Y: dialog.Y + 6, Width: inputW, Height: 1}
}

func (v *GameView) renderTradeDialog(ctx runtime.RenderContext) {
	bounds := v.Bounds()
	rect := v.tradeDialogRect(bounds)
	ctx.Buffer.Fill(rect, ' ', v.style)
	ctx.Buffer.DrawBox(rect, v.accentStyle)

	action := "Buy"
	if !v.tradeIsBuy {
		action = "Sell"
	}
	title := fmt.Sprintf(" %s %s ", action, v.tradeCandy)
	ctx.Buffer.SetString(rect.X+2, rect.Y, title, v.accentStyle)

	price := v.game.Prices.Get()[v.tradeCandy]
	if v.tradeIsBuy {
		price = v.game.buyPrice(v.tradeCandy)
	} else {
		price = v.game.sellPrice(v.tradeCandy)
	}
	inv := v.game.Inventory.Get()
	owned := inv[v.tradeCandy]

	ctx.Buffer.SetString(rect.X+2, rect.Y+2, fmt.Sprintf("Price: $%d each", price), v.style)
	if v.tradeIsBuy {
		stock := v.game.availableStock(v.tradeCandy)
		ctx.Buffer.SetString(rect.X+2, rect.Y+3, fmt.Sprintf("Stock: %d", stock), v.style)
		ctx.Buffer.SetString(rect.X+2, rect.Y+4, fmt.Sprintf("Cash: $%d", v.game.Cash.Get()), v.style)
	} else {
		ctx.Buffer.SetString(rect.X+2, rect.Y+3, fmt.Sprintf("You own: %d", owned), v.style)
		ctx.Buffer.SetString(rect.X+2, rect.Y+4, fmt.Sprintf("Cash: $%d", v.game.Cash.Get()), v.style)
	}

	label := "Qty: "
	hint := "[Enter] [Esc]"
	ctx.Buffer.SetString(rect.X+2, rect.Y+6, label, v.style)
	v.tradeInput.Layout(v.tradeInputRect(bounds))
	v.tradeInput.Render(ctx)

	hintX := rect.X + rect.Width - 2 - len(hint)
	ctx.Buffer.SetString(hintX, rect.Y+6, hint, v.dimStyle)
}

func (v *GameView) renderEventDialog(ctx runtime.RenderContext) {
	bounds := v.Bounds()
	dialogW := 50
	dialogH := 10
	x := (bounds.Width - dialogW) / 2
	y := (bounds.Height - dialogH) / 2

	rect := runtime.Rect{X: x, Y: y, Width: dialogW, Height: dialogH}
	ctx.Buffer.Fill(rect, ' ', v.style)
	ctx.Buffer.DrawBox(rect, v.accentStyle)

	title := " " + v.game.EventTitle.Get() + " "
	ctx.Buffer.SetString(x+2, y, title, v.accentStyle.Reverse(true))

	msg := v.game.EventMessage.Get()
	lines := splitLines(msg, dialogW-4)
	for i, line := range lines {
		if i < dialogH-3 {
			ctx.Buffer.SetString(x+2, y+2+i, line, v.style)
		}
	}

	if len(v.game.eventOptions) == 0 {
		ctx.Buffer.SetString(x+2, y+dialogH-2, "[Press any key to continue]", v.dimStyle)
		return
	}
	optionsLine := ""
	for _, option := range v.game.eventOptions {
		if optionsLine != "" {
			optionsLine += "  "
		}
		optionsLine += fmt.Sprintf("[%c] %s", unicode.ToUpper(option.Key), option.Label)
	}
	ctx.Buffer.SetString(x+2, y+dialogH-2, optionsLine, v.dimStyle)
}

func (v *GameView) renderGameOverDialog(ctx runtime.RenderContext) {
	bounds := v.Bounds()
	dialogW := 50
	dialogH := 14
	x := (bounds.Width - dialogW) / 2
	y := (bounds.Height - dialogH) / 2

	rect := runtime.Rect{X: x, Y: y, Width: dialogW, Height: dialogH}
	ctx.Buffer.Fill(rect, ' ', v.style)
	ctx.Buffer.DrawBox(rect, v.accentStyle)

	ctx.Buffer.SetString(x+2, y, " GAME OVER ", v.accentStyle.Reverse(true))

	msg := v.game.GameOverMsg.Get()
	lines := splitLines(msg, dialogW-4)
	for i, line := range lines {
		if i < dialogH-3 {
			ctx.Buffer.SetString(x+2, y+2+i, line, v.style)
		}
	}

	ctx.Buffer.SetString(x+2, y+dialogH-2, "[R]estart  [Q]uit", v.dimStyle)
}

func (v *GameView) renderBankDialog(ctx runtime.RenderContext) {
	bounds := v.Bounds()
	rect := v.bankDialogRect(bounds)
	ctx.Buffer.Fill(rect, ' ', v.style)
	ctx.Buffer.DrawBox(rect, v.accentStyle)

	title := " BANK "
	ctx.Buffer.SetString(rect.X+2, rect.Y, title, v.accentStyle)

	cash := v.game.Cash.Get()
	bank := v.game.Bank.Get()
	limit := v.game.BankLimit
	ctx.Buffer.SetString(rect.X+2, rect.Y+2, fmt.Sprintf("Cash: $%d", cash), v.style)
	ctx.Buffer.SetString(rect.X+2, rect.Y+3, fmt.Sprintf("Bank: $%d / $%d", bank, limit), v.style)

	mode := "Deposit"
	if v.bankAction == bankWithdraw {
		mode = "Withdraw"
	}
	ctx.Buffer.SetString(rect.X+2, rect.Y+4, fmt.Sprintf("Mode: %s", mode), v.style)

	label := "Amount: "
	ctx.Buffer.SetString(rect.X+2, rect.Y+6, label, v.style)
	v.bankInput.Layout(runtime.Rect{
		X:      rect.X + 2 + len(label),
		Y:      rect.Y + 6,
		Width:  rect.Width - len(label) - 4,
		Height: 1,
	})
	v.bankInput.Render(ctx)

	ctx.Buffer.SetString(rect.X+2, rect.Y+rect.Height-2, "[D]eposit  [W]ithdraw  [Enter] Apply  [Esc] Close", v.dimStyle)
}

func (v *GameView) renderLoanDialog(ctx runtime.RenderContext) {
	bounds := v.Bounds()
	rect := v.loanDialogRect(bounds)
	ctx.Buffer.Fill(rect, ' ', v.style)
	ctx.Buffer.DrawBox(rect, v.accentStyle)

	title := " LOAN SHARK "
	ctx.Buffer.SetString(rect.X+2, rect.Y, title, v.accentStyle)

	for i, tier := range LoanTiers {
		style := v.style
		line := ""
		if v.game.loanTierUnlocked(i) {
			line = fmt.Sprintf("[%d] %s $%d @ %d%%/day (Heat +%d)",
				i+1, tier.Name, tier.Amount, tier.InterestPercent, tier.HeatPenalty)
		} else {
			line = fmt.Sprintf("[%d] %s (LOCKED)", i+1, tier.Name)
			style = v.dimStyle
		}
		ctx.Buffer.SetString(rect.X+2, rect.Y+2+i, line, style)
	}

	loanY := rect.Y + 6
	totalLoans := v.game.TotalLoanDebt()
	ctx.Buffer.SetString(rect.X+2, loanY, fmt.Sprintf("Active loans: $%d", totalLoans), v.style)
	loansLine := "Loans: none"
	if len(v.game.Loans) > 0 {
		parts := make([]string, 0, len(v.game.Loans))
		for _, loan := range v.game.Loans {
			tier := LoanTiers[loan.Tier]
			parts = append(parts, fmt.Sprintf("%s $%d", tier.Name, loan.Balance))
		}
		loansLine = "Loans: " + strings.Join(parts, ", ")
	}
	ctx.Buffer.SetString(rect.X+2, loanY+1, truncPad(loansLine, rect.Width-4), v.style)

	label := "Repay: "
	ctx.Buffer.SetString(rect.X+2, rect.Y+rect.Height-3, label, v.style)
	v.loanInput.Layout(runtime.Rect{
		X:      rect.X + 2 + len(label),
		Y:      rect.Y + rect.Height - 3,
		Width:  rect.Width - len(label) - 4,
		Height: 1,
	})
	v.loanInput.Render(ctx)

	ctx.Buffer.SetString(rect.X+2, rect.Y+rect.Height-2, "[1-3] Take  [R]epay  [Enter] Apply  [Esc] Close", v.dimStyle)
}

func (v *GameView) renderCombatDialog(ctx runtime.RenderContext) {
	if !v.game.InCombat() {
		return
	}
	bounds := v.Bounds()
	rect := v.combatDialogRect(bounds)
	ctx.Buffer.Fill(rect, ' ', v.style)
	ctx.Buffer.DrawBox(rect, v.accentStyle)

	enemy := v.game.Combat.Enemy
	title := fmt.Sprintf(" COMBAT: %s ", enemy.Name)
	ctx.Buffer.SetString(rect.X+2, rect.Y, title, v.accentStyle)

	ctx.Buffer.SetString(rect.X+2, rect.Y+2, fmt.Sprintf("You HP: %d", v.game.HP.Get()), v.style)
	ctx.Buffer.SetString(rect.X+24, rect.Y+2, fmt.Sprintf("%s HP: %d", shortName(enemy.Name), v.game.Combat.EnemyHP), v.style)

	logStart := rect.Y + 4
	maxLines := rect.Height - 6
	logs := v.game.Combat.Log
	if len(logs) > maxLines {
		logs = logs[len(logs)-maxLines:]
	}
	for i, line := range logs {
		ctx.Buffer.SetString(rect.X+2, logStart+i, line, v.style)
	}

	ctx.Buffer.SetString(rect.X+2, rect.Y+rect.Height-2, "[A]ttack  [D]efend  [I]tem  [F]lee", v.dimStyle)
}

func (v *GameView) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if v.toastStack != nil {
		if result := v.toastStack.HandleMessage(msg); result.Handled {
			return result
		}
	}

	if v.showSave && v.saveDialog != nil {
		return v.saveDialog.HandleMessage(msg)
	}
	if v.showLoad && v.loadDialog != nil {
		return v.loadDialog.HandleMessage(msg)
	}

	// Handle game over state
	if v.game.GameOver.Get() {
		if key, ok := msg.(runtime.KeyMsg); ok {
			if key.Rune == 'q' || key.Rune == 'Q' {
				return runtime.WithCommand(runtime.Quit{})
			}
			if key.Rune == 'r' || key.Rune == 'R' {
				if v.onRequestNewGame != nil {
					v.onRequestNewGame()
					return runtime.Handled()
				}
				v.restartGame()
				return runtime.Handled()
			}
		}
		return runtime.Handled()
	}

	if v.eventModal != nil && !v.game.InCombat() {
		return v.eventModal.HandleMessage(msg)
	}

	if v.showItems {
		return v.handleItemsInput(msg)
	}

	if v.showCraft {
		return v.handleCraftInput(msg)
	}

	if v.showUpgrades {
		return v.handleUpgradesInput(msg)
	}

	if v.showIntel {
		return v.handleIntelInput(msg)
	}

	if v.showGear {
		return v.handleGearInput(msg)
	}

	if v.showBlackMarket {
		return v.handleBlackMarketInput(msg)
	}

	if v.showStats {
		return v.handleStatsInput(msg)
	}

	if v.showStash {
		return v.handleStashInput(msg)
	}

	if v.game.InCombat() {
		return v.handleCombatInput(msg)
	}

	if v.showBank {
		return v.handleBankInput(msg)
	}

	if v.showLoan {
		return v.handleLoanInput(msg)
	}

	// Handle trade dialog
	if v.showTrade {
		return v.handleTradeInput(msg)
	}

	if result := v.handleActiveTabInput(msg); result.Handled {
		return result
	}

	// Normal gameplay
	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		return runtime.Unhandled()
	}

	switch key.Key {
	case terminal.KeyF5:
		v.openSaveDialog()
		return runtime.Handled()
	case terminal.KeyF9:
		v.openLoadDialog()
		return runtime.Handled()
	case terminal.KeyLeft:
		v.tabs.SetSelected(v.tabs.SelectedIndex() - 1)
		v.Invalidate()
		return runtime.Handled()
	case terminal.KeyRight:
		v.tabs.SetSelected(v.tabs.SelectedIndex() + 1)
		v.Invalidate()
		return runtime.Handled()
	}

	switch key.Rune {
	case 'q', 'Q':
		return runtime.WithCommand(runtime.Quit{})

	case 'b', 'B':
		v.startTrade(true)
		return runtime.Handled()

	case 's', 'S':
		v.startTrade(false)
		return runtime.Handled()

	case 'c', 'C':
		v.openCraftDialog()
		return runtime.Handled()

	case 'u', 'U':
		v.openItemsDialog(false)
		return runtime.Handled()

	case 'g', 'G':
		v.openUpgradesDialog()
		return runtime.Handled()

	case 'w', 'W':
		v.openGearDialog()
		return runtime.Handled()

	case 'r', 'R':
		v.game.ChallengeRival()
		v.refresh()
		return runtime.Handled()

	case 'n', 'N':
		v.openIntelDialog()
		return runtime.Handled()

	case 'm', 'M':
		v.openBlackMarketDialog()
		return runtime.Handled()

	case 't', 'T':
		v.openStatsDialog()
		return runtime.Handled()

	case 'h', 'H':
		v.openStashDialog()
		return runtime.Handled()

	case 'p', 'P':
		v.game.PayDebt(v.game.Cash.Get())
		v.refresh()
		return runtime.Handled()

	case 'k', 'K':
		v.openBankDialog()
		return runtime.Handled()

	case 'l', 'L':
		v.openLoanDialog()
		return runtime.Handled()

	case 'e', 'E':
		v.game.EndDayEarly()
		v.refresh()
		return runtime.Handled()

	case '[':
		v.tabs.SetSelected(v.tabs.SelectedIndex() - 1)
		v.Invalidate()
		return runtime.Handled()
	case ']':
		v.tabs.SetSelected(v.tabs.SelectedIndex() + 1)
		v.Invalidate()
		return runtime.Handled()

	case '1', '2', '3', '4', '5', '6':
		idx := int(key.Rune - '1')
		if idx >= 0 && idx < len(Locations) {
			v.game.Travel(idx)
			v.refresh()
		}
		return runtime.Handled()
	}

	return runtime.Unhandled()
}

func (v *GameView) startTrade(isBuy bool) {
	// Use the currently selected candy from the market table.
	selected := 0
	if v.tradeTab != nil && v.tradeTab.marketTable != nil {
		selected = v.tradeTab.marketTable.SelectedIndex()
	}
	if selected < 0 || selected >= len(CandyTypes) {
		selected = 0
	}

	v.showBank = false
	v.showLoan = false
	v.showCraft = false
	v.showItems = false
	v.showUpgrades = false
	v.showIntel = false
	v.showGear = false
	v.showBlackMarket = false
	v.showStats = false
	v.showStash = false
	v.bankInput.Blur()
	v.loanInput.Blur()
	v.craftInput.Blur()
	v.stashInput.Blur()
	v.blackMarketInput.Blur()

	v.showTrade = true
	v.tradeIsBuy = isBuy
	v.tradeCandy = CandyTypes[selected].Name
	v.tradeQty = 1
	v.tradeInput.Clear()
	v.tradeInput.SetPlaceholder("1")
	v.tradeInput.Focus()
	v.Invalidate()
}

func (v *GameView) handleActiveTabInput(msg runtime.Message) runtime.HandleResult {
	if v.tabs == nil {
		return runtime.Unhandled()
	}
	switch v.tabs.SelectedIndex() {
	case 0:
		if v.tradeTab != nil {
			return v.tradeTab.HandleMessage(msg)
		}
	case 1:
		if v.inventoryTab != nil {
			return v.inventoryTab.HandleMessage(msg)
		}
	case 2:
		if v.statsTabView != nil {
			return v.statsTabView.HandleMessage(msg)
		}
	case 3:
		if v.mapTab != nil {
			return v.mapTab.HandleMessage(msg)
		}
	}
	return runtime.Unhandled()
}

func (v *GameView) openSaveDialog() {
	if v.game.GameOver.Get() || v.game.InCombat() {
		return
	}
	slots, err := ListSaveSlots()
	if err != nil {
		v.game.Message.Set("Unable to list save slots.")
		v.refresh()
		return
	}
	v.closeOverlays()
	v.saveDialog.UpdateSlots(slots)
	v.showSave = true
	v.showLoad = false
	v.Invalidate()
}

func (v *GameView) openLoadDialog() {
	if v.game.InCombat() {
		return
	}
	slots, err := ListSaveSlots()
	if err != nil {
		v.game.Message.Set("Unable to list save slots.")
		v.refresh()
		return
	}
	v.closeOverlays()
	v.loadDialog.UpdateSlots(slots)
	v.showLoad = true
	v.showSave = false
	v.Invalidate()
}

func (v *GameView) handleSaveSlot(slot int) {
	if v.saveDialog == nil {
		return
	}
	name := fmt.Sprintf("Slot %d", slot)
	for _, info := range v.saveDialog.slots {
		if info.Slot == slot && info.Name != "" {
			name = info.Name
			break
		}
	}
	if err := v.game.SaveToSlot(slot, name); err != nil {
		v.game.Message.Set("Save failed: " + err.Error())
	} else if v.toasts != nil {
		v.toasts.ShowSuccess("Game Saved", fmt.Sprintf("Slot %d", slot))
	}
	v.showSave = false
	v.refresh()
}

func (v *GameView) handleLoadSlot(slot int) {
	if err := v.game.LoadFromSlot(slot); err != nil {
		v.game.Message.Set("Load failed: " + err.Error())
	} else if v.toasts != nil {
		v.toasts.ShowSuccess("Game Loaded", fmt.Sprintf("Slot %d", slot))
	}
	v.lastPrices = nil
	v.lastHour = 0
	v.lastHeat = 0
	v.lastDay = 0
	v.lastLoc = -1
	v.showLoad = false
	v.refresh()
}

func (v *GameView) closeOverlays() {
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
	v.showStash = false
	v.showSave = false
	v.showLoad = false
	v.tradeInput.Blur()
	v.bankInput.Blur()
	v.loanInput.Blur()
	v.craftInput.Blur()
	v.stashInput.Blur()
	v.blackMarketInput.Blur()
}

func (v *GameView) syncToasts() {
	if v.toasts == nil {
		return
	}
	prices := v.game.Prices.Get()
	locIdx := v.game.Location.Get()
	if v.lastLoc == -1 {
		v.lastLoc = locIdx
	}
	if locIdx != v.lastLoc {
		v.lastPrices = copyMarketPrices(prices)
		v.lastLoc = locIdx
		return
	}
	if v.lastPrices == nil {
		v.lastPrices = copyMarketPrices(prices)
	} else {
		locName := Locations[locIdx].Name
		for _, candy := range CandyTypes {
			oldPrice := v.lastPrices[candy.Name]
			newPrice := prices[candy.Name]
			if oldPrice != 0 && newPrice != oldPrice {
				v.toasts.ShowPriceAlert(candy.Name, oldPrice, newPrice, locName)
			}
		}
		v.lastPrices = copyMarketPrices(prices)
	}

	hour := v.game.Hour.Get()
	if v.lastHour == 0 {
		v.lastHour = hour
	} else if hour != v.lastHour {
		hoursLeft := hoursPerDay - hour + 1
		v.toasts.ShowTimeWarning(hoursLeft)
		v.lastHour = hour
	}

	heat := v.game.Heat.Get()
	if (v.lastHeat < 50 && heat >= 50) || (v.lastHeat < 75 && heat >= 75) {
		v.toasts.ShowHeatWarning(heat)
	}
	v.lastHeat = heat

	day := v.game.Day.Get()
	if v.lastDay == 0 {
		v.lastDay = day
	} else if day != v.lastDay {
		daysLeft := v.game.MaxDays - day
		v.toasts.ShowDebtReminder(v.game.TotalDebt(), daysLeft)
		v.lastDay = day
	}
}

func (v *GameView) syncEventModal() {
	if v.game.ShowEvent.Get() && !v.game.InCombat() {
		if v.eventModal == nil || v.eventModal.title != v.game.EventTitle.Get() || v.eventModal.message != v.game.EventMessage.Get() {
			v.eventModal = v.buildEventModal()
		}
		return
	}
	v.eventModal = nil
}

func (v *GameView) buildEventModal() *EventModal {
	choices := make([]EventChoice, 0, len(v.game.eventOptions))
	for _, option := range v.game.eventOptions {
		opt := option
		choices = append(choices, EventChoice{
			Key:   opt.Key,
			Label: opt.Label,
			OnSelect: func() {
				opt.Action()
			},
		})
	}
	modal := NewEventModal(v.game.EventTitle.Get(), v.game.EventMessage.Get(), choices...)
	modal.OnDismiss(func() {
		v.game.DismissEvent()
		v.Invalidate()
	})
	return modal
}

func copyMarketPrices(src MarketPrices) MarketPrices {
	if src == nil {
		return nil
	}
	dst := make(MarketPrices, len(src))
	for key, value := range src {
		dst[key] = value
	}
	return dst
}

func (v *GameView) handleTradeInput(msg runtime.Message) runtime.HandleResult {
	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		return v.tradeInput.HandleMessage(msg)
	}

	switch key.Key {
	case terminal.KeyEscape:
		v.showTrade = false
		v.tradeInput.Blur()
		v.Invalidate()
		return runtime.Handled()

	case terminal.KeyEnter:
		qty := parseInputAmount(v.tradeInput.Text())
		if qty <= 0 {
			qty = 1
		}

		if v.tradeIsBuy {
			v.game.Buy(v.tradeCandy, qty)
		} else {
			v.game.Sell(v.tradeCandy, qty)
		}

		v.showTrade = false
		v.tradeInput.Blur()
		v.refresh()
		return runtime.Handled()
	}

	// Pass to input
	return v.tradeInput.HandleMessage(msg)
}

func (v *GameView) openBankDialog() {
	if !v.game.bankUnlocked() {
		v.game.Message.Set("Bank access locked. Reach Day 3 or earn $200 profit.")
		v.refresh()
		return
	}
	v.showTrade = false
	v.showLoan = false
	v.showCraft = false
	v.showItems = false
	v.showUpgrades = false
	v.showIntel = false
	v.showGear = false
	v.showBlackMarket = false
	v.showStats = false
	v.showStash = false
	v.tradeInput.Blur()
	v.loanInput.Blur()
	v.craftInput.Blur()
	v.stashInput.Blur()
	v.blackMarketInput.Blur()
	v.bankAction = bankDeposit
	v.showBank = true
	v.bankInput.Clear()
	v.bankInput.Focus()
	v.Invalidate()
}

func (v *GameView) handleBankInput(msg runtime.Message) runtime.HandleResult {
	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		return v.bankInput.HandleMessage(msg)
	}

	switch key.Key {
	case terminal.KeyEscape:
		v.showBank = false
		v.bankInput.Blur()
		v.Invalidate()
		return runtime.Handled()
	case terminal.KeyEnter:
		amount := parseInputAmount(v.bankInput.Text())
		if amount == 0 {
			if v.bankAction == bankDeposit {
				amount = v.game.Cash.Get()
			} else {
				amount = v.game.Bank.Get()
			}
		}
		if v.bankAction == bankDeposit {
			v.game.Deposit(amount)
		} else {
			v.game.Withdraw(amount)
		}
		v.bankInput.Clear()
		v.refresh()
		return runtime.Handled()
	}

	switch key.Rune {
	case 'd', 'D':
		v.bankAction = bankDeposit
		v.Invalidate()
		return runtime.Handled()
	case 'w', 'W':
		v.bankAction = bankWithdraw
		v.Invalidate()
		return runtime.Handled()
	}

	return v.bankInput.HandleMessage(msg)
}

func (v *GameView) openLoanDialog() {
	v.showTrade = false
	v.showBank = false
	v.showCraft = false
	v.showItems = false
	v.showUpgrades = false
	v.showIntel = false
	v.showGear = false
	v.showBlackMarket = false
	v.showStats = false
	v.showStash = false
	v.tradeInput.Blur()
	v.bankInput.Blur()
	v.craftInput.Blur()
	v.stashInput.Blur()
	v.blackMarketInput.Blur()
	v.showLoan = true
	v.loanRepay = false
	v.loanInput.Clear()
	v.loanInput.Blur()
	v.Invalidate()
}

func (v *GameView) handleLoanInput(msg runtime.Message) runtime.HandleResult {
	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		if v.loanRepay {
			return v.loanInput.HandleMessage(msg)
		}
		return runtime.Unhandled()
	}

	if key.Key == terminal.KeyEscape {
		if v.loanRepay {
			v.loanRepay = false
			v.loanInput.Blur()
			v.Invalidate()
			return runtime.Handled()
		}
		v.showLoan = false
		v.loanInput.Blur()
		v.Invalidate()
		return runtime.Handled()
	}

	if v.loanRepay {
		if key.Key == terminal.KeyEnter {
			amount := parseInputAmount(v.loanInput.Text())
			if amount == 0 {
				amount = v.game.Cash.Get()
			}
			v.game.RepayLoan(amount)
			v.loanInput.Clear()
			v.refresh()
			return runtime.Handled()
		}
		return v.loanInput.HandleMessage(msg)
	}

	switch key.Rune {
	case '1', '2', '3':
		tier := int(key.Rune - '1')
		v.game.TakeLoan(tier)
		v.refresh()
		return runtime.Handled()
	case 'r', 'R':
		v.loanRepay = true
		v.loanInput.Clear()
		v.loanInput.Focus()
		v.Invalidate()
		return runtime.Handled()
	}

	return runtime.Handled()
}

func (v *GameView) handleCombatInput(msg runtime.Message) runtime.HandleResult {
	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		return runtime.Handled()
	}
	switch key.Rune {
	case 'a', 'A':
		v.game.CombatAttack()
		v.refresh()
		return runtime.Handled()
	case 'd', 'D':
		v.game.CombatDefend()
		v.refresh()
		return runtime.Handled()
	case 'i', 'I':
		v.openItemsDialog(true)
		return runtime.Handled()
	case 'f', 'F':
		v.game.CombatFlee()
		v.refresh()
		return runtime.Handled()
	}
	return runtime.Handled()
}

func (v *GameView) restartGame() {
	v.showTrade = false
	v.tradeInput.Blur()
	v.tradeInput.Clear()
	v.showBank = false
	v.bankInput.Blur()
	v.bankInput.Clear()
	v.showLoan = false
	v.loanInput.Blur()
	v.loanInput.Clear()
	v.loanRepay = false
	v.showCraft = false
	v.craftInput.Blur()
	v.craftInput.Clear()
	v.showItems = false
	v.showUpgrades = false
	v.showStash = false
	v.showIntel = false
	v.showGear = false
	v.showBlackMarket = false
	v.showStats = false
	v.stashInput.Blur()
	v.stashInput.Clear()
	v.blackMarketInput.Blur()
	v.blackMarketInput.Clear()
	v.game.StartNewRun()
	v.lastPrices = nil
	v.lastHour = 0
	v.lastHeat = 0
	v.lastDay = 0
	v.lastLoc = -1
	v.refresh()
}

func (v *GameView) ChildWidgets() []runtime.Widget {
	widgets := []runtime.Widget{
		v.header,
		v.tabs,
		v.messageLabel,
		v.statsLabel,
		v.inventoryLbl,
		v.scheduleLbl,
		v.statusLbl,
		v.heatGauge,
	}
	if v.toastStack != nil {
		widgets = append(widgets, v.toastStack)
	}
	if v.saveDialog != nil {
		widgets = append(widgets, v.saveDialog)
	}
	if v.loadDialog != nil {
		widgets = append(widgets, v.loadDialog)
	}
	return widgets
}
