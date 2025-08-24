package display

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mikowitz/greyskull/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatWeight(t *testing.T) {
	tests := []struct {
		name     string
		weight   float64
		expected string
	}{
		{
			name:     "whole number weight",
			weight:   135.0,
			expected: "135",
		},
		{
			name:     "decimal weight",
			weight:   135.5,
			expected: "135.5",
		},
		{
			name:     "zero weight",
			weight:   0.0,
			expected: "0",
		},
		{
			name:     "small decimal",
			weight:   45.25,
			expected: "45.2", // Should truncate to 1 decimal place
		},
		{
			name:     "large weight",
			weight:   225.0,
			expected: "225",
		},
		{
			name:     "fractional weight",
			weight:   97.5,
			expected: "97.5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatWeight(tt.weight)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatLiftName(t *testing.T) {
	tests := []struct {
		name     string
		lift     models.LiftName
		expected string
	}{
		{
			name:     "squat",
			lift:     models.Squat,
			expected: "Squat",
		},
		{
			name:     "deadlift",
			lift:     models.Deadlift,
			expected: "Deadlift",
		},
		{
			name:     "bench press",
			lift:     models.BenchPress,
			expected: "Bench Press",
		},
		{
			name:     "overhead press",
			lift:     models.OverheadPress,
			expected: "Overhead Press",
		},
		{
			name:     "unknown lift",
			lift:     models.LiftName("UnknownLift"),
			expected: "UnknownLift",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatLiftName(tt.lift)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWorkoutFormatter_DisplayWorkout(t *testing.T) {
	tests := []struct {
		name            string
		workout         *models.Workout
		expectedContent []string // strings that should appear in output
	}{
		{
			name: "workout with warmup and working sets",
			workout: &models.Workout{
				ID:            uuid.Must(uuid.NewV7()),
				UserProgramID: uuid.Must(uuid.NewV7()),
				Day:           1,
				EnteredAt:     time.Now(),
				Exercises: []models.Lift{
					{
						ID:       uuid.Must(uuid.NewV7()),
						LiftName: models.OverheadPress,
						Sets: []models.Set{
							{
								ID:         uuid.Must(uuid.NewV7()),
								Weight:     45.0,
								TargetReps: 5,
								Type:       models.WarmupSet,
								Order:      1,
							},
							{
								ID:         uuid.Must(uuid.NewV7()),
								Weight:     52.5,
								TargetReps: 4,
								Type:       models.WarmupSet,
								Order:      2,
							},
							{
								ID:         uuid.Must(uuid.NewV7()),
								Weight:     95.0,
								TargetReps: 5,
								Type:       models.WorkingSet,
								Order:      3,
							},
							{
								ID:         uuid.Must(uuid.NewV7()),
								Weight:     95.0,
								TargetReps: 5,
								Type:       models.AMRAPSet,
								Order:      4,
							},
						},
					},
				},
			},
			expectedContent: []string{
				"Day 1 Workout:",
				"================",
				"Overhead Press:",
				"Warmup:",
				"5 reps @ 45 lbs",
				"4 reps @ 52.5 lbs",
				"Working Sets:",
				"Set 1: 5 reps @ 95 lbs",
				"Set 2: 5+ reps @ 95 lbs (AMRAP)",
			},
		},
		{
			name: "workout without warmup sets",
			workout: &models.Workout{
				ID:            uuid.Must(uuid.NewV7()),
				UserProgramID: uuid.Must(uuid.NewV7()),
				Day:           3,
				EnteredAt:     time.Now(),
				Exercises: []models.Lift{
					{
						ID:       uuid.Must(uuid.NewV7()),
						LiftName: models.Squat,
						Sets: []models.Set{
							{
								ID:         uuid.Must(uuid.NewV7()),
								Weight:     135.0,
								TargetReps: 5,
								Type:       models.WorkingSet,
								Order:      1,
							},
							{
								ID:         uuid.Must(uuid.NewV7()),
								Weight:     135.0,
								TargetReps: 5,
								Type:       models.WorkingSet,
								Order:      2,
							},
							{
								ID:         uuid.Must(uuid.NewV7()),
								Weight:     135.0,
								TargetReps: 5,
								Type:       models.AMRAPSet,
								Order:      3,
							},
						},
					},
				},
			},
			expectedContent: []string{
				"Day 3 Workout:",
				"================",
				"Squat:",
				"Working Sets:",
				"Set 1: 5 reps @ 135 lbs",
				"Set 2: 5 reps @ 135 lbs",
				"Set 3: 5+ reps @ 135 lbs (AMRAP)",
			},
		},
		{
			name: "multiple exercises workout",
			workout: &models.Workout{
				ID:            uuid.Must(uuid.NewV7()),
				UserProgramID: uuid.Must(uuid.NewV7()),
				Day:           2,
				EnteredAt:     time.Now(),
				Exercises: []models.Lift{
					{
						ID:       uuid.Must(uuid.NewV7()),
						LiftName: models.BenchPress,
						Sets: []models.Set{
							{
								ID:         uuid.Must(uuid.NewV7()),
								Weight:     185.0,
								TargetReps: 5,
								Type:       models.WorkingSet,
								Order:      1,
							},
						},
					},
					{
						ID:       uuid.Must(uuid.NewV7()),
						LiftName: models.Deadlift,
						Sets: []models.Set{
							{
								ID:         uuid.Must(uuid.NewV7()),
								Weight:     225.5,
								TargetReps: 5,
								Type:       models.AMRAPSet,
								Order:      1,
							},
						},
					},
				},
			},
			expectedContent: []string{
				"Day 2 Workout:",
				"Bench Press:",
				"Set 1: 5 reps @ 185 lbs",
				"Deadlift:",
				"Set 1: 5+ reps @ 225.5 lbs (AMRAP)",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			formatter := &WorkoutFormatter{out: &buf}

			formatter.DisplayWorkout(tt.workout)

			output := buf.String()
			for _, expected := range tt.expectedContent {
				assert.Contains(t, output, expected, "Output should contain: %s", expected)
			}
		})
	}
}

func TestWorkoutFormatter_DisplayWeightChanges(t *testing.T) {
	tests := []struct {
		name            string
		oldWeights      map[models.LiftName]float64
		newWeights      map[models.LiftName]float64
		expectedContent []string
		shouldBeEmpty   bool
	}{
		{
			name: "weight increases",
			oldWeights: map[models.LiftName]float64{
				models.OverheadPress: 95.0,
				models.Squat:         135.0,
			},
			newWeights: map[models.LiftName]float64{
				models.OverheadPress: 97.5,
				models.Squat:         140.0,
			},
			expectedContent: []string{
				"Weight Updates:",
				"Overhead Press: 95 → 97.5 lbs (+2.5)",
				"Squat: 135 → 140 lbs (+5.0)",
			},
		},
		{
			name: "weight decrease (deload)",
			oldWeights: map[models.LiftName]float64{
				models.BenchPress: 185.0,
			},
			newWeights: map[models.LiftName]float64{
				models.BenchPress: 166.5,
			},
			expectedContent: []string{
				"Weight Updates:",
				"Bench Press: 185 → 166.5 lbs (-18.5)",
			},
		},
		{
			name: "mixed changes",
			oldWeights: map[models.LiftName]float64{
				models.OverheadPress: 95.0,
				models.BenchPress:    185.0,
				models.Squat:         135.0,
				models.Deadlift:      225.0,
			},
			newWeights: map[models.LiftName]float64{
				models.OverheadPress: 100.0, // +5.0 (double progression)
				models.BenchPress:    166.5, // -18.5 (deload)
				models.Squat:         140.0, // +5.0 (normal)
				models.Deadlift:      225.0, // no change
			},
			expectedContent: []string{
				"Weight Updates:",
				"Overhead Press: 95 → 100 lbs (+5.0)",
				"Bench Press: 185 → 166.5 lbs (-18.5)",
				"Squat: 135 → 140 lbs (+5.0)",
			},
		},
		{
			name: "no changes",
			oldWeights: map[models.LiftName]float64{
				models.OverheadPress: 95.0,
				models.Squat:         135.0,
			},
			newWeights: map[models.LiftName]float64{
				models.OverheadPress: 95.0,
				models.Squat:         135.0,
			},
			shouldBeEmpty: true,
		},
		{
			name:          "empty weights",
			oldWeights:    map[models.LiftName]float64{},
			newWeights:    map[models.LiftName]float64{},
			shouldBeEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			formatter := &WorkoutFormatter{out: &buf}

			formatter.DisplayWeightChanges(tt.oldWeights, tt.newWeights)

			output := buf.String()

			if tt.shouldBeEmpty {
				assert.Empty(t, strings.TrimSpace(output), "Output should be empty for no changes")
				return
			}

			for _, expected := range tt.expectedContent {
				assert.Contains(t, output, expected, "Output should contain: %s", expected)
			}
		})
	}
}

func TestWorkoutFormatter_DisplayWorkoutSummary(t *testing.T) {
	tests := []struct {
		name            string
		workout         *models.Workout
		nextDay         int
		expectedContent []string
	}{
		{
			name: "basic completion summary",
			workout: &models.Workout{
				ID:            uuid.Must(uuid.NewV7()),
				UserProgramID: uuid.Must(uuid.NewV7()),
				Day:           1,
				EnteredAt:     time.Now(),
				Exercises: []models.Lift{
					{
						ID:       uuid.Must(uuid.NewV7()),
						LiftName: models.OverheadPress,
						Sets: []models.Set{
							{
								ID:         uuid.Must(uuid.NewV7()),
								Weight:     95.0,
								TargetReps: 5,
								ActualReps: 5,
								Type:       models.WorkingSet,
								Order:      1,
							},
							{
								ID:         uuid.Must(uuid.NewV7()),
								Weight:     95.0,
								TargetReps: 5,
								ActualReps: 7,
								Type:       models.AMRAPSet,
								Order:      2,
							},
						},
					},
				},
			},
			nextDay: 2,
			expectedContent: []string{
				"Workout logged successfully!",
				"Next workout: Day 2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			formatter := &WorkoutFormatter{out: &buf}

			formatter.DisplayWorkoutSummary(tt.workout, tt.nextDay)

			output := buf.String()
			for _, expected := range tt.expectedContent {
				assert.Contains(t, output, expected, "Output should contain: %s", expected)
			}
		})
	}
}

func TestWorkoutFormatter_FormatSetDisplay(t *testing.T) {
	tests := []struct {
		name     string
		set      models.Set
		setIndex int
		expected string
	}{
		{
			name: "working set",
			set: models.Set{
				Weight:     135.0,
				TargetReps: 5,
				Type:       models.WorkingSet,
			},
			setIndex: 1,
			expected: "Set 1: 5 reps @ 135 lbs",
		},
		{
			name: "AMRAP set",
			set: models.Set{
				Weight:     135.0,
				TargetReps: 5,
				Type:       models.AMRAPSet,
			},
			setIndex: 3,
			expected: "Set 3: 5+ reps @ 135 lbs (AMRAP)",
		},
		{
			name: "warmup set",
			set: models.Set{
				Weight:     45.0,
				TargetReps: 5,
				Type:       models.WarmupSet,
			},
			setIndex: 0, // Warmup sets don't use set index
			expected: "5 reps @ 45 lbs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatSetDisplay(tt.set, tt.setIndex)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWorkoutFormatter_IO_Integration(t *testing.T) {
	t.Run("output is written to provided writer", func(t *testing.T) {
		var buf bytes.Buffer
		formatter := &WorkoutFormatter{out: &buf}

		workout := &models.Workout{
			Day: 1,
			Exercises: []models.Lift{
				{
					LiftName: models.Squat,
					Sets: []models.Set{
						{
							Weight:     135.0,
							TargetReps: 5,
							Type:       models.WorkingSet,
						},
					},
				},
			},
		}

		formatter.DisplayWorkout(workout)

		output := buf.String()
		assert.NotEmpty(t, output)
		assert.Contains(t, output, "Day 1 Workout:")
		assert.Contains(t, output, "Squat:")
	})

	t.Run("formatter can be created with different writers", func(t *testing.T) {
		var buf1, buf2 bytes.Buffer

		formatter1 := &WorkoutFormatter{out: &buf1}
		formatter2 := &WorkoutFormatter{out: &buf2}

		oldWeights := map[models.LiftName]float64{models.Squat: 135.0}
		newWeights := map[models.LiftName]float64{models.Squat: 140.0}

		formatter1.DisplayWeightChanges(oldWeights, newWeights)
		formatter2.DisplayWeightChanges(oldWeights, newWeights)

		output1 := buf1.String()
		output2 := buf2.String()

		assert.Equal(t, output1, output2, "Both formatters should produce identical output")
		assert.Contains(t, output1, "Weight Updates:")
	})
}

// Test edge cases and error conditions
func TestWorkoutFormatter_EdgeCases(t *testing.T) {
	t.Run("empty workout", func(t *testing.T) {
		var buf bytes.Buffer
		formatter := &WorkoutFormatter{out: &buf}

		workout := &models.Workout{
			Day:       1,
			Exercises: []models.Lift{},
		}

		// Should not panic
		require.NotPanics(t, func() {
			formatter.DisplayWorkout(workout)
		})

		output := buf.String()
		assert.Contains(t, output, "Day 1 Workout:")
	})

	t.Run("nil workout", func(t *testing.T) {
		var buf bytes.Buffer
		formatter := &WorkoutFormatter{out: &buf}

		// Should handle nil gracefully or panic appropriately
		require.Panics(t, func() {
			formatter.DisplayWorkout(nil)
		})
	})

	t.Run("workout with no sets", func(t *testing.T) {
		var buf bytes.Buffer
		formatter := &WorkoutFormatter{out: &buf}

		workout := &models.Workout{
			Day: 1,
			Exercises: []models.Lift{
				{
					LiftName: models.Squat,
					Sets:     []models.Set{},
				},
			},
		}

		require.NotPanics(t, func() {
			formatter.DisplayWorkout(workout)
		})

		output := buf.String()
		assert.Contains(t, output, "Squat:")
		assert.Contains(t, output, "Working Sets:")
	})
}

