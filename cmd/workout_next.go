package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/mikowitz/greyskull/models"
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
	displayWorkout(cmd, nextWorkout)

	return nil
}

func displayWorkout(cmd *cobra.Command, workout *models.Workout) {
	cmd.Printf("Day %d Workout:\n", workout.Day)
	cmd.Printf("================\n\n")

	for _, lift := range workout.Exercises {
		cmd.Printf("%s:\n", formatLiftName(lift.LiftName))

		// Group sets by type
		warmupSets := []models.Set{}
		workingSets := []models.Set{}

		for _, set := range lift.Sets {
			if set.Type == models.WarmupSet {
				warmupSets = append(warmupSets, set)
			} else {
				workingSets = append(workingSets, set)
			}
		}

		// Display warmup sets if any
		if len(warmupSets) > 0 {
			cmd.Printf("  Warmup:\n")
			for _, set := range warmupSets {
				cmd.Printf("    %d reps @ %s lbs\n", set.TargetReps, formatWeight(set.Weight))
			}
		}

		// Display working sets
		cmd.Printf("  Working Sets:\n")
		for i, set := range workingSets {
			if set.Type == models.AMRAPSet {
				cmd.Printf("    Set %d: %d+ reps @ %s lbs (AMRAP)\n", i+1, set.TargetReps, formatWeight(set.Weight))
			} else {
				cmd.Printf("    Set %d: %d reps @ %s lbs\n", i+1, set.TargetReps, formatWeight(set.Weight))
			}
		}

		cmd.Printf("\n")
	}
}

func formatWeight(weight float64) string {
	// Remove decimal if it's a whole number
	if weight == float64(int(weight)) {
		return strconv.Itoa(int(weight))
	}
	return fmt.Sprintf("%.1f", weight)
}

func formatLiftName(lift models.LiftName) string {
	switch lift {
	case models.Squat:
		return "Squat"
	case models.Deadlift:
		return "Deadlift"
	case models.BenchPress:
		return "Bench Press"
	case models.OverheadPress:
		return "Overhead Press"
	default:
		return string(lift)
	}
}