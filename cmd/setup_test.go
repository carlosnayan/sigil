package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/carlos/sigil/internal/link"
)

func TestRunSetup_noSigilFile(t *testing.T) {
	cmdExecuteMu.Lock()
	defer cmdExecuteMu.Unlock()
	resetCmdState(t)

	home := t.TempDir()
	t.Setenv("HOME", home)
	proj := t.TempDir()

	oldGetwd := setupGetwd
	setupGetwd = func() (string, error) { return proj, nil }
	t.Cleanup(func() { setupGetwd = oldGetwd })

	rootCmd.SetArgs([]string{"setup"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "no sigil.yaml found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunSetup_missingSetupConfig(t *testing.T) {
	cmdExecuteMu.Lock()
	defer cmdExecuteMu.Unlock()
	resetCmdState(t)

	home := t.TempDir()
	t.Setenv("HOME", home)
	proj := t.TempDir()
	if err := os.WriteFile(filepath.Join(proj, "sigil.yaml"), []byte("setup: {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	oldGetwd := setupGetwd
	setupGetwd = func() (string, error) { return proj, nil }
	t.Cleanup(func() { setupGetwd = oldGetwd })

	rootCmd.SetArgs([]string{"setup"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "setup.config is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunSetup_encMissing(t *testing.T) {
	cmdExecuteMu.Lock()
	defer cmdExecuteMu.Unlock()
	resetCmdState(t)

	home := t.TempDir()
	t.Setenv("HOME", home)
	sigilDir := filepath.Join(home, ".sigil")
	if err := os.MkdirAll(filepath.Join(sigilDir, "vaults"), 0o700); err != nil {
		t.Fatal(err)
	}

	proj := t.TempDir()
	manifest := "setup:\n  config: dev\n"
	if err := os.WriteFile(filepath.Join(proj, "sigil.yaml"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}

	oldGetwd := setupGetwd
	setupGetwd = func() (string, error) { return proj, nil }
	t.Cleanup(func() { setupGetwd = oldGetwd })

	rootCmd.SetArgs([]string{"setup"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), `config "dev" not found`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunSetup_happyPath_sigilYml(t *testing.T) {
	cmdExecuteMu.Lock()
	defer cmdExecuteMu.Unlock()
	resetCmdState(t)

	home := t.TempDir()
	t.Setenv("HOME", home)
	sigilDir := filepath.Join(home, ".sigil")
	vaultsDir := filepath.Join(sigilDir, "vaults")
	if err := os.MkdirAll(vaultsDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(vaultsDir, "prod.enc"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}

	proj := t.TempDir()
	manifest := "setup:\n  config: prod\n"
	if err := os.WriteFile(filepath.Join(proj, "sigil.yml"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}

	projAbs, err := filepath.Abs(proj)
	if err != nil {
		t.Fatal(err)
	}

	oldGetwd := setupGetwd
	setupGetwd = func() (string, error) { return proj, nil }
	t.Cleanup(func() { setupGetwd = oldGetwd })

	rootCmd.SetArgs([]string{"setup"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}

	slug, ok := link.Get(sigilDir, projAbs)
	if !ok || slug != "prod" {
		t.Fatalf("link Get: %q %v", slug, ok)
	}
}
