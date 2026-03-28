package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/carlos/sigil/internal/crypto"
)

func TestDefaultConfigPath_underHomeSigil(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	resetCmdState(t)
	got, err := DefaultConfigPath()
	if err != nil {
		t.Fatal(err)
	}
	want, err := filepath.Abs(filepath.Join(home, ".sigil", "vault.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	gotAbs, err := filepath.Abs(got)
	if err != nil {
		t.Fatal(err)
	}
	if gotAbs != want {
		t.Fatalf("DefaultConfigPath: got %q, want %q", gotAbs, want)
	}
}

func TestRunInit_rejectsExistingVaultYAML(t *testing.T) {
	cmdExecuteMu.Lock()
	defer cmdExecuteMu.Unlock()
	resetCmdState(t)

	home := t.TempDir()
	t.Setenv("HOME", home)
	sigilDir := filepath.Join(home, ".sigil")
	if err := os.MkdirAll(sigilDir, 0o700); err != nil {
		t.Fatal(err)
	}
	vaultYAML := filepath.Join(sigilDir, "vault.yaml")
	if err := os.WriteFile(vaultYAML, []byte("project: x\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	vaultAbs, err := filepath.Abs(vaultYAML)
	if err != nil {
		t.Fatal(err)
	}

	rootCmd.SetArgs([]string{"init", "--config", vaultAbs})
	err = rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error when vault.yaml already exists")
	}
	if !strings.Contains(err.Error(), "already initialized") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunInit_rejectsExistingSecretEnc(t *testing.T) {
	cmdExecuteMu.Lock()
	defer cmdExecuteMu.Unlock()
	resetCmdState(t)

	home := t.TempDir()
	t.Setenv("HOME", home)
	sigilDir := filepath.Join(home, ".sigil")
	if err := os.MkdirAll(sigilDir, 0o700); err != nil {
		t.Fatal(err)
	}
	secretPath := filepath.Join(sigilDir, "secret.enc")
	if err := os.WriteFile(secretPath, []byte("dummy"), 0o600); err != nil {
		t.Fatal(err)
	}

	cfgDir := t.TempDir()
	cfgAbsent := filepath.Join(cfgDir, "vault.yaml")
	cfgAbs, err := filepath.Abs(cfgAbsent)
	if err != nil {
		t.Fatal(err)
	}

	rootCmd.SetArgs([]string{"init", "--config", cfgAbs})
	err = rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error when secret.enc already exists")
	}
	if !strings.Contains(err.Error(), "user data already exists") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunInit_createsVaultWithCustomConfigPath(t *testing.T) {
	cmdExecuteMu.Lock()
	defer cmdExecuteMu.Unlock()
	resetCmdState(t)

	home := t.TempDir()
	t.Setenv("HOME", home)
	cfgDir := t.TempDir()
	cfgPath := filepath.Join(cfgDir, "vault.yaml")
	cfgAbs, err := filepath.Abs(cfgPath)
	if err != nil {
		t.Fatal(err)
	}

	oldPC := initPromptConfirmPassword
	oldC := initConfirm
	oldPP := initPromptPassword
	initPromptConfirmPassword = func(_, _ string) (string, error) { return "unit-test-pass", nil }
	initConfirm = func(_ string, _ bool) (bool, error) { return false, nil }
	initPromptPassword = func(_ string) (string, error) { return "unit-test-team-secret", nil }
	t.Cleanup(func() {
		initPromptConfirmPassword = oldPC
		initConfirm = oldC
		initPromptPassword = oldPP
	})

	rootCmd.SetArgs([]string{"init", "--config", cfgAbs})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(cfgAbs); err != nil {
		t.Fatalf("vault.yaml: %v", err)
	}
	sigilDir := filepath.Join(home, ".sigil")
	secretPath := filepath.Join(sigilDir, "secret.enc")
	blob, err := os.ReadFile(secretPath)
	if err != nil {
		t.Fatalf("secret.enc: %v", err)
	}
	plain, err := crypto.UnwrapSecret("unit-test-pass", blob)
	if err != nil {
		t.Fatalf("UnwrapSecret: %v", err)
	}
	if plain != "unit-test-team-secret" {
		t.Fatalf("unexpected plaintext secret: %q", plain)
	}

	envEnc := filepath.Join(sigilDir, "vaults", "default.dev.env.enc")
	enc, err := os.ReadFile(envEnc)
	if err != nil {
		t.Fatalf("default.dev.env.enc: %v", err)
	}
	pt, err := crypto.DecryptVault(plain, enc)
	if err != nil {
		t.Fatalf("DecryptVault: %v", err)
	}
	if !strings.Contains(string(pt), "KEY=value") {
		t.Fatalf("unexpected vault plaintext: %q", string(pt))
	}
}

func TestRunInit_createsVaultWithDefaultConfigPath(t *testing.T) {
	cmdExecuteMu.Lock()
	defer cmdExecuteMu.Unlock()
	resetCmdState(t)

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Chdir(t.TempDir())

	oldPC := initPromptConfirmPassword
	oldC := initConfirm
	oldPP := initPromptPassword
	initPromptConfirmPassword = func(_, _ string) (string, error) { return "unit-pass-2", nil }
	initConfirm = func(_ string, _ bool) (bool, error) { return false, nil }
	initPromptPassword = func(_ string) (string, error) { return "team-secret-2", nil }
	t.Cleanup(func() {
		initPromptConfirmPassword = oldPC
		initConfirm = oldC
		initPromptPassword = oldPP
	})

	rootCmd.SetArgs([]string{"init"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}

	cfgPath := filepath.Join(home, ".sigil", "vault.yaml")
	if _, err := os.Stat(cfgPath); err != nil {
		t.Fatalf("default vault.yaml: %v", err)
	}
}
