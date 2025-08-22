package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "greyskull",
	Short: "A command-line workout tracker for Greyskull LP",
	Long: `greyskull is a command-line workout tracker specifically designed for the Greyskull LP program.
It helps you manage users, track workout programs, log completed workouts, and automatically 
calculate weight progressions based on your AMRAP performance.`,
	Version: "0.1.0",
	Run: func(cmd *cobra.Command, args []string) {
		// Show help when no subcommand is provided
		cmd.Help()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Add child commands
	rootCmd.AddCommand(userCmd)
}