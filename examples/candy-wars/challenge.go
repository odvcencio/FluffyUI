package main

func (g *Game) ChallengeRival() bool {
	if g.GameOver.Get() || g.InCombat() || g.ShowEvent.Get() {
		return false
	}
	g.addHeat(5, true, false)
	g.StartCombat(enemyRivalTrader)
	if g.InCombat() {
		g.appendCombatLog("You challenge a rival trader!")
		return true
	}
	return false
}
