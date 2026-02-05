package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	goruntime "runtime"
	"strings"
	"time"

	"github.com/mattn/go-runewidth"
)

type doctorStatus string

const (
	doctorPass doctorStatus = "pass"
	doctorWarn doctorStatus = "warn"
	doctorFail doctorStatus = "fail"
	doctorInfo doctorStatus = "info"
)

type doctorCheck struct {
	Name   string       `json:"name"`
	Status doctorStatus `json:"status"`
	Detail string       `json:"detail"`
}

type doctorReport struct {
	Timestamp string        `json:"timestamp"`
	OS        string        `json:"os"`
	Arch      string        `json:"arch"`
	Checks    []doctorCheck `json:"checks"`
}

func runDoctor(args []string) error {
	fs := flag.NewFlagSet("doctor", flag.ContinueOnError)
	jsonOutput := fs.Bool("json", false, "emit JSON output")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if len(fs.Args()) > 0 {
		return errors.New("doctor does not accept positional arguments")
	}

	report := doctorReport{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		OS:        goruntime.GOOS,
		Arch:      goruntime.GOARCH,
		Checks:    collectDoctorChecks(),
	}

	if *jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(report); err != nil {
			return err
		}
	} else {
		printDoctorReport(report)
	}

	_, _, fails, _ := doctorSummary(report.Checks)
	if fails > 0 {
		return fmt.Errorf("doctor found %d failing checks", fails)
	}
	return nil
}

func collectDoctorChecks() []doctorCheck {
	term := strings.TrimSpace(os.Getenv("TERM"))
	colorterm := strings.TrimSpace(os.Getenv("COLORTERM"))

	checks := []doctorCheck{
		checkTerminal(term),
		checkTrueColor(term, colorterm),
		checkMouse(term),
		checkUnicodeWidth(),
		checkKittyGraphics(term),
		checkSixelGraphics(term),
		checkAccessibility(),
	}
	return checks
}

func printDoctorReport(report doctorReport) {
	fmt.Printf("FluffyUI Doctor (%s)\n", report.Timestamp)
	fmt.Printf("Environment: %s/%s TERM=%q COLORTERM=%q\n", report.OS, report.Arch, os.Getenv("TERM"), os.Getenv("COLORTERM"))
	fmt.Println()
	for _, check := range report.Checks {
		fmt.Printf("[%s] %-14s %s\n", strings.ToUpper(string(check.Status)), check.Name, check.Detail)
	}
	fmt.Println()
	pass, warn, fail, info := doctorSummary(report.Checks)
	fmt.Printf("Summary: %d pass, %d warn, %d fail, %d info\n", pass, warn, fail, info)
}

func doctorSummary(checks []doctorCheck) (pass, warn, fail, info int) {
	for _, check := range checks {
		switch check.Status {
		case doctorPass:
			pass++
		case doctorWarn:
			warn++
		case doctorFail:
			fail++
		default:
			info++
		}
	}
	return pass, warn, fail, info
}

func checkTerminal(term string) doctorCheck {
	if term == "" {
		return doctorCheck{
			Name:   "Terminal",
			Status: doctorFail,
			Detail: "TERM is not set",
		}
	}
	if strings.EqualFold(term, "dumb") {
		return doctorCheck{
			Name:   "Terminal",
			Status: doctorFail,
			Detail: "TERM=dumb does not support interactive TUIs",
		}
	}
	return doctorCheck{
		Name:   "Terminal",
		Status: doctorPass,
		Detail: "TERM=" + term,
	}
}

func checkTrueColor(term string, colorterm string) doctorCheck {
	if detectTrueColor(term, colorterm) {
		detail := "true color detected"
		if colorterm != "" {
			detail += " via COLORTERM=" + colorterm
		} else {
			detail += " via TERM=" + term
		}
		return doctorCheck{Name: "TrueColor", Status: doctorPass, Detail: detail}
	}
	return doctorCheck{
		Name:   "TrueColor",
		Status: doctorWarn,
		Detail: "true color not detected; gradients may be downgraded",
	}
}

func checkMouse(term string) doctorCheck {
	if detectMouseSupport(term) {
		return doctorCheck{
			Name:   "Mouse",
			Status: doctorPass,
			Detail: "mouse-capable terminal family detected",
		}
	}
	return doctorCheck{
		Name:   "Mouse",
		Status: doctorWarn,
		Detail: "mouse support uncertain for TERM=" + term,
	}
}

func checkUnicodeWidth() doctorCheck {
	ascii := runewidth.StringWidth("A")
	wide := runewidth.StringWidth("界")
	emoji := runewidth.StringWidth("🙂")
	combining := runewidth.StringWidth("é")

	if ascii == 1 && wide == 2 && combining == 1 {
		if emoji == 2 {
			return doctorCheck{
				Name:   "Unicode",
				Status: doctorPass,
				Detail: "width checks passed (ASCII=1, CJK=2, emoji=2)",
			}
		}
		return doctorCheck{
			Name:   "Unicode",
			Status: doctorWarn,
			Detail: fmt.Sprintf("emoji width=%d (expected 2 in most terminals)", emoji),
		}
	}

	return doctorCheck{
		Name:   "Unicode",
		Status: doctorWarn,
		Detail: fmt.Sprintf("width mismatch (ASCII=%d, CJK=%d, combining=%d)", ascii, wide, combining),
	}
}

func checkKittyGraphics(term string) doctorCheck {
	if detectKittySupport(term) {
		return doctorCheck{
			Name:   "Kitty",
			Status: doctorPass,
			Detail: "kitty graphics protocol likely available",
		}
	}
	return doctorCheck{
		Name:   "Kitty",
		Status: doctorInfo,
		Detail: "kitty graphics protocol not detected",
	}
}

func checkSixelGraphics(term string) doctorCheck {
	if detectSixelSupport(term) {
		return doctorCheck{
			Name:   "Sixel",
			Status: doctorPass,
			Detail: "sixel-capable terminal detected",
		}
	}
	return doctorCheck{
		Name:   "Sixel",
		Status: doctorInfo,
		Detail: "sixel support not detected",
	}
}

func checkAccessibility() doctorCheck {
	switch goruntime.GOOS {
	case "linux":
		if strings.TrimSpace(os.Getenv("NO_AT_BRIDGE")) == "1" {
			return doctorCheck{
				Name:   "A11y",
				Status: doctorWarn,
				Detail: "NO_AT_BRIDGE=1 disables AT-SPI bridge integration",
			}
		}
		path, err := exec.LookPath("dbus-send")
		if err != nil {
			return doctorCheck{
				Name:   "A11y",
				Status: doctorWarn,
				Detail: "dbus-send not found; cannot probe org.a11y.Bus",
			}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		cmd := exec.CommandContext(
			ctx,
			path,
			"--session",
			"--dest=org.a11y.Bus",
			"/org/a11y/bus",
			"org.freedesktop.DBus.Peer.Ping",
		)
		if err := cmd.Run(); err != nil {
			return doctorCheck{
				Name:   "A11y",
				Status: doctorWarn,
				Detail: "org.a11y.Bus is unreachable on the session bus",
			}
		}
		return doctorCheck{
			Name:   "A11y",
			Status: doctorPass,
			Detail: "AT-SPI bus is reachable",
		}
	case "darwin":
		return doctorCheck{
			Name:   "A11y",
			Status: doctorPass,
			Detail: "NSAccessibility APIs available on macOS",
		}
	case "windows":
		return doctorCheck{
			Name:   "A11y",
			Status: doctorPass,
			Detail: "UI Automation APIs available on Windows",
		}
	default:
		return doctorCheck{
			Name:   "A11y",
			Status: doctorInfo,
			Detail: "no OS-specific accessibility probe for this platform",
		}
	}
}

func detectTrueColor(term string, colorterm string) bool {
	term = strings.ToLower(strings.TrimSpace(term))
	colorterm = strings.ToLower(strings.TrimSpace(colorterm))
	return strings.Contains(colorterm, "truecolor") ||
		strings.Contains(colorterm, "24bit") ||
		strings.Contains(term, "truecolor") ||
		strings.Contains(term, "direct")
}

func detectMouseSupport(term string) bool {
	term = strings.ToLower(strings.TrimSpace(term))
	if term == "" || term == "dumb" {
		return false
	}
	families := []string{
		"xterm",
		"screen",
		"tmux",
		"rxvt",
		"kitty",
		"alacritty",
		"wezterm",
		"foot",
		"st-",
		"vte",
	}
	for _, family := range families {
		if strings.Contains(term, family) {
			return true
		}
	}
	return false
}

func detectKittySupport(term string) bool {
	term = strings.ToLower(strings.TrimSpace(term))
	if strings.Contains(term, "kitty") {
		return true
	}
	return strings.TrimSpace(os.Getenv("KITTY_WINDOW_ID")) != ""
}

func detectSixelSupport(term string) bool {
	term = strings.ToLower(strings.TrimSpace(term))
	if strings.Contains(term, "sixel") {
		return true
	}
	if strings.Contains(term, "mlterm") {
		return true
	}
	return false
}
