package cmd

import (
	"errors"
	"fmt"

	"github.com/mikowitz/greyskull/repository"
	"github.com/mikowitz/greyskull/services"
	"github.com/spf13/cobra"
)

// switchCmd represents the user switch command
var switchCmd = &cobra.Command{
	Use:   "switch <username>",
	Short: "Switch to a different user",
	Long: `Switch to a different user (case-insensitive). The specified user becomes
the current active user for all workout tracking operations.`,
	Args: cobra.ExactArgs(1),
	RunE: switchUser,
}

func switchUser(cmd *cobra.Command, args []string) error {
	username := args[0]

	// Initialize command context with dependency injection
	ctx, err := services.NewCommandContextWithDefaults()
	if err != nil {
		return fmt.Errorf("failed to initialize context: %w", err)
	}

	// Validate user exists (case-insensitive lookup)
	user, err := ctx.UserRepo.Get(username)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return fmt.Errorf("user %q not found", username)
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Set as current user
	if err := ctx.UserRepo.SetCurrent(username); err != nil {
		return fmt.Errorf("failed to set current user: %w", err)
	}

	// Show confirmation with actual username casing
	fmt.Fprintf(cmd.OutOrStdout(), "Switched to user %q.\n", user.Username)
	return nil
}