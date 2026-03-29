package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/carlos/sigil/internal/config"
	"github.com/carlos/sigil/internal/secret"
	"github.com/carlos/sigil/internal/ui"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize Sigil in the current project",
	Long: `Creates ~/.sigil/vault.yaml (with a generated secret for encrypting vault files),
and an empty vaults directory. No .enc config files are created;
add them after sigil config via Manage configs.`,
	Args: cobra.NoArgs,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, _ []string) error {
	cfgPath, err := configPathForCmd(cmd)
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

	autoKey, err := initGenerateKey(32)
	if err != nil {
		return err
	}

	fmt.Println("Your auto-generated secret (encrypts vault files; share with teammates for the same .enc files):")
	fmt.Println(autoKey)
	fmt.Println()

	useAuto, err := initConfirm("Use this generated secret?", true)
	if err != nil {
		return err
	}

	chosen := autoKey
	if !useAuto {
		s, err := initPromptPassword("Enter your secret")
		if err != nil {
			return err
		}
		if s == "" {
			return fmt.Errorf("encryption key cannot be empty")
		}
		chosen = s
	}

	if err := os.MkdirAll(sigilHome, 0o700); err != nil {
		return err
	}

	vaultsDir := filepath.Join(sigilHome, "vaults")
	if err := os.MkdirAll(vaultsDir, 0o700); err != nil {
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

	defaultEnv := "dev"
	c := &config.Config{
		Project: proj,
		Env:     defaultEnv,
		Secret:  chosen,
		Vaults:  []string{},
		Inject:  map[string]string{},
	}
	if err := config.Save(cfgPath, c); err != nil {
		return err
	}

	ui.Success("Sigil initialized successfully.")
	return nil
}

var (
	initConfirm        = ui.Confirm
	initPromptPassword = ui.PromptPassword
	initGenerateKey    = secret.Generate
)
