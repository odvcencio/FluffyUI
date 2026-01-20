package main

import "fmt"

func (g *Game) StashCount() int {
	if !g.hasStash {
		return 0
	}
	count := 0
	for _, qty := range g.stash {
		count += qty
	}
	return count
}

func (g *Game) StashDeposit(name string, qty int) bool {
	if !g.hasStash {
		g.Message.Set("No stash available.")
		return false
	}
	if qty <= 0 {
		return false
	}
	inv := g.Inventory.Get()
	if inv[name] < qty {
		g.Message.Set("Not enough items to stash.")
		return false
	}
	if g.StashCount()+qty > g.stashCapacity {
		g.Message.Set("Stash is full.")
		return false
	}

	newInv := make(Inventory)
	for k, v := range inv {
		newInv[k] = v
	}
	newInv[name] -= qty
	if newInv[name] <= 0 {
		delete(newInv, name)
	}
	g.Inventory.Set(newInv)

	if g.stash == nil {
		g.stash = make(Inventory)
	}
	g.stash[name] += qty
	g.Message.Set(fmt.Sprintf("Stashed %d %s.", qty, name))
	return true
}

func (g *Game) StashWithdraw(name string, qty int) bool {
	if !g.hasStash {
		g.Message.Set("No stash available.")
		return false
	}
	if qty <= 0 {
		return false
	}
	if g.stash[name] < qty {
		g.Message.Set("Not enough stashed items.")
		return false
	}
	if g.InventoryCount()+qty > g.Capacity {
		g.Message.Set("Not enough space in your backpack.")
		return false
	}

	newInv := make(Inventory)
	for k, v := range g.Inventory.Get() {
		newInv[k] = v
	}
	newInv[name] += qty
	g.Inventory.Set(newInv)

	g.stash[name] -= qty
	if g.stash[name] <= 0 {
		delete(g.stash, name)
	}
	g.Message.Set(fmt.Sprintf("Withdrew %d %s.", qty, name))
	return true
}
