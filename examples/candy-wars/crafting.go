package main

import "fmt"

func (g *Game) CraftItem(index int, qty int) bool {
	if g.GameOver.Get() || g.InCombat() {
		return false
	}
	if !g.craftingUnlocked() {
		g.Message.Set("Crafting is locked. Reach Day 7 or earn $500 profit.")
		return false
	}
	if index < 0 || index >= len(CraftedItems) || qty <= 0 {
		return false
	}
	if !g.canSpendHours(qty) {
		g.Message.Set("Not enough hours left to craft that many.")
		return false
	}

	item := CraftedItems[index]
	inv := g.Inventory.Get()

	required := make(map[string]int)
	totalIngredients := 0
	for name, amount := range item.Ingredients {
		needed := amount * qty
		required[name] = needed
		totalIngredients += needed
	}

	for name, needed := range required {
		if inv[name] < needed {
			g.Message.Set(fmt.Sprintf("Need %d %s to craft %s.", needed, name, item.Name))
			return false
		}
	}

	currentCount := 0
	for _, count := range inv {
		currentCount += count
	}
	netCount := currentCount - totalIngredients + qty
	if netCount > g.Capacity {
		g.Message.Set("Not enough space in your backpack.")
		return false
	}

	newInv := make(Inventory)
	for k, v := range inv {
		newInv[k] = v
	}
	for name, needed := range required {
		newInv[name] -= needed
		if newInv[name] <= 0 {
			delete(newInv, name)
		}
	}
	newInv[item.Name] += qty
	g.Inventory.Set(newInv)

	dayRolled := g.advanceHours(qty)
	if dayRolled {
		g.StartNewDay()
	} else {
		g.resetLoiter()
	}
	g.RefreshPrices()
	g.Message.Set(fmt.Sprintf("Crafted %d %s.", qty, item.Name))
	g.recordCraft(qty)
	g.CheckEndConditions()
	return true
}
