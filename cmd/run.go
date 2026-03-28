package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run [flags] -- <command> [args...]",
	Short: "Run a command with vault variables injected",
	Long: strings.TrimSpace(`
Decrypts vaults, merges variables, and runs the child process.
Example: sigil run -- docker compose up

Full implementation in phase 8.`),
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("sigil run — stub (phase 8) args=%v\n", args)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
