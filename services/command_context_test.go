package services

import (
	"errors"
	"testing"

	"github.com/mikowitz/greyskull/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommandContext_NewCommandContext(t *testing.T) {
	factory := NewJSONRepositoryFactory()
	
	ctx, err := NewCommandContext(factory)
	
	require.NoError(t, err)
	require.NotNil(t, ctx)
	assert.NotNil(t, ctx.UserRepo)
	assert.NotNil(t, ctx.UserService)
	
	// Verify the user service has the correct repository
	user, err := ctx.UserRepo.GetCurrent()
	// We don't care about the result, just that it doesn't panic
	_ = user
	_ = err
}

func TestCommandContext_WithMockFactory(t *testing.T) {
	mockFactory := new(MockRepositoryFactory)
	mockRepo := new(MockUserRepository)
	
	mockFactory.On("NewUserRepository").Return(mockRepo, nil).Once()
	
	ctx, err := NewCommandContext(mockFactory)
	
	require.NoError(t, err)
	require.NotNil(t, ctx)
	assert.Equal(t, mockRepo, ctx.UserRepo)
	assert.NotNil(t, ctx.UserService)
	
	mockFactory.AssertExpectations(t)
}

func TestCommandContext_RepositoryCreationFailure(t *testing.T) {
	mockFactory := new(MockRepositoryFactory)
	expectedError := errors.New("repository creation failed")
	
	mockFactory.On("NewUserRepository").Return(nil, expectedError).Once()
	
	ctx, err := NewCommandContext(mockFactory)
	
	assert.Error(t, err)
	assert.Nil(t, ctx)
	assert.Contains(t, err.Error(), "failed to create user repository")
	assert.Contains(t, err.Error(), expectedError.Error())
	
	mockFactory.AssertExpectations(t)
}

func TestCommandContext_UserServiceIntegration(t *testing.T) {
	// This test verifies that the CommandContext properly integrates
	// the UserService with the repository
	
	mockFactory := new(MockRepositoryFactory)
	mockRepo := new(MockUserRepository)
	
	mockFactory.On("NewUserRepository").Return(mockRepo, nil).Once()
	
	// Setup user service expectations
	mockRepo.On("GetCurrent").Return("testuser", nil).Once()
	mockRepo.On("Get", "testuser").Return(&models.User{Username: "testuser"}, nil).Once()
	
	ctx, err := NewCommandContext(mockFactory)
	require.NoError(t, err)
	
	// Use the user service from the context
	user, err := ctx.UserService.RequireCurrentUser()
	
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "testuser", user.Username)
	
	mockFactory.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestCommandContext_DefaultFactory(t *testing.T) {
	// Test using the default factory
	ctx, err := NewCommandContextWithDefaults()
	
	require.NoError(t, err)
	require.NotNil(t, ctx)
	assert.NotNil(t, ctx.UserRepo)
	assert.NotNil(t, ctx.UserService)
}

func TestCommandContext_NilFactory(t *testing.T) {
	ctx, err := NewCommandContext(nil)
	
	assert.Error(t, err)
	assert.Nil(t, ctx)
	assert.Contains(t, err.Error(), "repository factory cannot be nil")
}

// Test helper to demonstrate how command contexts make command testing easier
func TestCommandContext_CommandTestingExample(t *testing.T) {
	// This demonstrates how a command would use the CommandContext for testing
	
	mockFactory := new(MockRepositoryFactory)
	mockRepo := new(MockUserRepository)
	
	// Setup factory to return our mock repository
	mockFactory.On("NewUserRepository").Return(mockRepo, nil).Once()
	
	// Setup repository expectations for a typical command
	mockRepo.On("GetCurrent").Return("testuser", nil).Once()
	mockRepo.On("Get", "testuser").Return(&models.User{
		Username: "testuser",
	}, nil).Once()
	
	// Create command context (what a command would do)
	ctx, err := NewCommandContext(mockFactory)
	require.NoError(t, err)
	
	// Simulate a command using the context
	user, err := ctx.UserService.RequireCurrentUser()
	
	// Verify the command worked
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "testuser", user.Username)
	
	// Verify all expectations were met
	mockFactory.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
	
	// This pattern eliminates the need for commands to:
	// 1. Create repositories directly
	// 2. Handle repository creation errors individually
	// 3. Set up complex test environments
}