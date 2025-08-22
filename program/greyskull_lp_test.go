package program

import (
	"fmt"
	"testing"

	"github.com/mikowitz/greyskull/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGreyskullLP_ProgramStructure(t *testing.T) {
	program := GreyskullLP

	// Test basic program properties
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", program.ID.String())
	assert.Equal(t, "OG Greyskull LP", program.Name)
	assert.Equal(t, "1.0.0", program.Version)

	// Test program has 6 workout days
	require.Len(t, program.Workouts, 6)

	// Test each day is numbered correctly
	for i, workout := range program.Workouts {
		assert.Equal(t, i+1, workout.Day)
	}
}

func TestGreyskullLP_WorkoutCycle(t *testing.T) {
	program := GreyskullLP

	expectedExercises := []struct {
		day       int
		exercises []models.LiftName
	}{
		{day: 1, exercises: []models.LiftName{models.OverheadPress, models.Squat}},
		{day: 2, exercises: []models.LiftName{models.BenchPress, models.Deadlift}},
		{day: 3, exercises: []models.LiftName{models.OverheadPress, models.Squat}},
		{day: 4, exercises: []models.LiftName{models.BenchPress, models.Squat}},
		{day: 5, exercises: []models.LiftName{models.OverheadPress, models.Deadlift}},
		{day: 6, exercises: []models.LiftName{models.BenchPress, models.Squat}},
	}

	for _, expected := range expectedExercises {
		t.Run(fmt.Sprintf("Day_%d", expected.day), func(t *testing.T) {
			workout := program.Workouts[expected.day-1]
			require.Len(t, workout.Lifts, 2, "Each day should have exactly 2 exercises")

			actualExercises := make([]models.LiftName, len(workout.Lifts))
			for i, lift := range workout.Lifts {
				actualExercises[i] = lift.LiftName
			}

			assert.Equal(t, expected.exercises, actualExercises)
		})
	}
}

func TestGreyskullLP_WarmupProtocol(t *testing.T) {
	program := GreyskullLP

	expectedWarmupSets := []models.SetTemplate{
		{Reps: 5, WeightPercentage: 0.0, Type: models.WarmupSet},   // Empty bar
		{Reps: 4, WeightPercentage: 0.55, Type: models.WarmupSet}, // 55%
		{Reps: 3, WeightPercentage: 0.70, Type: models.WarmupSet}, // 70%
		{Reps: 2, WeightPercentage: 0.85, Type: models.WarmupSet}, // 85%
	}

	// Test warmup protocol is consistent across all lifts and days
	for dayIdx, workout := range program.Workouts {
		for liftIdx, lift := range workout.Lifts {
			t.Run(fmt.Sprintf("Day_%d_Lift_%d_%s", dayIdx+1, liftIdx+1, string(lift.LiftName)), func(t *testing.T) {
				require.Len(t, lift.WarmupSets, 4, "Each lift should have 4 warmup sets")

				for i, actualSet := range lift.WarmupSets {
					expected := expectedWarmupSets[i]
					assert.Equal(t, expected.Reps, actualSet.Reps, "Warmup set %d reps", i+1)
					assert.Equal(t, expected.WeightPercentage, actualSet.WeightPercentage, "Warmup set %d weight percentage", i+1)
					assert.Equal(t, expected.Type, actualSet.Type, "Warmup set %d type", i+1)
				}
			})
		}
	}
}

func TestGreyskullLP_WorkingSets(t *testing.T) {
	program := GreyskullLP

	expectedWorkingSets := []models.SetTemplate{
		{Reps: 5, WeightPercentage: 1.0, Type: models.WorkingSet}, // First working set
		{Reps: 5, WeightPercentage: 1.0, Type: models.WorkingSet}, // Second working set
		{Reps: 5, WeightPercentage: 1.0, Type: models.AMRAPSet},   // AMRAP set
	}

	// Test working sets are consistent across all lifts and days
	for dayIdx, workout := range program.Workouts {
		for liftIdx, lift := range workout.Lifts {
			t.Run(fmt.Sprintf("Day_%d_Lift_%d_%s", dayIdx+1, liftIdx+1, string(lift.LiftName)), func(t *testing.T) {
				require.Len(t, lift.WorkingSets, 3, "Each lift should have 3 working sets")

				for i, actualSet := range lift.WorkingSets {
					expected := expectedWorkingSets[i]
					assert.Equal(t, expected.Reps, actualSet.Reps, "Working set %d reps", i+1)
					assert.Equal(t, expected.WeightPercentage, actualSet.WeightPercentage, "Working set %d weight percentage", i+1)
					assert.Equal(t, expected.Type, actualSet.Type, "Working set %d type", i+1)
				}

				// Verify the third set is AMRAP
				assert.Equal(t, models.AMRAPSet, lift.WorkingSets[2].Type, "Third working set should be AMRAP")
			})
		}
	}
}

func TestGreyskullLP_ProgressionRules(t *testing.T) {
	program := GreyskullLP
	rules := program.ProgressionRules

	// Test increase rules
	expectedIncreases := map[models.LiftName]float64{
		models.OverheadPress: 2.5,
		models.BenchPress:    2.5,
		models.Squat:        5.0,
		models.Deadlift:     5.0,
	}

	assert.Equal(t, expectedIncreases, rules.IncreaseRules, "Progression increase rules should match specification")

	// Test deload percentage
	assert.Equal(t, 0.9, rules.DeloadPercentage, "Deload should be to 90% of current weight")

	// Test double progression threshold
	assert.Equal(t, 10, rules.DoubleThreshold, "Double progression threshold should be 10 reps")

	// Verify all core lifts have progression rules
	coreLifts := []models.LiftName{models.Squat, models.Deadlift, models.BenchPress, models.OverheadPress}
	for _, lift := range coreLifts {
		_, exists := rules.IncreaseRules[lift]
		assert.True(t, exists, "Lift %s should have progression rule", string(lift))
	}
}

func TestGetByID(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		wantErr  error
		wantNil  bool
	}{
		{
			name:    "valid Greyskull LP ID",
			id:      "550e8400-e29b-41d4-a716-446655440000",
			wantErr: nil,
			wantNil: false,
		},
		{
			name:    "invalid ID",
			id:      "invalid-id",
			wantErr: ErrProgramNotFound,
			wantNil: true,
		},
		{
			name:    "empty ID",
			id:      "",
			wantErr: ErrProgramNotFound,
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program, err := GetByID(tt.id)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}

			if tt.wantNil {
				assert.Nil(t, program)
			} else {
				assert.NotNil(t, program)
				assert.Equal(t, GreyskullLP.ID.String(), program.ID.String())
				assert.Equal(t, GreyskullLP.Name, program.Name)
			}
		})
	}
}

func TestList(t *testing.T) {
	programs := List()

	// Should return exactly one program (Greyskull LP)
	require.Len(t, programs, 1)

	program := programs[0]
	assert.Equal(t, GreyskullLP.ID.String(), program.ID.String())
	assert.Equal(t, GreyskullLP.Name, program.Name)
	assert.Equal(t, GreyskullLP.Version, program.Version)

	// Verify it's the same instance
	assert.Same(t, GreyskullLP, program)
}

func TestGreyskullLP_AllLiftsPresent(t *testing.T) {
	program := GreyskullLP

	// Count occurrences of each lift across all days
	liftCounts := make(map[models.LiftName]int)

	for _, workout := range program.Workouts {
		for _, lift := range workout.Lifts {
			liftCounts[lift.LiftName]++
		}
	}

	// Verify expected frequencies based on the 6-day cycle
	expectedCounts := map[models.LiftName]int{
		models.OverheadPress: 3, // Days 1, 3, 5
		models.BenchPress:    3, // Days 2, 4, 6
		models.Squat:        4, // Days 1, 3, 4, 6
		models.Deadlift:     2, // Days 2, 5
	}

	assert.Equal(t, expectedCounts, liftCounts, "Lift frequencies should match expected pattern")
}

func TestGreyskullLP_ComprehensiveValidation(t *testing.T) {
	program := GreyskullLP

	// Verify program is not nil
	require.NotNil(t, program)

	// Verify all required fields are present
	assert.NotEmpty(t, program.ID)
	assert.NotEmpty(t, program.Name)
	assert.NotEmpty(t, program.Version)
	assert.NotEmpty(t, program.Workouts)

	// Verify progression rules
	assert.NotNil(t, program.ProgressionRules.IncreaseRules)
	assert.Greater(t, program.ProgressionRules.DeloadPercentage, 0.0)
	assert.Greater(t, program.ProgressionRules.DoubleThreshold, 0)

	// Verify each workout day
	for i, workout := range program.Workouts {
		assert.Equal(t, i+1, workout.Day, "Workout day should match index")
		assert.NotEmpty(t, workout.Lifts, "Each workout should have lifts")

		// Verify each lift in the workout
		for j, lift := range workout.Lifts {
			assert.NotEmpty(t, string(lift.LiftName), "Lift name should not be empty")
			assert.NotEmpty(t, lift.WarmupSets, "Lift should have warmup sets")
			assert.NotEmpty(t, lift.WorkingSets, "Lift should have working sets")

			// Verify warmup sets
			for k, warmupSet := range lift.WarmupSets {
				assert.Greater(t, warmupSet.Reps, 0, "Warmup set %d should have positive reps", k+1)
				assert.GreaterOrEqual(t, warmupSet.WeightPercentage, 0.0, "Warmup set %d weight percentage should be non-negative", k+1)
				assert.Equal(t, models.WarmupSet, warmupSet.Type, "Warmup set %d should have WarmupSet type", k+1)
			}

			// Verify working sets
			for k, workingSet := range lift.WorkingSets {
				assert.Greater(t, workingSet.Reps, 0, "Working set %d should have positive reps", k+1)
				assert.Equal(t, 1.0, workingSet.WeightPercentage, "Working set %d should use 100% weight", k+1)
				if k < len(lift.WorkingSets)-1 {
					assert.Equal(t, models.WorkingSet, workingSet.Type, "Non-final working set %d should have WorkingSet type", k+1)
				} else {
					assert.Equal(t, models.AMRAPSet, workingSet.Type, "Final working set should have AMRAPSet type")
				}
			}

			t.Logf("Day %d, Lift %d (%s): %d warmup sets, %d working sets", 
				i+1, j+1, string(lift.LiftName), len(lift.WarmupSets), len(lift.WorkingSets))
		}
	}
}