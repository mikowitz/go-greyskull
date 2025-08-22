package repository

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mikowitz/greyskull/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewJSONUserRepository(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalConfigDir := os.Getenv("XDG_CONFIG_HOME")
	if originalConfigDir == "" {
		originalConfigDir = os.Getenv("HOME")
	}

	// Set temporary config directory
	os.Setenv("XDG_CONFIG_HOME", tempDir)
	defer func() {
		if originalConfigDir != "" {
			os.Setenv("XDG_CONFIG_HOME", originalConfigDir)
		} else {
			os.Unsetenv("XDG_CONFIG_HOME")
		}
	}()

	repo, err := NewJSONUserRepository()
	require.NoError(t, err)
	require.NotNil(t, repo)

	// Verify directory structure was created
	jsonRepo := repo.(*JSONUserRepository)
	assert.DirExists(t, jsonRepo.usersDir)
}

func TestJSONUserRepository_Create(t *testing.T) {
	repo := setupTestRepository(t)

	user := createTestUser("TestUser")

	// Test successful creation
	err := repo.Create(user)
	assert.NoError(t, err)

	// Test duplicate creation
	err = repo.Create(user)
	assert.ErrorIs(t, err, ErrUserAlreadyExists)

	// Test case-insensitive duplicate detection
	userLower := createTestUser("testuser")
	err = repo.Create(userLower)
	assert.ErrorIs(t, err, ErrUserAlreadyExists)
}

func TestJSONUserRepository_Get(t *testing.T) {
	repo := setupTestRepository(t)

	originalUser := createTestUser("TestUser")
	err := repo.Create(originalUser)
	require.NoError(t, err)

	tests := []struct {
		name     string
		username string
		wantErr  error
	}{
		{
			name:     "exact case match",
			username: "TestUser",
			wantErr:  nil,
		},
		{
			name:     "lowercase match",
			username: "testuser",
			wantErr:  nil,
		},
		{
			name:     "uppercase match",
			username: "TESTUSER",
			wantErr:  nil,
		},
		{
			name:     "mixed case match",
			username: "tEsTuSeR",
			wantErr:  nil,
		},
		{
			name:     "non-existent user",
			username: "NonExistent",
			wantErr:  ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := repo.Get(tt.username)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, "TestUser", user.Username) // Original casing preserved
				assert.Equal(t, originalUser.ID, user.ID)
			}
		})
	}
}

func TestJSONUserRepository_Update(t *testing.T) {
	repo := setupTestRepository(t)

	user := createTestUser("TestUser")
	err := repo.Create(user)
	require.NoError(t, err)

	// Update user data
	user.Username = "TestUser" // Keep same username but modify other fields
	program := &models.UserProgram{
		ID:       uuid.New(),
		UserID:   user.ID,
		ProgramID: uuid.New(),
		StartingWeights: map[models.LiftName]float64{
			models.Squat: 135.0,
		},
		CurrentWeights: map[models.LiftName]float64{
			models.Squat: 140.0,
		},
		CurrentDay: 2,
		StartedAt:  time.Now(),
	}
	user.Programs[program.ID] = program

	// Test successful update
	err = repo.Update(user)
	assert.NoError(t, err)

	// Verify update was saved
	retrievedUser, err := repo.Get("TestUser")
	require.NoError(t, err)
	assert.Len(t, retrievedUser.Programs, 1)
	assert.Equal(t, 2, retrievedUser.Programs[program.ID].CurrentDay)

	// Test update non-existent user
	nonExistentUser := createTestUser("NonExistent")
	err = repo.Update(nonExistentUser)
	assert.ErrorIs(t, err, ErrUserNotFound)
}

func TestJSONUserRepository_List(t *testing.T) {
	repo := setupTestRepository(t)

	// Test empty repository
	usernames, err := repo.List()
	assert.NoError(t, err)
	assert.Empty(t, usernames)

	// Create multiple users with different casings
	users := []string{"Alice", "bob", "Charlie", "DAVE"}
	for _, username := range users {
		user := createTestUser(username)
		err := repo.Create(user)
		require.NoError(t, err)
	}

	// Test listing users
	usernames, err = repo.List()
	assert.NoError(t, err)
	assert.Len(t, usernames, 4)

	// Verify original casing is preserved
	expectedUsernames := []string{"Alice", "bob", "Charlie", "DAVE"}
	for _, expected := range expectedUsernames {
		assert.Contains(t, usernames, expected)
	}
}

func TestJSONUserRepository_CurrentUser(t *testing.T) {
	repo := setupTestRepository(t)

	// Test no current user
	current, err := repo.GetCurrent()
	assert.ErrorIs(t, err, ErrNoCurrentUser)
	assert.Empty(t, current)

	// Create a user
	user := createTestUser("TestUser")
	err = repo.Create(user)
	require.NoError(t, err)

	// Test setting current user
	err = repo.SetCurrent("testuser") // Case-insensitive
	assert.NoError(t, err)

	// Test getting current user
	current, err = repo.GetCurrent()
	assert.NoError(t, err)
	assert.Equal(t, "TestUser", current) // Original casing preserved

	// Test setting non-existent user as current
	err = repo.SetCurrent("NonExistent")
	assert.ErrorIs(t, err, ErrUserNotFound)

	// Verify current user unchanged
	current, err = repo.GetCurrent()
	assert.NoError(t, err)
	assert.Equal(t, "TestUser", current)
}

func TestJSONUserRepository_ConcurrentAccess(t *testing.T) {
	repo := setupTestRepository(t)

	const numGoroutines = 10
	const numUsers = 5

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*numUsers)

	// Test concurrent user creation
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < numUsers; j++ {
				user := createTestUser(fmt.Sprintf("User_%d_%d", goroutineID, j))
				err := repo.Create(user)
				if err != nil {
					errors <- err
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		assert.NoError(t, err)
	}

	// Verify all users were created
	usernames, err := repo.List()
	assert.NoError(t, err)
	assert.Len(t, usernames, numGoroutines*numUsers)
}

func TestJSONUserRepository_CaseInsensitiveHandling(t *testing.T) {
	repo := setupTestRepository(t)

	// Create user with mixed case
	originalUser := createTestUser("MixedCaseUser")
	err := repo.Create(originalUser)
	require.NoError(t, err)

	testCases := []string{
		"mixedcaseuser",
		"MIXEDCASEUSER", 
		"MixedCaseUser",
		"mIxEdCaSeUsEr",
	}

	for _, testCase := range testCases {
		t.Run("access_with_"+testCase, func(t *testing.T) {
			// Test Get
			user, err := repo.Get(testCase)
			assert.NoError(t, err)
			assert.Equal(t, "MixedCaseUser", user.Username)

			// Test SetCurrent
			err = repo.SetCurrent(testCase)
			assert.NoError(t, err)

			current, err := repo.GetCurrent()
			assert.NoError(t, err)
			assert.Equal(t, "MixedCaseUser", current)
		})
	}
}

// Helper functions

func setupTestRepository(t *testing.T) UserRepository {
	// Create temporary directory
	tempDir := t.TempDir()
	
	// Create a mock repository with temp directory
	repo := &JSONUserRepository{
		configDir:   tempDir,
		usersDir:    filepath.Join(tempDir, "users"),
		currentFile: filepath.Join(tempDir, "current_user.txt"),
	}

	// Create users directory
	err := os.MkdirAll(repo.usersDir, 0755)
	require.NoError(t, err)

	return repo
}

func createTestUser(username string) *models.User {
	return &models.User{
		ID:             uuid.New(),
		Username:       username,
		CurrentProgram: uuid.Nil,
		Programs:       make(map[uuid.UUID]*models.UserProgram),
		WorkoutHistory: []models.Workout{},
		CreatedAt:      time.Now(),
	}
}