package main

import "math/rand"

func timeOfDayPriceMultiplier(hour int) int {
	switch {
	case hour <= 2:
		return 110
	case hour <= 4:
		return 100
	case hour <= 6:
		return 115
	default:
		return 90
	}
}

func timeOfDayStockMultiplier(hour int) int {
	switch {
	case hour <= 2:
		return 90
	case hour <= 4:
		return 100
	case hour <= 6:
		return 100
	default:
		return 125
	}
}

func applyPercent(value, percent int) int {
	if value <= 0 || percent <= 0 {
		return 0
	}
	result := (value*percent + 50) / 100
	if result < 1 {
		result = 1
	}
	return result
}

func candyByName(name string) CandyType {
	for _, candy := range CandyTypes {
		if candy.Name == name {
			return candy
		}
	}
	return CandyType{Name: name, MinPrice: 1, MaxPrice: 1}
}

func (g *Game) addItemToInventory(name string, qty int) int {
	if qty <= 0 {
		return 0
	}
	space := g.Capacity - g.InventoryCount()
	if space <= 0 {
		return 0
	}
	if qty > space {
		qty = space
	}
	inv := g.Inventory.Get()
	newInv := make(Inventory)
	for k, v := range inv {
		newInv[k] = v
	}
	newInv[name] += qty
	g.Inventory.Set(newInv)
	return qty
}

func (g *Game) addRandomCandies(count int) int {
	if count <= 0 {
		return 0
	}
	space := g.Capacity - g.InventoryCount()
	if space <= 0 {
		return 0
	}
	if count > space {
		count = space
	}
	inv := g.Inventory.Get()
	newInv := make(Inventory)
	for k, v := range inv {
		newInv[k] = v
	}
	for i := 0; i < count; i++ {
		candy := CandyTypes[rand.Intn(len(CandyTypes))]
		newInv[candy.Name]++
	}
	g.Inventory.Set(newInv)
	return count
}
