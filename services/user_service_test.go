package services

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/mikowitz/greyskull/models"
	"github.com/mikowitz/greyskull/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockUserRepository implements repository.UserRepository for testing
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) Get(username string) (*models.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) List() ([]string, error) {
	args := m.Called()
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockUserRepository) GetCurrent() (string, error) {
	args := m.Called()
	return args.Get(0).(string), args.Error(1)
}

func (m *MockUserRepository) SetCurrent(username string) error {
	args := m.Called(username)
	return args.Error(0)
}

// MockProgramService implements program loading for testing
type MockProgramService struct {
	mock.Mock
}

func (m *MockProgramService) GetByID(id string) (*models.Program, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Program), args.Error(1)
}

func TestUserService_RequireCurrentUser(t *testing.T) {
	tests := []struct {
		name           string
		mockGetCurrent func() (string, error)
		mockGet        func(string) (*models.User, error)
		expectedUser   *models.User
		expectedError  string
	}{
		{
			name: "successful user loading",
			mockGetCurrent: func() (string, error) {
				return "testuser", nil
			},
			mockGet: func(username string) (*models.User, error) {
				require.Equal(t, "testuser", username)
				return &models.User{
					ID:       uuid.New(),
					Username: "testuser",
				}, nil
			},
			expectedUser: &models.User{
				Username: "testuser",
			},
		},
		{
			name: "no current user set",
			mockGetCurrent: func() (string, error) {
				return "", repository.ErrNoCurrentUser
			},
			mockGet: func(username string) (*models.User, error) {
				t.Fatal("Get should not be called when GetCurrent fails")
				return nil, nil
			},
			expectedError: "no current user set. Use 'greyskull user create' or 'greyskull user switch' first",
		},
		{
			name: "get current user fails with other error",
			mockGetCurrent: func() (string, error) {
				return "", errors.New("repository error")
			},
			mockGet: func(username string) (*models.User, error) {
				t.Fatal("Get should not be called when GetCurrent fails")
				return nil, nil
			},
			expectedError: "failed to get current user: repository error",
		},
		{
			name: "user not found",
			mockGetCurrent: func() (string, error) {
				return "missinguser", nil
			},
			mockGet: func(username string) (*models.User, error) {
				require.Equal(t, "missinguser", username)
				return nil, repository.ErrUserNotFound
			},
			expectedError: "failed to load current user: user not found",
		},
		{
			name: "user loading fails with other error",
			mockGetCurrent: func() (string, error) {
				return "testuser", nil
			},
			mockGet: func(username string) (*models.User, error) {
				return nil, errors.New("database error")
			},
			expectedError: "failed to load current user: database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			userService := NewUserService(mockRepo, nil)

			// Setup mock expectations
			currentUsername, currentErr := tt.mockGetCurrent()
			mockRepo.On("GetCurrent").Return(currentUsername, currentErr).Once()
			if tt.expectedError == "" || !errors.Is(currentErr, repository.ErrNoCurrentUser) {
				if currentErr == nil {
					user, userErr := tt.mockGet(currentUsername)
					mockRepo.On("Get", currentUsername).Return(user, userErr).Once()
				}
			}

			// Execute test
			user, err := userService.RequireCurrentUser()

			// Assert results
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, user)
			} else {
				require.NoError(t, err)
				require.NotNil(t, user)
				assert.Equal(t, tt.expectedUser.Username, user.Username)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_GetCurrentUserWithProgram(t *testing.T) {
	testProgramID := uuid.New()
	testUserProgramID := uuid.New()
	testProgram := &models.Program{
		ID:   testProgramID,
		Name: "Test Program",
	}

	tests := []struct {
		name               string
		mockGetCurrent     func() (string, error)
		mockGet            func(string) (*models.User, error)
		mockGetByID        func(string) (*models.Program, error)
		expectedUser       *models.User
		expectedUserProg   *models.UserProgram
		expectedProgram    *models.Program
		expectedError      string
	}{
		{
			name: "successful loading with active program",
			mockGetCurrent: func() (string, error) {
				return "testuser", nil
			},
			mockGet: func(username string) (*models.User, error) {
				userProgram := &models.UserProgram{
					ID:        testUserProgramID,
					ProgramID: testProgramID,
				}
				return &models.User{
					ID:             uuid.New(),
					Username:       "testuser",
					CurrentProgram: testUserProgramID,
					Programs:       map[uuid.UUID]*models.UserProgram{testUserProgramID: userProgram},
				}, nil
			},
			mockGetByID: func(id string) (*models.Program, error) {
				require.Equal(t, testProgramID.String(), id)
				return testProgram, nil
			},
			expectedUser: &models.User{Username: "testuser"},
			expectedUserProg: &models.UserProgram{
				ID:        testUserProgramID,
				ProgramID: testProgramID,
			},
			expectedProgram: testProgram,
		},
		{
			name: "user has no active program",
			mockGetCurrent: func() (string, error) {
				return "testuser", nil
			},
			mockGet: func(username string) (*models.User, error) {
				return &models.User{
					ID:             uuid.New(),
					Username:       "testuser",
					CurrentProgram: uuid.Nil, // No active program
					Programs:       make(map[uuid.UUID]*models.UserProgram),
				}, nil
			},
			mockGetByID: func(string) (*models.Program, error) {
				t.Fatal("GetByID should not be called when no program is active")
				return nil, nil
			},
			expectedError: "no active program. Use 'greyskull program start' to begin a program",
		},
		{
			name: "current program not found in user programs",
			mockGetCurrent: func() (string, error) {
				return "testuser", nil
			},
			mockGet: func(username string) (*models.User, error) {
				return &models.User{
					ID:             uuid.New(),
					Username:       "testuser",
					CurrentProgram: testUserProgramID,
					Programs:       make(map[uuid.UUID]*models.UserProgram), // Empty - program missing
				}, nil
			},
			mockGetByID: func(string) (*models.Program, error) {
				t.Fatal("GetByID should not be called when user program is missing")
				return nil, nil
			},
			expectedError: "current program not found in user programs",
		},
		{
			name: "program loading fails",
			mockGetCurrent: func() (string, error) {
				return "testuser", nil
			},
			mockGet: func(username string) (*models.User, error) {
				userProgram := &models.UserProgram{
					ID:        testUserProgramID,
					ProgramID: testProgramID,
				}
				return &models.User{
					ID:             uuid.New(),
					Username:       "testuser",
					CurrentProgram: testUserProgramID,
					Programs:       map[uuid.UUID]*models.UserProgram{testUserProgramID: userProgram},
				}, nil
			},
			mockGetByID: func(id string) (*models.Program, error) {
				return nil, errors.New("program not found")
			},
			expectedError: "failed to load program: program not found",
		},
		{
			name: "user loading fails",
			mockGetCurrent: func() (string, error) {
				return "testuser", nil
			},
			mockGet: func(username string) (*models.User, error) {
				return nil, errors.New("user load error")
			},
			mockGetByID: func(string) (*models.Program, error) {
				t.Fatal("GetByID should not be called when user loading fails")
				return nil, nil
			},
			expectedError: "failed to load current user: user load error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			mockProgramService := new(MockProgramService)
			userService := NewUserService(mockRepo, mockProgramService)

			// Setup mock expectations
			mockRepo.On("GetCurrent").Return(tt.mockGetCurrent()).Once()
			
			if username, err := tt.mockGetCurrent(); err == nil {
				user, userErr := tt.mockGet(username)
				mockRepo.On("Get", username).Return(user, userErr).Once()
				
				if userErr == nil && user != nil {
					// Check if user has active program and if program loading should be called
					if user.CurrentProgram != uuid.Nil {
						if userProgram, exists := user.Programs[user.CurrentProgram]; exists {
							mockProgramService.On("GetByID", userProgram.ProgramID.String()).Return(tt.mockGetByID(userProgram.ProgramID.String())).Once()
						}
					}
				}
			}

			// Execute test
			user, userProgram, program, err := userService.GetCurrentUserWithProgram()

			// Assert results
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, user)
				assert.Nil(t, userProgram)
				assert.Nil(t, program)
			} else {
				require.NoError(t, err)
				require.NotNil(t, user)
				require.NotNil(t, userProgram)
				require.NotNil(t, program)
				
				assert.Equal(t, tt.expectedUser.Username, user.Username)
				assert.Equal(t, tt.expectedUserProg.ID, userProgram.ID)
				assert.Equal(t, tt.expectedUserProg.ProgramID, userProgram.ProgramID)
				assert.Equal(t, tt.expectedProgram.Name, program.Name)
			}

			mockRepo.AssertExpectations(t)
			mockProgramService.AssertExpectations(t)
		})
	}
}

func TestUserService_NewUserService(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockProgramService := new(MockProgramService)

	t.Run("creates service with repository only", func(t *testing.T) {
		service := NewUserService(mockRepo, nil)
		assert.NotNil(t, service)
		assert.Equal(t, mockRepo, service.repo)
		assert.Nil(t, service.programService)
	})

	t.Run("creates service with both dependencies", func(t *testing.T) {
		service := NewUserService(mockRepo, mockProgramService)
		assert.NotNil(t, service)
		assert.Equal(t, mockRepo, service.repo)
		assert.Equal(t, mockProgramService, service.programService)
	})

	t.Run("panics with nil repository", func(t *testing.T) {
		assert.Panics(t, func() {
			NewUserService(nil, mockProgramService)
		})
	})
}

func TestUserService_GetCurrentUserWithProgram_UsesDefaultProgramService(t *testing.T) {
	// This test ensures that when programService is nil, it falls back to program.GetByID
	testProgramID := uuid.New()
	testUserProgramID := uuid.New()
	
	mockRepo := new(MockUserRepository)
	userService := NewUserService(mockRepo, nil) // No program service

	userProgram := &models.UserProgram{
		ID:        testUserProgramID,
		ProgramID: testProgramID,
	}
	user := &models.User{
		ID:             uuid.New(),
		Username:       "testuser",
		CurrentProgram: testUserProgramID,
		Programs:       map[uuid.UUID]*models.UserProgram{testUserProgramID: userProgram},
	}

	mockRepo.On("GetCurrent").Return("testuser", nil).Once()
	mockRepo.On("Get", "testuser").Return(user, nil).Once()

	// This should work even without programService by using program.GetByID directly
	// But we can't actually test this without mocking the program package, 
	// so this test just ensures the service doesn't panic
	_, _, _, err := userService.GetCurrentUserWithProgram()
	
	// We expect an error because program.GetByID won't find a test program
	// but we shouldn't get a panic
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load program")
	
	mockRepo.AssertExpectations(t)
}

func TestUserService_RequireCurrentUser_ErrorMessages(t *testing.T) {
	// Test that error messages match exactly what the existing commands use
	mockRepo := new(MockUserRepository)
	userService := NewUserService(mockRepo, nil)

	t.Run("no current user error message matches existing commands", func(t *testing.T) {
		mockRepo.On("GetCurrent").Return("", repository.ErrNoCurrentUser).Once()

		_, err := userService.RequireCurrentUser()

		assert.Error(t, err)
		// This should match the exact error message used in workout_log.go and workout_next.go
		assert.Equal(t, "no current user set. Use 'greyskull user create' or 'greyskull user switch' first", err.Error())
	})

	mockRepo.AssertExpectations(t)
}

func TestUserService_GetCurrentUserWithProgram_ErrorMessages(t *testing.T) {
	// Test that error messages match exactly what the existing commands use
	mockRepo := new(MockUserRepository)
	userService := NewUserService(mockRepo, nil)

	t.Run("no active program error message matches existing commands", func(t *testing.T) {
		user := &models.User{
			ID:             uuid.New(),
			Username:       "testuser", 
			CurrentProgram: uuid.Nil,
			Programs:       make(map[uuid.UUID]*models.UserProgram),
		}

		mockRepo.On("GetCurrent").Return("testuser", nil).Once()
		mockRepo.On("Get", "testuser").Return(user, nil).Once()

		_, _, _, err := userService.GetCurrentUserWithProgram()

		assert.Error(t, err)
		// This should match the exact error message used in workout_log.go and workout_next.go  
		assert.Equal(t, "no active program. Use 'greyskull program start' to begin a program", err.Error())
	})

	t.Run("current program not found error message matches existing commands", func(t *testing.T) {
		testUserProgramID := uuid.New()
		user := &models.User{
			ID:             uuid.New(),
			Username:       "testuser",
			CurrentProgram: testUserProgramID,
			Programs:       make(map[uuid.UUID]*models.UserProgram), // Empty
		}

		mockRepo.On("GetCurrent").Return("testuser", nil).Once()
		mockRepo.On("Get", "testuser").Return(user, nil).Once()

		_, _, _, err := userService.GetCurrentUserWithProgram()

		assert.Error(t, err)
		// This should match the exact error message used in workout_log.go and workout_next.go
		assert.Equal(t, "current program not found in user programs", err.Error())
	})

	mockRepo.AssertExpectations(t)
}