package services

import (
	"fmt"

	"github.com/mikowitz/greyskull/repository"
)

// CommandContext encapsulates all dependencies needed by CLI commands
// This provides a clean way to inject dependencies and makes testing easier
type CommandContext struct {
	// UserRepo provides direct access to user repository operations
	UserRepo repository.UserRepository
	
	// UserService provides high-level user operations (built on top of UserRepo)
	UserService *UserService
}

// NewCommandContext creates a new CommandContext with the specified repository factory
// This allows for dependency injection and better testability
func NewCommandContext(factory RepositoryFactory) (*CommandContext, error) {
	if factory == nil {
		return nil, fmt.Errorf("repository factory cannot be nil")
	}
	
	// Create the user repository using the factory
	userRepo, err := factory.NewUserRepository()
	if err != nil {
		return nil, fmt.Errorf("failed to create user repository: %w", err)
	}
	
	// Create the user service with the repository
	userService := NewUserService(userRepo, nil)
	
	return &CommandContext{
		UserRepo:    userRepo,
		UserService: userService,
	}, nil
}

// NewCommandContextWithDefaults creates a CommandContext using the default repository factory
// This is a convenience method for commands that don't need custom dependency injection
func NewCommandContextWithDefaults() (*CommandContext, error) {
	return NewCommandContext(GetDefaultRepositoryFactory())
}