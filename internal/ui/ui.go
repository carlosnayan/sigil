package ui

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
)

var (
	successColor = color.New(color.FgGreen).SprintFunc()
	warnColor    = color.New(color.FgYellow).SprintFunc()
	errColor     = color.New(color.FgRed).SprintFunc()
)

func Success(msg string) {
	fmt.Fprintln(os.Stdout, successColor("✓ "+msg))
}

func Warn(msg string) {
	fmt.Fprintln(os.Stderr, warnColor("! "+msg))
}

func Error(msg string) {
	fmt.Fprintln(os.Stderr, errColor("✗ "+msg))
}

func PromptPassword(label string) (string, error) {
	p := promptui.Prompt{
		Label: label,
		Mask:  '•',
	}
	return p.Run()
}

const maxPasswordConfirmAttempts = 3

func PromptConfirmPassword(label, confirmLabel string) (string, error) {
	for attempt := 0; attempt < maxPasswordConfirmAttempts; attempt++ {
		p1, err := PromptPassword(label)
		if err != nil {
			return "", err
		}
		p2, err := PromptPassword(confirmLabel)
		if err != nil {
			return "", err
		}
		if p1 == "" {
			Error("password cannot be empty")
			continue
		}
		if p1 != p2 {
			Error("passwords do not match")
			continue
		}
		return p1, nil
	}
	return "", errors.New("too many failed password confirmation attempts")
}

func PromptSelect(label string, items []string) (int, string, error) {
	p := promptui.Select{
		Label: label,
		Items: items,
	}
	return p.Run()
}

var VaultMenuItems = []string{
	"Manage secrets",
	"Import .env file",
	"Export .env file",
	"Rekey / Change passphrase",
	"Doctor / Status",
	"Exit",
}

func VaultMenu(unlocked bool) (int, string, error) {
	status := "🔒 Locked"
	if unlocked {
		status = "🔓 Unlocked"
	}
	bold := color.New(color.Bold)
	fmt.Fprintln(os.Stdout)
	fmt.Fprintln(os.Stdout, bold.Sprint("SIGIL — Vault Menu"))
	fmt.Fprintf(os.Stdout, "Status: %s\n", status)

	sel := promptui.Select{
		Label:    "",
		Items:    VaultMenuItems,
		Size:     len(VaultMenuItems),
		HideHelp: true,
		Templates: &promptui.SelectTemplates{
			Label:    "{{if false}}{{end}}",
			Active:   "❯ {{ . | cyan }}",
			Inactive: "  {{ . }}",
			Selected: "❯ {{ . | green }}",
		},
	}
	return sel.Run()
}

func Confirm(label string, defaultYes bool) (bool, error) {
	hint := "[Y/n]"
	if !defaultYes {
		hint = "[y/N]"
	}
	p := promptui.Prompt{
		Label: label + " " + hint,
	}
	s, err := p.Run()
	if err != nil {
		return false, err
	}
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return defaultYes, nil
	}
	return s == "y" || s == "yes", nil
}
