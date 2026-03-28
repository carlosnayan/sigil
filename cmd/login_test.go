package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/carlos/sigil/internal/crypto"
	"github.com/carlos/sigil/internal/ui"
	"github.com/manifoldco/promptui"
)

func TestRunLogin_noVault(t *testing.T) {
	cmdExecuteMu.Lock()
	defer cmdExecuteMu.Unlock()
	resetCmdState(t)

	home := t.TempDir()
	t.Setenv("HOME", home)

	rootCmd.SetArgs([]string{"login"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error when secret.enc is missing")
	}
	if !strings.Contains(err.Error(), "run `sigil init`") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunLogin_emptyPassword(t *testing.T) {
	cmdExecuteMu.Lock()
	defer cmdExecuteMu.Unlock()
	resetCmdState(t)

	home := t.TempDir()
	t.Setenv("HOME", home)
	sigilDir := filepath.Join(home, ".sigil")
	if err := os.MkdirAll(sigilDir, 0o700); err != nil {
		t.Fatal(err)
	}
	blob, err := crypto.WrapSecret("real", "team")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sigilDir, "secret.enc"), blob, 0o600); err != nil {
		t.Fatal(err)
	}

	old := loginPromptPassword
	loginPromptPassword = func(string) (string, error) { return "", nil }
	t.Cleanup(func() { loginPromptPassword = old })

	rootCmd.SetArgs([]string{"login"})
	err = rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for empty password")
	}
	if !strings.Contains(err.Error(), "cannot be empty") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunLogin_wrongPassword(t *testing.T) {
	cmdExecuteMu.Lock()
	defer cmdExecuteMu.Unlock()
	resetCmdState(t)

	home := t.TempDir()
	t.Setenv("HOME", home)
	sigilDir := filepath.Join(home, ".sigil")
	if err := os.MkdirAll(sigilDir, 0o700); err != nil {
		t.Fatal(err)
	}
	blob, err := crypto.WrapSecret("correct", "team")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sigilDir, "secret.enc"), blob, 0o600); err != nil {
		t.Fatal(err)
	}

	old := loginPromptPassword
	loginPromptPassword = func(string) (string, error) { return "wrong-pass", nil }
	t.Cleanup(func() { loginPromptPassword = old })

	rootCmd.SetArgs([]string{"login"})
	err = rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for wrong password")
	}
	if !strings.Contains(err.Error(), "wrong password") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunLogin_passwordInterruptReturnsNil(t *testing.T) {
	cmdExecuteMu.Lock()
	defer cmdExecuteMu.Unlock()
	resetCmdState(t)

	home := t.TempDir()
	t.Setenv("HOME", home)
	sigilDir := filepath.Join(home, ".sigil")
	if err := os.MkdirAll(sigilDir, 0o700); err != nil {
		t.Fatal(err)
	}
	blob, err := crypto.WrapSecret("p", "team")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sigilDir, "secret.enc"), blob, 0o600); err != nil {
		t.Fatal(err)
	}

	old := loginPromptPassword
	loginPromptPassword = func(string) (string, error) {
		return "", promptui.ErrInterrupt
	}
	t.Cleanup(func() { loginPromptPassword = old })

	rootCmd.SetArgs([]string{"login"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("expected nil on password interrupt, got %v", err)
	}
}

func TestRunLogin_menuExit(t *testing.T) {
	cmdExecuteMu.Lock()
	defer cmdExecuteMu.Unlock()
	resetCmdState(t)

	home := t.TempDir()
	t.Setenv("HOME", home)
	sigilDir := filepath.Join(home, ".sigil")
	if err := os.MkdirAll(sigilDir, 0o700); err != nil {
		t.Fatal(err)
	}
	blob, err := crypto.WrapSecret("pw", "team-secret")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sigilDir, "secret.enc"), blob, 0o600); err != nil {
		t.Fatal(err)
	}

	oldP := loginPromptPassword
	oldM := loginVaultMenu
	loginPromptPassword = func(string) (string, error) { return "pw", nil }
	loginVaultMenu = func(bool) (int, string, error) {
		return 5, ui.VaultMenuItems[5], nil
	}
	t.Cleanup(func() {
		loginPromptPassword = oldP
		loginVaultMenu = oldM
	})

	rootCmd.SetArgs([]string{"login"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
}

func TestRunLogin_menuInterruptReturnsNil(t *testing.T) {
	cmdExecuteMu.Lock()
	defer cmdExecuteMu.Unlock()
	resetCmdState(t)

	home := t.TempDir()
	t.Setenv("HOME", home)
	sigilDir := filepath.Join(home, ".sigil")
	if err := os.MkdirAll(sigilDir, 0o700); err != nil {
		t.Fatal(err)
	}
	blob, err := crypto.WrapSecret("pw", "team-secret")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sigilDir, "secret.enc"), blob, 0o600); err != nil {
		t.Fatal(err)
	}

	oldP := loginPromptPassword
	oldM := loginVaultMenu
	loginPromptPassword = func(string) (string, error) { return "pw", nil }
	loginVaultMenu = func(bool) (int, string, error) {
		return 0, "", promptui.ErrInterrupt
	}
	t.Cleanup(func() {
		loginPromptPassword = oldP
		loginVaultMenu = oldM
	})

	rootCmd.SetArgs([]string{"login"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("expected nil on menu interrupt, got %v", err)
	}
}
