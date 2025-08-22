package cmd

import (
	"errors"
	"fmt"

	"github.com/mikowitz/greyskull/repository"
	"github.com/spf13/cobra"
)

// listCmd represents the user list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all users",
	Long: `List all users in the system. The current active user is marked with an asterisk (*).
Original username casing is preserved in the display.`,
	RunE: listUsers,
}

func listUsers(cmd *cobra.Command, args []string) error {
	// Initialize repository
	repo, err := repository.NewJSONUserRepository()
	if err != nil {
		return fmt.Errorf("failed to initialize repository: %w", err)
	}

	// Get all users
	usernames, err := repo.List()
	if err != nil {
		return fmt.Errorf("failed to list users: %w", err)
	}

	// Check if no users exist
	if len(usernames) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No users found. Use 'greyskull user create' to create your first user.")
		return nil
	}

	// Get current user
	currentUser, err := repo.GetCurrent()
	var hasCurrentUser bool
	if err != nil && !errors.Is(err, repository.ErrNoCurrentUser) {
		return fmt.Errorf("failed to get current user: %w", err)
	}
	hasCurrentUser = err == nil

	// Display users
	fmt.Fprintln(cmd.OutOrStdout(), "Users:")
	for _, username := range usernames {
		marker := " "
		if hasCurrentUser && username == currentUser {
			marker = "*"
		}
		fmt.Fprintf(cmd.OutOrStdout(), "  %s %s\n", marker, username)
	}

	if hasCurrentUser {
		fmt.Fprintf(cmd.OutOrStdout(), "\n* Current user: %s\n", currentUser)
	} else {
		fmt.Fprintln(cmd.OutOrStdout(), "\nNo current user set. Use 'greyskull user switch <username>' to set one.")
	}

	return nil
}