package cmd

import (
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:              "sigil",
	Short:            "Sigil — encrypt environment files offline",
	Long:             "Environment files, sealed with a Sigil",
	Version:          "0.1.0",
	TraverseChildren: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().String("config", "", "path to vault.yaml (default: ~/.sigil/vault.yaml)")
}

func SigilHome() (string, error) {
	h, err := homedir.Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(h, ".sigil"), nil
}

func DefaultConfigPath() (string, error) {
	dir, err := SigilHome()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "vault.yaml"), nil
}
