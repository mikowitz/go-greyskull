package cmd

import (
	"github.com/spf13/cobra"
)

var programCmd = &cobra.Command{
	Use:   "program",
	Short: "Manage workout programs",
	Long:  "Manage workout programs including starting new programs and tracking progress.",
}

func init() {
	rootCmd.AddCommand(programCmd)
	
	// Child commands will be added here
	programCmd.AddCommand(programStartCmd)
}