package ui

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
)

var (
	successColor = color.New(color.FgGreen).SprintFunc()
	warnColor    = color.New(color.FgYellow).SprintFunc()
	errColor     = color.New(color.FgRed).SprintFunc()
)

// Success imprime mensagem de sucesso.
func Success(msg string) {
	fmt.Fprintln(os.Stdout, successColor("✓ "+msg))
}

// Warn imprime aviso.
func Warn(msg string) {
	fmt.Fprintln(os.Stderr, warnColor("! "+msg))
}

// Error imprime erro formatado.
func Error(msg string) {
	fmt.Fprintln(os.Stderr, errColor("✗ "+msg))
}

// PromptPassword solicita senha sem eco (stub interativo usando promptui).
func PromptPassword(label string) (string, error) {
	p := promptui.Prompt{
		Label: label,
		Mask:  '•',
	}
	return p.Run()
}
