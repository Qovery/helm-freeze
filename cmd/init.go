package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "create a simple helm-freeze.yaml file",
	Long:  `Create a configuration file with minimal content`,
	Run: func(cmd *cobra.Command, args []string) {
		var minimalConfig = `
charts:
  - name: prometheus-operator
    version: 9.3.1

repos:
  - name: stable
    url: https://charts.helm.sh/stable

destinations:
  - name: default
    path: ./
`
		configFile := "helm-freeze.yaml"

		_, err := os.Stat(configFile)
		if err == nil {
			fmt.Printf("Configuration file %s already exists\n", configFile)
			os.Exit(1)
		}

		f, err := os.Create(configFile)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		_, err = f.WriteString(minimalConfig)
		if err != nil {
			fmt.Println(err)
			f.Close()
			os.Exit(1)
		}

		err = f.Close()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Printf("Configuration file %s created\n", configFile)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
