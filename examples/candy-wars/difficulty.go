package main

// Difficulty represents a game difficulty level.
type Difficulty string

const (
	DifficultyNormal    Difficulty = "normal"
	DifficultyNightmare Difficulty = "nightmare"
	DifficultyHell      Difficulty = "hell"
)

// DifficultyModifiers contains all gameplay modifiers.
type DifficultyModifiers struct {
	Name        string
	Description string
	Tagline     string

	PriceVolatility float64
	EnemyDamage     float64
	EnemyHP         float64
	HeatGain        float64
	HeatDecay       float64
	DebtInterest    int
	StartingDebt    int

	HellBonusesApply bool
}

var DifficultySettings = map[Difficulty]DifficultyModifiers{
	DifficultyNormal: {
		Name:             "Normal",
		Description:      "The standard experience.",
		Tagline:          "Learn the ropes, pay your debts.",
		PriceVolatility:  1.0,
		EnemyDamage:      1.0,
		EnemyHP:          1.0,
		HeatGain:         1.0,
		HeatDecay:        1.0,
		DebtInterest:     3,
		StartingDebt:     500,
		HellBonusesApply: false,
	},
	DifficultyNightmare: {
		Name:             "Nightmare",
		Description:      "Higher prices, aggressive enemies, faster heat.",
		Tagline:          "The teachers are onto you.",
		PriceVolatility:  1.5,
		EnemyDamage:      1.5,
		EnemyHP:          1.25,
		HeatGain:         1.5,
		HeatDecay:        0.75,
		DebtInterest:     5,
		StartingDebt:     750,
		HellBonusesApply: false,
	},
	DifficultyHell: {
		Name:             "Hell",
		Description:      "Brutal economy, relentless enemies, no mercy.",
		Tagline:          "Only the worthy survive.",
		PriceVolatility:  2.0,
		EnemyDamage:      2.0,
		EnemyHP:          1.5,
		HeatGain:         2.0,
		HeatDecay:        0.5,
		DebtInterest:     8,
		StartingDebt:     1000,
		HellBonusesApply: true,
	},
}

// HellRank represents player progression in Hell mode.
type HellRank string

const (
	HellRankUnranked HellRank = "unranked"
	HellRankBronze   HellRank = "bronze"
	HellRankSilver   HellRank = "silver"
	HellRankGold     HellRank = "gold"
	HellRankPlatinum HellRank = "platinum"
	HellRankDiamond  HellRank = "diamond"
)

// HellRankInfo describes a rank and its requirements/bonuses.
type HellRankInfo struct {
	Rank         HellRank
	Name         string
	WinsRequired int
	BonusType    string
	BonusDesc    string
}

var HellRanks = []HellRankInfo{
	{HellRankUnranked, "Unranked", 0, "", ""},
	{HellRankBronze, "Bronze", 5, "sell_bonus", "+5% sell prices in Hell"},
	{HellRankSilver, "Silver", 15, "hp_bonus", "+10 starting HP in Hell"},
	{HellRankGold, "Gold", 30, "heat_decay", "Heat decays 10% faster in Hell"},
	{HellRankPlatinum, "Platinum", 50, "lucky_charm", "Start with Lucky Charm in Hell"},
	{HellRankDiamond, "Diamond", 100, "???", "???"},
}

// GetHellRank returns the player's current Hell rank.
func (g *Game) GetHellRank() HellRankInfo {
	wins := 0
	if g.meta != nil {
		wins = g.meta.HellWins
	}

	rank := HellRanks[0]
	for _, r := range HellRanks {
		if wins >= r.WinsRequired {
			rank = r
		}
	}
	return rank
}

// GetHellRankProgress returns progress toward next rank (current wins, required for next).
func (g *Game) GetHellRankProgress() (current, required int) {
	wins := 0
	if g.meta != nil {
		wins = g.meta.HellWins
	}

	currentRank := g.GetHellRank()
	for i, r := range HellRanks {
		if r.Rank == currentRank.Rank && i < len(HellRanks)-1 {
			return wins, HellRanks[i+1].WinsRequired
		}
	}
	return wins, wins // Max rank
}

// ApplyHellBonuses applies Hell rank bonuses to game state.
func (g *Game) ApplyHellBonuses() {
	rank := g.GetHellRank()

	// Apply cumulative bonuses based on rank
	switch rank.Rank {
	case HellRankDiamond:
		fallthrough
	case HellRankPlatinum:
		// Start with Lucky Charm
		g.tradeBuffUses = 1
		g.tradeBuffBuyMult = 110
		g.tradeBuffSellMult = 110
		fallthrough
	case HellRankGold:
		// Heat decays faster (applied in game logic)
		fallthrough
	case HellRankSilver:
		// +10 starting HP
		g.HP.Set(g.HP.Get() + 10)
		fallthrough
	case HellRankBronze:
		// +5% sell prices (applied in game logic)
	}
}
