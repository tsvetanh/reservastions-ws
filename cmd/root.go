package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

// Root command
var rootCmd = &cobra.Command{
	Use:   "admin-cli",
	Short: "CLI tool to manage halls and reservations",
	Long:  "A command-line tool for managing halls, reservations, utilization, and summary reports.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Use --help to see available commands.")
	},
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

// Register subcommands
func init() {
	rootCmd.AddCommand(HallCmd)
	rootCmd.AddCommand(ReservationCmd)
	rootCmd.AddCommand(SummaryCmd)
	rootCmd.AddCommand(UtilizationCmd)
}
