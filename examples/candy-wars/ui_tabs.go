package main

import (
	"fmt"

	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/state"
	"github.com/odvcencio/fluffyui/terminal"
	"github.com/odvcencio/fluffyui/widgets"
)

// TradeTabContent renders the trading interface.
type TradeTabContent struct {
	widgets.Component
	game *Game
	view *GameView

	marketTable   *widgets.Table
	priceChart    *widgets.BarChart
	priceChartSig *state.Signal[[]widgets.BarData]
	sparkline     *widgets.Sparkline
	selectedIdx   int

	// Price trend tracking
	lastPrices   MarketPrices
	lastLocation int

	style       backend.Style
	accentStyle backend.Style
}

func NewTradeTabContent(game *Game, view *GameView) *TradeTabContent {
	chartData := state.NewSignal([]widgets.BarData{})
	chartData.SetEqualFunc(func(a, b []widgets.BarData) bool {
		if len(a) != len(b) {
			return false
		}
		for i := range a {
			if a[i].Label != b[i].Label || a[i].Value != b[i].Value {
				return false
			}
		}
		return true
	})

	t := &TradeTabContent{
		game:          game,
		view:          view,
		style:         backend.DefaultStyle(),
		accentStyle:   backend.DefaultStyle().Bold(true),
		priceChartSig: chartData,
		lastLocation:  -1,
	}

	t.marketTable = widgets.NewTable(
		widgets.TableColumn{Title: "Candy", Width: 14},
		widgets.TableColumn{Title: "Price", Width: 7},
		widgets.TableColumn{Title: "Stock", Width: 5},
		widgets.TableColumn{Title: "Yours", Width: 5},
		widgets.TableColumn{Title: "Trend", Width: 7},
	)

	t.priceChart = widgets.NewBarChart(chartData)
	t.sparkline = widgets.NewSparkline(game.PriceHistory)

	return t
}

func (t *TradeTabContent) UpdateMarketTable() {
	prices := t.game.Prices.Get()
	inv := t.game.Inventory.Get()
	currentLoc := t.game.Location.Get()

	// Reset price tracking when location changes
	if currentLoc != t.lastLocation {
		t.lastPrices = nil
		t.lastLocation = currentLoc
	}

	rows := make([][]string, len(CandyTypes))
	for i, candy := range CandyTypes {
		price := prices[candy.Name]
		stock := t.game.availableStock(candy.Name)
		owned := inv[candy.Name]

		// Calculate trend based on previous price
		trend := "  ---"
		if t.lastPrices != nil {
			lastPrice := t.lastPrices[candy.Name]
			if lastPrice > 0 {
				diff := price - lastPrice
				if diff > 0 {
					trend = fmt.Sprintf(" ↑+%d", diff)
				} else if diff < 0 {
					trend = fmt.Sprintf(" ↓%d", diff)
				} else {
					trend = "  →0"
				}
			}
		}

		rows[i] = []string{
			candy.Emoji + " " + candy.Name,
			fmt.Sprintf("$%d", price),
			fmt.Sprintf("%d", stock),
			fmt.Sprintf("%d", owned),
			trend,
		}
	}
	t.marketTable.SetRows(rows)

	// Store current prices for next comparison
	t.lastPrices = make(MarketPrices)
	for k, v := range prices {
		t.lastPrices[k] = v
	}
}

func (t *TradeTabContent) UpdatePriceChart() {
	if t.selectedIdx < 0 || t.selectedIdx >= len(CandyTypes) {
		return
	}
	candy := CandyTypes[t.selectedIdx]
	bars := make([]widgets.BarData, 0, len(Locations))

	for i, loc := range Locations {
		price := 0
		known := t.game.hasInformant || i == t.game.Location.Get()
		if known {
			prices := t.game.PricesForLocation(i)
			price = prices[candy.Name]
		}

		label := loc.Name
		if len(label) > 8 {
			label = label[:8]
		}
		if i == t.game.Location.Get() {
			label += "*"
		}

		bars = append(bars, widgets.BarData{Label: label, Value: float64(price)})
	}
	t.priceChartSig.Set(bars)
}

func (t *TradeTabContent) Measure(c runtime.Constraints) runtime.Size {
	return c.MaxSize()
}

func (t *TradeTabContent) Layout(bounds runtime.Rect) {
	t.Component.Layout(bounds)

	leftWidth := bounds.Width * 55 / 100
	rightWidth := bounds.Width - leftWidth - 1
	tableHeight := len(CandyTypes) + 2

	t.marketTable.Layout(runtime.Rect{X: bounds.X, Y: bounds.Y, Width: leftWidth, Height: tableHeight})
	t.priceChart.Layout(runtime.Rect{X: bounds.X + leftWidth + 1, Y: bounds.Y, Width: rightWidth, Height: tableHeight})

	sparkY := bounds.Y + tableHeight + 1
	t.sparkline.Layout(runtime.Rect{X: bounds.X, Y: sparkY, Width: bounds.Width, Height: 1})
}

func (t *TradeTabContent) Render(ctx runtime.RenderContext) {
	t.UpdateMarketTable()
	t.UpdatePriceChart()

	t.marketTable.Render(ctx)
	t.priceChart.Render(ctx)

	sparkBounds := t.sparkline.Bounds()
	ctx.Buffer.SetString(sparkBounds.X, sparkBounds.Y, "Net Worth: ", t.style)
	t.sparkline.Render(ctx)
}

func (t *TradeTabContent) HandleMessage(msg runtime.Message) runtime.HandleResult {
	t.marketTable.Focus()
	result := t.marketTable.HandleMessage(msg)
	t.selectedIdx = t.marketTable.SelectedIndex()
	return result
}

func (t *TradeTabContent) ChildWidgets() []runtime.Widget {
	children := make([]runtime.Widget, 0, 3)
	if t.marketTable != nil {
		children = append(children, t.marketTable)
	}
	if t.priceChart != nil {
		children = append(children, t.priceChart)
	}
	if t.sparkline != nil {
		children = append(children, t.sparkline)
	}
	return children
}

// InventoryTabContent shows player inventory and equipment.
type InventoryTabContent struct {
	widgets.Component
	game *Game
	view *GameView

	candyTable *widgets.Table
	itemsTable *widgets.Table
	focusLeft  bool

	style       backend.Style
	dimStyle    backend.Style
	accentStyle backend.Style
}

func NewInventoryTabContent(game *Game, view *GameView) *InventoryTabContent {
	i := &InventoryTabContent{
		game:        game,
		view:        view,
		style:       backend.DefaultStyle(),
		dimStyle:    backend.DefaultStyle().Dim(true),
		accentStyle: backend.DefaultStyle().Bold(true),
		focusLeft:   true,
	}

	i.candyTable = widgets.NewTable(
		widgets.TableColumn{Title: "Candy", Width: 16},
		widgets.TableColumn{Title: "Qty", Width: 5},
		widgets.TableColumn{Title: "Value", Width: 8},
	)

	i.itemsTable = widgets.NewTable(
		widgets.TableColumn{Title: "Item", Width: 18},
		widgets.TableColumn{Title: "Qty", Width: 4},
		widgets.TableColumn{Title: "Use", Width: 10},
	)

	return i
}

func (i *InventoryTabContent) updateTables() {
	// Update candy table
	inv := i.game.Inventory.Get()
	prices := i.game.Prices.Get()

	candyRows := make([][]string, 0, len(CandyTypes))
	for _, candy := range CandyTypes {
		qty := inv[candy.Name]
		if qty > 0 {
			price := prices[candy.Name]
			value := qty * price
			candyRows = append(candyRows, []string{
				candy.Emoji + " " + candy.Name,
				fmt.Sprintf("%d", qty),
				fmt.Sprintf("$%d", value),
			})
		}
	}
	if len(candyRows) == 0 {
		candyRows = [][]string{{"(empty)", "-", "-"}}
	}
	i.candyTable.SetRows(candyRows)

	// Update items table (crafted items and black market items)
	itemRows := make([][]string, 0)
	for _, item := range CraftedItems {
		qty := inv[item.Name]
		if qty > 0 {
			useContext := "Combat"
			if item.UseOutOfCombat && !item.UseInCombat {
				useContext = "Out"
			} else if item.UseOutOfCombat && item.UseInCombat {
				useContext = "Any"
			}
			itemRows = append(itemRows, []string{item.Name, fmt.Sprintf("%d", qty), useContext})
		}
	}
	for _, item := range BlackMarketItems {
		qty := inv[item.Name]
		if qty > 0 {
			useContext := "Combat"
			if item.UseOutOfCombat && !item.UseInCombat {
				useContext = "Out"
			} else if item.UseOutOfCombat && item.UseInCombat {
				useContext = "Any"
			}
			itemRows = append(itemRows, []string{item.Name, fmt.Sprintf("%d", qty), useContext})
		}
	}
	if len(itemRows) == 0 {
		itemRows = [][]string{{"(no items)", "-", "-"}}
	}
	i.itemsTable.SetRows(itemRows)
}

func (i *InventoryTabContent) Measure(c runtime.Constraints) runtime.Size {
	return c.MaxSize()
}

func (i *InventoryTabContent) Layout(bounds runtime.Rect) {
	i.Component.Layout(bounds)

	// Left side: candy inventory
	leftWidth := bounds.Width / 2
	candyHeight := len(CandyTypes) + 3
	i.candyTable.Layout(runtime.Rect{X: bounds.X, Y: bounds.Y + 2, Width: leftWidth - 1, Height: candyHeight})

	// Right side: items
	rightX := bounds.X + leftWidth
	rightWidth := bounds.Width - leftWidth
	itemsHeight := len(CraftedItems) + len(BlackMarketItems) + 3
	i.itemsTable.Layout(runtime.Rect{X: rightX, Y: bounds.Y + 2, Width: rightWidth, Height: itemsHeight})
}

func (i *InventoryTabContent) Render(ctx runtime.RenderContext) {
	i.updateTables()
	bounds := i.Bounds()

	leftWidth := bounds.Width / 2
	rightX := bounds.X + leftWidth

	// Section headers
	ctx.Buffer.SetString(bounds.X, bounds.Y, "CANDIES", i.accentStyle)
	ctx.Buffer.SetString(rightX, bounds.Y, "ITEMS", i.accentStyle)

	// Capacity info
	capacityLine := fmt.Sprintf("Backpack: %d/%d", i.game.InventoryCount(), i.game.Capacity)
	ctx.Buffer.SetString(bounds.X, bounds.Y+1, capacityLine, i.dimStyle)

	// Stash info if available
	if i.game.hasStash {
		stashLine := fmt.Sprintf("Stash: %d/%d", i.game.StashCount(), i.game.stashCapacity)
		ctx.Buffer.SetString(rightX, bounds.Y+1, stashLine, i.dimStyle)
	}

	// Render tables
	i.candyTable.Render(ctx)
	i.itemsTable.Render(ctx)

	// Equipment section at bottom
	equipY := bounds.Y + len(CandyTypes) + 6
	ctx.Buffer.SetString(bounds.X, equipY, "EQUIPMENT", i.accentStyle)

	weapon := i.game.playerWeapon()
	armor := i.game.playerArmor()

	weaponLine := fmt.Sprintf("Weapon: %s (ATK +%d)", weapon.Name, weapon.AtkBonus)
	if weapon.StunChance > 0 {
		weaponLine += fmt.Sprintf(" [%d%% stun]", weapon.StunChance)
	}
	ctx.Buffer.SetString(bounds.X, equipY+1, weaponLine, i.style)

	armorLine := fmt.Sprintf("Armor:  %s (DEF +%d", armor.Name, armor.DefBonus)
	if armor.SPDMod != 0 {
		armorLine += fmt.Sprintf(", SPD %+d", armor.SPDMod)
	}
	armorLine += ")"
	ctx.Buffer.SetString(bounds.X, equipY+2, armorLine, i.style)

	// Stats summary
	statsLine := fmt.Sprintf("Stats: ATK %d | DEF %d | SPD %d | HP %d/%d",
		i.game.playerATK(), i.game.playerDEF(), i.game.playerSPD(),
		i.game.HP.Get(), maxHP)
	ctx.Buffer.SetString(bounds.X, equipY+4, statsLine, i.dimStyle)

	// Trade buff indicator
	if i.game.tradeBuffUses > 0 {
		buffLine := fmt.Sprintf("Trade Buff Active: %d uses (Buy -10%%, Sell +10%%)", i.game.tradeBuffUses)
		ctx.Buffer.SetString(bounds.X, equipY+5, buffLine, i.accentStyle.Foreground(backend.ColorGreen))
	}
}

func (i *InventoryTabContent) HandleMessage(msg runtime.Message) runtime.HandleResult {
	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		return runtime.Unhandled()
	}

	switch key.Key {
	case terminal.KeyLeft:
		i.focusLeft = true
		i.candyTable.Focus()
		i.itemsTable.Blur()
		return runtime.Handled()
	case terminal.KeyRight:
		i.focusLeft = false
		i.itemsTable.Focus()
		i.candyTable.Blur()
		return runtime.Handled()
	}

	if i.focusLeft {
		i.candyTable.Focus()
		return i.candyTable.HandleMessage(msg)
	}
	i.itemsTable.Focus()
	return i.itemsTable.HandleMessage(msg)
}

// StatsTabContent shows player stats, achievements, skill trees.
type StatsTabContent struct {
	widgets.Component
	game *Game
	view *GameView

	skillTree    *SkillTreeView
	achievements *widgets.Table

	style backend.Style
}

func NewStatsTabContent(game *Game, view *GameView) *StatsTabContent {
	s := &StatsTabContent{
		game:  game,
		view:  view,
		style: backend.DefaultStyle(),
	}

	s.skillTree = NewSkillTreeView(game)
	s.achievements = widgets.NewTable(
		widgets.TableColumn{Title: "Achievement", Width: 20},
		widgets.TableColumn{Title: "Progress", Width: 12},
	)

	return s
}

func (s *StatsTabContent) Measure(c runtime.Constraints) runtime.Size {
	return c.MaxSize()
}

func (s *StatsTabContent) Layout(bounds runtime.Rect) {
	s.Component.Layout(bounds)

	leftWidth := bounds.Width / 2
	leftRect := runtime.Rect{X: bounds.X, Y: bounds.Y, Width: leftWidth - 1, Height: bounds.Height}
	rightRect := runtime.Rect{X: bounds.X + leftWidth, Y: bounds.Y, Width: bounds.Width - leftWidth, Height: bounds.Height}

	s.skillTree.Layout(leftRect)
	s.achievements.Layout(rightRect)
}

func (s *StatsTabContent) Render(ctx runtime.RenderContext) {
	bounds := s.Bounds()
	if bounds.Width <= 0 || bounds.Height <= 0 {
		return
	}

	progress := s.game.GetAchievementProgress()
	rows := make([][]string, len(Achievements))
	for i, ach := range Achievements {
		label := ach.Name
		if ach.Hidden {
			label = "???"
		}
		status := "Locked"
		if i < len(progress) {
			if progress[i].Unlocked {
				status = "Unlocked"
			} else if ach.ProgressMax > 0 {
				status = fmt.Sprintf("%d/%d", progress[i].Progress, ach.ProgressMax)
			}
		}
		rows[i] = []string{label, status}
	}
	s.achievements.SetRows(rows)

	s.skillTree.Render(ctx)
	s.achievements.Render(ctx)
}

func (s *StatsTabContent) HandleMessage(msg runtime.Message) runtime.HandleResult {
	s.skillTree.Focus()
	if result := s.skillTree.HandleMessage(msg); result.Handled {
		return result
	}
	s.achievements.Focus()
	return s.achievements.HandleMessage(msg)
}

func (s *StatsTabContent) ChildWidgets() []runtime.Widget {
	children := make([]runtime.Widget, 0, 2)
	if s.skillTree != nil {
		children = append(children, s.skillTree)
	}
	if s.achievements != nil {
		children = append(children, s.achievements)
	}
	return children
}

// MapTabContent shows location map with travel options.
type MapTabContent struct {
	widgets.Component
	game *Game
	view *GameView

	locationList *widgets.List[Location]

	style backend.Style
}

func NewMapTabContent(game *Game, view *GameView) *MapTabContent {
	m := &MapTabContent{
		game:  game,
		view:  view,
		style: backend.DefaultStyle(),
	}

	adapter := widgets.NewSliceAdapter(Locations, func(loc Location, index int, selected bool, ctx runtime.RenderContext) {
		style := m.style
		if selected {
			style = style.Reverse(true)
		}
		line := fmt.Sprintf("%-12s Risk: %d", loc.Name, loc.RiskLevel)
		ctx.Buffer.SetString(ctx.Bounds.X, ctx.Bounds.Y, truncPad(line, ctx.Bounds.Width), style)
	})
	m.locationList = widgets.NewList(adapter)
	m.locationList.SetOnSelect(func(index int, loc Location) {
		m.game.Travel(index)
	})

	return m
}

func (m *MapTabContent) Measure(c runtime.Constraints) runtime.Size {
	return c.MaxSize()
}

func (m *MapTabContent) Layout(bounds runtime.Rect) {
	m.Component.Layout(bounds)
	m.locationList.Layout(runtime.Rect{X: bounds.X, Y: bounds.Y, Width: bounds.Width, Height: len(Locations) + 1})
}

func (m *MapTabContent) Render(ctx runtime.RenderContext) {
	m.locationList.Render(ctx)
}

func (m *MapTabContent) HandleMessage(msg runtime.Message) runtime.HandleResult {
	m.locationList.Focus()
	return m.locationList.HandleMessage(msg)
}

func (m *MapTabContent) ChildWidgets() []runtime.Widget {
	if m.locationList == nil {
		return nil
	}
	return []runtime.Widget{m.locationList}
}
