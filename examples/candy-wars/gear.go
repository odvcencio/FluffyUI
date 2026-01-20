package main

import "math"

func weaponByName(name string) (Weapon, bool) {
	for _, weapon := range Weapons {
		if weapon.Name == name {
			return weapon, true
		}
	}
	return Weapon{}, false
}

func armorByName(name string) (Armor, bool) {
	for _, armor := range Armors {
		if armor.Name == name {
			return armor, true
		}
	}
	return Armor{}, false
}

func (g *Game) playerWeapon() Weapon {
	if weapon, ok := weaponByName(g.equippedWeapon); ok {
		return weapon
	}
	return Weapons[0]
}

func (g *Game) playerArmor() Armor {
	if armor, ok := armorByName(g.equippedArmor); ok {
		return armor
	}
	return Armors[0]
}

func (g *Game) gearShopAvailable() bool {
	return g.Location.Get() == locationGymnasium
}

func (g *Game) weaponPrice(weapon Weapon) int {
	if weapon.Cost <= 0 {
		return 0
	}
	price := weapon.Cost
	if g.Location.Get() == locationGymnasium {
		price = applyPercent(price, 110)
	}
	return price
}

func (g *Game) armorPrice(armor Armor) int {
	if armor.Cost <= 0 {
		return 0
	}
	return armor.Cost
}

func (g *Game) BuyOrEquipWeapon(name string) bool {
	if !g.gearShopAvailable() {
		g.Message.Set("Gear shop is only at the Gymnasium.")
		return false
	}
	weapon, ok := weaponByName(name)
	if !ok {
		return false
	}
	if weapon.RequiresBlackMarket && !g.blackMarketUnlocked() {
		g.Message.Set("Black Market weapons are still locked.")
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
	price := g.weaponPrice(weapon)
	if !g.spendCash(price) {
		return false
	}
	g.ownedWeapons[weapon.Name] = true
	g.equippedWeapon = weapon.Name
	g.Message.Set("Bought and equipped " + weapon.Name + ".")
	return true
}

func (g *Game) BuyOrEquipArmor(name string) bool {
	if !g.gearShopAvailable() {
		g.Message.Set("Gear shop is only at the Gymnasium.")
		return false
	}
	armor, ok := armorByName(name)
	if !ok {
		return false
	}
	if armor.RequiresBlackMarket && !g.blackMarketUnlocked() {
		g.Message.Set("Black Market armor is still locked.")
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
	price := g.armorPrice(armor)
	if !g.spendCash(price) {
		return false
	}
	g.ownedArmors[armor.Name] = true
	g.equippedArmor = armor.Name
	g.Message.Set("Bought and equipped " + armor.Name + ".")
	return true
}

func weaponMultiplier(attacker WeaponClass, defender WeaponClass) float64 {
	if attacker == classNeutral || defender == classNeutral || attacker == defender {
		return 1.0
	}
	switch attacker {
	case classBlunt:
		if defender == classQuick {
			return 1.5
		}
		if defender == classSharp {
			return 0.5
		}
	case classSharp:
		if defender == classBlunt {
			return 1.5
		}
		if defender == classQuick {
			return 0.5
		}
	case classQuick:
		if defender == classSharp {
			return 1.5
		}
		if defender == classBlunt {
			return 0.5
		}
	}
	return 1.0
}

func applyWeaponMultiplier(raw int, attacker WeaponClass, defender WeaponClass) (int, float64) {
	if raw < 1 {
		raw = 1
	}
	mult := weaponMultiplier(attacker, defender)
	damage := int(math.Floor(float64(raw) * mult))
	if damage < 1 {
		damage = 1
	}
	return damage, mult
}

func weaponClassLabel(class WeaponClass) string {
	switch class {
	case classBlunt:
		return "Blunt"
	case classSharp:
		return "Sharp"
	case classQuick:
		return "Quick"
	default:
		return "Neutral"
	}
}
