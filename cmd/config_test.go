package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/carlos/sigil/internal/ui"
	"github.com/manifoldco/promptui"
)

func writeVaultYAML(t *testing.T, path, secret string) {
	t.Helper()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	body := "project: test\nenv: dev\nsecret: " + secret + "\nvaults: []\ninject: {}\n"
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestRunConfig_noVaultYAML(t *testing.T) {
	cmdExecuteMu.Lock()
	defer cmdExecuteMu.Unlock()
	resetCmdState(t)

	home := t.TempDir()
	t.Setenv("HOME", home)

	rootCmd.SetArgs([]string{"config"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error when vault.yaml is missing")
	}
	if !strings.Contains(err.Error(), "run `sigil init`") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunConfig_emptySecret(t *testing.T) {
	cmdExecuteMu.Lock()
	defer cmdExecuteMu.Unlock()
	resetCmdState(t)

	home := t.TempDir()
	t.Setenv("HOME", home)
	cfgPath := filepath.Join(home, ".sigil", "vault.yaml")
	if err := os.MkdirAll(filepath.Dir(cfgPath), 0o700); err != nil {
		t.Fatal(err)
	}
	body := "project: test\nenv: dev\nvaults: []\ninject: {}\n"
	if err := os.WriteFile(cfgPath, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}

	rootCmd.SetArgs([]string{"config"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for empty secret")
	}
	if !strings.Contains(err.Error(), "no secret") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunConfig_menuExit(t *testing.T) {
	cmdExecuteMu.Lock()
	defer cmdExecuteMu.Unlock()
	resetCmdState(t)

	home := t.TempDir()
	t.Setenv("HOME", home)
	cfgPath := filepath.Join(home, ".sigil", "vault.yaml")
	writeVaultYAML(t, cfgPath, "team-secret")

	oldM := configVaultMenu
	configVaultMenu = func() (int, string, error) {
		return 2, ui.VaultMenuItems[2], nil
	}
	t.Cleanup(func() { configVaultMenu = oldM })

	rootCmd.SetArgs([]string{"config"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
}

func TestRunConfig_menuInterruptReturnsNil(t *testing.T) {
	cmdExecuteMu.Lock()
	defer cmdExecuteMu.Unlock()
	resetCmdState(t)

	home := t.TempDir()
	t.Setenv("HOME", home)
	cfgPath := filepath.Join(home, ".sigil", "vault.yaml")
	writeVaultYAML(t, cfgPath, "team-secret")

	oldM := configVaultMenu
	configVaultMenu = func() (int, string, error) {
		return 0, "", promptui.ErrInterrupt
	}
	t.Cleanup(func() { configVaultMenu = oldM })

	rootCmd.SetArgs([]string{"config"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("expected nil on menu interrupt, got %v", err)
	}
}
