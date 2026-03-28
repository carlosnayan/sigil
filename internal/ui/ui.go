package ui

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
)

var (
	successColor = color.New(color.FgGreen).SprintFunc()
	warnColor    = color.New(color.FgYellow).SprintFunc()
	errColor     = color.New(color.FgRed).SprintFunc()
)

type pendingNotice struct {
	warn bool
	text string
}

var (
	menuMu           sync.Mutex
	menuSessionDepth int
	pendingNotices   []pendingNotice
)

// BeginMenuSession marks interactive menu mode so Warn/Error are queued for replay
// after the next PrepareMenuScreen (after ClearScreen). Nestable (e.g. config → manage configs).
func BeginMenuSession() {
	menuMu.Lock()
	menuSessionDepth++
	menuMu.Unlock()
}

// EndMenuSession ends one level of menu session; when depth reaches zero, pending notices are dropped.
func EndMenuSession() {
	menuMu.Lock()
	menuSessionDepth--
	if menuSessionDepth < 0 {
		menuSessionDepth = 0
	}
	if menuSessionDepth == 0 {
		pendingNotices = nil
	}
	menuMu.Unlock()
}

func queueNotice(warn bool, text string) {
	menuMu.Lock()
	if menuSessionDepth > 0 {
		pendingNotices = append(pendingNotices, pendingNotice{warn: warn, text: text})
	}
	menuMu.Unlock()
}

func Success(msg string) {
	fmt.Fprintln(os.Stdout, successColor("✓ "+msg))
}

func Warn(msg string) {
	fmt.Fprintln(os.Stderr, warnColor("! "+msg))
	queueNotice(true, msg)
}

func Error(msg string) {
	fmt.Fprintln(os.Stderr, errColor("✗ "+msg))
	queueNotice(false, msg)
}

// ClearScreen moves the cursor home and clears the visible terminal buffer.
func ClearScreen() {
	fmt.Fprint(os.Stdout, "\033[H\033[2J")
}

// PrepareMenuScreen clears the terminal and reprints Warn/Error lines queued since the last call
// (so messages are not lost when ClearScreen runs between menu iterations).
func PrepareMenuScreen() {
	ClearScreen()
	menuMu.Lock()
	q := pendingNotices
	pendingNotices = nil
	menuMu.Unlock()
	for _, n := range q {
		if n.warn {
			fmt.Fprintln(os.Stderr, warnColor("! "+n.text))
		} else {
			fmt.Fprintln(os.Stderr, errColor("✗ "+n.text))
		}
	}
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

// BackItemLabel is the standard label for returning to the previous menu.
const BackItemLabel = "← Back"

var VaultMenuItems = []string{
	"Add new config",
	"Manage configs",
	"Rekey / Change secret",
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
			Selected: " ",
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
