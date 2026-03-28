package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with the vault password",
	Long:  "Unlocks the GPG agent with the vault password. Full implementation in phase 4.",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("vault login — stub (phase 4)")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
