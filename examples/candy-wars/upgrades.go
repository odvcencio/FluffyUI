package main

import "fmt"

func (g *Game) canAccessCombatUpgrades() bool {
	return g.Location.Get() == locationGymnasium
}

func (g *Game) canAccessTradeUpgrades() bool {
	return g.Location.Get() == locationCafeteria
}

func (g *Game) spendCash(cost int) bool {
	if cost <= 0 {
		return true
	}
	cash := g.Cash.Get()
	if cash < cost {
		g.Message.Set("Not enough cash.")
		return false
	}
	g.Cash.Set(cash - cost)
	return true
}

func (g *Game) BuyWorkout() bool {
	if !g.canAccessCombatUpgrades() {
		g.Message.Set("Workout sessions are only at the Gymnasium.")
		return false
	}
	if g.workoutCount >= statUpgradeMax {
		g.Message.Set("Workout Sessions are maxed out.")
		return false
	}
	if !g.spendCash(statUpgradeCost) {
		return false
	}
	g.workoutCount++
	g.Message.Set(fmt.Sprintf("Workout complete. ATK +2 (%d/%d).", g.workoutCount, statUpgradeMax))
	return true
}

func (g *Game) BuyThickSkin() bool {
	if !g.canAccessCombatUpgrades() {
		g.Message.Set("Thick Skin training is only at the Gymnasium.")
		return false
	}
	if g.thickSkinCount >= statUpgradeMax {
		g.Message.Set("Thick Skin is maxed out.")
		return false
	}
	if !g.spendCash(statUpgradeCost) {
		return false
	}
	g.thickSkinCount++
	g.Message.Set(fmt.Sprintf("Training complete. DEF +2 (%d/%d).", g.thickSkinCount, statUpgradeMax))
	return true
}

func (g *Game) BuyTrackPractice() bool {
	if !g.canAccessCombatUpgrades() {
		g.Message.Set("Track Practice is only at the Gymnasium.")
		return false
	}
	if g.trackPracticeCount >= statUpgradeMax {
		g.Message.Set("Track Practice is maxed out.")
		return false
	}
	if !g.spendCash(statUpgradeCost) {
		return false
	}
	g.trackPracticeCount++
	g.Message.Set(fmt.Sprintf("You feel faster. SPD +2 (%d/%d).", g.trackPracticeCount, statUpgradeMax))
	return true
}

func (g *Game) BuyHireMuscle() bool {
	if !g.canAccessCombatUpgrades() {
		g.Message.Set("Hire Muscle is only at the Gymnasium.")
		return false
	}
	if g.hasMuscle {
		g.Message.Set("You've already hired muscle.")
		return false
	}
	if !g.spendCash(muscleCost) {
		return false
	}
	g.hasMuscle = true
	g.Message.Set("Muscle hired. +10 ATK in fights.")
	return true
}

func (g *Game) BuyIntimidation() bool {
	if !g.canAccessCombatUpgrades() {
		g.Message.Set("Intimidation training is only at the Gymnasium.")
		return false
	}
	if g.hasIntimidation {
		g.Message.Set("You're already intimidating enough.")
		return false
	}
	if !g.spendCash(intimidationCost) {
		return false
	}
	g.hasIntimidation = true
	g.Message.Set("Intimidation unlocked. Enemies may flee.")
	return true
}

func (g *Game) BuyBackpackUpgrade() bool {
	if !g.canAccessTradeUpgrades() {
		g.Message.Set("Backpack upgrades are only at the Cafeteria.")
		return false
	}
	if g.backpackTier >= len(backpackTierCosts) {
		g.Message.Set("Backpack is fully upgraded.")
		return false
	}
	cost := backpackTierCosts[g.backpackTier]
	gain := backpackTierGains[g.backpackTier]
	if !g.spendCash(cost) {
		return false
	}
	g.backpackTier++
	g.Capacity += gain
	g.Message.Set(fmt.Sprintf("Backpack expanded by %d slots.", gain))
	return true
}

func (g *Game) BuySecretStash() bool {
	if !g.canAccessTradeUpgrades() {
		g.Message.Set("Secret Stash is only at the Cafeteria.")
		return false
	}
	if g.hasStash {
		g.Message.Set("Secret Stash already unlocked.")
		return false
	}
	if !g.spendCash(stashCost) {
		return false
	}
	g.hasStash = true
	g.stashCapacity = stashCapacity
	if g.stash == nil {
		g.stash = make(Inventory)
	}
	g.Message.Set("Secret Stash unlocked. Keep 50 items safe.")
	return true
}

func (g *Game) BuyBike() bool {
	if !g.canAccessTradeUpgrades() {
		g.Message.Set("Bikes are only sold at the Cafeteria.")
		return false
	}
	if g.hasBike {
		g.Message.Set("You already have a bike.")
		return false
	}
	if !g.spendCash(bikeCost) {
		return false
	}
	g.hasBike = true
	g.Message.Set("Bike acquired. Travel now costs 1 hour.")
	return true
}

func (g *Game) BuyInformant() bool {
	if !g.canAccessTradeUpgrades() {
		g.Message.Set("Informant Network is only at the Cafeteria.")
		return false
	}
	if g.hasInformant {
		g.Message.Set("Informant Network already active.")
		return false
	}
	if !g.spendCash(informantCost) {
		return false
	}
	g.hasInformant = true
	g.Message.Set("Informant Network online. Adjacent prices unlocked.")
	return true
}

func (g *Game) BuyBankExpansion() bool {
	if !g.canAccessTradeUpgrades() {
		g.Message.Set("Bank expansions are only at the Cafeteria.")
		return false
	}
	if g.bankExpanded {
		g.Message.Set("Bank limit already expanded.")
		return false
	}
	if !g.spendCash(bankExpansionCost) {
		return false
	}
	g.bankExpanded = true
	g.BankLimit = bankLimitExpanded
	g.Message.Set("Bank limit expanded to $1500.")
	return true
}
