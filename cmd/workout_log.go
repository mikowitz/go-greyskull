package cmd

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mikowitz/greyskull/display"
	"github.com/mikowitz/greyskull/models"
	"github.com/mikowitz/greyskull/services"
	"github.com/mikowitz/greyskull/workout"
	"github.com/spf13/cobra"
)

var workoutLogCmd = &cobra.Command{
	Use:   "log",
	Short: "Log a completed workout",
	Long:  `Log a completed workout for your current program.

By default, assumes all non-AMRAP sets were completed successfully.
Use --fail flag to record individual reps for each set.`,
	RunE:  logWorkout,
}

func init() {
	workoutLogCmd.Flags().Bool("fail", false, "Record individual reps for each set")
}

func logWorkout(cmd *cobra.Command, args []string) error {
	// Initialize command context with dependency injection
	ctx, err := services.NewCommandContextWithDefaults()
	if err != nil {
		return fmt.Errorf("failed to initialize context: %w", err)
	}

	// Load current user, program, and user program in one call
	user, userProgram, program, err := ctx.UserService.GetCurrentUserWithProgram()
	if err != nil {
		return err
	}

	// Calculate and display the next workout
	nextWorkout, err := workout.CalculateNextWorkout(user, program)
	if err != nil {
		return fmt.Errorf("failed to calculate next workout: %w", err)
	}

	// Display the workout like the "next" command
	formatter := display.NewWorkoutFormatter(cmd.OutOrStdout())
	formatter.DisplayWorkout(nextWorkout)

	// Check for --fail flag to determine collection mode
	failMode, err := cmd.Flags().GetBool("fail")
	if err != nil {
		return fmt.Errorf("failed to get fail flag: %w", err)
	}

	var completedWorkout *models.Workout
	if failMode {
		// Collect reps for every set individually
		completedWorkout, err = collectWithFailure(cmd, nextWorkout)
		if err != nil {
			return fmt.Errorf("failed to collect workout data: %w", err)
		}
	} else {
		// Collect AMRAP reps only (normal mode)
		amrapReps, err := collectAMRAPReps(cmd, nextWorkout)
		if err != nil {
			return fmt.Errorf("failed to collect AMRAP reps: %w", err)
		}
		// Create completed workout with auto-completed sets
		completedWorkout = buildCompletedWorkout(nextWorkout, amrapReps)
	}

	// Add to user's workout history
	user.WorkoutHistory = append(user.WorkoutHistory, *completedWorkout)

	// Calculate weight progression based on AMRAP performance
	newWeights, err := workout.CalculateProgression(completedWorkout, userProgram.CurrentWeights, &program.ProgressionRules)
	if err != nil {
		return fmt.Errorf("failed to calculate progression: %w", err)
	}

	// Display weight changes
	formatter.DisplayWeightChanges(userProgram.CurrentWeights, newWeights)

	// Update current weights
	userProgram.CurrentWeights = newWeights

	// Increment CurrentDay (with wrapping)
	nextDay := userProgram.CurrentDay + 1
	if nextDay > len(program.Workouts) {
		nextDay = 1
	}
	userProgram.CurrentDay = nextDay

	// Save user
	err = ctx.UserRepo.Update(user)
	if err != nil {
		return fmt.Errorf("failed to save workout: %w", err)
	}

	// Show completion summary
	cmd.Printf("\nWorkout logged successfully!\n")
	cmd.Printf("Next workout: Day %d\n", nextDay)

	return nil
}


// collectAMRAPReps prompts user for AMRAP set completion
func collectAMRAPReps(cmd *cobra.Command, nextWorkout *models.Workout) (map[models.LiftName]int, error) {
	amrapReps := make(map[models.LiftName]int)

	// Create input reader for user interaction
	inputReader := NewCLIInputReader(cmd.InOrStdin(), cmd.OutOrStdout())

	for _, exercise := range nextWorkout.Exercises {
		// Find AMRAP sets
		for _, set := range exercise.Sets {
			if set.Type == models.AMRAPSet {
				prompt := fmt.Sprintf("How many reps did you complete for %s AMRAP set (%d+)? ", 
					display.FormatLiftName(exercise.LiftName), set.TargetReps)
				
				value, err := inputReader.ReadPositiveInt(prompt)
				if err != nil {
					return nil, fmt.Errorf("failed to read AMRAP reps for %s: %w", exercise.LiftName, err)
				}
				
				amrapReps[exercise.LiftName] = value
				break // Only one AMRAP set per exercise
			}
		}
	}

	return amrapReps, nil
}

// collectWithFailure prompts user for actual reps on every set
func collectWithFailure(cmd *cobra.Command, nextWorkout *models.Workout) (*models.Workout, error) {
	// Create input reader for user interaction
	inputReader := NewCLIInputReader(cmd.InOrStdin(), cmd.OutOrStdout())

	// Create completed workout structure
	completed := &models.Workout{
		ID:            uuid.Must(uuid.NewV7()),
		UserProgramID: nextWorkout.UserProgramID,
		Day:           nextWorkout.Day,
		Exercises:     make([]models.Lift, len(nextWorkout.Exercises)),
		EnteredAt:     time.Now(),
	}

	for i, exercise := range nextWorkout.Exercises {
		cmd.Printf("\n%s:\n", display.FormatLiftName(exercise.LiftName))
		
		completedExercise := models.Lift{
			ID:       uuid.Must(uuid.NewV7()),
			LiftName: exercise.LiftName,
			Sets:     make([]models.Set, len(exercise.Sets)),
		}

		for j, set := range exercise.Sets {
			// Format set type for display
			setTypeStr := "Working"
			if set.Type == models.WarmupSet {
				setTypeStr = "Warmup"
			} else if set.Type == models.AMRAPSet {
				setTypeStr = "AMRAP"
			}

			prompt := fmt.Sprintf("%s - Set %d (%s):\nTarget: %d reps @ %s lbs\nHow many reps completed? ", 
				display.FormatLiftName(exercise.LiftName), 
				set.Order,
				setTypeStr,
				set.TargetReps, 
				display.FormatWeight(set.Weight))
			
			value, err := inputReader.ReadInt(prompt)
			if err != nil {
				return nil, fmt.Errorf("failed to read reps for %s set %d: %w", exercise.LiftName, set.Order, err)
			}

			if value < 0 {
				return nil, fmt.Errorf("number cannot be negative for %s set %d", exercise.LiftName, set.Order)
			}
			
			// Create completed set
			completedSet := models.Set{
				ID:         uuid.Must(uuid.NewV7()),
				Weight:     set.Weight,
				TargetReps: set.TargetReps,
				ActualReps: value, // Use the actual reps entered by user
				Type:       set.Type,
				Order:      set.Order,
			}

			completedExercise.Sets[j] = completedSet
		}

		completed.Exercises[i] = completedExercise
	}

	return completed, nil
}


// buildCompletedWorkout creates a completed workout from template with AMRAP reps filled in
func buildCompletedWorkout(template *models.Workout, amrapReps map[models.LiftName]int) *models.Workout {
	completed := &models.Workout{
		ID:            uuid.Must(uuid.NewV7()),
		UserProgramID: template.UserProgramID,
		Day:           template.Day,
		Exercises:     make([]models.Lift, len(template.Exercises)),
		EnteredAt:     time.Now(),
	}

	for i, exercise := range template.Exercises {
		completedExercise := models.Lift{
			ID:       uuid.Must(uuid.NewV7()),
			LiftName: exercise.LiftName,
			Sets:     make([]models.Set, len(exercise.Sets)),
		}

		for j, set := range exercise.Sets {
			completedSet := models.Set{
				ID:         uuid.Must(uuid.NewV7()),
				Weight:     set.Weight,
				TargetReps: set.TargetReps,
				Type:       set.Type,
				Order:      set.Order,
			}

			// Set ActualReps based on set type
			if set.Type == models.AMRAPSet {
				// Use AMRAP reps from user input
				completedSet.ActualReps = amrapReps[exercise.LiftName]
			} else {
				// Auto-complete non-AMRAP sets
				completedSet.ActualReps = set.TargetReps
			}

			completedExercise.Sets[j] = completedSet
		}

		completed.Exercises[i] = completedExercise
	}

	return completed
}


