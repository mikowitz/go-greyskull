package repository

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/mikowitz/greyskull/models"
)

// JSONUserRepository implements UserRepository using JSON files for persistence
type JSONUserRepository struct {
	configDir   string
	usersDir    string
	currentFile string
	mutex       sync.Mutex
}

// NewJSONUserRepository creates a new JSONUserRepository instance
func NewJSONUserRepository() (UserRepository, error) {
	// Check for XDG_CONFIG_HOME first (for Linux/testing), fallback to OS default
	var configDir string
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		configDir = xdgConfig
	} else {
		var err error
		configDir, err = os.UserConfigDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user config directory: %w", err)
		}
	}

	greyskullDir := filepath.Join(configDir, "greyskull")
	usersDir := filepath.Join(greyskullDir, "users")
	currentFile := filepath.Join(greyskullDir, "current_user.txt")

	// Create directory structure
	if err := os.MkdirAll(usersDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create users directory: %w", err)
	}

	return &JSONUserRepository{
		configDir:   greyskullDir,
		usersDir:    usersDir,
		currentFile: currentFile,
	}, nil
}

// Create creates a new user
func (r *JSONUserRepository) Create(user *models.User) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Check if user already exists (case-insensitive)
	if r.userExists(user.Username) {
		return ErrUserAlreadyExists
	}

	// Save user to file
	filename := r.getUserFilename(user.Username)
	return r.saveUserToFile(user, filename)
}

// Get retrieves a user by username (case-insensitive)
func (r *JSONUserRepository) Get(username string) (*models.User, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	filename := r.findUserFile(username)
	if filename == "" {
		return nil, ErrUserNotFound
	}

	return r.loadUserFromFile(filename)
}

// Update updates an existing user
func (r *JSONUserRepository) Update(user *models.User) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	filename := r.findUserFile(user.Username)
	if filename == "" {
		return ErrUserNotFound
	}

	return r.saveUserToFile(user, filename)
}

// List returns all usernames in their original casing
func (r *JSONUserRepository) List() ([]string, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	entries, err := os.ReadDir(r.usersDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read users directory: %w", err)
	}

	var usernames []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			// Load user to get original username casing
			user, err := r.loadUserFromFile(filepath.Join(r.usersDir, entry.Name()))
			if err != nil {
				continue // Skip corrupted files
			}
			usernames = append(usernames, user.Username)
		}
	}

	return usernames, nil
}

// GetCurrent returns the current active username
func (r *JSONUserRepository) GetCurrent() (string, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	data, err := os.ReadFile(r.currentFile)
	if err != nil {
		if os.IsNotExist(err) {
			return "", ErrNoCurrentUser
		}
		return "", fmt.Errorf("failed to read current user file: %w", err)
	}

	username := strings.TrimSpace(string(data))
	if username == "" {
		return "", ErrNoCurrentUser
	}

	// Verify user still exists
	if !r.userExists(username) {
		return "", ErrNoCurrentUser
	}

	return username, nil
}

// SetCurrent sets the current active user
func (r *JSONUserRepository) SetCurrent(username string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Find user file (case-insensitive)
	filename := r.findUserFile(username)
	if filename == "" {
		return ErrUserNotFound
	}

	// Load user to get original username casing
	user, err := r.loadUserFromFile(filename)
	if err != nil {
		return err
	}

	// Save the original username casing
	err = os.WriteFile(r.currentFile, []byte(user.Username), 0644)
	if err != nil {
		return fmt.Errorf("failed to write current user file: %w", err)
	}

	return nil
}

// Helper methods

// getUserFilename returns the filename for a user (lowercase)
func (r *JSONUserRepository) getUserFilename(username string) string {
	return filepath.Join(r.usersDir, strings.ToLower(username)+".json")
}

// userExists checks if a user exists (case-insensitive)
func (r *JSONUserRepository) userExists(username string) bool {
	filename := r.getUserFilename(username)
	_, err := os.Stat(filename)
	return err == nil
}

// findUserFile finds the user file for a username (case-insensitive)
func (r *JSONUserRepository) findUserFile(username string) string {
	filename := r.getUserFilename(username)
	if _, err := os.Stat(filename); err == nil {
		return filename
	}
	return ""
}

// saveUserToFile saves a user to a JSON file
func (r *JSONUserRepository) saveUserToFile(user *models.User, filename string) error {
	data, err := json.MarshalIndent(user, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal user data: %w", err)
	}

	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write user file: %w", err)
	}

	return nil
}

// loadUserFromFile loads a user from a JSON file
func (r *JSONUserRepository) loadUserFromFile(filename string) (*models.User, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read user file: %w", err)
	}

	var user models.User
	err = json.Unmarshal(data, &user)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal user data: %w", err)
	}

	return &user, nil
}