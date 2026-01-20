package main

import (
	"fmt"
	"strconv"
)

func truncPad(s string, width int) string {
	if len(s) > width {
		return s[:width]
	}
	for len(s) < width {
		s += " "
	}
	return s
}

func shortName(name string) string {
	if len(name) > 6 {
		return name[:6]
	}
	return name
}

func splitLines(s string, maxWidth int) []string {
	var lines []string
	current := ""
	for _, r := range s {
		if r == '\n' {
			lines = append(lines, current)
			current = ""
			continue
		}
		current += string(r)
		if len(current) >= maxWidth {
			lines = append(lines, current)
			current = ""
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

func parseInputAmount(text string) int {
	if text == "" {
		return 0
	}
	value, err := strconv.Atoi(text)
	if err != nil || value < 0 {
		return 0
	}
	return value
}

func formatScheduleLine(game *Game) string {
	schedule := game.Schedule
	return fmt.Sprintf(
		"Schedule: P1 %s | P2 %s | P3 %s | P4 %s",
		locationShortName(Locations[schedule[0]].Name),
		locationShortName(Locations[schedule[1]].Name),
		locationShortName(Locations[schedule[2]].Name),
		locationShortName(Locations[schedule[3]].Name),
	)
}

func formatStatusLine(game *Game) string {
	period := game.currentPeriod()
	start, end := periodHourRange(period)
	scheduled := game.Schedule[period]
	loc := game.Location.Get()
	return fmt.Sprintf(
		"Period %d (%d-%d) %s | At %s [%s]",
		period+1,
		start,
		end,
		locationShortName(Locations[scheduled].Name),
		locationShortName(Locations[loc].Name),
		scheduleStatusLabel(game.scheduleStatus()),
	)
}

func scheduleStatusLabel(status ScheduleStatus) string {
	switch status {
	case statusBlendingIn:
		return "BLEND"
	case statusOffCampus:
		return "OFF"
	default:
		return "WRONG"
	}
}

func locationShortName(name string) string {
	switch name {
	case "Cafeteria":
		return "Cafe"
	case "Gymnasium":
		return "Gym"
	case "Library":
		return "Library"
	case "Playground":
		return "Play"
	case "Art Room":
		return "Art"
	case "Music Hall":
		return "Music"
	default:
		return name
	}
}
