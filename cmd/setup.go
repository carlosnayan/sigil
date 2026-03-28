package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Create empty encrypted vault files",
	Long:  "Creates ~/.vault/vaults/<name>.<env>.env.enc files. Full implementation in phase 5.",
	RunE: func(cmd *cobra.Command, args []string) error {
		env, _ := cmd.Flags().GetString("env")
		fmt.Printf("vault setup — stub (phase 5) [env=%q]\n", env)
		return nil
	},
}

func init() {
	setupCmd.Flags().String("env", "dev", "environment (e.g. dev, staging, prod)")
	rootCmd.AddCommand(setupCmd)
}
