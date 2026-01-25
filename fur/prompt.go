package fur

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"golang.org/x/term"
)

// Ask prompts for text input.
func Ask(prompt string) (string, error) {
	return Prompt(prompt, PromptOpts{})
}

// AskDefault prompts with a default value.
func AskDefault(prompt, def string) (string, error) {
	return Prompt(prompt, PromptOpts{Default: def})
}

// AskPassword prompts with hidden input.
func AskPassword(prompt string) (string, error) {
	return Prompt(prompt, PromptOpts{Password: true})
}

// Confirm prompts for yes/no.
func Confirm(prompt string) (bool, error) {
	for {
		answer, err := Prompt(prompt, PromptOpts{Default: "n"})
		if err != nil {
			return false, err
		}
		switch strings.ToLower(strings.TrimSpace(answer)) {
		case "y", "yes", "true", "1":
			return true, nil
		case "n", "no", "false", "0", "":
			return false, nil
		default:
			Default().Println("[red]Please answer yes or no.[/]")
		}
	}
}

// Select prompts with choices.
func Select(prompt string, choices []string) (string, error) {
	return Prompt(prompt, PromptOpts{Choices: choices})
}

// PromptOpts configures complex prompts.
type PromptOpts struct {
	Default   string
	Validator func(string) error
	Password  bool
	Choices   []string
}

// Prompt displays a prompt and returns the input.
func Prompt(message string, opts PromptOpts) (string, error) {
	console := Default()
	if len(opts.Choices) > 0 {
		for i, choice := range opts.Choices {
			console.Printf("  %d) %s\n", i+1, choice)
		}
	}
	for {
		prompt := buildPrompt(message, opts.Default, opts.Password)
		var input string
		var err error
		if opts.Password {
			input, err = readPassword(console, prompt)
		} else {
			input, err = readLine(console, prompt)
		}
		if err != nil {
			return "", err
		}
		if input == "" && opts.Default != "" {
			input = opts.Default
		}
		if len(opts.Choices) > 0 {
			resolved, ok := resolveChoice(input, opts.Choices)
			if !ok {
				console.Println("[red]Please select one of the listed options.[/]")
				continue
			}
			input = resolved
		}
		if opts.Validator != nil {
			if err := opts.Validator(input); err != nil {
				console.Println("[red]" + err.Error() + "[/]")
				continue
			}
		}
		return input, nil
	}
}

func buildPrompt(message, def string, password bool) string {
	prompt := message
	if def != "" {
		prompt = fmt.Sprintf("%s [%s]", prompt, def)
	}
	if password {
		prompt += " (hidden)"
	}
	return prompt + ": "
}

func readLine(console *Console, prompt string) (string, error) {
	console.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func readPassword(console *Console, prompt string) (string, error) {
	console.Print(prompt)
	bytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	console.Println()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(bytes)), nil
}

func resolveChoice(input string, choices []string) (string, bool) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return "", false
	}
	if idx, err := strconv.Atoi(trimmed); err == nil {
		if idx >= 1 && idx <= len(choices) {
			return choices[idx-1], true
		}
		return "", false
	}
	for _, choice := range choices {
		if strings.EqualFold(choice, trimmed) {
			return choice, true
		}
	}
	return "", false
}
