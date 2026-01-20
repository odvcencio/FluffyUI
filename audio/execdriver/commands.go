package execdriver

import (
	"os/exec"
	"runtime"
)

// DefaultCommand returns the first command candidate for this OS.
func DefaultCommand() (Command, bool) {
	candidates := DefaultCommandCandidates()
	if len(candidates) == 0 {
		return Command{}, false
	}
	return candidates[0], true
}

// DefaultCommandCandidates returns OS-specific command options.
func DefaultCommandCandidates() []Command {
	return commandsForOS(runtime.GOOS)
}

// DetectCommand returns the first available command in PATH.
func DetectCommand() (Command, bool) {
	for _, cmd := range DefaultCommandCandidates() {
		if cmd.Path == "" {
			continue
		}
		if _, err := exec.LookPath(cmd.Path); err == nil {
			return cmd, true
		}
	}
	return Command{}, false
}

func commandsForOS(goos string) []Command {
	switch goos {
	case "darwin":
		return []Command{
			{Path: "afplay", Args: []string{"{{path}}"}},
		}
	case "linux":
		return []Command{
			{Path: "paplay", Args: []string{"{{path}}"}},
			{Path: "aplay", Args: []string{"{{path}}"}},
			{Path: "ffplay", Args: []string{"-nodisp", "-autoexit", "{{path}}"}},
		}
	case "windows":
		return []Command{
			{Path: "powershell", Args: []string{"-c", "(New-Object Media.SoundPlayer '{{path}}').PlaySync()"}},
		}
	default:
		return nil
	}
}
