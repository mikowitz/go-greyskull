package repository

import (
	"errors"

	"github.com/mikowitz/greyskull/models"
)

// Sentinel errors for repository operations
var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrNoCurrentUser     = errors.New("no current user set")
)

// UserRepository defines the interface for user persistence operations
type UserRepository interface {
	// Create creates a new user. Returns ErrUserAlreadyExists if username already exists.
	Create(user *models.User) error

	// Get retrieves a user by username (case-insensitive). Returns ErrUserNotFound if user doesn't exist.
	Get(username string) (*models.User, error)

	// Update updates an existing user. Returns ErrUserNotFound if user doesn't exist.
	Update(user *models.User) error

	// List returns all usernames in their original casing.
	List() ([]string, error)

	// GetCurrent returns the current active username. Returns ErrNoCurrentUser if none is set.
	GetCurrent() (string, error)

	// SetCurrent sets the current active user. Returns ErrUserNotFound if user doesn't exist.
	SetCurrent(username string) error
}