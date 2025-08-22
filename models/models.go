package models

import (
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Type definitions
type (
	LiftName string
	SetType  string
)

// LiftName constants
const (
	Squat         LiftName = "Squat"
	Deadlift      LiftName = "Deadlift"
	BenchPress    LiftName = "BenchPress"
	OverheadPress LiftName = "OverheadPress"
)

// SetType constants
const (
	WarmupSet  SetType = "WarmupSet"
	WorkingSet SetType = "WorkingSet"
	AMRAPSet   SetType = "AMRAPSet"
)

// User domain structs
type User struct {
	ID             uuid.UUID                  `json:"id"`
	Username       string                     `json:"username"`
	CurrentProgram uuid.UUID                  `json:"current_program"` // UUID ref
	Programs       map[uuid.UUID]*UserProgram `json:"programs"`
	WorkoutHistory []Workout                  `json:"workout_history"`
	CreatedAt      time.Time                  `json:"created_at"`
}

type UserProgram struct {
	ID              uuid.UUID            `json:"id"`
	UserID          uuid.UUID            `json:"user_id"`
	ProgramID       uuid.UUID            `json:"program_id"`
	StartingWeights map[LiftName]float64 `json:"starting_weights"`
	CurrentWeights  map[LiftName]float64 `json:"current_weights"`
	CurrentDay      int                  `json:"current_day"`
	StartedAt       time.Time            `json:"started_at"`
}

type Workout struct {
	ID            uuid.UUID `json:"id"`
	UserProgramID uuid.UUID `json:"user_program_id"`
	Day           int       `json:"day"`
	Exercises     []Lift    `json:"exercises"`
	EnteredAt     time.Time `json:"entered_at"`
}

type Lift struct {
	ID       uuid.UUID `json:"id"`
	LiftName LiftName  `json:"lift_name"`
	Sets     []Set     `json:"sets"`
}

type Set struct {
	ID         uuid.UUID `json:"id"`
	Weight     float64   `json:"weight"`
	TargetReps int       `json:"target_reps"`
	ActualReps int       `json:"actual_reps"`
	Type       SetType   `json:"type"`
	Order      int       `json:"order"`
}

// Program template structs
type Program struct {
	ID               uuid.UUID         `json:"id"`
	Name             string            `json:"name"`
	Version          string            `json:"version"`
	Workouts         []WorkoutTemplate `json:"workouts"`
	ProgressionRules ProgressionRules  `json:"progression_rules"`
}

type WorkoutTemplate struct {
	Day   int            `json:"day"`
	Lifts []LiftTemplate `json:"lifts"`
}

type LiftTemplate struct {
	LiftName    LiftName      `json:"lift_name"`
	WarmupSets  []SetTemplate `json:"warmup_sets"`
	WorkingSets []SetTemplate `json:"working_sets"`
}

type SetTemplate struct {
	Reps             int     `json:"reps"`
	WeightPercentage float64 `json:"weight_percentage"`
	Type             SetType `json:"type"`
}

type ProgressionRules struct {
	IncreaseRules    map[LiftName]float64 `json:"increase_rules"`
	DeloadPercentage float64              `json:"deload_percentage"`
	DoubleThreshold  int                  `json:"double_threshold"`
}

// Validation methods
func (u *User) Validate() error {
	username := strings.TrimSpace(u.Username)
	if username == "" {
		return ErrUsernameEmpty
	}

	// Username must start with a letter and contain only letters, numbers, and dashes
	validUsername := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9-]*$`)
	if !validUsername.MatchString(username) {
		return ErrUsernameInvalid
	}

	return nil
}

func (s *Set) IsComplete() bool {
	return s.ActualReps > 0
}

// Helper function to generate UUID v7
func GenerateUUIDv7() uuid.UUID {
	return uuid.Must(uuid.NewV7())
}

// Custom errors
type ValidationError string

func (e ValidationError) Error() string {
	return string(e)
}

const (
	ErrUsernameEmpty   ValidationError = "username cannot be empty"
	ErrUsernameInvalid ValidationError = "username must start with a letter and contain only letters, numbers, and dashes"
)

