package workout

import (
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/mikowitz/greyskull/models"
)

func RoundDown2_5(input float64) float64 {
	return math.Floor(input/2.5) * 2.5
}

func CalculateWarmupSets(weight float64, setTemplates []models.SetTemplate) []models.Set {
	sets := []models.Set{}
	if weight <= 85.0 {
		return sets
	}
	for i, tpl := range setTemplates {
		setWeight := 45.0
		if tpl.WeightPercentage > 0.0 {
			setWeight = RoundDown2_5(weight * tpl.WeightPercentage)
		}
		set := models.Set{
			ID:         uuid.Must(uuid.NewV7()),
			Weight:     setWeight,
			TargetReps: tpl.Reps,
			Type:       tpl.Type,
			Order:      i + 1,
		}
		sets = append(sets, set)

	}
	return sets
}

func CalculateWorkingSets(weight float64, setTemplates []models.SetTemplate) []models.Set {
	sets := []models.Set{}
	weight = RoundDown2_5(weight)
	for i, tpl := range setTemplates {
		set := models.Set{
			ID:         uuid.Must(uuid.NewV7()),
			Weight:     weight,
			TargetReps: tpl.Reps,
			Type:       tpl.Type,
			Order:      i + 1,
		}
		sets = append(sets, set)
	}
	return sets
}

func GetWorkoutDay(currentDay, totalDays int) int {
	mod := currentDay % totalDays
	if mod == 0 {
		return totalDays
	}
	return mod
}

func CalculateNextWorkout(user *models.User, program *models.Program) (*models.Workout, error) {
	// Check if user has a current program
	if user.CurrentProgram == uuid.Nil {
		return nil, fmt.Errorf("no current program set for user")
	}

	// Get current UserProgram
	userProgram, exists := user.Programs[user.CurrentProgram]
	if !exists {
		return nil, fmt.Errorf("current program not found in user programs")
	}

	// Get current day and handle cycle wrapping
	workoutDay := GetWorkoutDay(userProgram.CurrentDay, len(program.Workouts))
	
	// Get WorkoutTemplate for that day (convert to 0-based index)
	workoutTemplate := program.Workouts[workoutDay-1]

	// Create the workout
	workout := &models.Workout{
		ID:            uuid.Must(uuid.NewV7()),
		UserProgramID: userProgram.ID,
		Day:           workoutDay,
		Exercises:     make([]models.Lift, 0, len(workoutTemplate.Lifts)),
		EnteredAt:     time.Now(),
	}

	// For each LiftTemplate, calculate sets and create Lift
	for _, liftTemplate := range workoutTemplate.Lifts {
		// Get current weight for this lift
		currentWeight, exists := userProgram.CurrentWeights[liftTemplate.LiftName]
		if !exists {
			return nil, fmt.Errorf("current weight not found for lift %s", liftTemplate.LiftName)
		}

		// Calculate warmup sets (may be empty if weight < 85 lbs)
		warmupSets := CalculateWarmupSets(currentWeight, liftTemplate.WarmupSets)

		// Calculate working sets
		workingSets := CalculateWorkingSets(currentWeight, liftTemplate.WorkingSets)

		// Combine all sets and adjust order for working sets
		allSets := make([]models.Set, 0, len(warmupSets)+len(workingSets))
		allSets = append(allSets, warmupSets...)
		
		// Adjust order for working sets to continue from warmup sets
		for i := range workingSets {
			workingSets[i].Order = len(warmupSets) + i + 1
			allSets = append(allSets, workingSets[i])
		}

		// Create Lift with all sets
		lift := models.Lift{
			ID:       uuid.Must(uuid.NewV7()),
			LiftName: liftTemplate.LiftName,
			Sets:     allSets,
		}

		workout.Exercises = append(workout.Exercises, lift)
	}

	return workout, nil
}

// GetAMRAPReps finds and returns the actual reps completed in the AMRAP set for a given lift
func GetAMRAPReps(lift *models.Lift) (int, error) {
	for _, set := range lift.Sets {
		if set.Type == models.AMRAPSet {
			return set.ActualReps, nil
		}
	}
	return 0, fmt.Errorf("no AMRAP set found for lift %s", lift.LiftName)
}

// CalculateNewWeight determines the new weight based on AMRAP performance
func CalculateNewWeight(currentWeight float64, amrapReps int, baseIncrement float64, rules *models.ProgressionRules) float64 {
	var newWeight float64
	
	if amrapReps < 5 {
		// Deload - reduce weight by deload percentage
		newWeight = currentWeight * rules.DeloadPercentage
	} else if amrapReps >= rules.DoubleThreshold {
		// Double progression - add double the base increment
		newWeight = currentWeight + (baseIncrement * 2)
	} else {
		// Normal progression - add base increment
		newWeight = currentWeight + baseIncrement
	}
	
	// Round down to 2.5 lbs
	return RoundDown2_5(newWeight)
}

// CalculateProgression calculates new weights for all lifts based on workout performance
func CalculateProgression(workout *models.Workout, currentWeights map[models.LiftName]float64, rules *models.ProgressionRules) (map[models.LiftName]float64, error) {
	newWeights := make(map[models.LiftName]float64)
	
	// Copy current weights first
	for liftName, weight := range currentWeights {
		newWeights[liftName] = weight
	}
	
	// Update weights for lifts that were performed in this workout
	for _, lift := range workout.Exercises {
		// Get AMRAP reps for this lift
		amrapReps, err := GetAMRAPReps(&lift)
		if err != nil {
			return nil, fmt.Errorf("failed to get AMRAP reps for %s: %w", lift.LiftName, err)
		}
		
		// Get base increment for this lift
		baseIncrement, exists := rules.IncreaseRules[lift.LiftName]
		if !exists {
			return nil, fmt.Errorf("no progression rule found for lift %s", lift.LiftName)
		}
		
		// Get current weight
		currentWeight, exists := currentWeights[lift.LiftName]
		if !exists {
			return nil, fmt.Errorf("current weight not found for lift %s", lift.LiftName)
		}
		
		// Calculate new weight
		newWeights[lift.LiftName] = CalculateNewWeight(currentWeight, amrapReps, baseIncrement, rules)
	}
	
	return newWeights, nil
}
