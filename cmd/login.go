package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/carlos/sigil/internal/crypto"
	"github.com/carlos/sigil/internal/ui"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate and open the interactive vault menu",
	Long: `Unlocks ~/.sigil/secret.enc with your vault password, then opens an interactive
menu (arrow keys) to manage configs, import/export, rekey, and diagnostics.`,
	RunE: runLogin,
}

func init() {
	rootCmd.AddCommand(loginCmd)
}

func runLogin(cmd *cobra.Command, args []string) error {
	sigilHome, err := SigilHome()
	if err != nil {
		return err
	}
	secretEncPath := filepath.Join(sigilHome, "secret.enc")

	blob, err := os.ReadFile(secretEncPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("sigil: no vault found at %s — run `sigil init` first", secretEncPath)
		}
		return fmt.Errorf("sigil: read secret.enc: %w", err)
	}

	password, err := loginPromptPassword("Enter vault password")
	if err != nil {
		if errors.Is(err, promptui.ErrInterrupt) || errors.Is(err, promptui.ErrEOF) {
			return nil
		}
		return err
	}
	if password == "" {
		return fmt.Errorf("sigil: password cannot be empty")
	}

	encryptionKey, err := crypto.UnwrapSecret(password, blob)
	if err != nil {
		if errors.Is(err, crypto.ErrDecrypt) {
			return fmt.Errorf("sigil: wrong password or corrupted key store (secret.enc)")
		}
		return fmt.Errorf("sigil: unlock key store: %w", err)
	}

	for {
		idx, _, err := loginVaultMenu(true)
		if err != nil {
			if errors.Is(err, promptui.ErrInterrupt) || errors.Is(err, promptui.ErrEOF) {
				fmt.Fprintln(os.Stdout, "🔒 Vault has been locked.")
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
			ui.Success("Session closed. Encryption key cleared from memory.")
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

var (
	loginPromptPassword = ui.PromptPassword
	loginVaultMenu      = ui.VaultMenu
)
