package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const metaSavePath = "examples/candy-wars/candy-wars-meta.json"

type MetaUnlocks struct {
	Bank          bool `json:"bank"`
	Crafting      bool `json:"crafting"`
	StandardLoans bool `json:"standard_loans"`
	BlackMarket   bool `json:"black_market"`
}

type MetaProgress struct {
	Unlocks                MetaUnlocks           `json:"unlocks"`
	Achievements           []AchievementProgress `json:"achievements"`
	TotalRuns              int                   `json:"total_runs"`
	Wins                   int                   `json:"wins"`
	Losses                 int                   `json:"losses"`
	BestWorth              int                   `json:"best_worth"`
	HellWins               int                   `json:"hell_wins"`
	TotalTradeValue        int                   `json:"total_trade_value"`
	TotalTradeCount        int                   `json:"total_trade_count"`
	TotalCrafted           int                   `json:"total_crafted"`
	TotalEnemiesDefeated   int                   `json:"total_enemies_defeated"`
	TotalLoanPaid          int                   `json:"total_loan_paid"`
	TotalPlaytimeSeconds   int64                 `json:"total_playtime_seconds"`
	TotalWinSeconds        int64                 `json:"total_win_seconds"`
	TotalLossSeconds       int64                 `json:"total_loss_seconds"`
	FastestWinSeconds      int64                 `json:"fastest_win_seconds"`
	SlowestWinSeconds      int64                 `json:"slowest_win_seconds"`
	FastestLossSeconds     int64                 `json:"fastest_loss_seconds"`
	SlowestLossSeconds     int64                 `json:"slowest_loss_seconds"`
	LongestRunSeconds      int64                 `json:"longest_run_seconds"`
	ShortestRunSeconds     int64                 `json:"shortest_run_seconds"`
	LastRunDurationSeconds int64                 `json:"last_run_duration_seconds"`
	LastRunEndedAtUnixSecs int64                 `json:"last_run_ended_at_unix_secs"`
	RunHistory             []RunRecord           `json:"run_history"`
	TradeHistory           []TradeRecord         `json:"trade_history"`
}

type RunRecord struct {
	Timestamp       int64  `json:"timestamp"`
	Outcome         string `json:"outcome"`
	Day             int    `json:"day"`
	NetWorth        int    `json:"net_worth"`
	Debt            int    `json:"debt"`
	DurationSeconds int64  `json:"duration_seconds"`
}

type TradeRecord struct {
	Timestamp int64  `json:"timestamp"`
	Action    string `json:"action"`
	Item      string `json:"item"`
	Qty       int    `json:"qty"`
	PriceEach int    `json:"price_each"`
	Total     int    `json:"total"`
	Location  string `json:"location"`
}

func defaultMeta() *MetaProgress {
	return &MetaProgress{
		Achievements: make([]AchievementProgress, 0),
		RunHistory:   make([]RunRecord, 0),
		TradeHistory: make([]TradeRecord, 0),
	}
}

func LoadMeta() *MetaProgress {
	data, err := os.ReadFile(metaSavePath)
	if err != nil {
		return defaultMeta()
	}
	var meta MetaProgress
	if err := json.Unmarshal(data, &meta); err != nil {
		return defaultMeta()
	}
	if meta.Achievements == nil {
		meta.Achievements = make([]AchievementProgress, 0)
	}
	if meta.RunHistory == nil {
		meta.RunHistory = make([]RunRecord, 0)
	}
	if meta.TradeHistory == nil {
		meta.TradeHistory = make([]TradeRecord, 0)
	}
	return &meta
}

func SaveMeta(meta *MetaProgress) {
	if meta == nil {
		return
	}
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return
	}
	dir := filepath.Dir(metaSavePath)
	if dir != "." {
		_ = os.MkdirAll(dir, 0o755)
	}
	_ = os.WriteFile(metaSavePath, data, 0o644)
}

func (g *Game) applyMeta() {
	if g.meta == nil {
		g.meta = defaultMeta()
	}
	g.Runs = g.meta.TotalRuns
	g.Wins = g.meta.Wins
	g.Losses = g.meta.Losses
	g.BestWorth = g.meta.BestWorth
}

func (g *Game) saveMeta() {
	if g.meta == nil {
		return
	}
	SaveMeta(g.meta)
}

func (g *Game) recordPlaytime(win bool) {
	if g.meta == nil || g.runStart.IsZero() {
		return
	}
	elapsed := time.Since(g.runStart)
	if elapsed < 0 {
		return
	}
	seconds := int64(elapsed.Seconds())
	if seconds < 0 {
		seconds = 0
	}
	g.meta.TotalPlaytimeSeconds += seconds
	g.meta.LastRunDurationSeconds = seconds
	if g.meta.LongestRunSeconds < seconds {
		g.meta.LongestRunSeconds = seconds
	}
	if g.meta.ShortestRunSeconds == 0 || seconds < g.meta.ShortestRunSeconds {
		g.meta.ShortestRunSeconds = seconds
	}
	if win {
		g.meta.TotalWinSeconds += seconds
		if g.meta.FastestWinSeconds == 0 || seconds < g.meta.FastestWinSeconds {
			g.meta.FastestWinSeconds = seconds
		}
		if seconds > g.meta.SlowestWinSeconds {
			g.meta.SlowestWinSeconds = seconds
		}
	} else {
		g.meta.TotalLossSeconds += seconds
		if g.meta.FastestLossSeconds == 0 || seconds < g.meta.FastestLossSeconds {
			g.meta.FastestLossSeconds = seconds
		}
		if seconds > g.meta.SlowestLossSeconds {
			g.meta.SlowestLossSeconds = seconds
		}
	}
	g.meta.LastRunEndedAtUnixSecs = time.Now().Unix()
}

func (g *Game) recordTrade(qty int, value int) {
	if g.meta == nil || qty <= 0 || value <= 0 {
		return
	}
	g.meta.TotalTradeCount += qty
	g.meta.TotalTradeValue += value
	g.updateUnlocks()
	g.saveMeta()
}

func (g *Game) recordTradeHistory(action, item string, qty, priceEach, total int, location string) {
	if g.meta == nil || qty <= 0 || total <= 0 {
		return
	}
	record := TradeRecord{
		Timestamp: time.Now().Unix(),
		Action:    action,
		Item:      item,
		Qty:       qty,
		PriceEach: priceEach,
		Total:     total,
		Location:  location,
	}
	g.meta.TradeHistory = append(g.meta.TradeHistory, record)
	if len(g.meta.TradeHistory) > 25 {
		g.meta.TradeHistory = g.meta.TradeHistory[len(g.meta.TradeHistory)-25:]
	}
	g.saveMeta()
}

func (g *Game) recordCraft(qty int) {
	if g.meta == nil || qty <= 0 {
		return
	}
	g.meta.TotalCrafted += qty
	g.saveMeta()
}

func (g *Game) recordEnemyDefeat() {
	if g.meta == nil {
		return
	}
	g.meta.TotalEnemiesDefeated++
	g.updateUnlocks()
	g.saveMeta()
}

func (g *Game) recordLoanPaid(amount int) {
	if g.meta == nil || amount <= 0 {
		return
	}
	g.meta.TotalLoanPaid += amount
	g.updateUnlocks()
	g.saveMeta()
}

func (g *Game) recordRunHistory(win bool, totalWorth int, debt int, day int) {
	if g.meta == nil {
		return
	}
	outcome := "Loss"
	if win {
		outcome = "Win"
	}
	record := RunRecord{
		Timestamp:       time.Now().Unix(),
		Outcome:         outcome,
		Day:             day,
		NetWorth:        totalWorth,
		Debt:            debt,
		DurationSeconds: g.meta.LastRunDurationSeconds,
	}
	g.meta.RunHistory = append(g.meta.RunHistory, record)
	if len(g.meta.RunHistory) > 10 {
		g.meta.RunHistory = g.meta.RunHistory[len(g.meta.RunHistory)-10:]
	}
	g.saveMeta()
}

func (g *Game) checkAchievements() []string {
	unlocked := g.CheckAndUnlockAchievements()
	if len(unlocked) == 0 {
		return nil
	}
	names := make([]string, len(unlocked))
	for i, ach := range unlocked {
		names[i] = ach.Name
	}
	g.saveMeta()
	return names
}

func (g *Game) updateUnlocks() []string {
	if g.meta == nil {
		return nil
	}
	var unlocked []string
	profit := g.TotalWorth() - startingCash
	day := g.Day.Get()

	if !g.meta.Unlocks.Bank && (day >= 3 || profit >= 200) {
		g.meta.Unlocks.Bank = true
		unlocked = append(unlocked, "Bank Access")
	}
	if !g.meta.Unlocks.Crafting && (day >= 7 || profit >= 500) {
		g.meta.Unlocks.Crafting = true
		unlocked = append(unlocked, "Crafting Kit")
	}
	if !g.meta.Unlocks.StandardLoans && (day >= 10 || g.meta.TotalLoanPaid >= 300) {
		g.meta.Unlocks.StandardLoans = true
		unlocked = append(unlocked, "Standard Loans")
	}
	if !g.meta.Unlocks.BlackMarket && (day >= 15 || g.meta.TotalTradeValue >= 1000) {
		g.meta.Unlocks.BlackMarket = true
		unlocked = append(unlocked, "Black Market")
	}

	if len(unlocked) > 0 {
		g.saveMeta()
	}
	return unlocked
}

func (g *Game) bankUnlocked() bool {
	if g.meta == nil {
		return true
	}
	if g.meta.Unlocks.Bank {
		return true
	}
	return len(g.updateUnlocks()) > 0 && g.meta.Unlocks.Bank
}

func (g *Game) craftingUnlocked() bool {
	if g.meta == nil {
		return true
	}
	if g.meta.Unlocks.Crafting {
		return true
	}
	return len(g.updateUnlocks()) > 0 && g.meta.Unlocks.Crafting
}

func (g *Game) standardLoansUnlocked() bool {
	if g.meta == nil {
		return true
	}
	if g.meta.Unlocks.StandardLoans {
		return true
	}
	return len(g.updateUnlocks()) > 0 && g.meta.Unlocks.StandardLoans
}

func (g *Game) blackMarketUnlocked() bool {
	if g.meta == nil {
		return true
	}
	if g.meta.Unlocks.BlackMarket {
		return true
	}
	return len(g.updateUnlocks()) > 0 && g.meta.Unlocks.BlackMarket
}

func (g *Game) loanTierUnlocked(tier int) bool {
	switch tier {
	case 0:
		return true
	case 1:
		return g.standardLoansUnlocked()
	case 2:
		return g.blackMarketUnlocked()
	default:
		return false
	}
}

func (g *Game) AchievementCount() (int, int) {
	total := len(Achievements)
	if g.meta == nil || g.meta.Achievements == nil {
		return 0, total
	}
	unlocked := 0
	for _, got := range g.meta.Achievements {
		if got.Unlocked {
			unlocked++
		}
	}
	return unlocked, total
}
