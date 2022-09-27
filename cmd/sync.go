package cmd

import (
	"fmt"
	"os"

	"github.com/Qovery/helm-freeze/cfg"
	"github.com/Qovery/helm-freeze/exec"
	"github.com/spf13/cobra"
)

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync locally from the helm-freeze.yaml file",
	Long: `Download charts and un compress in the desired folders from the given configuration file information.
Running a git diff then will help to see any differences`,
	Run: func(cmd *cobra.Command, args []string) {
		errorsDuringSync := false
		configFile, _ := cmd.Flags().GetString("config-file")
		config, err := cfg.ValidateConfig(configFile)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		onlyCharts, _ := cmd.Flags().GetStringSlice("only-charts")
		if len(onlyCharts) > 0 {
			config = cfg.FilterCharts(config, onlyCharts)
		}

		err = exec.AddAllRepos(config)
		if err != nil {
			fmt.Printf("Error message: %s", err)
			os.Exit(1)
		}

		err = exec.HelmRepoUpdate()
		if err != nil {
			fmt.Printf("Error message: %s", err)
			errorsDuringSync = true
		}

		err = exec.GetAllCharts(config, configFile)
		if err != nil {
			fmt.Printf("Error message: %s", err)
			os.Exit(1)
		}

		if errorsDuringSync {
			fmt.Println("\nSync ended with errors")
			os.Exit(1)
		}
		fmt.Println("\nSync succeed!")
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)

	syncCmd.Flags().StringP("config-file", "f", "./helm-freeze.yaml", "Configuration file")
	syncCmd.Flags().StringSlice("only-charts", []string{}, "Sync only specified charts, comma separated")
}
