package main

import "time"

// Achievement represents an unlockable achievement.
type Achievement struct {
	ID          string
	Name        string
	Description string
	Hidden      bool
	HellOnly    bool
	ProgressMax int
	Check       func(g *Game) (unlocked bool, progress int)
}

// AchievementProgress tracks player's achievement state.
type AchievementProgress struct {
	ID         string    `json:"id"`
	Unlocked   bool      `json:"unlocked"`
	UnlockedAt time.Time `json:"unlocked_at,omitempty"`
	Progress   int       `json:"progress"`
}

var Achievements = []Achievement{
	// Trading
	{ID: "first_sale", Name: "First Sale", Description: "Sell your first candy"},
	{ID: "debt_free", Name: "Debt Free", Description: "Pay off your starting debt"},
	{ID: "bulk_buyer", Name: "Bulk Buyer", Description: "Buy 50+ candy in one transaction"},
	{ID: "monopolist", Name: "Monopolist", Description: "Own all stock of one candy type"},
	{ID: "price_gouger", Name: "Price Gouger", Description: "Sell candy at 200%+ markup"},

	// Combat
	{ID: "first_blood", Name: "First Blood", Description: "Win your first fight"},
	{ID: "untouchable", Name: "Untouchable", Description: "Win a fight without taking damage", ProgressMax: 10},
	{ID: "bully_hunter", Name: "Bully Hunter", Description: "Defeat 20 bullies", ProgressMax: 20},
	{ID: "boss_slayer", Name: "Boss Slayer", Description: "Defeat the loan shark's goon"},

	// Survival
	{ID: "marathon", Name: "Marathon", Description: "Survive 20 days", ProgressMax: 20},
	{ID: "speed_run", Name: "Speed Run", Description: "Win in under 15 days"},
	{ID: "millionaire", Name: "Millionaire", Description: "Accumulate $10,000 total worth"},
	{ID: "low_heat", Name: "Under the Radar", Description: "Win with heat never exceeding 25"},

	// Secret
	{ID: "black_market_vip", Name: "???", Description: "???", Hidden: true},
	{ID: "all_weapons", Name: "???", Description: "???", Hidden: true},

	// Hell-exclusive
	{ID: "inferno_trader", Name: "Inferno Trader", Description: "Win a Hell run", HellOnly: true},
	{ID: "untouchable_demon", Name: "Untouchable Demon", Description: "Win Hell without taking damage", HellOnly: true},
	{ID: "speed_demon", Name: "Speed Demon", Description: "Win Hell in under 15 days", HellOnly: true},
	{ID: "debt_crusher", Name: "Debt Crusher", Description: "Pay off Hell debt by Day 10", HellOnly: true},
	{ID: "diamond_hands", Name: "Diamond Hands", Description: "Reach Diamond rank", HellOnly: true},
}

// GetAchievementProgress returns current progress for all achievements.
func (g *Game) GetAchievementProgress() []AchievementProgress {
	progress := make([]AchievementProgress, len(Achievements))

	for i, ach := range Achievements {
		prog := AchievementProgress{ID: ach.ID}

		// Check if already unlocked in meta
		if g.meta != nil {
			for _, unlocked := range g.meta.Achievements {
				if unlocked.ID == ach.ID {
					prog = unlocked
					break
				}
			}
		}

		// Check current progress if not unlocked
		if !prog.Unlocked && ach.Check != nil {
			unlocked, current := ach.Check(g)
			prog.Progress = current
			if unlocked {
				prog.Unlocked = true
				prog.UnlockedAt = time.Now()
			}
		}

		progress[i] = prog
	}

	return progress
}

// AchievementPercentage returns completion percentage.
func (g *Game) AchievementPercentage() float64 {
	unlocked, total := g.AchievementCount()
	if total == 0 {
		return 0
	}
	return float64(unlocked) / float64(total) * 100
}

// CheckAndUnlockAchievements checks for newly unlocked achievements.
func (g *Game) CheckAndUnlockAchievements() []Achievement {
	var newlyUnlocked []Achievement

	for _, ach := range Achievements {
		if ach.Check == nil {
			continue
		}

		// Skip if already unlocked
		alreadyUnlocked := false
		if g.meta != nil {
			for _, prog := range g.meta.Achievements {
				if prog.ID == ach.ID && prog.Unlocked {
					alreadyUnlocked = true
					break
				}
			}
		}
		if alreadyUnlocked {
			continue
		}

		unlocked, _ := ach.Check(g)
		if unlocked {
			newlyUnlocked = append(newlyUnlocked, ach)
			// Record in meta
			if g.meta != nil {
				g.meta.Achievements = append(g.meta.Achievements, AchievementProgress{
					ID:         ach.ID,
					Unlocked:   true,
					UnlockedAt: time.Now(),
				})
			}
		}
	}

	return newlyUnlocked
}
