package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGenerateUUIDv7(t *testing.T) {
	t.Run("generates valid UUID v7", func(t *testing.T) {
		id := GenerateUUIDv7()
		assert.NotEqual(t, uuid.Nil, id)
		
		// Check that it's version 7
		assert.Equal(t, uuid.Version(7), id.Version())
	})

	t.Run("generates unique UUIDs", func(t *testing.T) {
		id1 := GenerateUUIDv7()
		id2 := GenerateUUIDv7()
		assert.NotEqual(t, id1, id2)
	})
}

func TestUserValidate(t *testing.T) {
	tests := []struct {
		name     string
		username string
		wantErr  bool
		errType  error
	}{
		{
			name:     "valid username with letters only",
			username: "testuser",
			wantErr:  false,
		},
		{
			name:     "valid username with letters and numbers",
			username: "testuser123",
			wantErr:  false,
		},
		{
			name:     "valid username with letters, numbers, and dashes",
			username: "test-user-123",
			wantErr:  false,
		},
		{
			name:     "valid single letter username",
			username: "a",
			wantErr:  false,
		},
		{
			name:     "empty username",
			username: "",
			wantErr:  true,
			errType:  ErrUsernameEmpty,
		},
		{
			name:     "whitespace only username",
			username: "   ",
			wantErr:  true,
			errType:  ErrUsernameEmpty,
		},
		{
			name:     "username starting with number",
			username: "123user",
			wantErr:  true,
			errType:  ErrUsernameInvalid,
		},
		{
			name:     "username starting with dash",
			username: "-testuser",
			wantErr:  true,
			errType:  ErrUsernameInvalid,
		},
		{
			name:     "username with spaces",
			username: "test user",
			wantErr:  true,
			errType:  ErrUsernameInvalid,
		},
		{
			name:     "username with special characters",
			username: "test@user",
			wantErr:  true,
			errType:  ErrUsernameInvalid,
		},
		{
			name:     "username with underscore",
			username: "test_user",
			wantErr:  true,
			errType:  ErrUsernameInvalid,
		},
		{
			name:     "username with period",
			username: "test.user",
			wantErr:  true,
			errType:  ErrUsernameInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{
				ID:       GenerateUUIDv7(),
				Username: tt.username,
				Programs: make(map[uuid.UUID]*UserProgram),
				CreatedAt: time.Now(),
			}

			err := user.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.errType, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSetIsComplete(t *testing.T) {
	tests := []struct {
		name       string
		actualReps int
		want       bool
	}{
		{
			name:       "completed set with positive reps",
			actualReps: 5,
			want:       true,
		},
		{
			name:       "incomplete set with zero reps",
			actualReps: 0,
			want:       false,
		},
		{
			name:       "incomplete set with negative reps",
			actualReps: -1,
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			set := &Set{
				ID:         GenerateUUIDv7(),
				Weight:     135.0,
				TargetReps: 5,
				ActualReps: tt.actualReps,
				Type:       WorkingSet,
				Order:      1,
			}

			assert.Equal(t, tt.want, set.IsComplete())
		})
	}
}

func TestUserStructInitialization(t *testing.T) {
	t.Run("user struct with all fields", func(t *testing.T) {
		userID := GenerateUUIDv7()
		programID := GenerateUUIDv7()
		createdAt := time.Now()

		user := &User{
			ID:             userID,
			Username:       "testuser",
			CurrentProgram: programID,
			Programs:       make(map[uuid.UUID]*UserProgram),
			WorkoutHistory: []Workout{},
			CreatedAt:      createdAt,
		}

		assert.Equal(t, userID, user.ID)
		assert.Equal(t, "testuser", user.Username)
		assert.Equal(t, programID, user.CurrentProgram)
		assert.NotNil(t, user.Programs)
		assert.NotNil(t, user.WorkoutHistory)
		assert.Equal(t, createdAt, user.CreatedAt)
	})
}

func TestUserProgramStructInitialization(t *testing.T) {
	t.Run("user program struct with all fields", func(t *testing.T) {
		id := GenerateUUIDv7()
		userID := GenerateUUIDv7()
		programID := GenerateUUIDv7()
		startedAt := time.Now()

		startingWeights := map[LiftName]float64{
			Squat:         135.0,
			Deadlift:      185.0,
			BenchPress:    115.0,
			OverheadPress: 75.0,
		}

		currentWeights := map[LiftName]float64{
			Squat:         140.0,
			Deadlift:      190.0,
			BenchPress:    117.5,
			OverheadPress: 77.5,
		}

		userProgram := &UserProgram{
			ID:              id,
			UserID:          userID,
			ProgramID:       programID,
			StartingWeights: startingWeights,
			CurrentWeights:  currentWeights,
			CurrentDay:      3,
			StartedAt:       startedAt,
		}

		assert.Equal(t, id, userProgram.ID)
		assert.Equal(t, userID, userProgram.UserID)
		assert.Equal(t, programID, userProgram.ProgramID)
		assert.Equal(t, startingWeights, userProgram.StartingWeights)
		assert.Equal(t, currentWeights, userProgram.CurrentWeights)
		assert.Equal(t, 3, userProgram.CurrentDay)
		assert.Equal(t, startedAt, userProgram.StartedAt)
	})
}

func TestWorkoutStructInitialization(t *testing.T) {
	t.Run("workout struct with all fields", func(t *testing.T) {
		workoutID := GenerateUUIDv7()
		userProgramID := GenerateUUIDv7()
		enteredAt := time.Now()

		lift := Lift{
			ID:       GenerateUUIDv7(),
			LiftName: Squat,
			Sets:     []Set{},
		}

		workout := &Workout{
			ID:            workoutID,
			UserProgramID: userProgramID,
			Day:           1,
			Exercises:     []Lift{lift},
			EnteredAt:     enteredAt,
		}

		assert.Equal(t, workoutID, workout.ID)
		assert.Equal(t, userProgramID, workout.UserProgramID)
		assert.Equal(t, 1, workout.Day)
		assert.Len(t, workout.Exercises, 1)
		assert.Equal(t, enteredAt, workout.EnteredAt)
	})
}

func TestLiftStructInitialization(t *testing.T) {
	t.Run("lift struct with all fields", func(t *testing.T) {
		liftID := GenerateUUIDv7()

		set := Set{
			ID:         GenerateUUIDv7(),
			Weight:     135.0,
			TargetReps: 5,
			ActualReps: 5,
			Type:       WorkingSet,
			Order:      1,
		}

		lift := &Lift{
			ID:       liftID,
			LiftName: Squat,
			Sets:     []Set{set},
		}

		assert.Equal(t, liftID, lift.ID)
		assert.Equal(t, Squat, lift.LiftName)
		assert.Len(t, lift.Sets, 1)
	})
}

func TestSetStructInitialization(t *testing.T) {
	t.Run("set struct with all fields", func(t *testing.T) {
		setID := GenerateUUIDv7()

		set := &Set{
			ID:         setID,
			Weight:     135.0,
			TargetReps: 5,
			ActualReps: 5,
			Type:       WorkingSet,
			Order:      1,
		}

		assert.Equal(t, setID, set.ID)
		assert.Equal(t, 135.0, set.Weight)
		assert.Equal(t, 5, set.TargetReps)
		assert.Equal(t, 5, set.ActualReps)
		assert.Equal(t, WorkingSet, set.Type)
		assert.Equal(t, 1, set.Order)
	})
}

func TestProgramStructInitialization(t *testing.T) {
	t.Run("program struct with all fields", func(t *testing.T) {
		programID := GenerateUUIDv7()

		progressionRules := ProgressionRules{
			IncreaseRules: map[LiftName]float64{
				OverheadPress: 2.5,
				BenchPress:    2.5,
				Squat:         5.0,
				Deadlift:      5.0,
			},
			DeloadPercentage: 0.9,
			DoubleThreshold:  10,
		}

		workoutTemplate := WorkoutTemplate{
			Day: 1,
			Lifts: []LiftTemplate{
				{
					LiftName: OverheadPress,
					WarmupSets: []SetTemplate{
						{Reps: 5, WeightPercentage: 0.0, Type: WarmupSet},
					},
					WorkingSets: []SetTemplate{
						{Reps: 5, WeightPercentage: 1.0, Type: WorkingSet},
					},
				},
			},
		}

		program := &Program{
			ID:               programID,
			Name:             "OG Greyskull LP",
			Version:          "1.0.0",
			Workouts:         []WorkoutTemplate{workoutTemplate},
			ProgressionRules: progressionRules,
		}

		assert.Equal(t, programID, program.ID)
		assert.Equal(t, "OG Greyskull LP", program.Name)
		assert.Equal(t, "1.0.0", program.Version)
		assert.Len(t, program.Workouts, 1)
		assert.Equal(t, progressionRules, program.ProgressionRules)
	})
}

func TestLiftNameConstants(t *testing.T) {
	assert.Equal(t, LiftName("Squat"), Squat)
	assert.Equal(t, LiftName("Deadlift"), Deadlift)
	assert.Equal(t, LiftName("BenchPress"), BenchPress)
	assert.Equal(t, LiftName("OverheadPress"), OverheadPress)
}

func TestSetTypeConstants(t *testing.T) {
	assert.Equal(t, SetType("WarmupSet"), WarmupSet)
	assert.Equal(t, SetType("WorkingSet"), WorkingSet)
	assert.Equal(t, SetType("AMRAPSet"), AMRAPSet)
}

func TestValidationError(t *testing.T) {
	t.Run("validation errors implement error interface", func(t *testing.T) {
		err1 := ErrUsernameEmpty
		assert.Error(t, err1)
		assert.Equal(t, "username cannot be empty", err1.Error())

		err2 := ErrUsernameInvalid
		assert.Error(t, err2)
		assert.Equal(t, "username must start with a letter and contain only letters, numbers, and dashes", err2.Error())
	})
}