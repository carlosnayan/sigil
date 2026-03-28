package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Print effective environment variables from the vault",
	Long:  "Text output with --plain or --json. Full implementation in phase 9.",
	RunE: func(cmd *cobra.Command, args []string) error {
		plain, _ := cmd.Flags().GetBool("plain")
		jsonOut, _ := cmd.Flags().GetBool("json")
		envFlag, _ := cmd.Flags().GetString("env")
		fmt.Printf("vault env — stub (phase 9) [plain=%v json=%v env=%q]\n", plain, jsonOut, envFlag)
		return nil
	},
}

func init() {
	envCmd.Flags().Bool("plain", false, "KEY=VALUE per line")
	envCmd.Flags().Bool("json", false, "JSON output")
	envCmd.Flags().String("env", "", "environment (defaults to vault.yaml when empty)")
	rootCmd.AddCommand(envCmd)
}
