package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:              "vault",
	Short:            "Local secret manager — 100% offline",
	Long:             "Secrets, sealed with a Sigil",
	Version:          "0.1.0",
	TraverseChildren: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return loadProjectConfig(cmd)
	},
}

// Execute roda o comando raiz.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().String("config", "", "path to vault.yaml (default: ./vault.yaml)")
}

func loadProjectConfig(cmd *cobra.Command) error {
	if cmd.Name() == "help" {
		return nil
	}
	cfgPath, err := cmd.Flags().GetString("config")
	if err != nil {
		return err
	}
	if cfgPath == "" {
		cfgPath = filepath.Join(".", "vault.yaml")
	}
	abs, err := filepath.Abs(cfgPath)
	if err != nil {
		return err
	}
	if _, err := os.Stat(abs); os.IsNotExist(err) {
		viper.Reset()
		return nil
	}
	viper.SetConfigFile(abs)
	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		if verbose, _ := cmd.Flags().GetBool("verbose"); verbose {
			_, _ = fmt.Fprintf(os.Stderr, "warning: could not read %s: %v\n", abs, err)
		}
		return nil
	}
	return nil
}

// VaultHome retorna ~/.vault (criação lazy nas fases seguintes).
func VaultHome() (string, error) {
	h, err := homedir.Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(h, ".vault"), nil
}
