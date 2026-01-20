package main

import "fmt"

type ItemDef struct {
	Name           string
	Description    string
	Class          WeaponClass
	UseInCombat    bool
	UseOutOfCombat bool
}

func itemDefByName(name string) (ItemDef, bool) {
	for _, item := range CraftedItems {
		if item.Name == name {
			return ItemDef{
				Name:           item.Name,
				Description:    item.Description,
				Class:          item.Class,
				UseInCombat:    item.UseInCombat,
				UseOutOfCombat: item.UseOutOfCombat,
			}, true
		}
	}
	for _, item := range BlackMarketItems {
		if item.Name == name {
			return ItemDef{
				Name:           item.Name,
				Description:    item.Description,
				Class:          item.Class,
				UseInCombat:    item.UseInCombat,
				UseOutOfCombat: item.UseOutOfCombat,
			}, true
		}
	}
	return ItemDef{}, false
}

func allItemDefs() []ItemDef {
	items := make([]ItemDef, 0, len(CraftedItems)+len(BlackMarketItems))
	for _, item := range CraftedItems {
		items = append(items, ItemDef{
			Name:           item.Name,
			Description:    item.Description,
			Class:          item.Class,
			UseInCombat:    item.UseInCombat,
			UseOutOfCombat: item.UseOutOfCombat,
		})
	}
	for _, item := range BlackMarketItems {
		items = append(items, ItemDef{
			Name:           item.Name,
			Description:    item.Description,
			Class:          item.Class,
			UseInCombat:    item.UseInCombat,
			UseOutOfCombat: item.UseOutOfCombat,
		})
	}
	return items
}

func (g *Game) UseItem(name string, inCombat bool) bool {
	item, ok := itemDefByName(name)
	if !ok {
		g.itemUseMessage(inCombat, "You don't recognize that item.")
		return false
	}
	if inCombat && !item.UseInCombat {
		g.itemUseMessage(inCombat, "That item can't be used in combat.")
		return false
	}
	if !inCombat && !item.UseOutOfCombat {
		g.itemUseMessage(inCombat, "That item only works in combat.")
		return false
	}
	inv := g.Inventory.Get()
	if inv[name] <= 0 {
		g.itemUseMessage(inCombat, "You're out of that item.")
		return false
	}
	if item.Name == "Rare Remedy" && g.HP.Get() >= maxHP {
		g.itemUseMessage(inCombat, "You're already at full health.")
		return false
	}

	switch item.Name {
	case "Sugar Rush":
		if g.Combat == nil {
			return false
		}
		g.consumeInventoryItem(name, 1)
		g.Combat.SpeedBuff = 15
		if g.Combat.SpeedTurns < 3 {
			g.Combat.SpeedTurns = 3
		}
		g.appendCombatLog("You feel wired. Speed surges!")
		return true

	case "Jaw Lockdown":
		if g.Combat == nil {
			return false
		}
		g.consumeInventoryItem(name, 1)
		g.Combat.EnemyStun = 1
		g.appendCombatLog("%s is locked down!", g.Combat.Enemy.Name)
		return true

	case "Chocolate Shield":
		if g.Combat == nil {
			return false
		}
		g.consumeInventoryItem(name, 1)
		g.Combat.Shield += 30
		g.appendCombatLog("Chocolate shield up (+30).")
		return true

	case "Sour Bomb":
		if g.Combat == nil {
			return false
		}
		g.consumeInventoryItem(name, 1)
		raw := 25
		damage, mult := applyWeaponMultiplier(raw, item.Class, g.Combat.Enemy.WeaponType)
		g.Combat.EnemyHP -= damage
		g.appendCombatLog("Sour bomb hits for %d!", damage)
		g.logEffectiveness(mult)
		if g.Combat.EnemyHP <= 0 {
			g.winCombat()
			return true
		}
		return true

	case "Heat Sponge":
		if g.Combat != nil {
			return false
		}
		g.consumeInventoryItem(name, 1)
		g.reduceHeat(20)
		g.Message.Set("Heat reduced by 20.")
		return true

	case "Trade Voucher":
		if g.Combat != nil {
			return false
		}
		g.consumeInventoryItem(name, 1)
		g.addTradeBuff(2)
		g.Message.Set("Trade Voucher primed for 2 trades.")
		return true

	case "Pepper Spray":
		if g.Combat == nil {
			return false
		}
		g.consumeInventoryItem(name, 1)
		raw := 18
		damage, mult := applyWeaponMultiplier(raw, item.Class, g.Combat.Enemy.WeaponType)
		g.Combat.EnemyHP -= damage
		g.appendCombatLog("Pepper spray hits for %d!", damage)
		g.logEffectiveness(mult)
		g.Combat.EnemyStun = 1
		if g.Combat.EnemyHP <= 0 {
			g.winCombat()
			return true
		}
		return true

	case "Rare Remedy":
		healed := g.healPlayer(50)
		if healed <= 0 {
			g.itemUseMessage(inCombat, "You're already at full health.")
			return false
		}
		g.consumeInventoryItem(name, 1)
		g.itemUseMessage(inCombat, fmt.Sprintf("Healed %d HP.", healed))
		return true

	case "Lucky Charm":
		if g.Combat != nil {
			return false
		}
		g.consumeInventoryItem(name, 1)
		g.addTradeBuff(1)
		g.Message.Set("Lucky Charm primed. Next trade boosted.")
		return true

	case "Invisibility Pop":
		if g.Combat == nil {
			return false
		}
		g.consumeInventoryItem(name, 1)
		g.appendCombatLog("You vanish in a flash!")
		g.Combat = nil
		g.Message.Set("You slipped away using an Invisibility Pop.")
		return true
	}

	g.itemUseMessage(inCombat, "Nothing happens.")
	return false
}

func (g *Game) itemUseMessage(inCombat bool, msg string) {
	if inCombat && g.Combat != nil {
		g.appendCombatLog("%s", msg)
		return
	}
	g.Message.Set(msg)
}

func (g *Game) consumeInventoryItem(name string, qty int) bool {
	if qty <= 0 {
		return false
	}
	inv := g.Inventory.Get()
	if inv[name] < qty {
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
	return true
}

func (g *Game) addTradeBuff(uses int) {
	if uses <= 0 {
		return
	}
	g.tradeBuffUses += uses
	g.tradeBuffBuyMult = 90
	g.tradeBuffSellMult = 110
}

func (g *Game) consumeTradeBuff() {
	if g.tradeBuffUses <= 0 {
		return
	}
	g.tradeBuffUses--
	if g.tradeBuffUses <= 0 {
		g.tradeBuffUses = 0
		g.tradeBuffBuyMult = 100
		g.tradeBuffSellMult = 100
	}
}
