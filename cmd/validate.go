package cmd

import (
	"fmt"
	"github.com/Qovery/helm-freeze/cfg"
	"os"

	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration",
	Long:  `Ensure the configuration is valid`,
	Run: func(cmd *cobra.Command, args []string) {
		configFile, _ := cmd.Flags().GetString("config-file")
		_, err := cfg.ValidateConfig(configFile)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println("Configuration is valid")
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
	validateCmd.Flags().StringP("config-file", "f", "./helm-freeze.yaml", "Configuration file")
}
