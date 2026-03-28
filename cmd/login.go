package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with the vault password",
	Long:  "Unlocks secret.enc with the passphrase and caches the team secret in memory. Full implementation in phase 4.",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("sigil login — stub (phase 4)")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
