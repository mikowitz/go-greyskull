package workout

import (
	"testing"

	"github.com/google/uuid"
	"github.com/mikowitz/greyskull/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAMRAPReps(t *testing.T) {
	tests := []struct {
		name        string
		lift        models.Lift
		expected    int
		shouldError bool
	}{
		{
			name: "lift with AMRAP set",
			lift: models.Lift{
				ID:       uuid.New(),
				LiftName: models.OverheadPress,
				Sets: []models.Set{
					{Type: models.WarmupSet, ActualReps: 5},
					{Type: models.WorkingSet, ActualReps: 5},
					{Type: models.WorkingSet, ActualReps: 5},
					{Type: models.AMRAPSet, ActualReps: 8},
				},
			},
			expected:    8,
			shouldError: false,
		},
		{
			name: "lift without AMRAP set",
			lift: models.Lift{
				ID:       uuid.New(),
				LiftName: models.OverheadPress,
				Sets: []models.Set{
					{Type: models.WarmupSet, ActualReps: 5},
					{Type: models.WorkingSet, ActualReps: 5},
					{Type: models.WorkingSet, ActualReps: 5},
				},
			},
			expected:    0,
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetAMRAPReps(&tt.lift)
			
			if tt.shouldError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "no AMRAP set found")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestCalculateNewWeight(t *testing.T) {
	rules := &models.ProgressionRules{
		IncreaseRules: map[models.LiftName]float64{
			models.OverheadPress: 2.5,
			models.BenchPress:    2.5,
			models.Squat:         5.0,
			models.Deadlift:      5.0,
		},
		DeloadPercentage: 0.9,
		DoubleThreshold:  10,
	}

	tests := []struct {
		name           string
		currentWeight  float64
		amrapReps      int
		baseIncrement  float64
		expected       float64
		description    string
	}{
		{
			name:           "normal progression",
			currentWeight:  95.0,
			amrapReps:      6,
			baseIncrement:  2.5,
			expected:       97.5,
			description:    "5-9 reps should add base increment",
		},
		{
			name:           "double progression",
			currentWeight:  95.0,
			amrapReps:      12,
			baseIncrement:  2.5,
			expected:       100.0,
			description:    ">=10 reps should add double increment",
		},
		{
			name:           "deload",
			currentWeight:  95.0,
			amrapReps:      3,
			baseIncrement:  2.5,
			expected:       85.0,
			description:    "<5 reps should deload to 90%",
		},
		{
			name:           "edge case - exactly 5 reps",
			currentWeight:  95.0,
			amrapReps:      5,
			baseIncrement:  2.5,
			expected:       97.5,
			description:    "exactly 5 reps should be normal progression",
		},
		{
			name:           "edge case - exactly 10 reps",
			currentWeight:  95.0,
			amrapReps:      10,
			baseIncrement:  2.5,
			expected:       100.0,
			description:    "exactly 10 reps should be double progression",
		},
		{
			name:           "weight rounding",
			currentWeight:  97.3,
			amrapReps:      6,
			baseIncrement:  2.5,
			expected:       97.5,
			description:    "result should be rounded down to 2.5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateNewWeight(tt.currentWeight, tt.amrapReps, tt.baseIncrement, rules)
			assert.Equal(t, tt.expected, result, tt.description)
		})
	}
}

func TestCalculateProgression(t *testing.T) {
	// Create a sample completed workout
	workout := &models.Workout{
		ID:            uuid.New(),
		UserProgramID: uuid.New(),
		Day:           1,
		Exercises: []models.Lift{
			{
				ID:       uuid.New(),
				LiftName: models.OverheadPress,
				Sets: []models.Set{
					{Type: models.WarmupSet, ActualReps: 5},
					{Type: models.WorkingSet, ActualReps: 5},
					{Type: models.WorkingSet, ActualReps: 5},
					{Type: models.AMRAPSet, ActualReps: 8}, // Normal progression
				},
			},
			{
				ID:       uuid.New(),
				LiftName: models.Squat,
				Sets: []models.Set{
					{Type: models.WorkingSet, ActualReps: 5},
					{Type: models.WorkingSet, ActualReps: 5},
					{Type: models.AMRAPSet, ActualReps: 12}, // Double progression
				},
			},
		},
	}

	currentWeights := map[models.LiftName]float64{
		models.OverheadPress: 95.0,
		models.BenchPress:    125.0,
		models.Squat:         135.0,
		models.Deadlift:      185.0,
	}

	rules := &models.ProgressionRules{
		IncreaseRules: map[models.LiftName]float64{
			models.OverheadPress: 2.5,
			models.BenchPress:    2.5,
			models.Squat:         5.0,
			models.Deadlift:      5.0,
		},
		DeloadPercentage: 0.9,
		DoubleThreshold:  10,
	}

	newWeights, err := CalculateProgression(workout, currentWeights, rules)
	require.NoError(t, err)

	// Verify progressions
	assert.Equal(t, 97.5, newWeights[models.OverheadPress], "OverheadPress should have normal progression (+2.5)")
	assert.Equal(t, 145.0, newWeights[models.Squat], "Squat should have double progression (+10.0)")
	
	// Verify non-worked lifts remain unchanged
	assert.Equal(t, 125.0, newWeights[models.BenchPress], "BenchPress should remain unchanged")
	assert.Equal(t, 185.0, newWeights[models.Deadlift], "Deadlift should remain unchanged")
}

func TestCalculateProgression_ErrorCases(t *testing.T) {
	currentWeights := map[models.LiftName]float64{
		models.OverheadPress: 95.0,
		models.Squat:         135.0,
	}

	rules := &models.ProgressionRules{
		IncreaseRules: map[models.LiftName]float64{
			models.OverheadPress: 2.5,
			models.Squat:         5.0,
		},
		DeloadPercentage: 0.9,
		DoubleThreshold:  10,
	}

	t.Run("no AMRAP set", func(t *testing.T) {
		workout := &models.Workout{
			Exercises: []models.Lift{
				{
					LiftName: models.OverheadPress,
					Sets: []models.Set{
						{Type: models.WorkingSet, ActualReps: 5},
					},
				},
			},
		}

		_, err := CalculateProgression(workout, currentWeights, rules)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no AMRAP set found")
	})

	t.Run("missing progression rule", func(t *testing.T) {
		workout := &models.Workout{
			Exercises: []models.Lift{
				{
					LiftName: models.BenchPress, // Not in rules
					Sets: []models.Set{
						{Type: models.AMRAPSet, ActualReps: 6},
					},
				},
			},
		}

		_, err := CalculateProgression(workout, currentWeights, rules)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no progression rule found")
	})

	t.Run("missing current weight", func(t *testing.T) {
		workout := &models.Workout{
			Exercises: []models.Lift{
				{
					LiftName: models.Deadlift, // Not in currentWeights
					Sets: []models.Set{
						{Type: models.AMRAPSet, ActualReps: 6},
					},
				},
			},
		}

		incompleteRules := &models.ProgressionRules{
			IncreaseRules: map[models.LiftName]float64{
				models.Deadlift: 5.0, // Rule exists but weight doesn't
			},
			DeloadPercentage: 0.9,
			DoubleThreshold:  10,
		}

		_, err := CalculateProgression(workout, currentWeights, incompleteRules)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "current weight not found")
	})
}

// TestRoundDown2_5 is already tested in calculator_test.go