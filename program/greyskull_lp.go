package program

import (
	"errors"

	"github.com/google/uuid"
	"github.com/mikowitz/greyskull/models"
)

// Sentinel errors for program operations
var (
	ErrProgramNotFound = errors.New("program not found")
)

// GreyskullLP is the complete OG Greyskull LP program template
var GreyskullLP = &models.Program{
	ID:      uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"), // Fixed UUID for consistency
	Name:    "OG Greyskull LP",
	Version: "1.0.0",
	Workouts: []models.WorkoutTemplate{
		// Day 1: Overhead Press, Squat
		{
			Day: 1,
			Lifts: []models.LiftTemplate{
				{
					LiftName: models.OverheadPress,
					WarmupSets: []models.SetTemplate{
						{Reps: 5, WeightPercentage: 0.0, Type: models.WarmupSet},   // Empty bar
						{Reps: 4, WeightPercentage: 0.55, Type: models.WarmupSet}, // 55%
						{Reps: 3, WeightPercentage: 0.70, Type: models.WarmupSet}, // 70%
						{Reps: 2, WeightPercentage: 0.85, Type: models.WarmupSet}, // 85%
					},
					WorkingSets: []models.SetTemplate{
						{Reps: 5, WeightPercentage: 1.0, Type: models.WorkingSet},
						{Reps: 5, WeightPercentage: 1.0, Type: models.WorkingSet},
						{Reps: 5, WeightPercentage: 1.0, Type: models.AMRAPSet},
					},
				},
				{
					LiftName: models.Squat,
					WarmupSets: []models.SetTemplate{
						{Reps: 5, WeightPercentage: 0.0, Type: models.WarmupSet},
						{Reps: 4, WeightPercentage: 0.55, Type: models.WarmupSet},
						{Reps: 3, WeightPercentage: 0.70, Type: models.WarmupSet},
						{Reps: 2, WeightPercentage: 0.85, Type: models.WarmupSet},
					},
					WorkingSets: []models.SetTemplate{
						{Reps: 5, WeightPercentage: 1.0, Type: models.WorkingSet},
						{Reps: 5, WeightPercentage: 1.0, Type: models.WorkingSet},
						{Reps: 5, WeightPercentage: 1.0, Type: models.AMRAPSet},
					},
				},
			},
		},
		// Day 2: Bench Press, Deadlift
		{
			Day: 2,
			Lifts: []models.LiftTemplate{
				{
					LiftName: models.BenchPress,
					WarmupSets: []models.SetTemplate{
						{Reps: 5, WeightPercentage: 0.0, Type: models.WarmupSet},
						{Reps: 4, WeightPercentage: 0.55, Type: models.WarmupSet},
						{Reps: 3, WeightPercentage: 0.70, Type: models.WarmupSet},
						{Reps: 2, WeightPercentage: 0.85, Type: models.WarmupSet},
					},
					WorkingSets: []models.SetTemplate{
						{Reps: 5, WeightPercentage: 1.0, Type: models.WorkingSet},
						{Reps: 5, WeightPercentage: 1.0, Type: models.WorkingSet},
						{Reps: 5, WeightPercentage: 1.0, Type: models.AMRAPSet},
					},
				},
				{
					LiftName: models.Deadlift,
					WarmupSets: []models.SetTemplate{
						{Reps: 5, WeightPercentage: 0.0, Type: models.WarmupSet},
						{Reps: 4, WeightPercentage: 0.55, Type: models.WarmupSet},
						{Reps: 3, WeightPercentage: 0.70, Type: models.WarmupSet},
						{Reps: 2, WeightPercentage: 0.85, Type: models.WarmupSet},
					},
					WorkingSets: []models.SetTemplate{
						{Reps: 5, WeightPercentage: 1.0, Type: models.WorkingSet},
						{Reps: 5, WeightPercentage: 1.0, Type: models.WorkingSet},
						{Reps: 5, WeightPercentage: 1.0, Type: models.AMRAPSet},
					},
				},
			},
		},
		// Day 3: Overhead Press, Squat
		{
			Day: 3,
			Lifts: []models.LiftTemplate{
				{
					LiftName: models.OverheadPress,
					WarmupSets: []models.SetTemplate{
						{Reps: 5, WeightPercentage: 0.0, Type: models.WarmupSet},
						{Reps: 4, WeightPercentage: 0.55, Type: models.WarmupSet},
						{Reps: 3, WeightPercentage: 0.70, Type: models.WarmupSet},
						{Reps: 2, WeightPercentage: 0.85, Type: models.WarmupSet},
					},
					WorkingSets: []models.SetTemplate{
						{Reps: 5, WeightPercentage: 1.0, Type: models.WorkingSet},
						{Reps: 5, WeightPercentage: 1.0, Type: models.WorkingSet},
						{Reps: 5, WeightPercentage: 1.0, Type: models.AMRAPSet},
					},
				},
				{
					LiftName: models.Squat,
					WarmupSets: []models.SetTemplate{
						{Reps: 5, WeightPercentage: 0.0, Type: models.WarmupSet},
						{Reps: 4, WeightPercentage: 0.55, Type: models.WarmupSet},
						{Reps: 3, WeightPercentage: 0.70, Type: models.WarmupSet},
						{Reps: 2, WeightPercentage: 0.85, Type: models.WarmupSet},
					},
					WorkingSets: []models.SetTemplate{
						{Reps: 5, WeightPercentage: 1.0, Type: models.WorkingSet},
						{Reps: 5, WeightPercentage: 1.0, Type: models.WorkingSet},
						{Reps: 5, WeightPercentage: 1.0, Type: models.AMRAPSet},
					},
				},
			},
		},
		// Day 4: Bench Press, Squat
		{
			Day: 4,
			Lifts: []models.LiftTemplate{
				{
					LiftName: models.BenchPress,
					WarmupSets: []models.SetTemplate{
						{Reps: 5, WeightPercentage: 0.0, Type: models.WarmupSet},
						{Reps: 4, WeightPercentage: 0.55, Type: models.WarmupSet},
						{Reps: 3, WeightPercentage: 0.70, Type: models.WarmupSet},
						{Reps: 2, WeightPercentage: 0.85, Type: models.WarmupSet},
					},
					WorkingSets: []models.SetTemplate{
						{Reps: 5, WeightPercentage: 1.0, Type: models.WorkingSet},
						{Reps: 5, WeightPercentage: 1.0, Type: models.WorkingSet},
						{Reps: 5, WeightPercentage: 1.0, Type: models.AMRAPSet},
					},
				},
				{
					LiftName: models.Squat,
					WarmupSets: []models.SetTemplate{
						{Reps: 5, WeightPercentage: 0.0, Type: models.WarmupSet},
						{Reps: 4, WeightPercentage: 0.55, Type: models.WarmupSet},
						{Reps: 3, WeightPercentage: 0.70, Type: models.WarmupSet},
						{Reps: 2, WeightPercentage: 0.85, Type: models.WarmupSet},
					},
					WorkingSets: []models.SetTemplate{
						{Reps: 5, WeightPercentage: 1.0, Type: models.WorkingSet},
						{Reps: 5, WeightPercentage: 1.0, Type: models.WorkingSet},
						{Reps: 5, WeightPercentage: 1.0, Type: models.AMRAPSet},
					},
				},
			},
		},
		// Day 5: Overhead Press, Deadlift
		{
			Day: 5,
			Lifts: []models.LiftTemplate{
				{
					LiftName: models.OverheadPress,
					WarmupSets: []models.SetTemplate{
						{Reps: 5, WeightPercentage: 0.0, Type: models.WarmupSet},
						{Reps: 4, WeightPercentage: 0.55, Type: models.WarmupSet},
						{Reps: 3, WeightPercentage: 0.70, Type: models.WarmupSet},
						{Reps: 2, WeightPercentage: 0.85, Type: models.WarmupSet},
					},
					WorkingSets: []models.SetTemplate{
						{Reps: 5, WeightPercentage: 1.0, Type: models.WorkingSet},
						{Reps: 5, WeightPercentage: 1.0, Type: models.WorkingSet},
						{Reps: 5, WeightPercentage: 1.0, Type: models.AMRAPSet},
					},
				},
				{
					LiftName: models.Deadlift,
					WarmupSets: []models.SetTemplate{
						{Reps: 5, WeightPercentage: 0.0, Type: models.WarmupSet},
						{Reps: 4, WeightPercentage: 0.55, Type: models.WarmupSet},
						{Reps: 3, WeightPercentage: 0.70, Type: models.WarmupSet},
						{Reps: 2, WeightPercentage: 0.85, Type: models.WarmupSet},
					},
					WorkingSets: []models.SetTemplate{
						{Reps: 5, WeightPercentage: 1.0, Type: models.WorkingSet},
						{Reps: 5, WeightPercentage: 1.0, Type: models.WorkingSet},
						{Reps: 5, WeightPercentage: 1.0, Type: models.AMRAPSet},
					},
				},
			},
		},
		// Day 6: Bench Press, Squat
		{
			Day: 6,
			Lifts: []models.LiftTemplate{
				{
					LiftName: models.BenchPress,
					WarmupSets: []models.SetTemplate{
						{Reps: 5, WeightPercentage: 0.0, Type: models.WarmupSet},
						{Reps: 4, WeightPercentage: 0.55, Type: models.WarmupSet},
						{Reps: 3, WeightPercentage: 0.70, Type: models.WarmupSet},
						{Reps: 2, WeightPercentage: 0.85, Type: models.WarmupSet},
					},
					WorkingSets: []models.SetTemplate{
						{Reps: 5, WeightPercentage: 1.0, Type: models.WorkingSet},
						{Reps: 5, WeightPercentage: 1.0, Type: models.WorkingSet},
						{Reps: 5, WeightPercentage: 1.0, Type: models.AMRAPSet},
					},
				},
				{
					LiftName: models.Squat,
					WarmupSets: []models.SetTemplate{
						{Reps: 5, WeightPercentage: 0.0, Type: models.WarmupSet},
						{Reps: 4, WeightPercentage: 0.55, Type: models.WarmupSet},
						{Reps: 3, WeightPercentage: 0.70, Type: models.WarmupSet},
						{Reps: 2, WeightPercentage: 0.85, Type: models.WarmupSet},
					},
					WorkingSets: []models.SetTemplate{
						{Reps: 5, WeightPercentage: 1.0, Type: models.WorkingSet},
						{Reps: 5, WeightPercentage: 1.0, Type: models.WorkingSet},
						{Reps: 5, WeightPercentage: 1.0, Type: models.AMRAPSet},
					},
				},
			},
		},
	},
	ProgressionRules: models.ProgressionRules{
		IncreaseRules: map[models.LiftName]float64{
			models.OverheadPress: 2.5, // Upper body: +2.5 lbs
			models.BenchPress:    2.5, // Upper body: +2.5 lbs
			models.Squat:        5.0,  // Lower body: +5.0 lbs
			models.Deadlift:     5.0,  // Lower body: +5.0 lbs
		},
		DeloadPercentage: 0.9,  // Deload to 90% on failure
		DoubleThreshold:  10,   // Double progression at 10+ reps
	},
}

// GetByID retrieves a program by its ID
func GetByID(id string) (*models.Program, error) {
	if id == GreyskullLP.ID.String() {
		return GreyskullLP, nil
	}
	return nil, ErrProgramNotFound
}

// List returns all available programs
func List() []*models.Program {
	return []*models.Program{GreyskullLP}
}