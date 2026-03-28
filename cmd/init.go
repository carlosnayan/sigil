package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/carlos/sigil/internal/config"
	"github.com/carlos/sigil/internal/crypto"
	"github.com/carlos/sigil/internal/secret"
	"github.com/carlos/sigil/internal/ui"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new vault in the current project",
	Long:  "Creates ~/.vault/secret.enc (passphrase-wrapped secret), encrypted vault file, and vault.yaml.",
	RunE:  runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func initConfigPath(cmd *cobra.Command) (string, error) {
	p, err := cmd.Flags().GetString("config")
	if err != nil {
		return "", err
	}
	if p == "" {
		p = filepath.Join(".", "vault.yaml")
	}
	return filepath.Abs(p)
}

func runInit(cmd *cobra.Command, args []string) error {
	cfgPath, err := initConfigPath(cmd)
	if err != nil {
		return err
	}
	if _, err := os.Stat(cfgPath); err == nil {
		return fmt.Errorf("vault already initialized: %s exists", cfgPath)
	} else if !os.IsNotExist(err) {
		return err
	}

	vaultHome, err := VaultHome()
	if err != nil {
		return err
	}
	secretEncPath := filepath.Join(vaultHome, "secret.enc")
	if _, err := os.Stat(secretEncPath); err == nil {
		return fmt.Errorf("vault user data already exists at %s (secret.enc). Remove ~/.vault to re-init or use an existing project", vaultHome)
	} else if !os.IsNotExist(err) {
		return err
	}

	password, err := ui.PromptConfirmPassword("Enter vault password:", "Confirm vault password:")
	if err != nil {
		return err
	}

	autoSecret, err := secret.Generate(32)
	if err != nil {
		return err
	}

	fmt.Println("Your auto-generated secret (encrypts vault files; share with teammates for the same .env.enc):")
	fmt.Println(autoSecret)
	fmt.Println()

	useAuto, err := ui.Confirm("Use this generated secret?", true)
	if err != nil {
		return err
	}

	chosen := autoSecret
	if !useAuto {
		s, err := ui.PromptPassword("Enter your secret")
		if err != nil {
			return err
		}
		if s == "" {
			return fmt.Errorf("secret cannot be empty")
		}
		chosen = s
	}

	if err := os.MkdirAll(vaultHome, 0o700); err != nil {
		return err
	}

	wrapped, err := crypto.WrapSecret(password, chosen)
	if err != nil {
		return fmt.Errorf("wrap secret: %w", err)
	}
	if err := os.WriteFile(secretEncPath, wrapped, 0o600); err != nil {
		return err
	}

	vaultsDir := filepath.Join(vaultHome, "vaults")
	if err := os.MkdirAll(vaultsDir, 0o700); err != nil {
		return err
	}

	starter := []byte("# Add KEY=value lines below.\n")
	vaultCipher, err := crypto.EncryptVault(chosen, starter)
	if err != nil {
		return fmt.Errorf("encrypt vault: %w", err)
	}

	outFile := filepath.Join(vaultsDir, "default.dev.env.enc")
	if err := os.WriteFile(outFile, vaultCipher, 0o600); err != nil {
		return err
	}

	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	proj := filepath.Base(wd)
	if proj == "." || proj == "/" {
		proj = "my-project"
	}

	c := &config.Config{
		Project: proj,
		Env:     "dev",
		Vaults:  []string{"default"},
		Inject:  map[string]string{},
	}
	if err := config.Save(cfgPath, c); err != nil {
		return err
	}

	ui.Success("Vault initialized successfully.")
	return nil
}
