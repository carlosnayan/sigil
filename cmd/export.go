package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export secrets to stdout",
	Long:  "env or json format. Full implementation in phase 10.",
	RunE: func(cmd *cobra.Command, args []string) error {
		format, _ := cmd.Flags().GetString("format")
		fmt.Printf("sigil export — stub (phase 10) [format=%q]\n", format)
		return nil
	},
}

func init() {
	exportCmd.Flags().String("format", "env", "output format: env or json")
	rootCmd.AddCommand(exportCmd)
}
