package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/carlos/sigil/internal/config"
	"github.com/carlos/sigil/internal/crypto"
	"github.com/carlos/sigil/internal/env"
	"github.com/carlos/sigil/internal/link"
	"github.com/carlos/sigil/internal/runner"
	"github.com/carlos/sigil/internal/vaultstore"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var runCmd = &cobra.Command{
	Use:   "run [flags] -- <command> [args...]",
	Short: "Run a command with vault variables injected",
	Long: `Decrypts the linked vault config, merges variables with your shell
environment and vault.yaml inject, then runs the child process.

Example: sigil run -- docker compose up

Requires a linked directory (sigil setup) or sigil.yaml with setup.config in the current directory.`,
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		code, err := runRun(cmd, args)
		if err != nil {
			return err
		}
		runOsExit(code)
		return nil
	},
}

var (
	runGetwd        = os.Getwd
	runChildProcess = runner.Run
	runOsExit       = func(code int) { os.Exit(code) }
)

func init() {
	rootCmd.AddCommand(runCmd)
}

func resolveRunSlug(cfgPath, cwd string) (string, error) {
	slug, ok, err := link.Get(cfgPath, cwd)
	if err != nil {
		return "", fmt.Errorf("sigil run: read vault.yaml links: %w", err)
	}
	if ok {
		if err := vaultstore.ValidateSlug(slug); err != nil {
			return "", fmt.Errorf("sigil run: invalid linked config slug: %w", err)
		}
		return slug, nil
	}

	manifest, err := findSigilManifest(cwd)
	if err != nil {
		if errors.Is(err, errSigilNotFound) {
			return "", fmt.Errorf("sigil run: no linked config for this directory — run `sigil setup` first (or add sigil.yaml with setup.config)")
		}
		return "", err
	}

	data, err := os.ReadFile(manifest)
	if err != nil {
		return "", fmt.Errorf("sigil run: read %s: %w", filepath.Base(manifest), err)
	}

	var doc sigilYAML
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return "", fmt.Errorf("sigil run: parse %s: %w", filepath.Base(manifest), err)
	}

	configSlug := doc.Setup.Config
	if configSlug == "" {
		return "", fmt.Errorf("sigil run: setup.config is required in sigil.yaml")
	}
	if err := vaultstore.ValidateSlug(configSlug); err != nil {
		return "", fmt.Errorf("sigil run: invalid setup.config: %w", err)
	}
	return configSlug, nil
}

func runRun(cmd *cobra.Command, args []string) (int, error) {
	if len(args) < 1 {
		return 0, fmt.Errorf("sigil run: no command — use sigil run -- <command> [args...]")
	}

	cfgPath, err := configPathForCmd(cmd)
	if err != nil {
		return 0, err
	}

	if _, err := os.Stat(cfgPath); err != nil {
		if os.IsNotExist(err) {
			return 0, fmt.Errorf("sigil run: no vault.yaml at %s — run `sigil init` first", cfgPath)
		}
		return 0, fmt.Errorf("sigil run: vault.yaml: %w", err)
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return 0, fmt.Errorf("sigil run: read vault.yaml: %w", err)
	}
	if cfg.Secret == "" {
		return 0, fmt.Errorf("sigil run: vault.yaml has no secret — run `sigil init` first")
	}

	sigilHome, err := SigilHome()
	if err != nil {
		return 0, err
	}

	wd, err := runGetwd()
	if err != nil {
		return 0, fmt.Errorf("sigil run: get working directory: %w", err)
	}
	cwd, err := filepath.Abs(wd)
	if err != nil {
		return 0, fmt.Errorf("sigil run: resolve working directory: %w", err)
	}

	slug, err := resolveRunSlug(cfgPath, cwd)
	if err != nil {
		return 0, err
	}

	encPath := vaultstore.EncPath(sigilHome, slug)
	blob, err := os.ReadFile(encPath)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, fmt.Errorf("sigil run: config %q not found at %s — create it via sigil config", slug, encPath)
		}
		return 0, fmt.Errorf("sigil run: read vault: %w", err)
	}

	plaintext, err := crypto.DecryptVault(cfg.Secret, blob)
	if err != nil {
		if errors.Is(err, crypto.ErrDecrypt) {
			return 0, fmt.Errorf("sigil run: cannot decrypt vault (wrong secret or corrupt file)")
		}
		return 0, fmt.Errorf("sigil run: decrypt vault: %w", err)
	}

	vaultVars, err := env.Parse([]byte(plaintext))
	if err != nil {
		return 0, fmt.Errorf("sigil run: parse vault env: %w", err)
	}

	osMap := env.FromSlice(os.Environ())
	merged := env.Merge(osMap, vaultVars, cfg.Inject)
	environ := env.ToSlice(merged)

	return runChildProcess(args[0], args[1:], environ)
}
