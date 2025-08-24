package cmd

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mikowitz/greyskull/models"
	"github.com/mikowitz/greyskull/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkoutNext_NoCurrentUser(t *testing.T) {
	_ = setupTestEnv(t)

	cmd := workoutNextCmd
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.RunE(cmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no current user set")
}

func TestWorkoutNext_NoActiveProgram(t *testing.T) {
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

	cmd := workoutNextCmd
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err = cmd.RunE(cmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no active program")
}

func TestWorkoutNext_ValidProgramAndWeights(t *testing.T) {
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

	// Capture output
	var buf bytes.Buffer
	cmd := workoutNextCmd
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.RunE(cmd, []string{})
	assert.NoError(t, err)

	output := buf.String()

	// Verify day is displayed
	assert.Contains(t, output, "Day 1 Workout:")

	// Verify both exercises are displayed
	assert.Contains(t, output, "Overhead Press:")
	assert.Contains(t, output, "Squat:")

	// Verify warmup sets are displayed (both exercises have weights >= 85 lbs)
	assert.Contains(t, output, "Warmup:")
	assert.Contains(t, output, "5 reps @ 45 lbs")    // Empty bar
	assert.Contains(t, output, "4 reps @ 50 lbs")    // 55% of 95 = 52.25 → rounds down to 50.0
	assert.Contains(t, output, "3 reps @ 65 lbs")    // 70% of 95 = 66.5 → rounds down to 65.0
	assert.Contains(t, output, "2 reps @ 80 lbs")    // 85% of 95 = 80.75 → rounds down to 80.0

	// Verify working sets are displayed
	assert.Contains(t, output, "Working Sets:")
	assert.Contains(t, output, "Set 1: 5 reps @ 95 lbs")
	assert.Contains(t, output, "Set 2: 5 reps @ 95 lbs")
	assert.Contains(t, output, "Set 3: 5+ reps @ 95 lbs (AMRAP)")

	// Verify squat working sets
	assert.Contains(t, output, "Set 1: 5 reps @ 135 lbs")
	assert.Contains(t, output, "Set 2: 5 reps @ 135 lbs")
	assert.Contains(t, output, "Set 3: 5+ reps @ 135 lbs (AMRAP)")
}

func TestWorkoutNext_DisplayForDifferentDays(t *testing.T) {
	tests := []struct {
		name        string
		day         int
		expectedDay int
		exercise1   string
		exercise2   string
	}{
		{
			name:        "Day 1",
			day:         1,
			expectedDay: 1,
			exercise1:   "Overhead Press:",
			exercise2:   "Squat:",
		},
		{
			name:        "Day 2",
			day:         2,
			expectedDay: 2,
			exercise1:   "Bench Press:",
			exercise2:   "Deadlift:",
		},
		{
			name:        "Day 3",
			day:         3,
			expectedDay: 3,
			exercise1:   "Overhead Press:",
			exercise2:   "Squat:",
		},
		{
			name:        "Day 4",
			day:         4,
			expectedDay: 4,
			exercise1:   "Bench Press:",
			exercise2:   "Squat:",
		},
		{
			name:        "Day 5",
			day:         5,
			expectedDay: 5,
			exercise1:   "Overhead Press:",
			exercise2:   "Deadlift:",
		},
		{
			name:        "Day 6",
			day:         6,
			expectedDay: 6,
			exercise1:   "Bench Press:",
			exercise2:   "Squat:",
		},
		{
			name:        "Day 7 wraps to Day 1",
			day:         7,
			expectedDay: 1,
			exercise1:   "Overhead Press:",
			exercise2:   "Squat:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
				CurrentDay: tt.day,
				StartedAt:  time.Now(),
			}

			user.Programs[userProgram.ID] = userProgram
			user.CurrentProgram = userProgram.ID

			err = repo.Create(user)
			require.NoError(t, err)

			err = repo.SetCurrent("TestUser")
			require.NoError(t, err)

			// Capture output
			var buf bytes.Buffer
			cmd := workoutNextCmd
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			err = cmd.RunE(cmd, []string{})
			assert.NoError(t, err)

			output := buf.String()

			// Verify correct day is displayed
			assert.Contains(t, output, fmt.Sprintf("Day %d Workout:", tt.expectedDay))

			// Verify correct exercises are displayed
			assert.Contains(t, output, tt.exercise1)
			assert.Contains(t, output, tt.exercise2)
		})
	}
}

func TestWorkoutNext_WarmupDisplayVsNoWarmup(t *testing.T) {
	tests := []struct {
		name           string
		weights        map[models.LiftName]float64
		shouldHaveOPWarmup   bool
		shouldHaveSquatWarmup bool
	}{
		{
			name: "High weights - both should have warmups",
			weights: map[models.LiftName]float64{
				models.OverheadPress: 95.0,  // >= 85, should have warmup
				models.Squat:         135.0, // >= 85, should have warmup
			},
			shouldHaveOPWarmup:   true,
			shouldHaveSquatWarmup: true,
		},
		{
			name: "Low weights - neither should have warmups",
			weights: map[models.LiftName]float64{
				models.OverheadPress: 75.0, // < 85, no warmup
				models.Squat:         80.0, // < 85, no warmup
			},
			shouldHaveOPWarmup:   false,
			shouldHaveSquatWarmup: false,
		},
		{
			name: "Mixed weights - only squat should have warmup",
			weights: map[models.LiftName]float64{
				models.OverheadPress: 70.0,  // < 85, no warmup
				models.Squat:         100.0, // >= 85, should have warmup
			},
			shouldHaveOPWarmup:   false,
			shouldHaveSquatWarmup: true,
		},
		{
			name: "Edge case - exactly 85 lbs",
			weights: map[models.LiftName]float64{
				models.OverheadPress: 85.0, // == 85, no warmup (< 85 is the rule)
				models.Squat:         85.0, // == 85, no warmup 
			},
			shouldHaveOPWarmup:   false,
			shouldHaveSquatWarmup: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

			// Set all weights, using test values for OP and Squat
			allWeights := map[models.LiftName]float64{
				models.Squat:         tt.weights[models.Squat],
				models.Deadlift:      185.0,
				models.BenchPress:    125.0,
				models.OverheadPress: tt.weights[models.OverheadPress],
			}

			userProgram := &models.UserProgram{
				ID:              uuid.Must(uuid.NewV7()),
				UserID:          user.ID,
				ProgramID:       uuid.Must(uuid.Parse("550e8400-e29b-41d4-a716-446655440000")),
				StartingWeights: allWeights,
				CurrentWeights:  allWeights,
				CurrentDay:      1, // Day 1: OverheadPress, Squat
				StartedAt:       time.Now(),
			}

			user.Programs[userProgram.ID] = userProgram
			user.CurrentProgram = userProgram.ID

			err = repo.Create(user)
			require.NoError(t, err)

			err = repo.SetCurrent("TestUser")
			require.NoError(t, err)

			// Capture output
			var buf bytes.Buffer
			cmd := workoutNextCmd
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			err = cmd.RunE(cmd, []string{})
			assert.NoError(t, err)

			output := buf.String()

			// Check warmup display based on expectations
			if tt.shouldHaveOPWarmup {
				// Should find warmup section before Overhead Press working sets
				opIndex := strings.Index(output, "Overhead Press:")
				workingIndex := strings.Index(output[opIndex:], "Working Sets:")
				warmupIndex := strings.Index(output[opIndex:opIndex+workingIndex], "Warmup:")
				assert.True(t, warmupIndex >= 0, "Overhead Press should have warmup section")
			} else {
				// Should NOT find warmup section for Overhead Press
				opIndex := strings.Index(output, "Overhead Press:")
				squatIndex := strings.Index(output[opIndex:], "Squat:")
				warmupIndex := strings.Index(output[opIndex:opIndex+squatIndex], "Warmup:")
				assert.True(t, warmupIndex < 0, "Overhead Press should NOT have warmup section")
			}

			if tt.shouldHaveSquatWarmup {
				// Should find warmup section before Squat working sets
				squatIndex := strings.Index(output, "Squat:")
				workingIndex := strings.Index(output[squatIndex:], "Working Sets:")
				warmupIndex := strings.Index(output[squatIndex:squatIndex+workingIndex], "Warmup:")
				assert.True(t, warmupIndex >= 0, "Squat should have warmup section")
			} else {
				// Should NOT find warmup section for Squat (after Squat title, before end of output)
				squatIndex := strings.Index(output, "Squat:")
				warmupIndex := strings.Index(output[squatIndex:], "Warmup:")
				assert.True(t, warmupIndex < 0, "Squat should NOT have warmup section")
			}
		})
	}
}

func TestWorkoutNext_VerifyAMRAPSetsMarked(t *testing.T) {
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

	// Capture output
	var buf bytes.Buffer
	cmd := workoutNextCmd
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.RunE(cmd, []string{})
	assert.NoError(t, err)

	output := buf.String()

	// Verify AMRAP sets are clearly marked
	assert.Contains(t, output, "Set 3: 5+ reps @ 95 lbs (AMRAP)")   // Overhead Press AMRAP
	assert.Contains(t, output, "Set 3: 5+ reps @ 135 lbs (AMRAP)")  // Squat AMRAP

	// Verify non-AMRAP sets are NOT marked as AMRAP
	assert.Contains(t, output, "Set 1: 5 reps @ 95 lbs")    // Not marked as AMRAP
	assert.Contains(t, output, "Set 2: 5 reps @ 95 lbs")    // Not marked as AMRAP
	assert.Contains(t, output, "Set 1: 5 reps @ 135 lbs")   // Not marked as AMRAP
	assert.Contains(t, output, "Set 2: 5 reps @ 135 lbs")   // Not marked as AMRAP

	// Count AMRAP occurrences - should be exactly 2 (one per exercise)
	amrapCount := strings.Count(output, "(AMRAP)")
	assert.Equal(t, 2, amrapCount, "Should have exactly 2 AMRAP sets marked")
}

