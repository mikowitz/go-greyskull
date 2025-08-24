package services

import (
	"errors"
	"testing"

	"github.com/mikowitz/greyskull/models"
	"github.com/mikowitz/greyskull/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockRepositoryFactory implements RepositoryFactory for testing
type MockRepositoryFactory struct {
	mock.Mock
}

func (m *MockRepositoryFactory) NewUserRepository() (repository.UserRepository, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(repository.UserRepository), args.Error(1)
}

func TestRepositoryFactory_Interface(t *testing.T) {
	// Ensure JSONRepositoryFactory implements RepositoryFactory interface
	factory := NewJSONRepositoryFactory()
	var _ RepositoryFactory = factory
	assert.NotNil(t, factory)
}

func TestJSONRepositoryFactory_NewUserRepository(t *testing.T) {
	factory := NewJSONRepositoryFactory()
	
	repo, err := factory.NewUserRepository()
	
	require.NoError(t, err)
	require.NotNil(t, repo)
	
	// Verify it returns a proper UserRepository interface implementation
	var _ repository.UserRepository = repo
}

func TestJSONRepositoryFactory_ErrorHandling(t *testing.T) {
	// This test verifies that repository creation errors are properly propagated
	// We can't easily test this with the real JSONRepositoryFactory without 
	// mocking the filesystem, but we can test the pattern with a mock
	
	mockFactory := new(MockRepositoryFactory)
	expectedError := errors.New("repository creation failed")
	
	mockFactory.On("NewUserRepository").Return(nil, expectedError).Once()
	
	repo, err := mockFactory.NewUserRepository()
	
	assert.Error(t, err)
	assert.Nil(t, repo)
	assert.Equal(t, expectedError, err)
	
	mockFactory.AssertExpectations(t)
}

func TestRepositoryFactory_Pattern(t *testing.T) {
	tests := []struct {
		name          string
		mockSetup     func(*MockRepositoryFactory)
		expectedError string
		expectRepo    bool
	}{
		{
			name: "successful repository creation",
			mockSetup: func(m *MockRepositoryFactory) {
				mockRepo := new(MockUserRepository)
				m.On("NewUserRepository").Return(mockRepo, nil).Once()
			},
			expectRepo: true,
		},
		{
			name: "repository creation fails",
			mockSetup: func(m *MockRepositoryFactory) {
				m.On("NewUserRepository").Return(nil, errors.New("creation failed")).Once()
			},
			expectedError: "creation failed",
		},
		{
			name: "repository creation returns nil without error",
			mockSetup: func(m *MockRepositoryFactory) {
				m.On("NewUserRepository").Return(nil, nil).Once()
			},
			expectRepo: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFactory := new(MockRepositoryFactory)
			tt.mockSetup(mockFactory)
			
			repo, err := mockFactory.NewUserRepository()
			
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, repo)
			} else if tt.expectRepo {
				assert.NoError(t, err)
				assert.NotNil(t, repo)
			} else {
				assert.NoError(t, err)
				assert.Nil(t, repo)
			}
			
			mockFactory.AssertExpectations(t)
		})
	}
}

// Test the command service pattern that uses repository factory
func TestCommandService_WithRepositoryFactory(t *testing.T) {
	tests := []struct {
		name            string
		factorySetup    func(*MockRepositoryFactory)
		userSetup       func(*MockUserRepository)
		expectedError   string
		expectUser      bool
	}{
		{
			name: "successful command execution with repository factory",
			factorySetup: func(mf *MockRepositoryFactory) {
				mockRepo := new(MockUserRepository)
				mf.On("NewUserRepository").Return(mockRepo, nil).Once()
			},
			userSetup: func(mr *MockUserRepository) {
				mr.On("GetCurrent").Return("testuser", nil).Once()
				mr.On("Get", "testuser").Return(&models.User{Username: "testuser"}, nil).Once()
			},
			expectUser: true,
		},
		{
			name: "repository creation fails",
			factorySetup: func(mf *MockRepositoryFactory) {
				mf.On("NewUserRepository").Return(nil, errors.New("factory failed")).Once()
			},
			expectedError: "failed to initialize repository: factory failed",
		},
		{
			name: "user service fails with repository from factory",
			factorySetup: func(mf *MockRepositoryFactory) {
				mockRepo := new(MockUserRepository)
				mf.On("NewUserRepository").Return(mockRepo, nil).Once()
			},
			userSetup: func(mr *MockUserRepository) {
				mr.On("GetCurrent").Return("", repository.ErrNoCurrentUser).Once()
			},
			expectedError: "no current user set",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFactory := new(MockRepositoryFactory)
			tt.factorySetup(mockFactory)
			
			// This simulates what a command would do
			repo, err := mockFactory.NewUserRepository()
			if err != nil {
				err = errors.New("failed to initialize repository: " + err.Error())
				if tt.expectedError != "" {
					assert.Error(t, err)
					assert.Contains(t, err.Error(), tt.expectedError)
				}
				return
			}
			
			// If we got a repository, set up user service expectations
			if repo != nil && tt.userSetup != nil {
				mockRepo, ok := repo.(*MockUserRepository)
				require.True(t, ok)
				tt.userSetup(mockRepo)
				
				// Simulate UserService usage
				userService := NewUserService(repo, nil)
				user, err := userService.RequireCurrentUser()
				
				if tt.expectedError != "" {
					assert.Error(t, err)
					assert.Contains(t, err.Error(), tt.expectedError)
					assert.Nil(t, user)
				} else if tt.expectUser {
					assert.NoError(t, err)
					assert.NotNil(t, user)
					assert.Equal(t, "testuser", user.Username)
				}
			}
			
			mockFactory.AssertExpectations(t)
		})
	}
}

// Test that validates the dependency injection makes testing easier
func TestRepositoryDependencyInjection_TestingBenefits(t *testing.T) {
	// This test demonstrates how dependency injection makes testing easier
	// by allowing us to inject mock repositories instead of relying on filesystem
	
	mockRepo := new(MockUserRepository)
	userService := NewUserService(mockRepo, nil)
	
	// Setup mock expectations
	mockRepo.On("GetCurrent").Return("testuser", nil).Once()
	mockRepo.On("Get", "testuser").Return(&models.User{Username: "testuser"}, nil).Once()
	
	// Execute the service method
	user, err := userService.RequireCurrentUser()
	
	// Verify results
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "testuser", user.Username)
	
	// Verify all mock expectations were met
	mockRepo.AssertExpectations(t)
	
	// This demonstrates the testing benefit: no filesystem setup required,
	// no temporary directories, just clean mock-based testing
}