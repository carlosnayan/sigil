package ui

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"

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
	"Manage configs",
	"Import .env file",
	"Export .env file",
	"Rekey / Change secret",
	"Doctor / Status",
	"Exit",
}

func promptSelectItems(items []string) *promptui.Select {
	return &promptui.Select{
		Label:    "",
		Items:    items,
		Size:     len(items),
		HideHelp: true,
		Templates: &promptui.SelectTemplates{
			Label:    "{{if false}}{{end}}",
			Active:   "❯ {{ . | cyan }}",
			Inactive: "  {{ . }}",
			Selected: "❯ {{ . | green }}",
		},
	}
}

func SelectList(header string, items []string) (int, string, error) {
	bold := color.New(color.Bold)
	fmt.Fprintln(os.Stdout)
	if header != "" {
		fmt.Fprintln(os.Stdout, bold.Sprint(header))
		fmt.Fprintln(os.Stdout)
	}
	sel := promptSelectItems(items)
	return sel.Run()
}

func VaultMenu() (int, string, error) {
	bold := color.New(color.Bold)
	fmt.Fprintln(os.Stdout)
	fmt.Fprintln(os.Stdout, bold.Sprint("SIGIL — Vault Menu"))

	sel := promptSelectItems(VaultMenuItems)
	return sel.Run()
}

func PromptText(label string) (string, error) {
	p := promptui.Prompt{
		Label: label,
	}
	return p.Run()
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

func runeWidth(s string) int {
	return utf8.RuneCountInString(s)
}

func padCell(s string, width int) string {
	w := runeWidth(s)
	if w >= width {
		rs := []rune(s)
		if len(rs) <= width {
			return s
		}
		return string(rs[:width])
	}
	return s + strings.Repeat(" ", width-w)
}

func PrintTable(headers []string, rows [][]string) {
	printTable(os.Stdout, headers, rows)
}

func printTable(w io.Writer, headers []string, rows [][]string) {
	if len(headers) == 0 {
		return
	}
	n := len(headers)
	widths := make([]int, n)
	for i, h := range headers {
		widths[i] = runeWidth(h)
	}
	for _, row := range rows {
		for i := 0; i < n; i++ {
			cell := ""
			if i < len(row) {
				cell = row[i]
			}
			if rw := runeWidth(cell); rw > widths[i] {
				widths[i] = rw
			}
		}
	}

	hr := func(left, mid, right string) {
		_, _ = fmt.Fprint(w, left)
		for i := 0; i < n; i++ {
			if i > 0 {
				_, _ = fmt.Fprint(w, mid)
			}
			_, _ = fmt.Fprint(w, strings.Repeat("─", widths[i]+2))
		}
		_, _ = fmt.Fprintln(w, right)
	}

	hr("┌", "┬", "┐")

	_, _ = fmt.Fprint(w, "│")
	for i := 0; i < n; i++ {
		if i > 0 {
			_, _ = fmt.Fprint(w, "│")
		}
		_, _ = fmt.Fprintf(w, " %s ", padCell(headers[i], widths[i]))
	}
	_, _ = fmt.Fprintln(w, "│")

	hr("├", "┼", "┤")

	for _, row := range rows {
		_, _ = fmt.Fprint(w, "│")
		for i := 0; i < n; i++ {
			if i > 0 {
				_, _ = fmt.Fprint(w, "│")
			}
			cell := ""
			if i < len(row) {
				cell = row[i]
			}
			_, _ = fmt.Fprintf(w, " %s ", padCell(cell, widths[i]))
		}
		_, _ = fmt.Fprintln(w, "│")
	}

	hr("└", "┴", "┘")
}
