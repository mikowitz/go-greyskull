package cmd

import (
	"github.com/spf13/cobra"
)

// userCmd represents the user command
var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Manage users",
	Long: `Manage users in the Greyskull LP tracker. Users store their workout programs,
progress, and history. Commands include creating new users, switching between users,
and listing existing users.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Show help when no subcommand is provided
		cmd.Help()
	},
}

func init() {
	// Add child commands
	userCmd.AddCommand(createCmd)
	userCmd.AddCommand(switchCmd) 
	userCmd.AddCommand(listCmd)
}