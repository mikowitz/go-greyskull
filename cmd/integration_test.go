package cmd

import (
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mikowitz/greyskull/models"
	"github.com/mikowitz/greyskull/repository"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserWorkflow_Integration(t *testing.T) {
	// Setup isolated test environment with unique temp directory per test
	tempDir := t.TempDir()
	originalConfigDir := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tempDir)
	defer func() {
		if originalConfigDir != "" {
			os.Setenv("XDG_CONFIG_HOME", originalConfigDir)
		} else {
			os.Unsetenv("XDG_CONFIG_HOME")
		}
	}()

	// Initialize repository
	repo, err := repository.NewJSONUserRepository()
	require.NoError(t, err)

	// Test 1: Initially no users should exist
	usernames, err := repo.List()
	require.NoError(t, err)
	assert.Empty(t, usernames)

	// Test 2: Create first user directly (avoid stdin mocking complexity)
	user1 := &models.User{
		ID:             uuid.New(),
		Username:       "TestUser",
		CurrentProgram: uuid.Nil,
		Programs:       make(map[uuid.UUID]*models.UserProgram),
		WorkoutHistory: []models.Workout{},
		CreatedAt:      time.Now(),
	}
	err = repo.Create(user1)
	require.NoError(t, err)

	// Test 3: Set as current user
	err = repo.SetCurrent("TestUser")
	require.NoError(t, err)

	// Test 4: Verify current user is set
	currentUser, err := repo.GetCurrent()
	require.NoError(t, err)
	assert.Equal(t, "TestUser", currentUser)

	// Test 5: Create second user with different casing
	user2 := &models.User{
		ID:             uuid.New(),
		Username:       "alice",
		CurrentProgram: uuid.Nil,
		Programs:       make(map[uuid.UUID]*models.UserProgram),
		WorkoutHistory: []models.Workout{},
		CreatedAt:      time.Now(),
	}
	err = repo.Create(user2)
	require.NoError(t, err)

	// Test 6: Test user listing functionality
	var buf bytes.Buffer
	listCmd.SetOut(&buf)
	err = listUsers(listCmd, []string{})
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Users:")
	assert.Contains(t, output, "TestUser")
	assert.Contains(t, output, "alice")
	assert.Contains(t, output, "* Current user: TestUser")

	// Test 7: Test user switching functionality
	buf.Reset()
	switchCmd.SetOut(&buf)
	err = switchUser(switchCmd, []string{"alice"})
	require.NoError(t, err)

	output = buf.String()
	assert.Contains(t, output, "Switched to user \"alice\"")

	// Test 8: Verify current user changed
	currentUser, err = repo.GetCurrent()
	require.NoError(t, err)
	assert.Equal(t, "alice", currentUser)

	// Test 9: Test case-insensitive switching
	buf.Reset()
	switchCmd.SetOut(&buf)
	err = switchUser(switchCmd, []string{"TESTUSER"})
	require.NoError(t, err)

	output = buf.String()
	assert.Contains(t, output, "Switched to user \"TestUser\"") // Original casing preserved

	// Test 10: Test error cases
	err = switchUser(switchCmd, []string{"NonExistent"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Test 11: Test duplicate user creation
	duplicateUser := &models.User{
		ID:             uuid.New(),
		Username:       "TestUser", // Same as existing user
		CurrentProgram: uuid.Nil,
		Programs:       make(map[uuid.UUID]*models.UserProgram),
		WorkoutHistory: []models.Workout{},
		CreatedAt:      time.Now(),
	}
	err = repo.Create(duplicateUser)
	assert.Error(t, err)
	assert.ErrorIs(t, err, repository.ErrUserAlreadyExists)

	// Test 12: Test case-insensitive duplicate detection
	duplicateUser2 := &models.User{
		ID:             uuid.New(),
		Username:       "testuser", // Different case but same user
		CurrentProgram: uuid.Nil,
		Programs:       make(map[uuid.UUID]*models.UserProgram),
		WorkoutHistory: []models.Workout{},
		CreatedAt:      time.Now(),
	}
	err = repo.Create(duplicateUser2)
	assert.Error(t, err)
	assert.ErrorIs(t, err, repository.ErrUserAlreadyExists)
}

func TestListUsers_EmptyRepository(t *testing.T) {
	// Setup isolated test environment with unique temp directory
	tempDir := t.TempDir()
	originalConfigDir := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tempDir)
	defer func() {
		if originalConfigDir != "" {
			os.Setenv("XDG_CONFIG_HOME", originalConfigDir)
		} else {
			os.Unsetenv("XDG_CONFIG_HOME")
		}
	}()

	var buf bytes.Buffer
	listCmd.SetOut(&buf)
	err := listUsers(listCmd, []string{})
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "No users found")
	assert.Contains(t, output, "Use 'greyskull user create'")
}

func TestValidateUsername_Comprehensive(t *testing.T) {
	tests := []struct {
		name     string
		username string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid simple username",
			username: "user",
			wantErr:  false,
		},
		{
			name:     "valid username with mixed case",
			username: "TestUser",
			wantErr:  false,
		},
		{
			name:     "valid username with numbers",
			username: "user123",
			wantErr:  false,
		},
		{
			name:     "valid username with dash",
			username: "test-user",
			wantErr:  false,
		},
		{
			name:     "valid username with underscore",
			username: "test_user",
			wantErr:  false,
		},
		{
			name:     "empty username",
			username: "",
			wantErr:  true,
			errMsg:   "username cannot be empty",
		},
		{
			name:     "username with forward slash",
			username: "user/name",
			wantErr:  true,
			errMsg:   "username contains invalid characters",
		},
		{
			name:     "username with backslash",
			username: "user\\name",
			wantErr:  true,
			errMsg:   "username contains invalid characters",
		},
		{
			name:     "username with colon",
			username: "user:name",
			wantErr:  true,
			errMsg:   "username contains invalid characters",
		},
		{
			name:     "username with asterisk",
			username: "user*name",
			wantErr:  true,
			errMsg:   "username contains invalid characters",
		},
		{
			name:     "username with question mark",
			username: "user?name",
			wantErr:  true,
			errMsg:   "username contains invalid characters",
		},
		{
			name:     "username with double quotes",
			username: "user\"name",
			wantErr:  true,
			errMsg:   "username contains invalid characters",
		},
		{
			name:     "username with less than",
			username: "user<name",
			wantErr:  true,
			errMsg:   "username contains invalid characters",
		},
		{
			name:     "username with greater than",
			username: "user>name",
			wantErr:  true,
			errMsg:   "username contains invalid characters",
		},
		{
			name:     "username with pipe",
			username: "user|name",
			wantErr:  true,
			errMsg:   "username contains invalid characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateUsername(tt.username)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUserCommandsHelp(t *testing.T) {
	tests := []struct {
		name        string
		cmd         *cobra.Command
		expectUsage string
		expectShort string
	}{
		{
			name:        "root command",
			cmd:         rootCmd,
			expectUsage: "greyskull",
			expectShort: "A command-line workout tracker for Greyskull LP",
		},
		{
			name:        "user command",
			cmd:         userCmd,
			expectUsage: "user",
			expectShort: "Manage users",
		},
		{
			name:        "user create command",
			cmd:         createCmd,
			expectUsage: "create",
			expectShort: "Create a new user",
		},
		{
			name:        "user switch command", 
			cmd:         switchCmd,
			expectUsage: "switch <username>",
			expectShort: "Switch to a different user",
		},
		{
			name:        "user list command",
			cmd:         listCmd,
			expectUsage: "list",
			expectShort: "List all users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectUsage, tt.cmd.Use)
			assert.Equal(t, tt.expectShort, tt.cmd.Short)
			assert.NotEmpty(t, tt.cmd.Long)
		})
	}
}