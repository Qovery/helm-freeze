package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("v%s\n", GetCurrentVersion())
	},
}

func GetCurrentVersion() string {
	return "0.4.1" // ci-version-check
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
