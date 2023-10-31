package main

import (
	"fmt"
	"os"

	"github.com/bb-consent/api/internal/cmd"
	"github.com/spf13/cobra"
)

func main() {

	var rootCmd = &cobra.Command{Use: "bb-consent-api"}

	// Define the "start-api" command
	var startAPICmd = &cobra.Command{
		Use:   "start-api",
		Short: "Starts the bb consent api server",
		Run:   cmd.StartApiCmdHandler,
	}

	// Define the "config" flag
	startAPICmd.Flags().StringVarP(&cmd.ConfigFileName, "config", "c", "config-development.json", "configuration file")

	// Add the "start-api" commands to the root command
	rootCmd.AddCommand(startAPICmd)

	// Execute the CLI
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
