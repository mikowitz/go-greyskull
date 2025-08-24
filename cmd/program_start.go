package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mikowitz/greyskull/models"
	"github.com/mikowitz/greyskull/program"
	"github.com/mikowitz/greyskull/repository"
	"github.com/spf13/cobra"
)

var programStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a new workout program",
	Long:  "Initialize a new workout program for the current user, setting starting weights for all lifts.",
	RunE:  startProgram,
}

func startProgram(cmd *cobra.Command, args []string) error {
	// Create input reader for user interaction
	inputReader := NewCLIInputReader(cmd.InOrStdin(), cmd.OutOrStdout())
	// Initialize repository
	repo, err := repository.NewJSONUserRepository()
	if err != nil {
		return fmt.Errorf("failed to initialize repository: %w", err)
	}

	// Check for current user
	currentUsername, err := repo.GetCurrent()
	if err != nil {
		if err == repository.ErrNoCurrentUser {
			return fmt.Errorf("no current user set. Please create a user first with 'greyskull user create' or switch to an existing user with 'greyskull user switch <username>'")
		}
		return fmt.Errorf("failed to get current user: %w", err)
	}

	// Load current user
	user, err := repo.Get(currentUsername)
	if err != nil {
		return fmt.Errorf("failed to load user: %w", err)
	}

	// List available programs
	programs := program.List()
	if len(programs) == 0 {
		return fmt.Errorf("no programs available")
	}

	fmt.Fprintln(cmd.OutOrStdout(), "Available programs:")
	for i, prog := range programs {
		fmt.Fprintf(cmd.OutOrStdout(), "%d. %s\n", i+1, prog.Name)
	}

	// Prompt for program selection
	var selection int
	for {
		num, err := inputReader.ReadInt("Select a program (enter number): ")
		if err != nil {
			fmt.Fprintf(cmd.OutOrStdout(), "Invalid input: %v. Please try again.\n", err)
			continue
		}
		if num < 1 || num > len(programs) {
			fmt.Fprintf(cmd.OutOrStdout(), "Invalid selection. Please enter a number between 1 and %d.\n", len(programs))
			continue
		}
		selection = num
		break
	}

	selectedProgram := programs[selection-1]

	// Define core lifts in display order
	lifts := []models.LiftName{
		models.Squat,
		models.Deadlift,
		models.BenchPress,
		models.OverheadPress,
	}

	// Prompt for starting weights
	startingWeights := make(map[models.LiftName]float64)
	for _, lift := range lifts {
		prompt := fmt.Sprintf("Enter starting weight for %s (lbs): ", liftDisplayName(lift))
		weight, err := inputReader.ReadPositiveFloat(prompt)
		if err != nil {
			return fmt.Errorf("failed to get weight for %s: %v", lift, err)
		}
		startingWeights[lift] = weight
	}

	// Create UserProgram
	userProgram := &models.UserProgram{
		ID:              uuid.Must(uuid.NewV7()),
		UserID:          user.ID,
		ProgramID:       selectedProgram.ID,
		StartingWeights: startingWeights,
		CurrentWeights:  make(map[models.LiftName]float64),
		CurrentDay:      1,
		StartedAt:       time.Now(),
	}

	// Copy starting weights to current weights
	for lift, weight := range startingWeights {
		userProgram.CurrentWeights[lift] = weight
	}

	// Update user
	if user.Programs == nil {
		user.Programs = make(map[uuid.UUID]*models.UserProgram)
	}
	user.Programs[userProgram.ID] = userProgram
	user.CurrentProgram = userProgram.ID

	// Save user
	if err := repo.Update(user); err != nil {
		return fmt.Errorf("failed to save user: %w", err)
	}

	// Show success message with day 1 preview
	fmt.Fprintf(cmd.OutOrStdout(), "Program started! %s\n", selectedProgram.Name)
	
	// Get day 1 exercises for preview
	if len(selectedProgram.Workouts) > 0 {
		day1 := selectedProgram.Workouts[0]
		exercises := make([]string, len(day1.Lifts))
		for i, lift := range day1.Lifts {
			exercises[i] = liftDisplayName(lift.LiftName)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Day 1 will be: %s\n", strings.Join(exercises, ", "))
	}

	return nil
}


// liftDisplayName converts LiftName to display-friendly format
func liftDisplayName(lift models.LiftName) string {
	switch lift {
	case models.BenchPress:
		return "Bench Press"
	case models.OverheadPress:
		return "Overhead Press"
	default:
		return string(lift)
	}
}