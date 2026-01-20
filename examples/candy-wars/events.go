package main

import (
	"fmt"
	"math/rand"
)

type eventCategory int

const (
	eventMarket eventCategory = iota
	eventLucky
	eventDanger
)

func (g *Game) CheckTravelEvent() {
	if g.ShowEvent.Get() {
		return
	}
	if !g.rollEventChance(g.travelEventChance()) {
		return
	}
	g.triggerEventCategory(true)
}

func (g *Game) CheckLoiterEvent() {
	if g.ShowEvent.Get() {
		return
	}
	chance := g.travelEventChance() / 2
	if chance > loiterEventCap {
		chance = loiterEventCap
	}
	if !g.rollEventChance(chance) {
		return
	}
	g.triggerEventCategory(false)
}

func (g *Game) travelEventChance() int {
	loc := Locations[g.Location.Get()]
	heat := g.Heat.Get()
	chance := loc.RiskLevel * baseEventRiskFactor
	if heatEventDivisor > 0 {
		chance += heat / heatEventDivisor
	}
	if chance > travelEventCap {
		chance = travelEventCap
	}
	return chance
}

func (g *Game) rollEventChance(chance int) bool {
	if chance <= 0 {
		return false
	}
	return rand.Intn(100) < chance
}

func (g *Game) triggerEventCategory(allowMarket bool) {
	category := g.rollEventCategory(allowMarket)
	switch category {
	case eventMarket:
		g.TriggerMarketEvent()
	case eventLucky:
		g.TriggerLuckyEvent()
	case eventDanger:
		g.TriggerDangerEvent()
	}
}

func (g *Game) rollEventCategory(allowMarket bool) eventCategory {
	goodMult, badMult := g.loiterStageMultipliers()
	switch g.scheduleStatus() {
	case statusBlendingIn:
		goodMult *= 1.25
	case statusOffCampus:
		badMult *= 1.5
	}

	marketWeight := 35.0
	luckyWeight := 35.0
	dangerWeight := 30.0

	if !allowMarket {
		marketWeight = 0
	}
	if g.Location.Get() == locationMusicHall {
		marketWeight *= 0.75
	}

	marketWeight *= goodMult
	luckyWeight *= goodMult
	dangerWeight *= badMult

	total := marketWeight + luckyWeight + dangerWeight
	if total <= 0 {
		return eventLucky
	}
	roll := rand.Float64() * total
	if roll < marketWeight {
		return eventMarket
	}
	roll -= marketWeight
	if roll < luckyWeight {
		return eventLucky
	}
	return eventDanger
}

func (g *Game) loiterStageMultipliers() (goodMult, badMult float64) {
	seconds := g.loiterTicks * tickSeconds
	switch {
	case seconds < 30:
		return 1.0, 1.0
	case seconds < 60:
		return 0.75, 1.0
	case seconds < 90:
		return 0.5, 1.25
	default:
		return 0.25, 1.5
	}
}

func (g *Game) TriggerLuckyEvent() {
	loc := g.Location.Get()
	events := []struct {
		title  string
		action func() string
	}{
		{
			title: "Lucky Find!",
			action: func() string {
				amount := rand.Intn(36) + 20
				g.Cash.Update(func(c int) int { return c + amount })
				return fmt.Sprintf("You found $%d in the hallway!", amount)
			},
		},
		{
			title: "Forgotten Stash",
			action: func() string {
				count := rand.Intn(6) + 3
				added := g.addRandomCandies(count)
				if added == 0 {
					return "You found a stash, but your bag is full."
				}
				if added < count {
					return fmt.Sprintf("You found a stash with %d candies, but only grabbed %d.", count, added)
				}
				return fmt.Sprintf("You found %d candies in a locker!", added)
			},
		},
		{
			title: "Grateful Customer",
			action: func() string {
				if g.SellBonuses[loc] < 105 {
					g.SellBonuses[loc] = 105
				}
				return "For the rest of today, you can sell for +5% here."
			},
		},
		{
			title: "Teacher's Pet",
			action: func() string {
				if g.heatImmunityDays < 2 {
					g.heatImmunityDays = 2
				}
				return "Heat gains are suppressed for 2 days."
			},
		},
		{
			title: "First Aid",
			action: func() string {
				healed := g.healPlayer(20)
				if healed == 0 {
					return "The nurse waves you off. You're already fine."
				}
				return fmt.Sprintf("A bandage and a pep talk. +%d HP.", healed)
			},
		},
		{
			title: "Study Hall",
			action: func() string {
				g.reduceHeat(10)
				return "You lay low for a bit.\nHeat drops by 10."
			},
		},
		{
			title: "Contraband Sample",
			action: func() string {
				item := BlackMarketItems[rand.Intn(len(BlackMarketItems))]
				if g.addItemToInventory(item.Name, 1) == 0 {
					return "A shady kid offers a sample.\nYour bag is too full."
				}
				return fmt.Sprintf("You got a %s.", item.Name)
			},
		},
		{
			title: "Rumor Mill",
			action: func() string {
				candy := CandyTypes[rand.Intn(len(CandyTypes))]
				tip := g.buildHotTip(candy.Name)
				return "Rumor mill:\n" + tip
			},
		},
		{
			title: "Study Buddy",
			action: func() string {
				g.addTradeBuff(1)
				return "A classmate covers for you.\nNext trade gets a better rate."
			},
		},
	}

	event := events[rand.Intn(len(events))]
	msg := event.action()
	g.updateUnlocks()
	g.TriggerEvent(event.title, msg)
}

func (g *Game) TriggerMarketEvent() {
	loc := g.Location.Get()
	candy := CandyTypes[rand.Intn(len(CandyTypes))]
	base := g.Markets[loc]
	if base == nil {
		base = g.rollBasePrices()
	}
	newBase := make(MarketPrices)
	for k, v := range base {
		newBase[k] = v
	}

	roll := rand.Intn(7)
	switch roll {
	case 0:
		newBase[candy.Name] = candy.MinPrice
		g.Markets[loc] = newBase
		g.RefreshPrices()
		g.TriggerEvent("Sugar Glut!", fmt.Sprintf("%s prices have crashed!\nBuying opportunity?", candy.Name))
	case 1:
		newBase[candy.Name] = candy.MaxPrice
		g.Markets[loc] = newBase
		g.RefreshPrices()
		g.TriggerEvent("Candy Bust!", fmt.Sprintf("%s is in high demand!\nSell now for maximum profit!", candy.Name))
	case 2:
		g.addShortage(loc, candy.Name, 3)
		g.TriggerEvent("Shortage!", fmt.Sprintf("%s is gone for a few days.", candy.Name))
	case 3:
		g.StockBonuses[loc] = 150
		g.TriggerEvent("Supply Rush!", "Stock surged +50% here today.")
	case 4:
		g.BuyBonuses[loc] = 90
		g.TriggerEvent("Flash Sale!", "Buy prices are 10% cheaper here today.")
	case 5:
		if g.SellBonuses[loc] < 110 {
			g.SellBonuses[loc] = 110
		}
		g.TriggerEvent("Pop-up Buyer!", "Sell prices are 10% higher here today.")
	default:
		g.hotTipEvent(candy.Name)
	}
}

func (g *Game) TriggerDangerEvent() {
	status := g.scheduleStatus()
	events := []struct {
		title       string
		action      func() string
		startCombat bool
	}{
		{
			title: "Snitch",
			action: func() string {
				g.addHeat(24, true, false)
				return "Someone ratted you out.\nHeat is rising fast."
			},
		},
		{
			title: "Mugging",
			action: func() string {
				cash := g.Cash.Get()
				percent := rand.Intn(13) + 12
				loss := cash * percent / 100
				g.Cash.Set(cash - loss)
				g.addHeat(3, true, false)
				if loss > 0 {
					return fmt.Sprintf("Bullies jumped you!\nLost $%d.", loss)
				}
				return "Bullies jumped you!\nYou had no cash to lose."
			},
		},
		{
			title: "Ambush",
			action: func() string {
				g.StartCombat(enemyBully)
				return "A bully blocks your path!"
			},
			startCombat: true,
		},
		{
			title: "Foot Chase",
			action: func() string {
				target := 10 + rand.Intn(9)
				roll := g.playerSPD() + rand.Intn(10)
				if roll >= target {
					g.addHeat(4, true, false)
					return "You outran the hall monitor.\nHeat +4."
				}
				g.StartCombat(enemyHallMonitor)
				return ""
			},
		},
		{
			title: "Confiscation",
			action: func() string {
				cash := g.Cash.Get()
				percent := rand.Intn(11) + 8
				loss := cash * percent / 100
				g.Cash.Set(cash - loss)
				lostItems := g.loseRandomCandy(rand.Intn(2) + 1)
				g.addHeat(6, true, false)
				if loss > 0 && lostItems > 0 {
					return fmt.Sprintf("A teacher confiscates $%d and %d candy.", loss, lostItems)
				}
				if loss > 0 {
					return fmt.Sprintf("A teacher confiscates $%d.", loss)
				}
				if lostItems > 0 {
					return fmt.Sprintf("A teacher confiscates %d candy.", lostItems)
				}
				return "A teacher searches your bag, but finds nothing."
			},
		},
		{
			title: "Bag Snatch",
			action: func() string {
				lostItems := g.loseRandomCandy(rand.Intn(3) + 1)
				g.addHeat(5, true, false)
				if lostItems > 0 {
					return fmt.Sprintf("A bully snatched %d candies from your bag.", lostItems)
				}
				return "A bully tried to snatch your bag, but it was empty."
			},
		},
		{
			title: "Rival Trader",
			action: func() string {
				g.TriggerEventWithOptions(
					"Rival Trader",
					"A rival trader sizes you up.\nFight for their stash?",
					[]EventOption{
						{
							Key:   'c',
							Label: "Challenge",
							Action: func() {
								g.ChallengeRival()
							},
						},
						{
							Key:   'b',
							Label: "Back off",
							Action: func() {
								g.addHeat(2, true, false)
								g.Message.Set("You back off and keep your head down.")
							},
						},
					},
				)
				return ""
			},
		},
	}

	if status != statusBlendingIn {
		events = append(events,
			struct {
				title       string
				action      func() string
				startCombat bool
			}{
				title: "Shakedown",
				action: func() string {
					if status == statusWrongClass && rand.Intn(2) == 0 {
						g.addHeat(5, true, false)
						return "A teacher questions you.\nYou get a warning."
					}
					amount := rand.Intn(61) + 30
					g.TriggerEventWithOptions(
						"Shakedown",
						fmt.Sprintf("A teacher catches you.\nPay $%d or fight?", amount),
						[]EventOption{
							{
								Key:   'p',
								Label: "Pay",
								Action: func() {
									cash := g.Cash.Get()
									paid := amount
									if cash < amount {
										paid = cash
									}
									g.Cash.Set(cash - paid)
									if paid < amount {
										g.loseRandomCandy(1)
									}
									g.addHeat(8, true, false)
									g.Message.Set(fmt.Sprintf("Paid $%d to keep quiet.", paid))
								},
							},
							{
								Key:   'f',
								Label: "Fight",
								Action: func() {
									g.addHeat(10, true, false)
									g.StartCombat(enemyHallMonitor)
								},
							},
						},
					)
					return ""
				},
			},
			struct {
				title       string
				action      func() string
				startCombat bool
			}{
				title: "Locker Search",
				action: func() string {
					if status == statusWrongClass && rand.Intn(2) == 0 {
						g.addHeat(5, true, false)
						return "They check your locker.\nYou get a warning."
					}
					lost := 0
					inv := g.Inventory.Get()
					if g.InventoryCount() > g.Capacity/2 {
						newInv := make(Inventory)
						for name, qty := range inv {
							loseQty := qty / 5
							remaining := qty - loseQty
							if remaining > 0 {
								newInv[name] = remaining
							}
							lost += loseQty
						}
						g.Inventory.Set(newInv)
					}
					g.addHeat(12, true, false)
					if lost > 0 {
						return fmt.Sprintf("Locker search!\nThey confiscated %d items.", lost)
					}
					return "Locker search!\nThey didn't find anything."
				},
			},
		)
	}

	event := events[rand.Intn(len(events))]
	msg := event.action()
	if event.startCombat {
		return
	}
	if msg == "" {
		return
	}
	g.TriggerEvent(event.title, msg)
}

func (g *Game) hotTipEvent(candyName string) {
	cost := 25
	g.TriggerEventWithOptions(
		"Hot Tip",
		fmt.Sprintf("A kid offers a tip on %s prices for $%d.", candyName, cost),
		[]EventOption{
			{
				Key:   'p',
				Label: "Pay",
				Action: func() {
					cash := g.Cash.Get()
					if cash < cost {
						g.Message.Set("Not enough cash for the tip.")
						return
					}
					g.Cash.Set(cash - cost)
					msg := g.buildHotTip(candyName)
					g.Message.Set(msg)
				},
			},
			{
				Key:   'n',
				Label: "No thanks",
				Action: func() {
					g.Message.Set("You pass on the tip.")
				},
			},
		},
	)
}

func (g *Game) buildHotTip(candyName string) string {
	bestBuyLoc := -1
	bestBuy := 0
	bestSellLoc := -1
	bestSell := 0

	for loc := range Locations {
		buy := g.buyPriceAtLocation(loc, candyName)
		if buy > 0 && (bestBuy == 0 || buy < bestBuy) {
			bestBuy = buy
			bestBuyLoc = loc
		}
		sell := g.sellPriceAtLocation(loc, candyName)
		if sell > bestSell {
			bestSell = sell
			bestSellLoc = loc
		}
	}

	if bestSellLoc < 0 {
		return "No tip available right now."
	}

	tip := ""
	if bestBuyLoc >= 0 {
		tip = fmt.Sprintf("Tip: Buy at %s ($%d)", Locations[bestBuyLoc].Name, bestBuy)
	} else {
		tip = "Tip: No buys available (shortage everywhere)"
	}
	tip += fmt.Sprintf("; Sell at %s ($%d).", Locations[bestSellLoc].Name, bestSell)
	return tip
}

func (g *Game) TriggerEvent(title, msg string) {
	g.TriggerEventWithOptions(title, msg, nil)
}

func (g *Game) TriggerEventWithOptions(title, msg string, options []EventOption) {
	if msg == "" {
		return
	}
	g.EventTitle.Set(title)
	g.EventMessage.Set(msg)
	g.eventOptions = options
	g.ShowEvent.Set(true)
}

func (g *Game) DismissEvent() {
	g.ShowEvent.Set(false)
	g.eventOptions = nil
}
