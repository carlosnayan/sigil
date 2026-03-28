package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var importCmd = &cobra.Command{
	Use:   "import <file>",
	Short: "Import variables from a .env file",
	Long:  "Reads a file and updates vaults. Full implementation in phase 10.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("sigil import — stub (phase 10) [file=%q]\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(importCmd)
}
