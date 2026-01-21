package main

import (
	"fmt"
	"math"
	"math/rand"
)

func (g *Game) StartCombat(enemy Combatant) {
	if g.InCombat() || g.GameOver.Get() {
		return
	}
	enemy = g.applyDifficultyToEnemy(enemy)
	if g.hasIntimidation && !enemy.IsGoon && rand.Intn(100) < 20 {
		g.Message.Set(fmt.Sprintf("%s backs off from your glare.", enemy.Name))
		return
	}
	g.Combat = &CombatState{
		Enemy:      enemy,
		EnemyHP:    enemy.HP,
		Log:        []string{fmt.Sprintf("%s blocks your path!", enemy.Name)},
		Shield:     0,
		SpeedBuff:  0,
		SpeedTurns: 0,
		EnemyStun:  0,
	}
	g.addHeat(combatHeatGain, true, false)
	g.ShowEvent.Set(false)
	g.eventOptions = nil

	if enemy.SPD > g.playerSPD() {
		g.enemyAttack()
	}
}

func (g *Game) playerATK() int {
	atk := 10 + g.workoutCount*2
	if g.hasMuscle {
		atk += 10
	}
	atk += g.playerWeapon().AtkBonus
	return atk
}

func (g *Game) playerDEF() int {
	return 10 + g.thickSkinCount*2 + g.playerArmor().DefBonus
}

func (g *Game) playerSPD() int {
	spd := 10 + g.trackPracticeCount*2 + g.playerArmor().SPDMod
	if g.Combat != nil && g.Combat.SpeedTurns > 0 {
		spd += g.Combat.SpeedBuff
	}
	if spd < 1 {
		spd = 1
	}
	return spd
}

func (g *Game) appendCombatLog(format string, args ...interface{}) {
	if g.Combat == nil {
		return
	}
	line := fmt.Sprintf(format, args...)
	g.Combat.Log = append(g.Combat.Log, line)
	if len(g.Combat.Log) > 8 {
		g.Combat.Log = g.Combat.Log[len(g.Combat.Log)-8:]
	}
}

func (g *Game) CombatAttack() {
	if g.Combat == nil {
		return
	}
	raw := g.playerATK() - g.Combat.Enemy.DEF
	damage, mult := applyWeaponMultiplier(raw, g.playerWeapon().Class, g.Combat.Enemy.WeaponType)
	g.Combat.EnemyHP -= damage
	weaponName := g.playerWeapon().Name
	if weaponName == "" {
		weaponName = "attack"
	}
	g.appendCombatLog("You swing your %s for %d damage.", weaponName, damage)
	g.logEffectiveness(mult)
	if g.Combat.EnemyHP <= 0 {
		g.winCombat()
		return
	}
	g.tryWeaponStun()
	g.endPlayerTurn()
	g.enemyAttack()
}

func (g *Game) CombatDefend() {
	if g.Combat == nil {
		return
	}
	g.Combat.Defending = true
	g.appendCombatLog("You brace for the next hit.")
	g.endPlayerTurn()
	g.enemyAttack()
}

func (g *Game) CombatFlee() {
	if g.Combat == nil {
		return
	}
	enemySPD := g.Combat.Enemy.SPD
	chance := 40 + (g.playerSPD() - enemySPD)
	if g.Combat.Defending {
		chance += 25
	}
	if chance < 5 {
		chance = 5
	}
	if chance > 95 {
		chance = 95
	}
	g.Combat.Defending = false
	if rand.Intn(100) < chance {
		g.appendCombatLog("You got away!")
		g.Combat = nil
		return
	}
	g.appendCombatLog("You failed to escape!")
	g.endPlayerTurn()
	g.enemyAttack()
}

func (g *Game) enemyAttack() {
	if g.Combat == nil {
		return
	}
	if g.enemyTryFlee() {
		return
	}
	if g.Combat.EnemyStun > 0 {
		g.Combat.EnemyStun--
		g.appendCombatLog("%s is stunned!", g.Combat.Enemy.Name)
		return
	}
	raw := g.Combat.Enemy.ATK - g.playerDEF()
	damage, mult := applyWeaponMultiplier(raw, g.Combat.Enemy.WeaponType, g.playerWeapon().Class)
	if g.Combat.Defending {
		damage = int(math.Ceil(float64(damage) / 2.0))
		g.Combat.Defending = false
	}
	if g.Combat.Shield > 0 && damage > 0 {
		absorbed := damage
		if absorbed > g.Combat.Shield {
			absorbed = g.Combat.Shield
		}
		g.Combat.Shield -= absorbed
		damage -= absorbed
		g.appendCombatLog("Shield absorbs %d damage.", absorbed)
	}
	if damage <= 0 {
		g.appendCombatLog("Your shield blocks the hit.")
		return
	}
	g.HP.Update(func(h int) int {
		return h - damage
	})
	weaponName := g.Combat.Enemy.Weapon
	if weaponName == "" {
		weaponName = "attack"
	}
	g.appendCombatLog("%s hits with %s for %d.", g.Combat.Enemy.Name, weaponName, damage)
	g.logEffectiveness(mult)
	if g.HP.Get() <= 0 {
		g.loseCombat()
		return
	}
}

func (g *Game) winCombat() {
	if g.Combat == nil {
		return
	}
	enemy := g.Combat.Enemy
	cash := rand.Intn(enemy.CashMax-enemy.CashMin+1) + enemy.CashMin
	g.Cash.Update(func(c int) int { return c + cash })
	if enemy.CandyMax > 0 {
		count := rand.Intn(enemy.CandyMax-enemy.CandyMin+1) + enemy.CandyMin
		if count > 0 {
			added := g.addRandomCandies(count)
			if added > 0 {
				g.appendCombatLog("You grabbed %d candies.", added)
			}
			if added < count {
				g.appendCombatLog("Your bag was too full to take it all.")
			}
		}
	}
	if enemy.RareChance > 0 && rand.Float64() < enemy.RareChance {
		if g.addItemToInventory("Rare Import", 1) > 0 {
			g.appendCombatLog("You scored a Rare Import.")
		} else {
			g.appendCombatLog("No room for a Rare Import.")
		}
	}
	if enemy.IsGoon {
		reduced := g.reduceLoanDebt(100)
		if reduced > 0 {
			g.appendCombatLog("Debt reduced by $%d.", reduced)
		}
	}
	g.maybeDropWeapon(enemy)
	g.maybeDropArmor(enemy)
	g.rewardRivalLoot(enemy)
	g.appendCombatLog("You won the fight!")
	g.Combat = nil
	g.Message.Set(fmt.Sprintf("Won the fight and earned $%d.", cash))
	g.recordEnemyDefeat()
	g.CheckEndConditions()
}

func (g *Game) loseCombat() {
	g.Combat = nil
	g.SendToNurse()
}

func (g *Game) SendToNurse() {
	if g.GameOver.Get() {
		return
	}
	cash := g.Cash.Get()
	g.Cash.Set(cash / 2)
	g.HP.Set(maxHP / 2)
	g.reduceHeat(25)
	g.Hour.Set(1)
	g.StartNewDay()
	g.RefreshPrices()
	g.Message.Set("You were knocked out and sent to the nurse.")
	g.CheckEndConditions()
}

func (g *Game) maybeTriggerGoonFight() bool {
	if !g.goonDue || g.InCombat() {
		return false
	}
	g.goonDue = false
	g.StartCombat(enemyGoon)
	return g.InCombat()
}

func (g *Game) CombatUseItem(name string) bool {
	if g.Combat == nil {
		return false
	}
	if !g.UseItem(name, true) {
		return false
	}
	if g.Combat == nil {
		return true
	}
	g.endPlayerTurn()
	g.enemyAttack()
	return true
}

func (g *Game) endPlayerTurn() {
	if g.Combat == nil {
		return
	}
	if g.Combat.SpeedTurns > 0 {
		g.Combat.SpeedTurns--
		if g.Combat.SpeedTurns <= 0 {
			g.Combat.SpeedTurns = 0
			g.Combat.SpeedBuff = 0
		}
	}
}

func (g *Game) tryWeaponStun() {
	if g.Combat == nil {
		return
	}
	weapon := g.playerWeapon()
	if weapon.StunChance <= 0 {
		return
	}
	if rand.Intn(100) < weapon.StunChance {
		g.Combat.EnemyStun = 1
		g.appendCombatLog("%s is stunned by your %s!", g.Combat.Enemy.Name, weapon.Name)
	}
}

func (g *Game) logEffectiveness(mult float64) {
	if mult >= 1.25 {
		g.appendCombatLog("It's super effective!")
	} else if mult <= 0.75 {
		g.appendCombatLog("It's not very effective.")
	}
}

func (g *Game) enemyTryFlee() bool {
	if g.Combat == nil {
		return false
	}
	enemy := g.Combat.Enemy
	if !enemy.WillFlee {
		return false
	}
	threshold := enemy.HP * enemy.FleeAtPct / 100
	if threshold <= 0 {
		threshold = enemy.HP / 3
	}
	if g.Combat.EnemyHP > threshold {
		return false
	}
	if rand.Intn(100) < enemy.FleeChance {
		g.Message.Set(fmt.Sprintf("%s fled the fight.", enemy.Name))
		g.Combat = nil
		return true
	}
	g.appendCombatLog("%s tries to flee!", enemy.Name)
	return false
}

func (g *Game) maybeDropWeapon(enemy Combatant) {
	if enemy.Weapon == "" || enemy.WeaponDrop <= 0 {
		return
	}
	weapon, ok := weaponByName(enemy.Weapon)
	if !ok {
		return
	}
	if weapon.RequiresBlackMarket && !g.blackMarketUnlocked() {
		return
	}
	if g.ownedWeapons == nil {
		g.ownedWeapons = make(map[string]bool)
	}
	if g.ownedWeapons[weapon.Name] {
		return
	}
	if rand.Intn(100) < enemy.WeaponDrop {
		g.ownedWeapons[weapon.Name] = true
		g.appendCombatLog("You picked up %s.", weapon.Name)
	}
}

func (g *Game) maybeDropArmor(enemy Combatant) {
	if enemy.Armor == "" || enemy.ArmorDrop <= 0 {
		return
	}
	armor, ok := armorByName(enemy.Armor)
	if !ok {
		return
	}
	if armor.RequiresBlackMarket && !g.blackMarketUnlocked() {
		return
	}
	if g.ownedArmors == nil {
		g.ownedArmors = make(map[string]bool)
	}
	if g.ownedArmors[armor.Name] {
		return
	}
	if rand.Intn(100) < enemy.ArmorDrop {
		g.ownedArmors[armor.Name] = true
		g.appendCombatLog("You picked up %s.", armor.Name)
	}
}

func (g *Game) rewardRivalLoot(enemy Combatant) {
	if enemy.Name != enemyRivalTrader.Name {
		return
	}
	bonusCash := rand.Intn(21) + 10
	g.Cash.Update(func(c int) int { return c + bonusCash })
	bonusCandies := rand.Intn(3) + 2
	added := g.addRandomCandies(bonusCandies)
	loc := g.Location.Get()
	if g.SellBonuses[loc] < 110 {
		g.SellBonuses[loc] = 110
	}
	if added > 0 {
		g.appendCombatLog("Rival stash: +$%d and %d candies.", bonusCash, added)
		if added < bonusCandies {
			g.appendCombatLog("You couldn't carry everything.")
		}
	} else {
		g.appendCombatLog("Rival stash: +$%d, but no room for candies.", bonusCash)
	}
	g.appendCombatLog("Local buzz boosts sells here today.")
}

func (g *Game) applyDifficultyToEnemy(enemy Combatant) Combatant {
	mods := g.difficultySettings()
	if mods.EnemyHP > 0 {
		enemy.HP = int(math.Round(float64(enemy.HP) * mods.EnemyHP))
		if enemy.HP < 1 {
			enemy.HP = 1
		}
	}
	if mods.EnemyDamage > 0 {
		enemy.ATK = int(math.Round(float64(enemy.ATK) * mods.EnemyDamage))
		if enemy.ATK < 1 {
			enemy.ATK = 1
		}
	}
	return enemy
}
