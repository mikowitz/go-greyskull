package workout

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mikowitz/greyskull/models"
	"github.com/mikowitz/greyskull/program"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoundDown2_5(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected float64
	}{
		{
			name:     "round down 42.7 to 42.5",
			input:    42.7,
			expected: 42.5,
		},
		{
			name:     "keep 45.0 as 45.0",
			input:    45.0,
			expected: 45.0,
		},
		{
			name:     "round down 47.3 to 45.0",
			input:    47.3,
			expected: 45.0,
		},
		{
			name:     "round down 49.9 to 47.5",
			input:    49.9,
			expected: 47.5,
		},
		{
			name:     "keep exact multiple 50.0",
			input:    50.0,
			expected: 50.0,
		},
		{
			name:     "round down 52.4 to 50.0",
			input:    52.4,
			expected: 50.0,
		},
		{
			name:     "keep exact half 52.5",
			input:    52.5,
			expected: 52.5,
		},
		{
			name:     "round down 52.6 to 52.5",
			input:    52.6,
			expected: 52.5,
		},
		{
			name:     "handle zero",
			input:    0.0,
			expected: 0.0,
		},
		{
			name:     "round down small number 1.3 to 0.0",
			input:    1.3,
			expected: 0.0,
		},
		{
			name:     "round down 2.6 to 2.5",
			input:    2.6,
			expected: 2.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RoundDown2_5(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCalculateWarmupSets(t *testing.T) {
	// Create warmup templates similar to Greyskull LP
	warmupTemplates := []models.SetTemplate{
		{Reps: 5, WeightPercentage: 0.0, Type: models.WarmupSet},  // Empty bar
		{Reps: 4, WeightPercentage: 0.55, Type: models.WarmupSet}, // 55%
		{Reps: 3, WeightPercentage: 0.70, Type: models.WarmupSet}, // 70%
		{Reps: 2, WeightPercentage: 0.85, Type: models.WarmupSet}, // 85%
	}

	t.Run("skip warmup for weight less than 85 lbs", func(t *testing.T) {
		result := CalculateWarmupSets(80.0, warmupTemplates)
		assert.Empty(t, result)
	})

	t.Run("skip warmup for exactly 85 lbs", func(t *testing.T) {
		result := CalculateWarmupSets(85.0, warmupTemplates)
		assert.Empty(t, result)
	})

	t.Run("calculate warmup for 100 lbs working weight", func(t *testing.T) {
		result := CalculateWarmupSets(100.0, warmupTemplates)

		require.Len(t, result, 4)

		// Check empty bar (45 lbs)
		assert.Equal(t, 45.0, result[0].Weight)
		assert.Equal(t, 5, result[0].TargetReps)
		assert.Equal(t, models.WarmupSet, result[0].Type)
		assert.Equal(t, 1, result[0].Order)
		assert.NotEqual(t, uuid.Nil, result[0].ID)

		// Check 55% (55 lbs, rounded down to 55.0)
		assert.Equal(t, 55.0, result[1].Weight)
		assert.Equal(t, 4, result[1].TargetReps)
		assert.Equal(t, models.WarmupSet, result[1].Type)
		assert.Equal(t, 2, result[1].Order)

		// Check 70% (70 lbs)
		assert.Equal(t, 70.0, result[2].Weight)
		assert.Equal(t, 3, result[2].TargetReps)
		assert.Equal(t, models.WarmupSet, result[2].Type)
		assert.Equal(t, 3, result[2].Order)

		// Check 85% (85 lbs)
		assert.Equal(t, 85.0, result[3].Weight)
		assert.Equal(t, 2, result[3].TargetReps)
		assert.Equal(t, models.WarmupSet, result[3].Type)
		assert.Equal(t, 4, result[3].Order)
	})

	t.Run("calculate warmup with rounding for 97.5 lbs working weight", func(t *testing.T) {
		result := CalculateWarmupSets(97.5, warmupTemplates)

		require.Len(t, result, 4)

		// Check empty bar (45 lbs)
		assert.Equal(t, 45.0, result[0].Weight)

		// Check 55% (53.625 rounded down to 52.5)
		assert.Equal(t, 52.5, result[1].Weight)

		// Check 70% (68.25 rounded down to 67.5)
		assert.Equal(t, 67.5, result[2].Weight)

		// Check 85% (82.875 rounded down to 82.5)
		assert.Equal(t, 82.5, result[3].Weight)
	})

	t.Run("empty templates returns empty slice", func(t *testing.T) {
		result := CalculateWarmupSets(100.0, []models.SetTemplate{})
		assert.Empty(t, result)
	})
}

func TestCalculateWorkingSets(t *testing.T) {
	// Create working set templates similar to Greyskull LP
	workingTemplates := []models.SetTemplate{
		{Reps: 5, WeightPercentage: 1.0, Type: models.WorkingSet},
		{Reps: 5, WeightPercentage: 1.0, Type: models.WorkingSet},
		{Reps: 5, WeightPercentage: 1.0, Type: models.AMRAPSet},
	}

	t.Run("calculate working sets for 135 lbs", func(t *testing.T) {
		result := CalculateWorkingSets(135.0, workingTemplates)

		require.Len(t, result, 3)

		// Check first working set
		assert.Equal(t, 135.0, result[0].Weight)
		assert.Equal(t, 5, result[0].TargetReps)
		assert.Equal(t, models.WorkingSet, result[0].Type)
		assert.Equal(t, 1, result[0].Order)
		assert.NotEqual(t, uuid.Nil, result[0].ID)

		// Check second working set
		assert.Equal(t, 135.0, result[1].Weight)
		assert.Equal(t, 5, result[1].TargetReps)
		assert.Equal(t, models.WorkingSet, result[1].Type)
		assert.Equal(t, 2, result[1].Order)

		// Check AMRAP set
		assert.Equal(t, 135.0, result[2].Weight)
		assert.Equal(t, 5, result[2].TargetReps)
		assert.Equal(t, models.AMRAPSet, result[2].Type)
		assert.Equal(t, 3, result[2].Order)
	})

	t.Run("calculate working sets with rounding for 42.7 lbs", func(t *testing.T) {
		result := CalculateWorkingSets(42.7, workingTemplates)

		require.Len(t, result, 3)

		// All sets should be rounded down to 42.5
		for i, set := range result {
			assert.Equal(t, 42.5, set.Weight)
			assert.Equal(t, 5, set.TargetReps)
			assert.Equal(t, i+1, set.Order)
		}

		// Check types
		assert.Equal(t, models.WorkingSet, result[0].Type)
		assert.Equal(t, models.WorkingSet, result[1].Type)
		assert.Equal(t, models.AMRAPSet, result[2].Type)
	})

	t.Run("handle weight less than 45 lbs", func(t *testing.T) {
		result := CalculateWorkingSets(30.0, workingTemplates)

		require.Len(t, result, 3)

		// All sets should use 30.0 (not restricted to 45 lbs minimum)
		for _, set := range result {
			assert.Equal(t, 30.0, set.Weight)
		}
	})

	t.Run("empty templates returns empty slice", func(t *testing.T) {
		result := CalculateWorkingSets(135.0, []models.SetTemplate{})
		assert.Empty(t, result)
	})
}

func TestGetWorkoutDay(t *testing.T) {
	tests := []struct {
		name       string
		currentDay int
		totalDays  int
		expected   int
	}{
		{
			name:       "day 1 of 6",
			currentDay: 1,
			totalDays:  6,
			expected:   1,
		},
		{
			name:       "day 6 of 6",
			currentDay: 6,
			totalDays:  6,
			expected:   6,
		},
		{
			name:       "day 7 wraps to day 1",
			currentDay: 7,
			totalDays:  6,
			expected:   1,
		},
		{
			name:       "day 8 wraps to day 2",
			currentDay: 8,
			totalDays:  6,
			expected:   2,
		},
		{
			name:       "day 12 wraps to day 6",
			currentDay: 12,
			totalDays:  6,
			expected:   6,
		},
		{
			name:       "day 13 wraps to day 1",
			currentDay: 13,
			totalDays:  6,
			expected:   1,
		},
		{
			name:       "3-day program cycling",
			currentDay: 4,
			totalDays:  3,
			expected:   1,
		},
		{
			name:       "large day number wrapping",
			currentDay: 100,
			totalDays:  6,
			expected:   4, // 100 % 6 = 4, but since we use 1-based: ((100-1) % 6) + 1 = 4
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetWorkoutDay(tt.currentDay, tt.totalDays)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCalculateNextWorkout(t *testing.T) {
	// Get the Greyskull LP program for testing
	greyskullProgram := program.GreyskullLP

	t.Run("calculate day 1 workout", func(t *testing.T) {
		user := createTestUser(1, map[models.LiftName]float64{
			models.OverheadPress: 95.0,
			models.Squat:         135.0,
			models.BenchPress:    125.0,
			models.Deadlift:      185.0,
		})

		result, err := CalculateNextWorkout(user, greyskullProgram)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Check workout structure
		assert.Equal(t, 1, result.Day)
		assert.NotEqual(t, uuid.Nil, result.ID)
		assert.Equal(t, user.Programs[user.CurrentProgram].ID, result.UserProgramID)
		assert.WithinDuration(t, time.Now(), result.EnteredAt, time.Second)

		// Day 1 should have OverheadPress and Squat
		require.Len(t, result.Exercises, 2)

		// Check OverheadPress (95 lbs should have warmup)
		ohpLift := result.Exercises[0]
		assert.Equal(t, models.OverheadPress, ohpLift.LiftName)
		assert.NotEqual(t, uuid.Nil, ohpLift.ID)

		// Should have 4 warmup sets + 3 working sets = 7 total
		require.Len(t, ohpLift.Sets, 7)

		// Check warmup sets
		assert.Equal(t, 45.0, ohpLift.Sets[0].Weight) // Empty bar
		assert.Equal(t, 50.0, ohpLift.Sets[1].Weight) // 55% of 95 = 52.25 → 50.0 (rounded DOWN)
		assert.Equal(t, 65.0, ohpLift.Sets[2].Weight) // 70% of 95 = 66.5 → 65.0 (rounded DOWN)
		assert.Equal(t, 80.0, ohpLift.Sets[3].Weight) // 85% of 95 = 80.75 → 80.0

		// Check working sets
		assert.Equal(t, 95.0, ohpLift.Sets[4].Weight)
		assert.Equal(t, models.WorkingSet, ohpLift.Sets[4].Type)
		assert.Equal(t, 95.0, ohpLift.Sets[5].Weight)
		assert.Equal(t, models.WorkingSet, ohpLift.Sets[5].Type)
		assert.Equal(t, 95.0, ohpLift.Sets[6].Weight)
		assert.Equal(t, models.AMRAPSet, ohpLift.Sets[6].Type)

		// Check Squat (135 lbs should have warmup)
		squatLift := result.Exercises[1]
		assert.Equal(t, models.Squat, squatLift.LiftName)

		// Should have 4 warmup sets + 3 working sets = 7 total
		require.Len(t, squatLift.Sets, 7)

		// Check squat working weight
		assert.Equal(t, 135.0, squatLift.Sets[4].Weight)
		assert.Equal(t, 135.0, squatLift.Sets[5].Weight)
		assert.Equal(t, 135.0, squatLift.Sets[6].Weight)
	})

	t.Run("calculate day 2 workout", func(t *testing.T) {
		user := createTestUser(2, map[models.LiftName]float64{
			models.OverheadPress: 95.0,
			models.Squat:         135.0,
			models.BenchPress:    125.0,
			models.Deadlift:      185.0,
		})

		result, err := CalculateNextWorkout(user, greyskullProgram)
		require.NoError(t, err)

		assert.Equal(t, 2, result.Day)

		// Day 2 should have BenchPress and Deadlift
		require.Len(t, result.Exercises, 2)
		assert.Equal(t, models.BenchPress, result.Exercises[0].LiftName)
		assert.Equal(t, models.Deadlift, result.Exercises[1].LiftName)

		// Check bench press working weight
		benchSets := result.Exercises[0].Sets
		workingSets := benchSets[len(benchSets)-3:] // Last 3 sets are working sets
		for _, set := range workingSets {
			assert.Equal(t, 125.0, set.Weight)
		}

		// Check deadlift working weight
		deadliftSets := result.Exercises[1].Sets
		workingSets = deadliftSets[len(deadliftSets)-3:] // Last 3 sets are working sets
		for _, set := range workingSets {
			assert.Equal(t, 185.0, set.Weight)
		}
	})

	t.Run("calculate workout with day cycling", func(t *testing.T) {
		user := createTestUser(7, map[models.LiftName]float64{
			models.OverheadPress: 95.0,
			models.Squat:         135.0,
			models.BenchPress:    125.0,
			models.Deadlift:      185.0,
		})

		result, err := CalculateNextWorkout(user, greyskullProgram)
		require.NoError(t, err)

		// Day 7 should wrap to day 1
		assert.Equal(t, 1, result.Day)

		// Should have OverheadPress and Squat like day 1
		require.Len(t, result.Exercises, 2)
		assert.Equal(t, models.OverheadPress, result.Exercises[0].LiftName)
		assert.Equal(t, models.Squat, result.Exercises[1].LiftName)
	})

	t.Run("calculate workout with low weights (no warmup)", func(t *testing.T) {
		user := createTestUser(1, map[models.LiftName]float64{
			models.OverheadPress: 65.0, // Below 85 lbs threshold
			models.Squat:         80.0, // Below 85 lbs threshold
			models.BenchPress:    70.0,
			models.Deadlift:      85.0, // Exactly 85 lbs (no warmup)
		})

		result, err := CalculateNextWorkout(user, greyskullProgram)
		require.NoError(t, err)

		// Day 1 should have OverheadPress and Squat
		require.Len(t, result.Exercises, 2)

		// OverheadPress should have only 3 working sets (no warmup)
		ohpLift := result.Exercises[0]
		assert.Equal(t, models.OverheadPress, ohpLift.LiftName)
		require.Len(t, ohpLift.Sets, 3)

		for i, set := range ohpLift.Sets {
			assert.Equal(t, 65.0, set.Weight)
			assert.Equal(t, i+1, set.Order)
		}

		// Squat should have only 3 working sets (no warmup)
		squatLift := result.Exercises[1]
		assert.Equal(t, models.Squat, squatLift.LiftName)
		require.Len(t, squatLift.Sets, 3)

		for i, set := range squatLift.Sets {
			assert.Equal(t, 80.0, set.Weight)
			assert.Equal(t, i+1, set.Order)
		}
	})

	t.Run("error when no current program", func(t *testing.T) {
		user := &models.User{
			ID:             uuid.New(),
			Username:       "testuser",
			CurrentProgram: uuid.Nil, // No current program
			Programs:       make(map[uuid.UUID]*models.UserProgram),
			WorkoutHistory: []models.Workout{},
			CreatedAt:      time.Now(),
		}

		result, err := CalculateNextWorkout(user, greyskullProgram)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "no current program")
	})

	t.Run("error when current program not found", func(t *testing.T) {
		nonExistentProgramID := uuid.New()
		user := &models.User{
			ID:             uuid.New(),
			Username:       "testuser",
			CurrentProgram: nonExistentProgramID,
			Programs:       make(map[uuid.UUID]*models.UserProgram),
			WorkoutHistory: []models.Workout{},
			CreatedAt:      time.Now(),
		}

		result, err := CalculateNextWorkout(user, greyskullProgram)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "current program not found")
	})
}

func TestCalculateNextWorkout_OrderValues(t *testing.T) {
	user := createTestUser(1, map[models.LiftName]float64{
		models.OverheadPress: 100.0,
		models.Squat:         150.0,
		models.BenchPress:    125.0,
		models.Deadlift:      185.0,
	})

	result, err := CalculateNextWorkout(user, program.GreyskullLP)
	require.NoError(t, err)

	// Check that all sets have proper Order values
	for _, lift := range result.Exercises {
		for i, set := range lift.Sets {
			assert.Equal(t, i+1, set.Order, "Set order should be sequential starting from 1")
		}
	}
}

// Helper function to create a test user with a program
func createTestUser(currentDay int, weights map[models.LiftName]float64) *models.User {
	userProgram := &models.UserProgram{
		ID:              uuid.New(),
		UserID:          uuid.New(),
		ProgramID:       program.GreyskullLP.ID,
		StartingWeights: weights,
		CurrentWeights:  weights,
		CurrentDay:      currentDay,
		StartedAt:       time.Now(),
	}

	user := &models.User{
		ID:             userProgram.UserID,
		Username:       "testuser",
		CurrentProgram: userProgram.ID,
		Programs:       map[uuid.UUID]*models.UserProgram{userProgram.ID: userProgram},
		WorkoutHistory: []models.Workout{},
		CreatedAt:      time.Now(),
	}

	return user
}

