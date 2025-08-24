package cmd

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/mikowitz/greyskull/models"
	"github.com/mikowitz/greyskull/repository"
	"github.com/spf13/cobra"
)

// createCmd represents the user create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new user",
	Long: `Create a new user for tracking workouts. The username must be filesystem-safe
and will be stored case-insensitively. After creation, the user will be set as the current user.`,
	RunE: createUser,
}

func createUser(cmd *cobra.Command, args []string) error {
	// Create input reader for user interaction
	inputReader := NewCLIInputReader(cmd.InOrStdin(), cmd.OutOrStdout())
	// Initialize repository
	repo, err := repository.NewJSONUserRepository()
	if err != nil {
		return fmt.Errorf("failed to initialize repository: %w", err)
	}

	// Prompt for username
	username, err := inputReader.ReadLine("Enter username: ")
	if err != nil {
		return fmt.Errorf("failed to read username: %w", err)
	}

	// Validate username
	if err := validateUsername(username); err != nil {
		return err
	}

	// Check for case-insensitive duplicates
	if _, err := repo.Get(username); err == nil {
		return fmt.Errorf("user %q already exists (case-insensitive)", username)
	} else if !errors.Is(err, repository.ErrUserNotFound) {
		return fmt.Errorf("failed to check for existing user: %w", err)
	}

	// Create user
	user := &models.User{
		ID:             uuid.New(),
		Username:       username,
		CurrentProgram: uuid.Nil,
		Programs:       make(map[uuid.UUID]*models.UserProgram),
		WorkoutHistory: []models.Workout{},
		CreatedAt:      time.Now(),
	}

	// Validate user
	if err := user.Validate(); err != nil {
		return fmt.Errorf("invalid user data: %w", err)
	}

	// Save user
	if err := repo.Create(user); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	// Set as current user
	if err := repo.SetCurrent(username); err != nil {
		return fmt.Errorf("failed to set current user: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "User %q created successfully and set as current user.\n", username)
	return nil
}

// validateUsername ensures the username is not empty and filesystem-safe
func validateUsername(username string) error {
	if username == "" {
		return errors.New("username cannot be empty")
	}

	// Check for filesystem-unsafe characters: / \ : * ? " < > |
	unsafeChars := regexp.MustCompile(`[/\\:*?"<>|]`)
	if unsafeChars.MatchString(username) {
		return errors.New("username contains invalid characters (/, \\, :, *, ?, \", <, >, |)")
	}

	return nil
}

