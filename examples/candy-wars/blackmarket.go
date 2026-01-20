package main

import "fmt"

func blackMarketWeaponPrice(weapon Weapon) int {
	if weapon.Cost <= 0 {
		return 0
	}
	return applyPercent(weapon.Cost, blackMarketGearMarkup)
}

func blackMarketArmorPrice(armor Armor) int {
	if armor.Cost <= 0 {
		return 0
	}
	return applyPercent(armor.Cost, blackMarketGearMarkup)
}

func (g *Game) BuyBlackMarketWeapon(name string) bool {
	if !g.blackMarketUnlocked() {
		g.Message.Set("Black Market is still locked.")
		return false
	}
	weapon, ok := weaponByName(name)
	if !ok || !weapon.RequiresBlackMarket {
		return false
	}
	if g.ownedWeapons == nil {
		g.ownedWeapons = make(map[string]bool)
	}
	if g.ownedWeapons[weapon.Name] {
		g.equippedWeapon = weapon.Name
		g.Message.Set("Equipped " + weapon.Name + ".")
		return true
	}
	price := blackMarketWeaponPrice(weapon)
	if !g.spendCash(price) {
		return false
	}
	g.ownedWeapons[weapon.Name] = true
	g.equippedWeapon = weapon.Name
	g.Message.Set("Bought and equipped " + weapon.Name + ".")
	return true
}

func (g *Game) BuyBlackMarketArmor(name string) bool {
	if !g.blackMarketUnlocked() {
		g.Message.Set("Black Market is still locked.")
		return false
	}
	armor, ok := armorByName(name)
	if !ok || !armor.RequiresBlackMarket {
		return false
	}
	if g.ownedArmors == nil {
		g.ownedArmors = make(map[string]bool)
	}
	if g.ownedArmors[armor.Name] {
		g.equippedArmor = armor.Name
		g.Message.Set("Equipped " + armor.Name + ".")
		return true
	}
	price := blackMarketArmorPrice(armor)
	if !g.spendCash(price) {
		return false
	}
	g.ownedArmors[armor.Name] = true
	g.equippedArmor = armor.Name
	g.Message.Set("Bought and equipped " + armor.Name + ".")
	return true
}

func (g *Game) BuyBlackMarketRareImport(qty int) bool {
	if !g.blackMarketUnlocked() {
		g.Message.Set("Black Market is still locked.")
		return false
	}
	if qty <= 0 {
		return false
	}
	totalCost := blackMarketRarePrice * qty
	if g.Cash.Get() < totalCost {
		g.Message.Set("Not enough cash.")
		return false
	}
	if g.InventoryCount()+qty > g.Capacity {
		g.Message.Set("Not enough space in your backpack.")
		return false
	}

	inv := g.Inventory.Get()
	newInv := make(Inventory)
	for k, v := range inv {
		newInv[k] = v
	}
	newInv["Rare Import"] += qty
	g.Inventory.Set(newInv)
	g.Cash.Update(func(c int) int { return c - totalCost })
	g.recordTrade(qty, totalCost)
	g.recordTradeHistory("Buy", "Rare Import", qty, blackMarketRarePrice, totalCost, "Black Market")
	g.Message.Set(fmt.Sprintf("Bought %d Rare Imports for $%d.", qty, totalCost))
	return true
}

func blackMarketItemByName(name string) (SpecialItem, bool) {
	for _, item := range BlackMarketItems {
		if item.Name == name {
			return item, true
		}
	}
	return SpecialItem{}, false
}

func (g *Game) BuyBlackMarketItem(name string, qty int) bool {
	if !g.blackMarketUnlocked() {
		g.Message.Set("Black Market is still locked.")
		return false
	}
	item, ok := blackMarketItemByName(name)
	if !ok {
		return false
	}
	if qty <= 0 {
		return false
	}
	totalCost := item.Price * qty
	if g.Cash.Get() < totalCost {
		g.Message.Set("Not enough cash.")
		return false
	}
	if g.InventoryCount()+qty > g.Capacity {
		g.Message.Set("Not enough space in your backpack.")
		return false
	}
	inv := g.Inventory.Get()
	newInv := make(Inventory)
	for k, v := range inv {
		newInv[k] = v
	}
	newInv[item.Name] += qty
	g.Inventory.Set(newInv)
	g.Cash.Update(func(c int) int { return c - totalCost })
	g.recordTradeHistory("Buy", item.Name, qty, item.Price, totalCost, "Black Market")
	g.Message.Set(fmt.Sprintf("Bought %d %s for $%d.", qty, item.Name, totalCost))
	return true
}
