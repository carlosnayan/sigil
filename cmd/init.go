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
	Short: "Initialize Sigil in the current project",
	Long:  "Creates ~/.sigil/secret.enc (passphrase-wrapped secret), encrypted vault files, and ~/.sigil/vault.yaml.",
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
		p, err = DefaultConfigPath()
		if err != nil {
			return "", err
		}
	}
	return filepath.Abs(p)
}

func runInit(cmd *cobra.Command, args []string) error {
	cfgPath, err := initConfigPath(cmd)
	if err != nil {
		return err
	}
	if _, err := os.Stat(cfgPath); err == nil {
		return fmt.Errorf("sigil: project already initialized: %s exists", cfgPath)
	} else if !os.IsNotExist(err) {
		return err
	}

	sigilHome, err := SigilHome()
	if err != nil {
		return err
	}
	secretEncPath := filepath.Join(sigilHome, "secret.enc")
	if _, err := os.Stat(secretEncPath); err == nil {
		return fmt.Errorf("sigil: user data already exists at %s (secret.enc). Remove ~/.sigil to re-init or use an existing project", sigilHome)
	} else if !os.IsNotExist(err) {
		return err
	}

	password, err := initPromptConfirmPassword("Enter vault password:", "Confirm vault password:")
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

	useAuto, err := initConfirm("Use this generated secret?", true)
	if err != nil {
		return err
	}

	chosen := autoSecret
	if !useAuto {
		s, err := initPromptPassword("Enter your secret")
		if err != nil {
			return err
		}
		if s == "" {
			return fmt.Errorf("secret cannot be empty")
		}
		chosen = s
	}

	if err := os.MkdirAll(sigilHome, 0o700); err != nil {
		return err
	}

	wrapped, err := crypto.WrapSecret(password, chosen)
	if err != nil {
		return fmt.Errorf("wrap secret: %w", err)
	}
	if err := os.WriteFile(secretEncPath, wrapped, 0o600); err != nil {
		return err
	}

	vaultsDir := filepath.Join(sigilHome, "vaults")
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

	ui.Success("Sigil initialized successfully.")
	return nil
}

var (
	initPromptConfirmPassword = ui.PromptConfirmPassword
	initConfirm               = ui.Confirm
	initPromptPassword        = ui.PromptPassword
)
