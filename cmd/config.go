package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/carlos/sigil/internal/config"
	"github.com/carlos/sigil/internal/ui"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Open the interactive vault menu",
	Long: `Reads the secret from ~/.sigil/vault.yaml and opens an interactive menu
(arrow keys) to manage configs, import/export, rekey, and diagnostics.`,
	RunE: runConfig,
}

func init() {
	rootCmd.AddCommand(configCmd)
}

func configPathForCmd(cmd *cobra.Command) (string, error) {
	p, err := cmd.Flags().GetString("config")
	if err != nil {
		return "", err
	}
	if p == "" {
		p, err = DefaultConfigPath()
		if err != nil {
			return "", err
		}
	}
	return filepath.Abs(p)
}

func runConfig(cmd *cobra.Command, _ []string) error {
	cfgPath, err := configPathForCmd(cmd)
	if err != nil {
		return err
	}

	if _, err := os.Stat(cfgPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("sigil: no vault.yaml at %s — run `sigil init` first", cfgPath)
		}
		return fmt.Errorf("sigil: vault.yaml: %w", err)
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("sigil: read vault.yaml: %w", err)
	}
	if cfg.Secret == "" {
		return fmt.Errorf("sigil: vault.yaml has no secret — run `sigil init` first")
	}

	sigilHome, err := SigilHome()
	if err != nil {
		return err
	}

	encryptionKey := cfg.Secret

	for {
		idx, _, err := configVaultMenu()
		if err != nil {
			if errors.Is(err, promptui.ErrInterrupt) || errors.Is(err, promptui.ErrEOF) {
				fmt.Fprintln(os.Stdout, "Session ended.")
				return nil
			}
			return err
		}

		switch idx {
		case 0:
			if err := runManageSecrets(sigilHome, encryptionKey); err != nil {
				return err
			}
		case 1, 2, 3, 4:
			menuStub(ui.VaultMenuItems[idx], encryptionKey)
		case 5:
			ui.Success("Session closed.")
			return nil
		default:
			ui.Warn("Unknown menu option")
		}
	}
}

func menuStub(action string, encryptionKey string) {
	_ = encryptionKey
	ui.Warn(action + " — not implemented yet")
}

var configVaultMenu = ui.VaultMenu
