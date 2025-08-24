package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/mikowitz/greyskull/display"
	"github.com/mikowitz/greyskull/program"
	"github.com/mikowitz/greyskull/repository"
	"github.com/mikowitz/greyskull/workout"
)

var workoutNextCmd = &cobra.Command{
	Use:   "next",
	Short: "Display the next workout",
	Long:  "Display the next workout based on your current program and progress.",
	RunE:  showNextWorkout,
}

func showNextWorkout(cmd *cobra.Command, args []string) error {
	// Load current user
	repo, err := repository.NewJSONUserRepository()
	if err != nil {
		return fmt.Errorf("failed to initialize repository: %w", err)
	}

	currentUsername, err := repo.GetCurrent()
	if err != nil {
		return fmt.Errorf("no current user set. Use 'greyskull user create' or 'greyskull user switch' first")
	}

	user, err := repo.Get(currentUsername)
	if err != nil {
		return fmt.Errorf("failed to load current user: %w", err)
	}

	// Check if user has a current program
	if user.CurrentProgram.String() == "00000000-0000-0000-0000-000000000000" {
		return fmt.Errorf("no active program. Use 'greyskull program start' to begin a program")
	}

	// Get UserProgram and Program
	userProgram, exists := user.Programs[user.CurrentProgram]
	if !exists {
		return fmt.Errorf("current program not found in user programs")
	}

	program, err := program.GetByID(userProgram.ProgramID.String())
	if err != nil {
		return fmt.Errorf("failed to load program: %w", err)
	}

	// Calculate next workout
	nextWorkout, err := workout.CalculateNextWorkout(user, program)
	if err != nil {
		return fmt.Errorf("failed to calculate next workout: %w", err)
	}

	// Display workout
	formatter := display.NewWorkoutFormatter(cmd.OutOrStdout())
	formatter.DisplayWorkout(nextWorkout)

	return nil
}

