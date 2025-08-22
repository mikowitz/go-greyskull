package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mikowitz/greyskull/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserCreate(t *testing.T) {
	// Setup temp environment
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

	tests := []struct {
		name           string
		input          string
		expectedOutput string
		expectedError  string
		shouldSucceed  bool
	}{
		{
			name:           "valid username",
			input:          "TestUser\n",
			expectedOutput: "User \"TestUser\" created successfully and set as current user.",
			shouldSucceed:  true,
		},
		{
			name:          "empty username",
			input:         "\n",
			expectedError: "username cannot be empty",
			shouldSucceed: false,
		},
		{
			name:          "username with invalid chars",
			input:         "user/name\n",
			expectedError: "username contains invalid characters",
			shouldSucceed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up any existing state
			usersDir := filepath.Join(tempDir, "greyskull", "users")
			os.RemoveAll(usersDir)

			// Mock stdin
			oldStdin := os.Stdin
			r, w, _ := os.Pipe()
			os.Stdin = r

			// Setup output capture
			var buf bytes.Buffer
			createCmd.SetOut(&buf)
			createCmd.SetErr(&buf)

			go func() {
				defer w.Close()
				w.Write([]byte(tt.input))
			}()

			// Execute command
			err := createCmd.RunE(createCmd, []string{})

			// Restore stdin
			os.Stdin = oldStdin

			// Check results
			output := buf.String()
			if tt.shouldSucceed {
				assert.NoError(t, err)
				assert.Contains(t, output, tt.expectedOutput)

				// Verify user was created and set as current
				repo, err := repository.NewJSONUserRepository()
				require.NoError(t, err)
				_ = repo

				currentUser, err := repo.GetCurrent()
				assert.NoError(t, err)
				assert.Equal(t, strings.TrimSpace(strings.Split(tt.input, "\n")[0]), currentUser)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			}
		})
	}
}

func TestUserCreateDuplicate(t *testing.T) {
	// Setup temp environment
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

	// Create first user
	repo, err := repository.NewJSONUserRepository()
	require.NoError(t, err)
	_ = repo

	// Mock first user creation
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r

	var buf bytes.Buffer
	createCmd.SetOut(&buf)
	createCmd.SetErr(&buf)

	go func() {
		defer w.Close()
		w.Write([]byte("TestUser\n"))
	}()

	err = createCmd.RunE(createCmd, []string{})
	os.Stdin = oldStdin
	require.NoError(t, err)

	// Try to create duplicate (case-insensitive)
	r2, w2, _ := os.Pipe()
	os.Stdin = r2

	var buf2 bytes.Buffer
	createCmd.SetOut(&buf2)
	createCmd.SetErr(&buf2)

	go func() {
		defer w2.Close()
		w2.Write([]byte("testuser\n"))
	}()

	err = createCmd.RunE(createCmd, []string{})
	os.Stdin = oldStdin

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestUserSwitch(t *testing.T) {
	// Setup temp environment and users
	tempDir := t.TempDir()
	setupTestUsers(t, tempDir, []string{"Alice", "Bob"})

	tests := []struct {
		name           string
		username       string
		expectedOutput string
		expectedError  string
		shouldSucceed  bool
	}{
		{
			name:           "valid user exact case",
			username:       "Alice",
			expectedOutput: "Switched to user \"Alice\".",
			shouldSucceed:  true,
		},
		{
			name:           "valid user case insensitive",
			username:       "bob",
			expectedOutput: "Switched to user \"Bob\".",
			shouldSucceed:  true,
		},
		{
			name:          "non-existent user",
			username:      "Charlie",
			expectedError: "user \"Charlie\" not found",
			shouldSucceed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			switchCmd.SetOut(&buf)
			switchCmd.SetErr(&buf)

			err := switchCmd.RunE(switchCmd, []string{tt.username})
			output := buf.String()

			if tt.shouldSucceed {
				assert.NoError(t, err)
				assert.Contains(t, output, tt.expectedOutput)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			}
		})
	}
}

func TestUserList(t *testing.T) {
	// Setup temp environment
	tempDir := t.TempDir()

	tests := []struct {
		name           string
		users          []string
		currentUser    string
		expectedOutput []string
	}{
		{
			name:           "no users",
			users:          []string{},
			expectedOutput: []string{"No users found", "Use 'greyskull user create'"},
		},
		{
			name:        "multiple users with current",
			users:       []string{"Alice", "bob", "Charlie"},
			currentUser: "Alice",
			expectedOutput: []string{
				"Users:",
				"* Alice",
				"  bob", 
				"  Charlie",
				"* Current user: Alice",
			},
		},
		{
			name:  "multiple users no current",
			users: []string{"Alice", "Bob"},
			expectedOutput: []string{
				"Users:",
				"  Alice",
				"  Bob",
				"No current user set",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupTestUsers(t, tempDir, tt.users)

			// Set current user if specified
			if tt.currentUser != "" {
				originalConfigDir := os.Getenv("XDG_CONFIG_HOME")
				os.Setenv("XDG_CONFIG_HOME", tempDir)
				defer func() {
					if originalConfigDir != "" {
						os.Setenv("XDG_CONFIG_HOME", originalConfigDir)
					} else {
						os.Unsetenv("XDG_CONFIG_HOME")
					}
				}()

				repo, err := repository.NewJSONUserRepository()
				require.NoError(t, err)
				err = repo.SetCurrent(tt.currentUser)
				require.NoError(t, err)
			}

			var buf bytes.Buffer
			listCmd.SetOut(&buf)
			listCmd.SetErr(&buf)

			err := listCmd.RunE(listCmd, []string{})
			assert.NoError(t, err)

			output := buf.String()
			for _, expected := range tt.expectedOutput {
				assert.Contains(t, output, expected)
			}
		})
	}
}

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid username",
			username: "TestUser",
			wantErr:  false,
		},
		{
			name:     "username with numbers",
			username: "User123",
			wantErr:  false,
		},
		{
			name:     "username with dashes",
			username: "test-user",
			wantErr:  false,
		},
		{
			name:     "empty username",
			username: "",
			wantErr:  true,
			errMsg:   "username cannot be empty",
		},
		{
			name:     "username with slash",
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
			name:     "username with quotes",
			username: "user\"name",
			wantErr:  true,
			errMsg:   "username contains invalid characters",
		},
		{
			name:     "username with angle brackets",
			username: "user<name>",
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

// Helper function to setup test users
func setupTestUsers(t *testing.T, tempDir string, usernames []string) {
	originalConfigDir := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tempDir)
	defer func() {
		if originalConfigDir != "" {
			os.Setenv("XDG_CONFIG_HOME", originalConfigDir)
		} else {
			os.Unsetenv("XDG_CONFIG_HOME")
		}
	}()

	repo, err := repository.NewJSONUserRepository()
	require.NoError(t, err)
	_ = repo

	for _, username := range usernames {
		// Mock stdin for each user creation
		oldStdin := os.Stdin
		r, w, _ := os.Pipe()
		os.Stdin = r

		var buf bytes.Buffer
		createCmd.SetOut(&buf)
		createCmd.SetErr(&buf)

		go func(user string) {
			defer w.Close()
			w.Write([]byte(fmt.Sprintf("%s\n", user)))
		}(username)

		err := createCmd.RunE(createCmd, []string{})
		os.Stdin = oldStdin
		require.NoError(t, err)
	}
}