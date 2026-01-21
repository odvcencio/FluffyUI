package main

import (
	"fmt"

	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/runtime"
	"github.com/odvcencio/fluffy-ui/state"
	"github.com/odvcencio/fluffy-ui/widgets"
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

	rows := make([][]string, len(CandyTypes))
	for i, candy := range CandyTypes {
		price := prices[candy.Name]
		stock := t.game.availableStock(candy.Name)
		owned := inv[candy.Name]
		trend := "---" // TODO: price history

		rows[i] = []string{
			candy.Emoji + " " + candy.Name,
			fmt.Sprintf("$%d", price),
			fmt.Sprintf("%d", stock),
			fmt.Sprintf("%d", owned),
			trend,
		}
	}
	t.marketTable.SetRows(rows)
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

// InventoryTabContent shows player inventory and equipment.
type InventoryTabContent struct {
	widgets.Component
	game *Game
	view *GameView

	candyList    *widgets.List[string]
	equipPanel   *widgets.Panel
	craftSection *widgets.Section

	style backend.Style
}

func NewInventoryTabContent(game *Game, view *GameView) *InventoryTabContent {
	i := &InventoryTabContent{
		game:  game,
		view:  view,
		style: backend.DefaultStyle(),
	}
	return i
}

func (i *InventoryTabContent) Measure(c runtime.Constraints) runtime.Size {
	return c.MaxSize()
}

func (i *InventoryTabContent) Layout(bounds runtime.Rect) {
	i.Component.Layout(bounds)
}

func (i *InventoryTabContent) Render(ctx runtime.RenderContext) {
	bounds := i.Bounds()
	ctx.Buffer.SetString(bounds.X, bounds.Y, "Inventory Tab - TODO", i.style)
}

func (i *InventoryTabContent) HandleMessage(msg runtime.Message) runtime.HandleResult {
	return runtime.Unhandled()
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
	m.locationList.OnSelect(func(index int, loc Location) {
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
