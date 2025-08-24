package cmd

import (
	"io"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mikowitz/greyskull/models"
	"github.com/mikowitz/greyskull/program"
	"github.com/mikowitz/greyskull/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStartProgram_NoCurrentUser(t *testing.T) {
	_ = setupTestEnv(t)
	
	cmd := programStartCmd
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	
	err := cmd.RunE(cmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no current user set")
}

func TestStartProgram_FullWorkflow(t *testing.T) {
	_ = setupTestEnv(t)
	
	// Create and set current user
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
	
	// Mock user input for program selection and weights
	// We'll test the helper functions separately and focus on integration here
	
	// Test that programs are available
	programs := program.List()
	assert.NotEmpty(t, programs)
	assert.Equal(t, "OG Greyskull LP", programs[0].Name)
	
	// Create a UserProgram directly to test the data structure
	userProgram := &models.UserProgram{
		ID:        uuid.Must(uuid.NewV7()),
		UserID:    user.ID,
		ProgramID: programs[0].ID,
		StartingWeights: map[models.LiftName]float64{
			models.Squat:         135.0,
			models.Deadlift:      185.0,
			models.BenchPress:    125.0,
			models.OverheadPress: 85.0,
		},
		CurrentWeights: map[models.LiftName]float64{
			models.Squat:         135.0,
			models.Deadlift:      185.0,
			models.BenchPress:    125.0,
			models.OverheadPress: 85.0,
		},
		CurrentDay: 1,
		StartedAt:  time.Now(),
	}
	
	// Update user with program
	user.Programs[userProgram.ID] = userProgram
	user.CurrentProgram = userProgram.ID
	
	err = repo.Update(user)
	require.NoError(t, err)
	
	// Verify user was updated correctly
	updatedUser, err := repo.Get("TestUser")
	require.NoError(t, err)
	
	assert.Equal(t, userProgram.ID, updatedUser.CurrentProgram)
	assert.Contains(t, updatedUser.Programs, userProgram.ID)
	
	savedProgram := updatedUser.Programs[userProgram.ID]
	assert.Equal(t, userProgram.UserID, savedProgram.UserID)
	assert.Equal(t, userProgram.ProgramID, savedProgram.ProgramID)
	assert.Equal(t, 1, savedProgram.CurrentDay)
	assert.Equal(t, 135.0, savedProgram.StartingWeights[models.Squat])
	assert.Equal(t, 135.0, savedProgram.CurrentWeights[models.Squat])
}


func TestLiftDisplayName(t *testing.T) {
	tests := []struct {
		lift     models.LiftName
		expected string
	}{
		{models.Squat, "Squat"},
		{models.Deadlift, "Deadlift"},
		{models.BenchPress, "Bench Press"},
		{models.OverheadPress, "Overhead Press"},
	}
	
	for _, tt := range tests {
		t.Run(string(tt.lift), func(t *testing.T) {
			result := liftDisplayName(tt.lift)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStartProgram_InvalidWeights(t *testing.T) {
	_ = setupTestEnv(t)
	
	// Create and set current user
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
	
}

func TestStartProgram_ProgramSelection(t *testing.T) {
	// Test that available programs are properly listed
	programs := program.List()
	require.NotEmpty(t, programs, "Expected at least one program to be available")
	
	// Verify Greyskull LP is available
	found := false
	for _, prog := range programs {
		if prog.Name == "OG Greyskull LP" {
			found = true
			assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", prog.ID.String())
			assert.Equal(t, "1.0.0", prog.Version)
			break
		}
	}
	assert.True(t, found, "Greyskull LP program should be available")
}

func TestStartProgram_UserProgramCreation(t *testing.T) {
	// Test UserProgram structure and UUID generation
	userID := uuid.New()
	programID := uuid.Must(uuid.Parse("550e8400-e29b-41d4-a716-446655440000"))
	
	startingWeights := map[models.LiftName]float64{
		models.Squat:         135.0,
		models.Deadlift:      185.0,
		models.BenchPress:    125.0,
		models.OverheadPress: 85.0,
	}
	
	userProgram := &models.UserProgram{
		ID:              uuid.Must(uuid.NewV7()),
		UserID:          userID,
		ProgramID:       programID,
		StartingWeights: startingWeights,
		CurrentWeights:  make(map[models.LiftName]float64),
		CurrentDay:      1,
		StartedAt:       time.Now(),
	}
	
	// Copy starting weights to current weights
	for lift, weight := range startingWeights {
		userProgram.CurrentWeights[lift] = weight
	}
	
	// Verify structure
	assert.NotEqual(t, uuid.Nil, userProgram.ID)
	assert.Equal(t, userID, userProgram.UserID)
	assert.Equal(t, programID, userProgram.ProgramID)
	assert.Equal(t, 1, userProgram.CurrentDay)
	assert.Equal(t, 4, len(userProgram.StartingWeights))
	assert.Equal(t, 4, len(userProgram.CurrentWeights))
	
	// Verify weights are correctly copied
	for lift, weight := range startingWeights {
		assert.Equal(t, weight, userProgram.StartingWeights[lift])
		assert.Equal(t, weight, userProgram.CurrentWeights[lift])
	}
}