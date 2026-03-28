package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new vault in the current project",
	Long:  "Creates initial project configuration (vault.yaml). Full implementation in phase 3.",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("vault init — stub (phase 3)")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
