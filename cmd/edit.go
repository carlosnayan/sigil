package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Open an editor to edit secrets",
	Long:  "Decrypts, merges envs, opens $EDITOR, and re-encrypts. Full implementation in phase 6.",
	RunE: func(cmd *cobra.Command, args []string) error {
		env, _ := cmd.Flags().GetString("env")
		fmt.Printf("sigil edit — stub (phase 6) [env=%q]\n", env)
		return nil
	},
}

func init() {
	editCmd.Flags().String("env", "", "environment (defaults to vault.yaml when empty)")
	rootCmd.AddCommand(editCmd)
}
