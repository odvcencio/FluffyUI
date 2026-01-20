package main

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/odvcencio/fluffy-ui/state"
)

// Game State

type Game struct {
	Cash         *state.Signal[int]
	Debt         *state.Signal[int]
	Bank         *state.Signal[int]
	BankLimit    int
	Day          *state.Signal[int]
	Hour         *state.Signal[int]
	MaxDays      int
	Location     *state.Signal[int]
	Inventory    *state.Signal[Inventory]
	Capacity     int
	Prices       *state.Signal[MarketPrices]
	PriceHistory *state.Signal[[]float64]
	Message      *state.Signal[string]
	GameOver     *state.Signal[bool]
	GameOverMsg  *state.Signal[string]
	ShowEvent    *state.Signal[bool]
	EventTitle   *state.Signal[string]
	EventMessage *state.Signal[string]
	Heat         *state.Signal[int] // 0-100, teacher suspicion level
	HP           *state.Signal[int]

	Markets            map[int]MarketPrices
	Stocks             map[int]MarketStock
	Shortages          map[int]map[string]int
	SellBonuses        map[int]int
	BuyBonuses         map[int]int
	StockBonuses       map[int]int
	loiterTicks        int
	heatImmunityDays   int
	Schedule           [4]int
	Loans              []Loan
	goonDue            bool
	Combat             *CombatState
	eventOptions       []EventOption
	tradeBuffUses      int
	tradeBuffBuyMult   int
	tradeBuffSellMult  int
	workoutCount       int
	thickSkinCount     int
	trackPracticeCount int
	hasMuscle          bool
	hasIntimidation    bool
	backpackTier       int
	hasStash           bool
	stash              Inventory
	stashCapacity      int
	hasBike            bool
	hasInformant       bool
	bankExpanded       bool
	ownedWeapons       map[string]bool
	ownedArmors        map[string]bool
	equippedWeapon     string
	equippedArmor      string

	// Session stats (long-lived across runs).
	Runs      int
	Wins      int
	Losses    int
	BestWorth int

	meta     *MetaProgress
	runStart time.Time
}

func NewGame() *Game {
	g := &Game{
		Cash:         state.NewSignal(startingCash),
		Debt:         state.NewSignal(startingDebt),
		Bank:         state.NewSignal(0),
		BankLimit:    bankLimitInitial,
		Day:          state.NewSignal(1),
		Hour:         state.NewSignal(1),
		MaxDays:      maxDays,
		Location:     state.NewSignal(0),
		Inventory:    state.NewSignal(make(Inventory)),
		Capacity:     startingCapacity,
		Prices:       state.NewSignal(make(MarketPrices)),
		PriceHistory: state.NewSignal([]float64{float64(startingCash)}),
		Message:      state.NewSignal(""),
		GameOver:     state.NewSignal(false),
		GameOverMsg:  state.NewSignal(""),
		ShowEvent:    state.NewSignal(false),
		EventTitle:   state.NewSignal(""),
		EventMessage: state.NewSignal(""),
		Heat:         state.NewSignal(0),
		HP:           state.NewSignal(startingHP),
		Markets:      make(map[int]MarketPrices),
		Stocks:       make(map[int]MarketStock),
		Shortages:    make(map[int]map[string]int),
		SellBonuses:  make(map[int]int),
	}

	g.Cash.SetEqualFunc(state.EqualComparable[int])
	g.Debt.SetEqualFunc(state.EqualComparable[int])
	g.Bank.SetEqualFunc(state.EqualComparable[int])
	g.Day.SetEqualFunc(state.EqualComparable[int])
	g.Hour.SetEqualFunc(state.EqualComparable[int])
	g.Location.SetEqualFunc(state.EqualComparable[int])
	g.Message.SetEqualFunc(state.EqualComparable[string])
	g.GameOver.SetEqualFunc(state.EqualComparable[bool])
	g.GameOverMsg.SetEqualFunc(state.EqualComparable[string])
	g.ShowEvent.SetEqualFunc(state.EqualComparable[bool])
	g.EventTitle.SetEqualFunc(state.EqualComparable[string])
	g.EventMessage.SetEqualFunc(state.EqualComparable[string])
	g.Heat.SetEqualFunc(state.EqualComparable[int])
	g.HP.SetEqualFunc(state.EqualComparable[int])

	g.meta = LoadMeta()
	g.applyMeta()
	g.StartNewRun()
	return g
}

func (g *Game) StartNewRun() {
	g.Runs++
	if g.meta != nil {
		g.meta.TotalRuns = g.Runs
		g.saveMeta()
	}
	g.runStart = time.Now()
	g.GameOver.Set(false)
	g.GameOverMsg.Set("")
	g.ShowEvent.Set(false)
	g.EventTitle.Set("")
	g.EventMessage.Set("")
	g.Cash.Set(startingCash)
	g.Debt.Set(startingDebt)
	g.Bank.Set(0)
	g.BankLimit = bankLimitInitial
	g.Day.Set(1)
	g.Hour.Set(1)
	g.Location.Set(0)
	g.Inventory.Set(make(Inventory))
	g.Heat.Set(0)
	g.HP.Set(startingHP)
	g.heatImmunityDays = 0
	g.loiterTicks = 0
	g.Schedule = g.generateSchedule()
	g.Loans = nil
	g.goonDue = false
	g.Combat = nil
	g.eventOptions = nil
	g.tradeBuffUses = 0
	g.tradeBuffBuyMult = 100
	g.tradeBuffSellMult = 100
	g.workoutCount = 0
	g.thickSkinCount = 0
	g.trackPracticeCount = 0
	g.hasMuscle = false
	g.hasIntimidation = false
	g.backpackTier = 0
	g.hasStash = false
	g.stash = make(Inventory)
	g.stashCapacity = 0
	g.hasBike = false
	g.hasInformant = false
	g.bankExpanded = false
	g.ownedWeapons = map[string]bool{"Fists": true}
	g.ownedArmors = map[string]bool{"Hoodie": true}
	g.equippedWeapon = "Fists"
	g.equippedArmor = "Hoodie"
	g.Shortages = make(map[int]map[string]int)
	g.Message.Set("Welcome to Jefferson Middle School! Trade candy to pay off your debt.")
	g.RollDailyMarkets()
	g.RollDailyStocks()
	g.ResetDailyBonuses()
	g.RefreshPrices()
	g.PriceHistory.Set([]float64{float64(startingCash)})
}

func (g *Game) rollBasePrices() MarketPrices {
	prices := make(MarketPrices)
	for _, candy := range CandyTypes {
		prices[candy.Name] = candy.MinPrice + rand.Intn(candy.MaxPrice-candy.MinPrice+1)
	}
	return prices
}

func (g *Game) rollBaseStocks(loc Location) MarketStock {
	stocks := make(MarketStock)
	for _, candy := range CandyTypes {
		base := rand.Intn(6) + 6 // 6-11
		qty := base + (loc.RiskLevel - 3)
		if qty < 2 {
			qty = 2
		}
		stocks[candy.Name] = qty
	}
	return stocks
}

func (g *Game) RollDailyMarkets() {
	g.Markets = make(map[int]MarketPrices, len(Locations))
	for i := range Locations {
		g.Markets[i] = g.rollBasePrices()
	}
}

func (g *Game) RollDailyStocks() {
	g.Stocks = make(map[int]MarketStock, len(Locations))
	for i, loc := range Locations {
		g.Stocks[i] = g.rollBaseStocks(loc)
	}
}

func (g *Game) ResetDailyBonuses() {
	g.SellBonuses = make(map[int]int, len(Locations))
	g.BuyBonuses = make(map[int]int, len(Locations))
	g.StockBonuses = make(map[int]int, len(Locations))
	for i := range Locations {
		g.SellBonuses[i] = 100
		g.BuyBonuses[i] = 100
		g.StockBonuses[i] = 100
	}
}

func (g *Game) generateSchedule() [4]int {
	periods := []int{locationLibrary, locationGymnasium, locationArtRoom, locationMusicHall}
	rand.Shuffle(len(periods), func(i, j int) {
		periods[i], periods[j] = periods[j], periods[i]
	})
	var schedule [4]int
	copy(schedule[:], periods)
	return schedule
}

func (g *Game) currentPeriod() int {
	hour := g.Hour.Get()
	if hour <= 0 {
		return 0
	}
	period := (hour - 1) / 2
	if period < 0 {
		return 0
	}
	if period > 3 {
		return 3
	}
	return period
}

func periodHourRange(period int) (int, int) {
	if period < 0 {
		period = 0
	}
	if period > 3 {
		period = 3
	}
	start := period*2 + 1
	return start, start + 1
}

func (g *Game) scheduleStatus() ScheduleStatus {
	return g.scheduleStatusAt(g.Location.Get())
}

func (g *Game) scheduleStatusAt(loc int) ScheduleStatus {
	period := g.currentPeriod()
	scheduled := g.Schedule[period]
	if loc == scheduled {
		return statusBlendingIn
	}
	if loc == locationCafeteria || loc == locationPlayground {
		return statusOffCampus
	}
	return statusWrongClass
}

func (g *Game) scheduleHeatMultiplier() int {
	switch g.scheduleStatus() {
	case statusBlendingIn:
		return 50
	case statusOffCampus:
		return 200
	default:
		return 100
	}
}

func (g *Game) schedulePriceMultipliers() (buyMult, sellMult int) {
	return g.schedulePriceMultipliersAt(g.Location.Get())
}

func (g *Game) schedulePriceMultipliersAt(loc int) (buyMult, sellMult int) {
	if g.scheduleStatusAt(loc) == statusOffCampus {
		return 90, 105
	}
	return 100, 100
}

func (g *Game) currentStockMultiplier(loc int) float64 {
	multiplier := float64(timeOfDayStockMultiplier(g.Hour.Get())) / 100.0
	if g.scheduleStatusAt(loc) == statusOffCampus {
		multiplier *= 1.5
	}
	bonus := g.StockBonuses[loc]
	if bonus <= 0 {
		bonus = 100
	}
	multiplier *= float64(bonus) / 100.0
	return multiplier
}

func (g *Game) availableStock(candyName string) int {
	loc := g.Location.Get()
	if g.isShortage(loc, candyName) {
		return 0
	}
	stock := g.Stocks[loc]
	if stock == nil {
		return 0
	}
	base := stock[candyName]
	multiplier := g.currentStockMultiplier(loc)
	available := int(math.Floor(float64(base) * multiplier))
	if available < 0 {
		return 0
	}
	return available
}

func (g *Game) consumeStock(candyName string, qty int) {
	if qty <= 0 {
		return
	}
	loc := g.Location.Get()
	stock := g.Stocks[loc]
	if stock == nil {
		return
	}
	base := stock[candyName]
	multiplier := g.currentStockMultiplier(loc)
	if multiplier >= 1.0 {
		reduction := int(math.Ceil(float64(qty) / multiplier))
		base -= reduction
	} else {
		base -= qty
	}
	if base < 0 {
		base = 0
	}
	stock[candyName] = base
}

func (g *Game) addShortage(loc int, candyName string, days int) {
	if days <= 0 {
		return
	}
	if g.Shortages == nil {
		g.Shortages = make(map[int]map[string]int)
	}
	if g.Shortages[loc] == nil {
		g.Shortages[loc] = make(map[string]int)
	}
	if g.Shortages[loc][candyName] < days {
		g.Shortages[loc][candyName] = days
	}
}

func (g *Game) isShortage(loc int, candyName string) bool {
	if g.Shortages == nil {
		return false
	}
	return g.Shortages[loc][candyName] > 0
}

func (g *Game) advanceShortages() {
	if g.Shortages == nil {
		return
	}
	for loc, candies := range g.Shortages {
		for name, days := range candies {
			days--
			if days <= 0 {
				delete(candies, name)
			} else {
				candies[name] = days
			}
		}
		if len(candies) == 0 {
			delete(g.Shortages, loc)
		}
	}
}

func (g *Game) RefreshPrices() {
	loc := g.Location.Get()
	base, ok := g.Markets[loc]
	if !ok {
		base = g.rollBasePrices()
		g.Markets[loc] = base
	}
	multiplier := timeOfDayPriceMultiplier(g.Hour.Get())
	prices := make(MarketPrices)
	for name, price := range base {
		prices[name] = applyPercent(price, multiplier)
	}
	g.Prices.Set(prices)
}

func (g *Game) PricesForLocation(loc int) MarketPrices {
	base, ok := g.Markets[loc]
	if !ok {
		base = g.rollBasePrices()
		g.Markets[loc] = base
	}
	multiplier := timeOfDayPriceMultiplier(g.Hour.Get())
	prices := make(MarketPrices)
	for name, price := range base {
		prices[name] = applyPercent(price, multiplier)
	}
	return prices
}

func (g *Game) canSpendHours(cost int) bool {
	current := g.Hour.Get()
	return current+cost-1 <= hoursPerDay
}

func (g *Game) InCombat() bool {
	return g.Combat != nil
}

func (g *Game) currentTravelHours() int {
	if g.hasBike {
		return 1
	}
	return travelHours
}

func (g *Game) advanceHours(cost int) bool {
	current := g.Hour.Get()
	next := current + cost
	dayRolled := false
	if next > hoursPerDay {
		next -= hoursPerDay
		dayRolled = true
	}
	g.Hour.Set(next)
	return dayRolled
}

func (g *Game) StartNewDay() {
	g.Day.Update(func(d int) int { return d + 1 })
	g.applyDebtInterest()
	g.applyBankInterest()
	g.applyLoanInterest()
	g.RollDailyMarkets()
	g.RollDailyStocks()
	g.ResetDailyBonuses()
	if g.heatImmunityDays > 0 {
		g.heatImmunityDays--
	}
	g.advanceShortages()
	g.reduceHeat(dayHeatDecay)
	g.resetLoiter()
	g.updateUnlocks()
}

func (g *Game) EndDayEarly() {
	if g.GameOver.Get() {
		return
	}
	if g.InCombat() {
		return
	}
	g.Hour.Set(1)
	g.StartNewDay()
	g.reduceHeat(endDayHeatReduction)
	healed := g.healPlayer(10)
	g.RefreshPrices()
	if healed > 0 {
		g.Message.Set(fmt.Sprintf("You called it a day early and rested. +%d HP.", healed))
	} else {
		g.Message.Set("You called it a day early and headed home.")
	}
	g.CheckEndConditions()
}

func (g *Game) resetLoiter() {
	g.loiterTicks = 0
}

func (g *Game) addHeat(amount int, applySchedule bool, ignoreImmunity bool) {
	if amount <= 0 {
		return
	}
	if g.heatImmunityDays > 0 && !ignoreImmunity {
		return
	}
	if applySchedule {
		amount = amount * g.scheduleHeatMultiplier() / 100
		if amount <= 0 {
			return
		}
	}
	g.Heat.Update(func(h int) int {
		newHeat := h + amount
		if newHeat > 100 {
			newHeat = 100
		}
		return newHeat
	})
}

func (g *Game) reduceHeat(amount int) {
	if amount <= 0 {
		return
	}
	g.Heat.Update(func(h int) int {
		newHeat := h - amount
		if newHeat < 0 {
			newHeat = 0
		}
		return newHeat
	})
}

func (g *Game) healPlayer(amount int) int {
	if amount <= 0 {
		return 0
	}
	current := g.HP.Get()
	if current >= maxHP {
		return 0
	}
	newHP := current + amount
	if newHP > maxHP {
		newHP = maxHP
	}
	g.HP.Set(newHP)
	return newHP - current
}

func (g *Game) applyDebtInterest() {
	debt := g.Debt.Get()
	if debt <= 0 {
		return
	}
	interest := debt * debtInterestRatePercent / 100
	if interest <= 0 {
		return
	}
	g.Debt.Set(debt + interest)
}

func (g *Game) applyBankInterest() {
	balance := g.Bank.Get()
	if balance <= 0 {
		return
	}
	interest := balance * bankInterestPercent / 100
	if interest <= 0 {
		return
	}
	newBalance := balance + interest
	if newBalance > g.BankLimit {
		newBalance = g.BankLimit
	}
	g.Bank.Set(newBalance)
}

func (g *Game) applyLoanInterest() {
	if len(g.Loans) == 0 {
		g.goonDue = false
		return
	}
	goonDue := false
	for i := range g.Loans {
		loan := g.Loans[i]
		tier := LoanTiers[loan.Tier]
		interest := loan.Balance * tier.InterestPercent / 100
		if interest > 0 {
			loan.Balance += interest
		}
		if tier.HeatPenalty > 0 {
			g.addHeat(tier.HeatPenalty, false, true)
		}
		if tier.ItemLoss > 0 {
			g.loseRandomCandy(tier.ItemLoss)
		}
		if tier.GoonFight {
			goonDue = true
		}
		g.Loans[i] = loan
	}
	g.goonDue = goonDue
}

func (g *Game) Tick() {
	if g.InCombat() {
		return
	}
	// Fluctuate prices slightly for the current location.
	loc := g.Location.Get()
	base := g.Markets[loc]
	if base == nil {
		base = g.rollBasePrices()
		g.Markets[loc] = base
	}
	newBase := make(MarketPrices)
	for name, price := range base {
		candy := candyByName(name)
		deltaRange := 2
		switch loc {
		case locationCafeteria:
			deltaRange = 3
		case locationMusicHall:
			deltaRange = 1
		}
		delta := rand.Intn(deltaRange*2+1) - deltaRange
		newPrice := price + delta
		if newPrice < candy.MinPrice {
			newPrice = candy.MinPrice
		}
		if newPrice > candy.MaxPrice {
			newPrice = candy.MaxPrice
		}
		newBase[name] = newPrice
	}
	g.Markets[loc] = newBase
	g.RefreshPrices()

	// Update price history (track total inventory value).
	history := g.PriceHistory.Get()
	totalValue := float64(g.Cash.Get())
	inv := g.Inventory.Get()
	prices := g.Prices.Get()
	for name, qty := range inv {
		if price, ok := prices[name]; ok {
			totalValue += float64(qty * price)
		}
	}
	if g.hasStash {
		for name, qty := range g.stash {
			if price, ok := prices[name]; ok {
				totalValue += float64(qty * price)
			}
		}
	}
	history = append(history, totalValue)
	if len(history) > 30 {
		history = history[len(history)-30:]
	}
	g.PriceHistory.Set(history)

	if loc == locationCafeteria {
		g.loiterTicks += 2
	} else {
		g.loiterTicks++
	}
	if g.loiterTicks%ticksPerLoiterCheck == 0 {
		g.reduceHeat(1)
		if !g.ShowEvent.Get() {
			g.CheckLoiterEvent()
		}
	}
}

func (g *Game) Travel(locationIndex int) {
	if g.GameOver.Get() {
		return
	}
	if g.InCombat() {
		return
	}
	if locationIndex < 0 || locationIndex >= len(Locations) {
		return
	}

	travelCost := g.currentTravelHours()
	if !g.canSpendHours(travelCost) {
		g.Message.Set("Too late to travel. End the day early.")
		return
	}

	dayRolled := g.advanceHours(travelCost)
	g.Location.Set(locationIndex)
	g.resetLoiter()
	if dayRolled {
		g.StartNewDay()
	}
	g.RefreshPrices()
	g.addHeat(1, true, false)

	loc := Locations[locationIndex]
	g.Message.Set(fmt.Sprintf("You arrived at the %s. %s", loc.Name, loc.Description))

	if g.maybeTriggerGoonFight() {
		return
	}

	// Check for random events
	g.CheckTravelEvent()

	// Check win/lose conditions
	g.CheckEndConditions()
}
