package services

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/mikowitz/greyskull/models"
	"github.com/mikowitz/greyskull/program"
	"github.com/mikowitz/greyskull/repository"
)

// ProgramService defines the interface for loading programs
type ProgramService interface {
	GetByID(id string) (*models.Program, error)
}

// UserService encapsulates common user operations used across CLI commands
type UserService struct {
	repo           repository.UserRepository
	programService ProgramService
}

// NewUserService creates a new UserService instance
// If programService is nil, it will use program.GetByID directly
func NewUserService(repo repository.UserRepository, programService ProgramService) *UserService {
	if repo == nil {
		panic("repository cannot be nil")
	}
	return &UserService{
		repo:           repo,
		programService: programService,
	}
}

// RequireCurrentUser loads the current user, handling all common error cases
// This consolidates the repository setup and user loading logic used by all commands
func (s *UserService) RequireCurrentUser() (*models.User, error) {
	// Get current username
	currentUsername, err := s.repo.GetCurrent()
	if err != nil {
		if err == repository.ErrNoCurrentUser {
			return nil, fmt.Errorf("no current user set. Use 'greyskull user create' or 'greyskull user switch' first")
		}
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}

	// Load user
	user, err := s.repo.Get(currentUsername)
	if err != nil {
		return nil, fmt.Errorf("failed to load current user: %w", err)
	}

	return user, nil
}

// GetCurrentUserWithProgram loads the current user, their active UserProgram, and Program
// This consolidates the complete user + program loading logic used by workout commands
func (s *UserService) GetCurrentUserWithProgram() (*models.User, *models.UserProgram, *models.Program, error) {
	// Load current user first
	user, err := s.RequireCurrentUser()
	if err != nil {
		return nil, nil, nil, err
	}

	// Check if user has a current program
	if user.CurrentProgram == uuid.Nil {
		return nil, nil, nil, fmt.Errorf("no active program. Use 'greyskull program start' to begin a program")
	}

	// Get UserProgram
	userProgram, exists := user.Programs[user.CurrentProgram]
	if !exists {
		return nil, nil, nil, fmt.Errorf("current program not found in user programs")
	}

	// Load Program definition
	var programDef *models.Program
	if s.programService != nil {
		programDef, err = s.programService.GetByID(userProgram.ProgramID.String())
	} else {
		// Fallback to direct program.GetByID call if no service is injected
		programDef, err = program.GetByID(userProgram.ProgramID.String())
	}
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to load program: %w", err)
	}

	return user, userProgram, programDef, nil
}