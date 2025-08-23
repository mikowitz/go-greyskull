package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mikowitz/greyskull/models"
	"github.com/mikowitz/greyskull/program"
	"github.com/mikowitz/greyskull/repository"
	"github.com/mikowitz/greyskull/workout"
	"github.com/spf13/cobra"
)

var workoutLogCmd = &cobra.Command{
	Use:   "log",
	Short: "Log a completed workout",
	Long:  "Log a completed workout for your current program.",
	RunE:  logWorkout,
}

func logWorkout(cmd *cobra.Command, args []string) error {
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

	// Calculate and display the next workout
	nextWorkout, err := workout.CalculateNextWorkout(user, program)
	if err != nil {
		return fmt.Errorf("failed to calculate next workout: %w", err)
	}

	// Display the workout like the "next" command
	displayWorkout(cmd, nextWorkout)

	// Collect AMRAP reps from user input
	amrapReps, err := collectAMRAPReps(cmd, nextWorkout)
	if err != nil {
		return fmt.Errorf("failed to collect AMRAP reps: %w", err)
	}

	// Create completed workout
	completedWorkout := buildCompletedWorkout(nextWorkout, amrapReps)

	// Add to user's workout history
	user.WorkoutHistory = append(user.WorkoutHistory, *completedWorkout)

	// Calculate weight progression based on AMRAP performance
	newWeights, err := workout.CalculateProgression(completedWorkout, userProgram.CurrentWeights, &program.ProgressionRules)
	if err != nil {
		return fmt.Errorf("failed to calculate progression: %w", err)
	}

	// Display weight changes
	displayWeightChanges(cmd, userProgram.CurrentWeights, newWeights)

	// Update current weights
	userProgram.CurrentWeights = newWeights

	// Increment CurrentDay (with wrapping)
	nextDay := userProgram.CurrentDay + 1
	if nextDay > len(program.Workouts) {
		nextDay = 1
	}
	userProgram.CurrentDay = nextDay

	// Save user
	err = repo.Update(user)
	if err != nil {
		return fmt.Errorf("failed to save workout: %w", err)
	}

	// Show completion summary
	cmd.Printf("\nWorkout logged successfully!\n")
	cmd.Printf("Next workout: Day %d\n", nextDay)

	return nil
}

// displayWorkout is defined in workout_next.go and can be reused here

// collectAMRAPReps prompts user for AMRAP set completion
func collectAMRAPReps(cmd *cobra.Command, nextWorkout *models.Workout) (map[models.LiftName]int, error) {
	amrapReps := make(map[models.LiftName]int)

	// Use a single scanner for the entire input stream
	scanner := bufio.NewScanner(cmd.InOrStdin())

	for _, exercise := range nextWorkout.Exercises {
		// Find AMRAP sets
		for _, set := range exercise.Sets {
			if set.Type == models.AMRAPSet {
				prompt := fmt.Sprintf("How many reps did you complete for %s AMRAP set (%d+)? ", 
					formatLiftName(exercise.LiftName), set.TargetReps)
				
				// Display prompt to command output
				cmd.Print(prompt)
				
				// Read from scanner
				if !scanner.Scan() {
					if err := scanner.Err(); err != nil {
						return nil, fmt.Errorf("failed to read input for %s: %w", exercise.LiftName, err)
					}
					return nil, fmt.Errorf("no input available for %s", exercise.LiftName)
				}

				input := strings.TrimSpace(scanner.Text())
				if input == "" {
					return nil, fmt.Errorf("input cannot be empty for %s", exercise.LiftName)
				}

				value, err := strconv.Atoi(input)
				if err != nil {
					return nil, fmt.Errorf("invalid number for %s: %s", exercise.LiftName, input)
				}

				if value <= 0 {
					return nil, fmt.Errorf("number must be positive for %s", exercise.LiftName)
				}
				
				amrapReps[exercise.LiftName] = value
				break // Only one AMRAP set per exercise
			}
		}
	}

	return amrapReps, nil
}

// promptInt prompts for a positive integer input
func promptInt(prompt string) (int, error) {
	reader := bufio.NewReader(os.Stdin)
	
	if prompt != "" {
		fmt.Print(prompt)
	}
	input, err := reader.ReadString('\n')
	if err != nil {
		return 0, fmt.Errorf("failed to read input: %w", err)
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return 0, fmt.Errorf("input cannot be empty")
	}

	value, err := strconv.Atoi(input)
	if err != nil {
		return 0, fmt.Errorf("invalid number: %s", input)
	}

	if value <= 0 {
		return 0, fmt.Errorf("number must be positive")
	}

	return value, nil
}

// promptIntWithReader is a testable version of promptInt that accepts an io.Reader
func promptIntWithReader(prompt string, reader io.Reader) (int, error) {
	scanner := bufio.NewScanner(reader)
	
	if prompt != "" {
		fmt.Print(prompt)
	}
	
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return 0, fmt.Errorf("failed to read input: %w", err)
		}
		return 0, fmt.Errorf("no input available")
	}

	input := strings.TrimSpace(scanner.Text())
	if input == "" {
		return 0, fmt.Errorf("input cannot be empty")
	}

	value, err := strconv.Atoi(input)
	if err != nil {
		return 0, fmt.Errorf("invalid number: %s", input)
	}

	if value <= 0 {
		return 0, fmt.Errorf("number must be positive")
	}

	return value, nil
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

// displayWeightChanges shows the progression changes to the user
func displayWeightChanges(cmd *cobra.Command, oldWeights, newWeights map[models.LiftName]float64) {
	hasChanges := false
	
	// Check if any weights changed
	for liftName, newWeight := range newWeights {
		if oldWeight, exists := oldWeights[liftName]; exists && oldWeight != newWeight {
			hasChanges = true
			break
		}
	}
	
	if !hasChanges {
		return // No changes to display
	}
	
	cmd.Printf("\nWeight Updates:\n")
	
	// Display changes for each lift that was worked
	lifts := []models.LiftName{models.OverheadPress, models.BenchPress, models.Squat, models.Deadlift}
	for _, liftName := range lifts {
		oldWeight, oldExists := oldWeights[liftName]
		newWeight, newExists := newWeights[liftName]
		
		if oldExists && newExists && oldWeight != newWeight {
			difference := newWeight - oldWeight
			var sign string
			if difference > 0 {
				sign = "+"
			}
			
			cmd.Printf("%s: %s → %s lbs (%s%.1f)\n", 
				formatLiftName(liftName),
				formatWeight(oldWeight),
				formatWeight(newWeight),
				sign,
				difference)
		}
	}
}

// formatWeight and formatLiftName are defined in workout_next.go
