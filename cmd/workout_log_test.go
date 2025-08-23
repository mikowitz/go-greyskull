package cmd

import (
	"bytes"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mikowitz/greyskull/models"
	"github.com/mikowitz/greyskull/program"
	"github.com/mikowitz/greyskull/repository"
	"github.com/mikowitz/greyskull/workout"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkoutLog_NoCurrentUser(t *testing.T) {
	_ = setupTestEnv(t)

	cmd := workoutLogCmd
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.RunE(cmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no current user")
}

func TestWorkoutLog_NoActiveProgram(t *testing.T) {
	_ = setupTestEnv(t)

	// Create and set current user without a program
	repo, err := repository.NewJSONUserRepository()
	require.NoError(t, err)

	user := &models.User{
		ID:             uuid.New(),
		Username:       "TestUser",
		CurrentProgram: uuid.Nil,
		Programs:       make(map[uuid.UUID]*models.UserProgram),
		WorkoutHistory: []models.Workout{},
		CreatedAt:      time.Now(),
	}

	err = repo.Create(user)
	require.NoError(t, err)

	err = repo.SetCurrent("TestUser")
	require.NoError(t, err)

	cmd := workoutLogCmd
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err = cmd.RunE(cmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no active program")
}

func TestWorkoutLog_SuccessfulLoggingFlow(t *testing.T) {
	_ = setupTestEnv(t)

	// Create and set current user with active program
	repo, err := repository.NewJSONUserRepository()
	require.NoError(t, err)

	user := &models.User{
		ID:             uuid.New(),
		Username:       "TestUser",
		CurrentProgram: uuid.Nil,
		Programs:       make(map[uuid.UUID]*models.UserProgram),
		WorkoutHistory: []models.Workout{},
		CreatedAt:      time.Now(),
	}

	// Create UserProgram with good weights (should trigger warmups)
	userProgram := &models.UserProgram{
		ID:        uuid.Must(uuid.NewV7()),
		UserID:    user.ID,
		ProgramID: uuid.Must(uuid.Parse("550e8400-e29b-41d4-a716-446655440000")), // Greyskull LP
		StartingWeights: map[models.LiftName]float64{
			models.Squat:         135.0,
			models.Deadlift:      185.0,
			models.BenchPress:    125.0,
			models.OverheadPress: 95.0,
		},
		CurrentWeights: map[models.LiftName]float64{
			models.Squat:         135.0,
			models.Deadlift:      185.0,
			models.BenchPress:    125.0,
			models.OverheadPress: 95.0,
		},
		CurrentDay: 1, // Day 1: OverheadPress, Squat
		StartedAt:  time.Now(),
	}

	user.Programs[userProgram.ID] = userProgram
	user.CurrentProgram = userProgram.ID

	err = repo.Create(user)
	require.NoError(t, err)

	err = repo.SetCurrent("TestUser")
	require.NoError(t, err)

	// The command should exist and be callable
	assert.NotNil(t, workoutLogCmd)
	assert.Equal(t, "log", workoutLogCmd.Use)

	// Command should run without error when user inputs valid AMRAP reps
	// Mock AMRAP input: 8 reps for OverheadPress, 7 reps for Squat
	input := strings.NewReader("8\n7\n")
	
	var buf bytes.Buffer
	cmd := workoutLogCmd
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetIn(input)

	err = cmd.RunE(cmd, []string{})
	assert.NoError(t, err, "Workout log command should run successfully")
}

func TestWorkoutLog_AMRAPInputValidation(t *testing.T) {
	// Test the promptInt helper function for AMRAP input validation
	tests := []struct {
		name        string
		input       string
		expected    int
		shouldError bool
	}{
		{
			name:        "valid positive integer",
			input:       "8",
			expected:    8,
			shouldError: false,
		},
		{
			name:        "valid single digit",
			input:       "5",
			expected:    5,
			shouldError: false,
		},
		{
			name:        "valid large number",
			input:       "15",
			expected:    15,
			shouldError: false,
		},
		{
			name:        "zero should be invalid",
			input:       "0",
			expected:    0,
			shouldError: true,
		},
		{
			name:        "negative number should be invalid",
			input:       "-1",
			expected:    0,
			shouldError: true,
		},
		{
			name:        "non-numeric input should be invalid",
			input:       "abc",
			expected:    0,
			shouldError: true,
		},
		{
			name:        "decimal input should be invalid",
			input:       "5.5",
			expected:    0,
			shouldError: true,
		},
		{
			name:        "empty input should be invalid",
			input:       "",
			expected:    0,
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a strings.Reader to simulate user input
			input := strings.NewReader(tt.input + "\n")

			// This will test the promptInt function that should be implemented
			result, err := promptIntWithReader("Enter reps: ", input)

			if tt.shouldError {
				assert.Error(t, err, "promptInt should return error for invalid input: %s", tt.input)
			} else {
				assert.NoError(t, err, "promptInt should not return error for valid input: %s", tt.input)
				assert.Equal(t, tt.expected, result, "promptInt should return correct value for input: %s", tt.input)
			}
		})
	}
}

func TestWorkoutLog_WorkoutSavedToHistory(t *testing.T) {
	_ = setupTestEnv(t)

	// Create and set current user with active program
	repo, err := repository.NewJSONUserRepository()
	require.NoError(t, err)

	user := &models.User{
		ID:             uuid.New(),
		Username:       "TestUser",
		CurrentProgram: uuid.Nil,
		Programs:       make(map[uuid.UUID]*models.UserProgram),
		WorkoutHistory: []models.Workout{}, // Initially empty
		CreatedAt:      time.Now(),
	}

	userProgram := &models.UserProgram{
		ID:        uuid.Must(uuid.NewV7()),
		UserID:    user.ID,
		ProgramID: uuid.Must(uuid.Parse("550e8400-e29b-41d4-a716-446655440000")),
		StartingWeights: map[models.LiftName]float64{
			models.Squat:         135.0,
			models.Deadlift:      185.0,
			models.BenchPress:    125.0,
			models.OverheadPress: 95.0,
		},
		CurrentWeights: map[models.LiftName]float64{
			models.Squat:         135.0,
			models.Deadlift:      185.0,
			models.BenchPress:    125.0,
			models.OverheadPress: 95.0,
		},
		CurrentDay: 1,
		StartedAt:  time.Now(),
	}

	user.Programs[userProgram.ID] = userProgram
	user.CurrentProgram = userProgram.ID

	err = repo.Create(user)
	require.NoError(t, err)

	err = repo.SetCurrent("TestUser")
	require.NoError(t, err)

	// Verify initial state
	assert.Len(t, user.WorkoutHistory, 0, "WorkoutHistory should initially be empty")

	// Mock AMRAP input: 8 reps for OverheadPress, 7 reps for Squat
	input := strings.NewReader("8\n7\n")

	var buf bytes.Buffer
	cmd := workoutLogCmd
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetIn(input)

	err = cmd.RunE(cmd, []string{})
	require.NoError(t, err, "Workout log command should complete successfully")

	// Reload user from repository to check saved state
	updatedUser, err := repo.Get("TestUser")
	require.NoError(t, err)

	// Verify workout was saved to history
	assert.Len(t, updatedUser.WorkoutHistory, 1, "WorkoutHistory should have 1 workout after logging")

	savedWorkout := updatedUser.WorkoutHistory[0]
	assert.Equal(t, userProgram.ID, savedWorkout.UserProgramID, "Workout should have correct UserProgramID")
	assert.Equal(t, 1, savedWorkout.Day, "Workout should have correct Day")
	assert.Len(t, savedWorkout.Exercises, 2, "Workout should have 2 exercises (OverheadPress, Squat)")

	// Check that EnteredAt timestamp is recent (within last 5 seconds)
	assert.WithinDuration(t, time.Now(), savedWorkout.EnteredAt, 5*time.Second, "EnteredAt should be recent")

	// Verify exercises are correct for Day 1
	exerciseNames := make([]models.LiftName, len(savedWorkout.Exercises))
	for i, exercise := range savedWorkout.Exercises {
		exerciseNames[i] = exercise.LiftName
	}
	assert.Contains(t, exerciseNames, models.OverheadPress, "Day 1 should include OverheadPress")
	assert.Contains(t, exerciseNames, models.Squat, "Day 1 should include Squat")

	// Verify set completion for each exercise
	for _, exercise := range savedWorkout.Exercises {
		assert.NotEmpty(t, exercise.Sets, "Exercise should have sets")

		for _, set := range exercise.Sets {
			assert.True(t, set.IsComplete(), "All sets should be marked as complete (ActualReps > 0)")

			if set.Type == models.AMRAPSet {
				// AMRAP sets should have the mocked input values
				if exercise.LiftName == models.OverheadPress {
					assert.Equal(t, 8, set.ActualReps, "OverheadPress AMRAP should have 8 reps")
				} else if exercise.LiftName == models.Squat {
					assert.Equal(t, 7, set.ActualReps, "Squat AMRAP should have 7 reps")
				}
			} else {
				// Non-AMRAP sets should have ActualReps = TargetReps
				assert.Equal(t, set.TargetReps, set.ActualReps, "Non-AMRAP sets should have ActualReps = TargetReps")
			}
		}
	}
}

func TestWorkoutLog_CurrentDayIncrements(t *testing.T) {
	_ = setupTestEnv(t)

	repo, err := repository.NewJSONUserRepository()
	require.NoError(t, err)

	user := &models.User{
		ID:             uuid.New(),
		Username:       "TestUser",
		CurrentProgram: uuid.Nil,
		Programs:       make(map[uuid.UUID]*models.UserProgram),
		WorkoutHistory: []models.Workout{},
		CreatedAt:      time.Now(),
	}

	userProgram := &models.UserProgram{
		ID:        uuid.Must(uuid.NewV7()),
		UserID:    user.ID,
		ProgramID: uuid.Must(uuid.Parse("550e8400-e29b-41d4-a716-446655440000")),
		StartingWeights: map[models.LiftName]float64{
			models.Squat:         135.0,
			models.Deadlift:      185.0,
			models.BenchPress:    125.0,
			models.OverheadPress: 95.0,
		},
		CurrentWeights: map[models.LiftName]float64{
			models.Squat:         135.0,
			models.Deadlift:      185.0,
			models.BenchPress:    125.0,
			models.OverheadPress: 95.0,
		},
		CurrentDay: 3, // Start at day 3
		StartedAt:  time.Now(),
	}

	user.Programs[userProgram.ID] = userProgram
	user.CurrentProgram = userProgram.ID

	err = repo.Create(user)
	require.NoError(t, err)

	err = repo.SetCurrent("TestUser")
	require.NoError(t, err)

	// Verify initial state
	assert.Equal(t, 3, userProgram.CurrentDay, "CurrentDay should initially be 3")

	// Mock AMRAP input for Day 3 exercises (OverheadPress, Squat)
	input := strings.NewReader("6\n8\n")

	var buf bytes.Buffer
	cmd := workoutLogCmd
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetIn(input)

	err = cmd.RunE(cmd, []string{})
	require.NoError(t, err)

	// Reload user to check updated state
	updatedUser, err := repo.Get("TestUser")
	require.NoError(t, err)

	updatedProgram := updatedUser.Programs[userProgram.ID]
	assert.Equal(t, 4, updatedProgram.CurrentDay, "CurrentDay should increment from 3 to 4")
}

func TestWorkoutLog_CurrentDayWrapsAroundAfterDay6(t *testing.T) {
	_ = setupTestEnv(t)

	repo, err := repository.NewJSONUserRepository()
	require.NoError(t, err)

	user := &models.User{
		ID:             uuid.New(),
		Username:       "TestUser",
		CurrentProgram: uuid.Nil,
		Programs:       make(map[uuid.UUID]*models.UserProgram),
		WorkoutHistory: []models.Workout{},
		CreatedAt:      time.Now(),
	}

	userProgram := &models.UserProgram{
		ID:        uuid.Must(uuid.NewV7()),
		UserID:    user.ID,
		ProgramID: uuid.Must(uuid.Parse("550e8400-e29b-41d4-a716-446655440000")),
		StartingWeights: map[models.LiftName]float64{
			models.Squat:         135.0,
			models.Deadlift:      185.0,
			models.BenchPress:    125.0,
			models.OverheadPress: 95.0,
		},
		CurrentWeights: map[models.LiftName]float64{
			models.Squat:         135.0,
			models.Deadlift:      185.0,
			models.BenchPress:    125.0,
			models.OverheadPress: 95.0,
		},
		CurrentDay: 6, // Start at day 6 (last day of cycle)
		StartedAt:  time.Now(),
	}

	user.Programs[userProgram.ID] = userProgram
	user.CurrentProgram = userProgram.ID

	err = repo.Create(user)
	require.NoError(t, err)

	err = repo.SetCurrent("TestUser")
	require.NoError(t, err)

	// Verify initial state
	assert.Equal(t, 6, userProgram.CurrentDay, "CurrentDay should initially be 6")

	// Mock AMRAP input for Day 6 exercises (BenchPress, Squat)
	input := strings.NewReader("9\n10\n")

	var buf bytes.Buffer
	cmd := workoutLogCmd
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetIn(input)

	err = cmd.RunE(cmd, []string{})
	require.NoError(t, err)

	// Reload user to check updated state
	updatedUser, err := repo.Get("TestUser")
	require.NoError(t, err)

	updatedProgram := updatedUser.Programs[userProgram.ID]
	assert.Equal(t, 1, updatedProgram.CurrentDay, "CurrentDay should wrap from 6 to 1")
}

func TestWorkoutLog_CompletionSummary(t *testing.T) {
	_ = setupTestEnv(t)

	repo, err := repository.NewJSONUserRepository()
	require.NoError(t, err)

	user := &models.User{
		ID:             uuid.New(),
		Username:       "TestUser",
		CurrentProgram: uuid.Nil,
		Programs:       make(map[uuid.UUID]*models.UserProgram),
		WorkoutHistory: []models.Workout{},
		CreatedAt:      time.Now(),
	}

	userProgram := &models.UserProgram{
		ID:        uuid.Must(uuid.NewV7()),
		UserID:    user.ID,
		ProgramID: uuid.Must(uuid.Parse("550e8400-e29b-41d4-a716-446655440000")),
		StartingWeights: map[models.LiftName]float64{
			models.Squat:         135.0,
			models.Deadlift:      185.0,
			models.BenchPress:    125.0,
			models.OverheadPress: 95.0,
		},
		CurrentWeights: map[models.LiftName]float64{
			models.Squat:         135.0,
			models.Deadlift:      185.0,
			models.BenchPress:    125.0,
			models.OverheadPress: 95.0,
		},
		CurrentDay: 5, // Day 5 -> Day 6 after logging
		StartedAt:  time.Now(),
	}

	user.Programs[userProgram.ID] = userProgram
	user.CurrentProgram = userProgram.ID

	err = repo.Create(user)
	require.NoError(t, err)

	err = repo.SetCurrent("TestUser")
	require.NoError(t, err)

	// Mock AMRAP input for Day 5 exercises (OverheadPress, Deadlift)
	input := strings.NewReader("7\n5\n")

	// Capture output to verify completion summary
	var buf bytes.Buffer
	cmd := workoutLogCmd
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetIn(input)

	err = cmd.RunE(cmd, []string{})
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Workout logged successfully!", "Should show success message")
	assert.Contains(t, output, "Next workout: Day 6", "Should show next workout day")
}

func TestWorkoutLog_DisplaysWorkoutLikeNextCommand(t *testing.T) {
	_ = setupTestEnv(t)

	repo, err := repository.NewJSONUserRepository()
	require.NoError(t, err)

	user := &models.User{
		ID:             uuid.New(),
		Username:       "TestUser",
		CurrentProgram: uuid.Nil,
		Programs:       make(map[uuid.UUID]*models.UserProgram),
		WorkoutHistory: []models.Workout{},
		CreatedAt:      time.Now(),
	}

	userProgram := &models.UserProgram{
		ID:        uuid.Must(uuid.NewV7()),
		UserID:    user.ID,
		ProgramID: uuid.Must(uuid.Parse("550e8400-e29b-41d4-a716-446655440000")),
		StartingWeights: map[models.LiftName]float64{
			models.Squat:         135.0,
			models.Deadlift:      185.0,
			models.BenchPress:    125.0,
			models.OverheadPress: 95.0,
		},
		CurrentWeights: map[models.LiftName]float64{
			models.Squat:         135.0,
			models.Deadlift:      185.0,
			models.BenchPress:    125.0,
			models.OverheadPress: 95.0,
		},
		CurrentDay: 1, // Day 1: OverheadPress, Squat
		StartedAt:  time.Now(),
	}

	user.Programs[userProgram.ID] = userProgram
	user.CurrentProgram = userProgram.ID

	err = repo.Create(user)
	require.NoError(t, err)

	err = repo.SetCurrent("TestUser")
	require.NoError(t, err)

	// Mock AMRAP input
	input := strings.NewReader("8\n7\n")

	// Capture output to verify workout display
	var buf bytes.Buffer
	cmd := workoutLogCmd
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetIn(input)

	err = cmd.RunE(cmd, []string{})
	require.NoError(t, err)

	// The command should display the workout details like the "next" command
	output := buf.String()
	assert.Contains(t, output, "Day 1 Workout:", "Should show day and workout header")
	assert.Contains(t, output, "Overhead Press:", "Should show OverheadPress exercise")
	assert.Contains(t, output, "Squat:", "Should show Squat exercise")
	assert.Contains(t, output, "Warmup:", "Should show warmup section")
	assert.Contains(t, output, "Working Sets:", "Should show working sets section")
	assert.Contains(t, output, "5+ reps @ 95 lbs (AMRAP)", "Should show OverheadPress AMRAP set")
	assert.Contains(t, output, "5+ reps @ 135 lbs (AMRAP)", "Should show Squat AMRAP set")
}

func TestBuildCompletedWorkout(t *testing.T) {
	// Create a template workout from calculator
	userProgram := &models.UserProgram{
		ID:        uuid.Must(uuid.NewV7()),
		UserID:    uuid.New(),
		ProgramID: uuid.Must(uuid.Parse("550e8400-e29b-41d4-a716-446655440000")),
		CurrentWeights: map[models.LiftName]float64{
			models.OverheadPress: 95.0,
			models.Squat:         135.0,
		},
		CurrentDay: 1,
	}

	user := &models.User{
		ID:             userProgram.UserID,
		Username:       "TestUser",
		CurrentProgram: userProgram.ID,
		Programs:       map[uuid.UUID]*models.UserProgram{userProgram.ID: userProgram},
	}

	// Create AMRAP reps map
	amrapReps := map[models.LiftName]int{
		models.OverheadPress: 8,
		models.Squat:         7,
	}

	// Get template workout from calculator (this should work since calculator exists)
	program := getGreyskullLP() // Helper function to get program
	templateWorkout, err := calculateNextWorkout(user, program)
	require.NoError(t, err)

	// Test buildCompletedWorkout function
	completedWorkout := buildCompletedWorkout(templateWorkout, amrapReps)

	// Verify basic structure
	assert.NotNil(t, completedWorkout, "Should return a completed workout")
	assert.Equal(t, templateWorkout.UserProgramID, completedWorkout.UserProgramID, "Should preserve UserProgramID")
	assert.Equal(t, templateWorkout.Day, completedWorkout.Day, "Should preserve Day")
	assert.Len(t, completedWorkout.Exercises, len(templateWorkout.Exercises), "Should have same number of exercises")

	// Verify EnteredAt is recent
	assert.WithinDuration(t, time.Now(), completedWorkout.EnteredAt, 5*time.Second, "EnteredAt should be recent")

	// Verify all sets are completed correctly
	for i, exercise := range completedWorkout.Exercises {
		templateExercise := templateWorkout.Exercises[i]
		assert.Equal(t, templateExercise.LiftName, exercise.LiftName, "Should preserve lift name")
		assert.Len(t, exercise.Sets, len(templateExercise.Sets), "Should have same number of sets")

		for j, set := range exercise.Sets {
			templateSet := templateExercise.Sets[j]

			// Verify all fields are preserved except ActualReps
			assert.Equal(t, templateSet.Weight, set.Weight, "Should preserve weight")
			assert.Equal(t, templateSet.TargetReps, set.TargetReps, "Should preserve target reps")
			assert.Equal(t, templateSet.Type, set.Type, "Should preserve set type")
			assert.Equal(t, templateSet.Order, set.Order, "Should preserve order")

			// Verify ActualReps is set correctly
			assert.True(t, set.IsComplete(), "All sets should be marked complete")

			if set.Type == models.AMRAPSet {
				expectedReps := amrapReps[exercise.LiftName]
				assert.Equal(t, expectedReps, set.ActualReps, "AMRAP sets should use provided reps")
			} else {
				assert.Equal(t, set.TargetReps, set.ActualReps, "Non-AMRAP sets should have ActualReps = TargetReps")
			}
		}
	}
}

func TestPromptInt(t *testing.T) {
	tests := []struct {
		name        string
		prompt      string
		input       string
		expected    int
		shouldError bool
	}{
		{
			name:        "valid input",
			prompt:      "Enter reps: ",
			input:       "8",
			expected:    8,
			shouldError: false,
		},
		{
			name:        "valid large number",
			prompt:      "Enter reps: ",
			input:       "15",
			expected:    15,
			shouldError: false,
		},
		{
			name:        "invalid input - zero",
			prompt:      "Enter reps: ",
			input:       "0",
			expected:    0,
			shouldError: true,
		},
		{
			name:        "invalid input - negative",
			prompt:      "Enter reps: ",
			input:       "-5",
			expected:    0,
			shouldError: true,
		},
		{
			name:        "invalid input - non-numeric",
			prompt:      "Enter reps: ",
			input:       "abc",
			expected:    0,
			shouldError: true,
		},
		{
			name:        "invalid input - decimal",
			prompt:      "Enter reps: ",
			input:       "5.5",
			expected:    0,
			shouldError: true,
		},
		{
			name:        "invalid input - empty",
			prompt:      "Enter reps: ",
			input:       "",
			expected:    0,
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a strings.Reader to simulate user input
			input := strings.NewReader(tt.input + "\n")

			result, err := promptIntWithReader(tt.prompt, input)

			if tt.shouldError {
				assert.Error(t, err, "promptInt should return error for invalid input: %s", tt.input)
			} else {
				assert.NoError(t, err, "promptInt should not return error for valid input: %s", tt.input)
				assert.Equal(t, tt.expected, result, "promptInt should return correct value for input: %s", tt.input)
			}
		})
	}
}

func TestWorkoutLog_AMRAPPromptText(t *testing.T) {
	_ = setupTestEnv(t)

	repo, err := repository.NewJSONUserRepository()
	require.NoError(t, err)

	user := &models.User{
		ID:             uuid.New(),
		Username:       "TestUser",
		CurrentProgram: uuid.Nil,
		Programs:       make(map[uuid.UUID]*models.UserProgram),
		WorkoutHistory: []models.Workout{},
		CreatedAt:      time.Now(),
	}

	userProgram := &models.UserProgram{
		ID:        uuid.Must(uuid.NewV7()),
		UserID:    user.ID,
		ProgramID: uuid.Must(uuid.Parse("550e8400-e29b-41d4-a716-446655440000")),
		StartingWeights: map[models.LiftName]float64{
			models.Squat:         135.0,
			models.Deadlift:      185.0,
			models.BenchPress:    125.0,
			models.OverheadPress: 95.0,
		},
		CurrentWeights: map[models.LiftName]float64{
			models.Squat:         135.0,
			models.Deadlift:      185.0,
			models.BenchPress:    125.0,
			models.OverheadPress: 95.0,
		},
		CurrentDay: 1, // Day 1: OverheadPress, Squat
		StartedAt:  time.Now(),
	}

	user.Programs[userProgram.ID] = userProgram
	user.CurrentProgram = userProgram.ID

	err = repo.Create(user)
	require.NoError(t, err)

	err = repo.SetCurrent("TestUser")
	require.NoError(t, err)

	// Mock AMRAP input
	input := strings.NewReader("8\n7\n")

	// Capture output to verify AMRAP prompts
	var buf bytes.Buffer
	cmd := workoutLogCmd
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetIn(input)

	err = cmd.RunE(cmd, []string{})
	require.NoError(t, err)

	output := buf.String()

	// The command should prompt for AMRAP reps with specific text format
	assert.Contains(t, output, "How many reps did you complete for Overhead Press AMRAP set (5+)?", "Should prompt for OverheadPress AMRAP")
	assert.Contains(t, output, "How many reps did you complete for Squat AMRAP set (5+)?", "Should prompt for Squat AMRAP")
}

func TestWorkoutLog_AutoCompleteNonAMRAPSets(t *testing.T) {
	_ = setupTestEnv(t)

	repo, err := repository.NewJSONUserRepository()
	require.NoError(t, err)

	user := &models.User{
		ID:             uuid.New(),
		Username:       "TestUser",
		CurrentProgram: uuid.Nil,
		Programs:       make(map[uuid.UUID]*models.UserProgram),
		WorkoutHistory: []models.Workout{},
		CreatedAt:      time.Now(),
	}

	userProgram := &models.UserProgram{
		ID:        uuid.Must(uuid.NewV7()),
		UserID:    user.ID,
		ProgramID: uuid.Must(uuid.Parse("550e8400-e29b-41d4-a716-446655440000")),
		StartingWeights: map[models.LiftName]float64{
			models.Squat:         135.0,
			models.Deadlift:      185.0,
			models.BenchPress:    125.0,
			models.OverheadPress: 95.0,
		},
		CurrentWeights: map[models.LiftName]float64{
			models.Squat:         135.0,
			models.Deadlift:      185.0,
			models.BenchPress:    125.0,
			models.OverheadPress: 95.0,
		},
		CurrentDay: 1,
		StartedAt:  time.Now(),
	}

	user.Programs[userProgram.ID] = userProgram
	user.CurrentProgram = userProgram.ID

	err = repo.Create(user)
	require.NoError(t, err)

	err = repo.SetCurrent("TestUser")
	require.NoError(t, err)

	// Mock AMRAP input only (should only prompt for AMRAP sets)
	input := strings.NewReader("8\n7\n")

	var buf bytes.Buffer
	cmd := workoutLogCmd
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetIn(input)

	err = cmd.RunE(cmd, []string{})
	require.NoError(t, err)

	// Reload user to check saved workout
	updatedUser, err := repo.Get("TestUser")
	require.NoError(t, err)

	require.Len(t, updatedUser.WorkoutHistory, 1)
	workout := updatedUser.WorkoutHistory[0]

	// Verify all warmup sets and working sets are auto-completed
	for _, exercise := range workout.Exercises {
		for _, set := range exercise.Sets {
			if set.Type == models.WarmupSet || set.Type == models.WorkingSet {
				assert.Equal(t, set.TargetReps, set.ActualReps, "Warmup and working sets should be auto-completed")
			}
		}
	}

	output := buf.String()

	// Should NOT prompt for warmup or working sets, only AMRAP
	warmupPromptCount := strings.Count(output, "How many reps did you complete for")
	assert.Equal(t, 2, warmupPromptCount, "Should only prompt for 2 AMRAP sets, not warmup or working sets")
}

func TestWorkoutLog_WeightProgressionNormalIncrease(t *testing.T) {
	// Test normal progression (AMRAP reps 5-9)
	_ = setupTestEnv(t)

	repo, err := repository.NewJSONUserRepository()
	require.NoError(t, err)

	user := &models.User{
		ID:             uuid.New(),
		Username:       "TestUser",
		CurrentProgram: uuid.Nil,
		Programs:       make(map[uuid.UUID]*models.UserProgram),
		WorkoutHistory: []models.Workout{},
		CreatedAt:      time.Now(),
	}

	initialWeights := map[models.LiftName]float64{
		models.Squat:         135.0,
		models.Deadlift:      185.0,
		models.BenchPress:    125.0,
		models.OverheadPress: 95.0,
	}

	userProgram := &models.UserProgram{
		ID:              uuid.Must(uuid.NewV7()),
		UserID:          user.ID,
		ProgramID:       uuid.Must(uuid.Parse("550e8400-e29b-41d4-a716-446655440000")),
		StartingWeights: initialWeights,
		CurrentWeights:  make(map[models.LiftName]float64),
		CurrentDay:      1, // Day 1: OverheadPress, Squat
		StartedAt:       time.Now(),
	}

	// Copy initial weights
	for k, v := range initialWeights {
		userProgram.CurrentWeights[k] = v
	}

	user.Programs[userProgram.ID] = userProgram
	user.CurrentProgram = userProgram.ID

	err = repo.Create(user)
	require.NoError(t, err)

	err = repo.SetCurrent("TestUser")
	require.NoError(t, err)

	// Mock AMRAP input: 6 reps for OverheadPress, 7 reps for Squat (normal progression)
	input := strings.NewReader("6\n7\n")

	var buf bytes.Buffer
	cmd := workoutLogCmd
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetIn(input)

	err = cmd.RunE(cmd, []string{})
	require.NoError(t, err)

	// Reload user to check progression
	updatedUser, err := repo.Get("TestUser")
	require.NoError(t, err)

	updatedProgram := updatedUser.Programs[userProgram.ID]

	// Verify normal progression applied
	assert.Equal(t, 97.5, updatedProgram.CurrentWeights[models.OverheadPress], "OverheadPress should increase by 2.5 (95 + 2.5)")
	assert.Equal(t, 140.0, updatedProgram.CurrentWeights[models.Squat], "Squat should increase by 5 (135 + 5)")
	// Other lifts should remain unchanged
	assert.Equal(t, 185.0, updatedProgram.CurrentWeights[models.Deadlift], "Deadlift should remain unchanged")
	assert.Equal(t, 125.0, updatedProgram.CurrentWeights[models.BenchPress], "BenchPress should remain unchanged")

	// Verify weight changes are displayed
	output := buf.String()
	assert.Contains(t, output, "Weight Updates:", "Should show weight updates section")
	assert.Contains(t, output, "Overhead Press: 95 → 97.5 lbs (+2.5)", "Should show OverheadPress progression")
	assert.Contains(t, output, "Squat: 135 → 140 lbs (+5.0)", "Should show Squat progression")
}

func TestWorkoutLog_WeightProgressionDoubleIncrease(t *testing.T) {
	// Test double progression (AMRAP reps >= 10)
	_ = setupTestEnv(t)

	repo, err := repository.NewJSONUserRepository()
	require.NoError(t, err)

	user := &models.User{
		ID:             uuid.New(),
		Username:       "TestUser",
		CurrentProgram: uuid.Nil,
		Programs:       make(map[uuid.UUID]*models.UserProgram),
		WorkoutHistory: []models.Workout{},
		CreatedAt:      time.Now(),
	}

	initialWeights := map[models.LiftName]float64{
		models.Squat:         135.0,
		models.Deadlift:      185.0,
		models.BenchPress:    125.0,
		models.OverheadPress: 95.0,
	}

	userProgram := &models.UserProgram{
		ID:              uuid.Must(uuid.NewV7()),
		UserID:          user.ID,
		ProgramID:       uuid.Must(uuid.Parse("550e8400-e29b-41d4-a716-446655440000")),
		StartingWeights: initialWeights,
		CurrentWeights:  make(map[models.LiftName]float64),
		CurrentDay:      1, // Day 1: OverheadPress, Squat
		StartedAt:       time.Now(),
	}

	// Copy initial weights
	for k, v := range initialWeights {
		userProgram.CurrentWeights[k] = v
	}

	user.Programs[userProgram.ID] = userProgram
	user.CurrentProgram = userProgram.ID

	err = repo.Create(user)
	require.NoError(t, err)

	err = repo.SetCurrent("TestUser")
	require.NoError(t, err)

	// Mock AMRAP input: 12 reps for OverheadPress, 15 reps for Squat (double progression)
	input := strings.NewReader("12\n15\n")

	var buf bytes.Buffer
	cmd := workoutLogCmd
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetIn(input)

	err = cmd.RunE(cmd, []string{})
	require.NoError(t, err)

	// Reload user to check progression
	updatedUser, err := repo.Get("TestUser")
	require.NoError(t, err)

	updatedProgram := updatedUser.Programs[userProgram.ID]

	// Verify double progression applied (2x base increment)
	assert.Equal(t, 100.0, updatedProgram.CurrentWeights[models.OverheadPress], "OverheadPress should increase by 5.0 (95 + 2.5*2)")
	assert.Equal(t, 145.0, updatedProgram.CurrentWeights[models.Squat], "Squat should increase by 10.0 (135 + 5*2)")

	// Verify weight changes are displayed
	output := buf.String()
	assert.Contains(t, output, "Overhead Press: 95 → 100 lbs (+5.0)", "Should show OverheadPress double progression")
	assert.Contains(t, output, "Squat: 135 → 145 lbs (+10.0)", "Should show Squat double progression")
}

func TestWorkoutLog_WeightProgressionDeload(t *testing.T) {
	// Test deload (AMRAP reps < 5)
	_ = setupTestEnv(t)

	repo, err := repository.NewJSONUserRepository()
	require.NoError(t, err)

	user := &models.User{
		ID:             uuid.New(),
		Username:       "TestUser",
		CurrentProgram: uuid.Nil,
		Programs:       make(map[uuid.UUID]*models.UserProgram),
		WorkoutHistory: []models.Workout{},
		CreatedAt:      time.Now(),
	}

	initialWeights := map[models.LiftName]float64{
		models.Squat:         135.0,
		models.Deadlift:      185.0,
		models.BenchPress:    125.0,
		models.OverheadPress: 95.0,
	}

	userProgram := &models.UserProgram{
		ID:              uuid.Must(uuid.NewV7()),
		UserID:          user.ID,
		ProgramID:       uuid.Must(uuid.Parse("550e8400-e29b-41d4-a716-446655440000")),
		StartingWeights: initialWeights,
		CurrentWeights:  make(map[models.LiftName]float64),
		CurrentDay:      1, // Day 1: OverheadPress, Squat
		StartedAt:       time.Now(),
	}

	// Copy initial weights
	for k, v := range initialWeights {
		userProgram.CurrentWeights[k] = v
	}

	user.Programs[userProgram.ID] = userProgram
	user.CurrentProgram = userProgram.ID

	err = repo.Create(user)
	require.NoError(t, err)

	err = repo.SetCurrent("TestUser")
	require.NoError(t, err)

	// Mock AMRAP input: 3 reps for OverheadPress, 4 reps for Squat (deload)
	input := strings.NewReader("3\n4\n")

	var buf bytes.Buffer
	cmd := workoutLogCmd
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetIn(input)

	err = cmd.RunE(cmd, []string{})
	require.NoError(t, err)

	// Reload user to check progression
	updatedUser, err := repo.Get("TestUser")
	require.NoError(t, err)

	updatedProgram := updatedUser.Programs[userProgram.ID]

	// Verify deload applied (90% of current weight)
	assert.Equal(t, 85.0, updatedProgram.CurrentWeights[models.OverheadPress], "OverheadPress should deload to 85.5 → 85.0 (95 * 0.9)")
	assert.Equal(t, 120.0, updatedProgram.CurrentWeights[models.Squat], "Squat should deload to 121.5 → 120.0 (135 * 0.9)")

	// Verify weight changes are displayed with negative values
	output := buf.String()
	assert.Contains(t, output, "Overhead Press: 95 → 85 lbs (-10.0)", "Should show OverheadPress deload")
	assert.Contains(t, output, "Squat: 135 → 120 lbs (-15.0)", "Should show Squat deload")
}

func TestWorkoutLog_SetsHaveUUIDsAndCorrectData(t *testing.T) {
	_ = setupTestEnv(t)

	repo, err := repository.NewJSONUserRepository()
	require.NoError(t, err)

	user := &models.User{
		ID:             uuid.New(),
		Username:       "TestUser",
		CurrentProgram: uuid.Nil,
		Programs:       make(map[uuid.UUID]*models.UserProgram),
		WorkoutHistory: []models.Workout{},
		CreatedAt:      time.Now(),
	}

	userProgram := &models.UserProgram{
		ID:        uuid.Must(uuid.NewV7()),
		UserID:    user.ID,
		ProgramID: uuid.Must(uuid.Parse("550e8400-e29b-41d4-a716-446655440000")),
		StartingWeights: map[models.LiftName]float64{
			models.Squat:         135.0,
			models.Deadlift:      185.0,
			models.BenchPress:    125.0,
			models.OverheadPress: 95.0,
		},
		CurrentWeights: map[models.LiftName]float64{
			models.Squat:         135.0,
			models.Deadlift:      185.0,
			models.BenchPress:    125.0,
			models.OverheadPress: 95.0,
		},
		CurrentDay: 1,
		StartedAt:  time.Now(),
	}

	user.Programs[userProgram.ID] = userProgram
	user.CurrentProgram = userProgram.ID

	err = repo.Create(user)
	require.NoError(t, err)

	err = repo.SetCurrent("TestUser")
	require.NoError(t, err)

	// Mock AMRAP input
	input := strings.NewReader("8\n7\n")

	var buf bytes.Buffer
	cmd := workoutLogCmd
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetIn(input)

	err = cmd.RunE(cmd, []string{})
	require.NoError(t, err)

	// Reload user to check saved workout
	updatedUser, err := repo.Get("TestUser")
	require.NoError(t, err)

	require.Len(t, updatedUser.WorkoutHistory, 1)
	savedWorkout := updatedUser.WorkoutHistory[0]

	// Verify workout has valid UUID and correct data
	assert.NotEqual(t, uuid.Nil, savedWorkout.ID, "Workout should have valid UUID")
	assert.Equal(t, userProgram.ID, savedWorkout.UserProgramID, "UserProgramID should match")
	assert.Equal(t, 1, savedWorkout.Day, "Day should match current day")
	assert.WithinDuration(t, time.Now(), savedWorkout.EnteredAt, 5*time.Second, "EnteredAt should be recent")

	// Verify each exercise and set has valid data
	for _, exercise := range savedWorkout.Exercises {
		assert.NotEqual(t, uuid.Nil, exercise.ID, "Exercise should have valid UUID")
		assert.NotEmpty(t, exercise.LiftName, "Exercise should have lift name")
		assert.NotEmpty(t, exercise.Sets, "Exercise should have sets")

		for _, set := range exercise.Sets {
			assert.NotEqual(t, uuid.Nil, set.ID, "Set should have valid UUID")
			assert.Greater(t, set.Weight, 0.0, "Set should have positive weight")
			assert.Greater(t, set.TargetReps, 0, "Set should have positive target reps")
			assert.Greater(t, set.ActualReps, 0, "Set should have positive actual reps")
			assert.NotEmpty(t, set.Type, "Set should have type")
			assert.Greater(t, set.Order, 0, "Set should have positive order")
		}
	}
}

// Helper functions that need to be implemented

// Use the implementation from workout_log.go directly - no wrapper needed

// buildCompletedWorkout is implemented in workout_log.go - no need to redefine here

func getGreyskullLP() *models.Program {
	// Helper to get the Greyskull LP program
	program, _ := program.GetByID("550e8400-e29b-41d4-a716-446655440000")
	return program
}

func calculateNextWorkout(user *models.User, program *models.Program) (*models.Workout, error) {
	// Call the actual calculator
	return workout.CalculateNextWorkout(user, program)
}

// promptInt is implemented in workout_log.go - no need to redefine here

// getByID is available via program.GetByID - no need to redefine here

