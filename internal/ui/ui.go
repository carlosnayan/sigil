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

const maxPasswordConfirmAttempts = 3

// PromptConfirmPassword pede a senha duas vezes; repete até maxPasswordConfirmAttempts se não coincidirem ou estiverem vazias.
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

// PromptSelect mostra um menu e devolve índice e valor escolhidos.
func PromptSelect(label string, items []string) (int, string, error) {
	p := promptui.Select{
		Label: label,
		Items: items,
	}
	return p.Run()
}

// Confirm pergunta Y/n (ou Enter para o default). defaultYes define o comportamento quando a resposta está vazia.
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
