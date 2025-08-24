package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/mikowitz/greyskull/display"
	"github.com/mikowitz/greyskull/services"
	"github.com/mikowitz/greyskull/workout"
)

var workoutNextCmd = &cobra.Command{
	Use:   "next",
	Short: "Display the next workout",
	Long:  "Display the next workout based on your current program and progress.",
	RunE:  showNextWorkout,
}

func showNextWorkout(cmd *cobra.Command, args []string) error {
	// Initialize command context with dependency injection
	ctx, err := services.NewCommandContextWithDefaults()
	if err != nil {
		return fmt.Errorf("failed to initialize context: %w", err)
	}

	// Load current user, program, and user program in one call
	user, _, program, err := ctx.UserService.GetCurrentUserWithProgram()
	if err != nil {
		return err
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

