package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/carlos/sigil/internal/crypto"
	"github.com/carlos/sigil/internal/env"
	"github.com/carlos/sigil/internal/link"
	"github.com/carlos/sigil/internal/vaultstore"
)

func writeVaultYAMLRun(t *testing.T, path, secret string) {
	t.Helper()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	body := `project: test
env: dev
secret: ` + secret + `
vaults: []
inject:
  INJECT_ONLY: injected
  SHARED: from_inject
`
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestRunCmd_noArgs(t *testing.T) {
	cmdExecuteMu.Lock()
	defer cmdExecuteMu.Unlock()
	resetCmdState(t)

	home := t.TempDir()
	t.Setenv("HOME", home)

	rootCmd.SetArgs([]string{"run"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error when no command after run")
	}
	if !strings.Contains(err.Error(), "no command") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunCmd_noSetup(t *testing.T) {
	cmdExecuteMu.Lock()
	defer cmdExecuteMu.Unlock()
	resetCmdState(t)

	home := t.TempDir()
	t.Setenv("HOME", home)
	sigilHome := filepath.Join(home, ".sigil")
	writeVaultYAMLRun(t, filepath.Join(sigilHome, "vault.yaml"), "secret-key")

	enc, err := crypto.EncryptVault("secret-key", []byte("X=1\n"))
	if err != nil {
		t.Fatal(err)
	}
	vdir := vaultstore.Dir(sigilHome)
	if err := os.MkdirAll(vdir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(vaultstore.EncPath(sigilHome, "dev"), enc, 0o600); err != nil {
		t.Fatal(err)
	}

	projectDir := t.TempDir()
	oldWd := runGetwd
	runGetwd = func() (string, error) { return projectDir, nil }
	defer func() { runGetwd = oldWd }()

	oldExit := runOsExit
	var exitRecorded int
	exitSeen := false
	runOsExit = func(code int) {
		exitRecorded = code
		exitSeen = true
	}
	defer func() { runOsExit = oldExit }()

	cfgAbs := filepath.Join(sigilHome, "vault.yaml")
	rootCmd.SetArgs([]string{"--config", cfgAbs, "run", "--", "true"})
	err = rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error when no setup / sigil.yaml")
	}
	if strings.Contains(err.Error(), "no command") {
		t.Fatalf("wrong error: %v", err)
	}
	if !strings.Contains(strings.ToLower(err.Error()), "setup") && !strings.Contains(err.Error(), "sigil.yaml") {
		t.Fatalf("expected setup hint, got: %v", err)
	}
	if exitSeen {
		t.Fatal("should not exit on error path")
	}
	_ = exitRecorded
}

func TestRunCmd_injectsEnv(t *testing.T) {
	cmdExecuteMu.Lock()
	defer cmdExecuteMu.Unlock()
	resetCmdState(t)

	home := t.TempDir()
	t.Setenv("HOME", home)
	sigilHome := filepath.Join(home, ".sigil")
	cfgPath := filepath.Join(sigilHome, "vault.yaml")
	writeVaultYAMLRun(t, cfgPath, "secret-key")

	plain := "FROM_VAULT=vault_value\nSHARED=from_vault\n"
	enc, err := crypto.EncryptVault("secret-key", []byte(plain))
	if err != nil {
		t.Fatal(err)
	}
	vdir := vaultstore.Dir(sigilHome)
	if err := os.MkdirAll(vdir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(vaultstore.EncPath(sigilHome, "dev"), enc, 0o600); err != nil {
		t.Fatal(err)
	}

	projectDir := t.TempDir()
	sigilManifest := filepath.Join(projectDir, "sigil.yaml")
	if err := os.WriteFile(sigilManifest, []byte("setup:\n  config: dev\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	oldWd := runGetwd
	runGetwd = func() (string, error) { return projectDir, nil }
	defer func() { runGetwd = oldWd }()

	oldChild := runChildProcess
	var capturedEnv []string
	runChildProcess = func(name string, args []string, environ []string) (int, error) {
		capturedEnv = append([]string(nil), environ...)
		return 0, nil
	}
	defer func() { runChildProcess = oldChild }()

	oldExit := runOsExit
	var exitCode int
	runOsExit = func(code int) { exitCode = code }
	defer func() { runOsExit = oldExit }()

	cfgAbs, _ := filepath.Abs(cfgPath)
	rootCmd.SetArgs([]string{"--config", cfgAbs, "run", "--", "true", "arg2"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if exitCode != 0 {
		t.Fatalf("exit code %d", exitCode)
	}

	m := env.FromSlice(capturedEnv)
	if m["FROM_VAULT"] != "vault_value" {
		t.Fatalf("FROM_VAULT: got %q", m["FROM_VAULT"])
	}
	if m["INJECT_ONLY"] != "injected" {
		t.Fatalf("INJECT_ONLY: got %q", m["INJECT_ONLY"])
	}
	if m["SHARED"] != "from_inject" {
		t.Fatalf("SHARED should be overridden by inject: got %q", m["SHARED"])
	}
}

func TestRunCmd_viaLinksYaml(t *testing.T) {
	cmdExecuteMu.Lock()
	defer cmdExecuteMu.Unlock()
	resetCmdState(t)

	home := t.TempDir()
	t.Setenv("HOME", home)
	sigilHome := filepath.Join(home, ".sigil")
	cfgPath := filepath.Join(sigilHome, "vault.yaml")
	writeVaultYAMLRun(t, cfgPath, "secret-key")

	enc, err := crypto.EncryptVault("secret-key", []byte("LINKED=1\n"))
	if err != nil {
		t.Fatal(err)
	}
	vdir := vaultstore.Dir(sigilHome)
	_ = os.MkdirAll(vdir, 0o700)
	_ = os.WriteFile(vaultstore.EncPath(sigilHome, "staging"), enc, 0o600)

	projectDir := t.TempDir()
	if err := link.Set(sigilHome, projectDir, "staging"); err != nil {
		t.Fatal(err)
	}

	oldWd := runGetwd
	runGetwd = func() (string, error) { return projectDir, nil }
	defer func() { runGetwd = oldWd }()

	oldChild := runChildProcess
	var capturedEnv []string
	runChildProcess = func(name string, args []string, environ []string) (int, error) {
		capturedEnv = append([]string(nil), environ...)
		return 0, nil
	}
	defer func() { runChildProcess = oldChild }()

	oldExit := runOsExit
	runOsExit = func(code int) {}
	defer func() { runOsExit = oldExit }()

	cfgAbs, _ := filepath.Abs(cfgPath)
	rootCmd.SetArgs([]string{"--config", cfgAbs, "run", "--", "true"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if env.FromSlice(capturedEnv)["LINKED"] != "1" {
		t.Fatalf("env missing LINKED: %v", capturedEnv)
	}
}

func TestRunCmd_exitCodePropagation(t *testing.T) {
	cmdExecuteMu.Lock()
	defer cmdExecuteMu.Unlock()
	resetCmdState(t)

	home := t.TempDir()
	t.Setenv("HOME", home)
	sigilHome := filepath.Join(home, ".sigil")
	cfgPath := filepath.Join(sigilHome, "vault.yaml")
	writeVaultYAMLRun(t, cfgPath, "secret-key")

	enc, err := crypto.EncryptVault("secret-key", []byte("A=b\n"))
	if err != nil {
		t.Fatal(err)
	}
	_ = os.MkdirAll(vaultstore.Dir(sigilHome), 0o700)
	_ = os.WriteFile(vaultstore.EncPath(sigilHome, "dev"), enc, 0o600)

	projectDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(projectDir, "sigil.yaml"), []byte("setup:\n  config: dev\n"), 0o600)

	oldWd := runGetwd
	runGetwd = func() (string, error) { return projectDir, nil }
	defer func() { runGetwd = oldWd }()

	oldChild := runChildProcess
	runChildProcess = func(name string, args []string, environ []string) (int, error) {
		return 17, nil
	}
	defer func() { runChildProcess = oldChild }()

	oldExit := runOsExit
	var exitCode int
	runOsExit = func(code int) { exitCode = code }
	defer func() { runOsExit = oldExit }()

	cfgAbs, _ := filepath.Abs(cfgPath)
	rootCmd.SetArgs([]string{"--config", cfgAbs, "run", "--", "x"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if exitCode != 17 {
		t.Fatalf("want exit 17, got %d", exitCode)
	}
}
