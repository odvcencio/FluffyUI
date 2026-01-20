package main

import (
	"fmt"
	"time"

	"github.com/odvcencio/fluffy-ui/runtime"
	"github.com/odvcencio/fluffy-ui/terminal"
)

type statsTab int

const (
	statsCareer statsTab = iota
	statsAchievements
	statsHistory
)

func (v *GameView) openStatsDialog() {
	if v.game.GameOver.Get() || v.game.InCombat() {
		return
	}
	v.showTrade = false
	v.showBank = false
	v.showLoan = false
	v.showCraft = false
	v.showItems = false
	v.showUpgrades = false
	v.showIntel = false
	v.showGear = false
	v.showBlackMarket = false
	v.showStash = false
	v.tradeInput.Blur()
	v.bankInput.Blur()
	v.loanInput.Blur()
	v.craftInput.Blur()
	v.stashInput.Blur()
	v.blackMarketInput.Blur()

	v.showStats = true
	v.Invalidate()
}

func (v *GameView) statsDialogRect(bounds runtime.Rect, lines int) runtime.Rect {
	dialogW := 70
	dialogH := lines + 6
	if dialogH < 12 {
		dialogH = 12
	}
	x := bounds.X + (bounds.Width-dialogW)/2
	y := bounds.Y + (bounds.Height-dialogH)/2
	return runtime.Rect{X: x, Y: y, Width: dialogW, Height: dialogH}
}

func (v *GameView) renderStatsDialog(ctx runtime.RenderContext) {
	lines := v.statsLines()
	bounds := v.Bounds()
	rect := v.statsDialogRect(bounds, len(lines))
	ctx.Buffer.Fill(rect, ' ', v.style)
	ctx.Buffer.DrawBox(rect, v.accentStyle)

	title := " CAREER STATS "
	if v.statsTab == statsAchievements {
		title = " ACHIEVEMENTS "
	} else if v.statsTab == statsHistory {
		title = " RUN HISTORY "
	}
	ctx.Buffer.SetString(rect.X+2, rect.Y, title, v.accentStyle)

	for i, line := range lines {
		if i >= rect.Height-4 {
			break
		}
		ctx.Buffer.SetString(rect.X+2, rect.Y+2+i, truncPad(line, rect.Width-4), v.style)
	}

	ctx.Buffer.SetString(rect.X+2, rect.Y+rect.Height-2, "[1] Career  [2] Achievements  [3] History  [Esc] Close", v.dimStyle)
}

func (v *GameView) handleStatsInput(msg runtime.Message) runtime.HandleResult {
	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		return runtime.Handled()
	}
	switch key.Key {
	case terminal.KeyEscape:
		v.showStats = false
		v.Invalidate()
		return runtime.Handled()
	}
	switch key.Rune {
	case '1':
		v.statsTab = statsCareer
		v.Invalidate()
		return runtime.Handled()
	case '2':
		v.statsTab = statsAchievements
		v.Invalidate()
		return runtime.Handled()
	case '3':
		v.statsTab = statsHistory
		v.Invalidate()
		return runtime.Handled()
	}
	return runtime.Handled()
}

func (v *GameView) statsLines() []string {
	if v.statsTab == statsAchievements {
		return v.achievementLines()
	}
	if v.statsTab == statsHistory {
		return v.historyLines()
	}
	return v.careerLines()
}

func (v *GameView) careerLines() []string {
	meta := v.game.meta
	if meta == nil {
		meta = defaultMeta()
	}
	winRate := 0
	if meta.TotalRuns > 0 {
		winRate = meta.Wins * 100 / meta.TotalRuns
	}
	avgRun := int64(0)
	if meta.TotalRuns > 0 {
		avgRun = meta.TotalPlaytimeSeconds / int64(meta.TotalRuns)
	}
	avgWin := int64(0)
	if meta.Wins > 0 {
		avgWin = meta.TotalWinSeconds / int64(meta.Wins)
	}
	avgLoss := int64(0)
	if meta.Losses > 0 {
		avgLoss = meta.TotalLossSeconds / int64(meta.Losses)
	}

	lines := []string{
		fmt.Sprintf("Total Runs: %d  Wins: %d  Losses: %d  Win Rate: %d%%", meta.TotalRuns, meta.Wins, meta.Losses, winRate),
		fmt.Sprintf("Best Worth: $%d  Trades: %d ($%d)", meta.BestWorth, meta.TotalTradeCount, meta.TotalTradeValue),
		fmt.Sprintf("Crafted: %d  Enemies: %d  Loan Paid: $%d", meta.TotalCrafted, meta.TotalEnemiesDefeated, meta.TotalLoanPaid),
		fmt.Sprintf("Playtime: %s  Avg Run: %s", formatDuration(meta.TotalPlaytimeSeconds), formatDuration(avgRun)),
		fmt.Sprintf("Longest Run: %s  Shortest Run: %s", formatDuration(meta.LongestRunSeconds), formatDuration(meta.ShortestRunSeconds)),
		fmt.Sprintf("Fastest Win: %s  Slowest Win: %s", formatDuration(meta.FastestWinSeconds), formatDuration(meta.SlowestWinSeconds)),
		fmt.Sprintf("Fastest Loss: %s  Slowest Loss: %s", formatDuration(meta.FastestLossSeconds), formatDuration(meta.SlowestLossSeconds)),
		fmt.Sprintf("Avg Win: %s  Avg Loss: %s", formatDuration(avgWin), formatDuration(avgLoss)),
		fmt.Sprintf("Last Run: %s ago (%s)", formatSince(meta.LastRunEndedAtUnixSecs), formatDuration(meta.LastRunDurationSeconds)),
		fmt.Sprintf("Unlocks: Bank %s  Craft %s  Loans %s  Market %s",
			unlockFlag(meta.Unlocks.Bank),
			unlockFlag(meta.Unlocks.Crafting),
			unlockFlag(meta.Unlocks.StandardLoans),
			unlockFlag(meta.Unlocks.BlackMarket),
		),
	}
	return lines
}

func (v *GameView) achievementLines() []string {
	meta := v.game.meta
	if meta == nil {
		meta = defaultMeta()
	}
	lines := make([]string, 0, len(achievementDefs)+1)
	unlocked, total := v.game.AchievementCount()
	lines = append(lines, fmt.Sprintf("Unlocked: %d/%d", unlocked, total))
	for _, def := range achievementDefs {
		status := "LOCKED"
		if meta.Achievements[def.ID] {
			status = "UNLOCKED"
		}
		lines = append(lines, fmt.Sprintf("[%s] %s", status, def.Name))
	}
	return lines
}

func (v *GameView) historyLines() []string {
	meta := v.game.meta
	if meta == nil {
		meta = defaultMeta()
	}
	lines := []string{"Recent Runs:"}
	if len(meta.RunHistory) == 0 {
		lines = append(lines, "  none yet")
	} else {
		start := 0
		if len(meta.RunHistory) > 5 {
			start = len(meta.RunHistory) - 5
		}
		for _, record := range meta.RunHistory[start:] {
			lines = append(lines, "  "+formatRunRecord(record))
		}
	}
	lines = append(lines, "")
	lines = append(lines, "Recent Trades:")
	if len(meta.TradeHistory) == 0 {
		lines = append(lines, "  none yet")
	} else {
		start := 0
		if len(meta.TradeHistory) > 5 {
			start = len(meta.TradeHistory) - 5
		}
		for _, record := range meta.TradeHistory[start:] {
			lines = append(lines, "  "+formatTradeRecord(record))
		}
	}
	return lines
}

func formatDuration(seconds int64) string {
	if seconds <= 0 {
		return "0m"
	}
	d := time.Duration(seconds) * time.Second
	hours := int(d.Hours())
	mins := int(d.Minutes()) % 60
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, mins)
	}
	secs := int(d.Seconds()) % 60
	if mins > 0 {
		return fmt.Sprintf("%dm %ds", mins, secs)
	}
	return fmt.Sprintf("%ds", secs)
}

func formatSince(unixSeconds int64) string {
	if unixSeconds <= 0 {
		return "never"
	}
	diff := time.Since(time.Unix(unixSeconds, 0))
	if diff < 0 {
		diff = 0
	}
	if diff.Hours() >= 24 {
		return fmt.Sprintf("%dd", int(diff.Hours()/24))
	}
	if diff.Hours() >= 1 {
		return fmt.Sprintf("%dh", int(diff.Hours()))
	}
	if diff.Minutes() >= 1 {
		return fmt.Sprintf("%dm", int(diff.Minutes()))
	}
	return fmt.Sprintf("%ds", int(diff.Seconds()))
}

func unlockFlag(on bool) string {
	if on {
		return "Y"
	}
	return "N"
}

func formatRunRecord(record RunRecord) string {
	when := time.Unix(record.Timestamp, 0).Format("2006-01-02")
	return fmt.Sprintf("%s %s Day %d Worth $%d Debt $%d %s",
		when,
		record.Outcome,
		record.Day,
		record.NetWorth,
		record.Debt,
		formatDuration(record.DurationSeconds),
	)
}

func formatTradeRecord(record TradeRecord) string {
	when := time.Unix(record.Timestamp, 0).Format("15:04")
	return fmt.Sprintf("%s %s %d %s @ $%d (%s)",
		when,
		record.Action,
		record.Qty,
		record.Item,
		record.PriceEach,
		record.Location,
	)
}
