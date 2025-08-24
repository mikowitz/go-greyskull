package display

import (
	"fmt"
	"io"
	"strconv"

	"github.com/mikowitz/greyskull/models"
)

type WorkoutFormatter struct {
	out io.Writer
}

func NewWorkoutFormatter(out io.Writer) *WorkoutFormatter {
	return &WorkoutFormatter{out: out}
}

func (f *WorkoutFormatter) Printf(format string, a ...any) {
	f.out.Write(fmt.Appendf([]byte{}, format, a...))
}

func (f *WorkoutFormatter) DisplayWorkout(workout *models.Workout) {
	f.Printf("Day %d Workout:\n", workout.Day)
	f.Printf("================\n\n")

	for _, lift := range workout.Exercises {
		f.Printf("%s:\n", FormatLiftName(lift.LiftName))

		// Group sets by type
		warmupSets := []models.Set{}
		workingSets := []models.Set{}

		for _, set := range lift.Sets {
			if set.Type == models.WarmupSet {
				warmupSets = append(warmupSets, set)
			} else {
				workingSets = append(workingSets, set)
			}
		}

		// Display warmup sets if any
		if len(warmupSets) > 0 {
			f.Printf("  Warmup:\n")
			for _, set := range warmupSets {
				f.Printf("    %d reps @ %s lbs\n", set.TargetReps, FormatWeight(set.Weight))
			}
		}

		// Display working sets
		f.Printf("  Working Sets:\n")
		for i, set := range workingSets {
			if set.Type == models.AMRAPSet {
				f.Printf("    Set %d: %d+ reps @ %s lbs (AMRAP)\n", i+1, set.TargetReps, FormatWeight(set.Weight))
			} else {
				f.Printf("    Set %d: %d reps @ %s lbs\n", i+1, set.TargetReps, FormatWeight(set.Weight))
			}
		}

		f.Printf("\n")
	}
}

func (f *WorkoutFormatter) DisplayWeightChanges(old, new map[models.LiftName]float64) {
	hasChanges := false

	// Check if any weights changed
	for liftName, newWeight := range new {
		if oldWeight, exists := old[liftName]; exists && oldWeight != newWeight {
			hasChanges = true
			break
		}
	}

	if !hasChanges {
		return // No changes to display
	}

	f.Printf("\nWeight Updates:\n")

	// Display changes for each lift that was worked
	lifts := []models.LiftName{models.OverheadPress, models.BenchPress, models.Squat, models.Deadlift}
	for _, liftName := range lifts {
		oldWeight, oldExists := old[liftName]
		newWeight, newExists := new[liftName]

		if oldExists && newExists && oldWeight != newWeight {
			difference := newWeight - oldWeight
			var sign string
			if difference > 0 {
				sign = "+"
			}

			f.Printf("%s: %s → %s lbs (%s%.1f)\n",
				FormatLiftName(liftName),
				FormatWeight(oldWeight),
				FormatWeight(newWeight),
				sign,
				difference)
		}
	}
}

func (f *WorkoutFormatter) DisplayWorkoutSummary(workout *models.Workout, nextDay int) {
	f.DisplayWorkout(workout)

	f.Printf("\nWorkout logged successfully!\n")
	f.Printf("Next workout: Day %d\n", nextDay)
}

func FormatWeight(weight float64) string {
	// Remove decimal if it's a whole number
	if weight == float64(int(weight)) {
		return strconv.Itoa(int(weight))
	}
	return fmt.Sprintf("%.1f", weight)
}

func FormatLiftName(lift models.LiftName) string {
	switch lift {
	case models.Squat:
		return "Squat"
	case models.Deadlift:
		return "Deadlift"
	case models.BenchPress:
		return "Bench Press"
	case models.OverheadPress:
		return "Overhead Press"
	default:
		return string(lift)
	}
}

func FormatSetDisplay(set models.Set, index int) string {
	switch set.Type {
	case models.WarmupSet:
		return fmt.Sprintf("%d reps @ %s lbs", set.TargetReps, FormatWeight(set.Weight))
	case models.AMRAPSet:
		return fmt.Sprintf("Set %d: %d+ reps @ %s lbs (AMRAP)", index, set.TargetReps, FormatWeight(set.Weight))
	default:
		return fmt.Sprintf("Set %d: %d reps @ %s lbs", index, set.TargetReps, FormatWeight(set.Weight))
	}
}
