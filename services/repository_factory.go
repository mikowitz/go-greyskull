package services

import (
	"github.com/mikowitz/greyskull/repository"
)

// RepositoryFactory defines an interface for creating repositories
// This abstraction allows for dependency injection and easier testing
type RepositoryFactory interface {
	// NewUserRepository creates a new UserRepository instance
	NewUserRepository() (repository.UserRepository, error)
}

// JSONRepositoryFactory implements RepositoryFactory for JSON-based storage
type JSONRepositoryFactory struct{}

// NewJSONRepositoryFactory creates a new JSONRepositoryFactory instance
func NewJSONRepositoryFactory() RepositoryFactory {
	return &JSONRepositoryFactory{}
}

// NewUserRepository creates a new JSON-based UserRepository
func (f *JSONRepositoryFactory) NewUserRepository() (repository.UserRepository, error) {
	return repository.NewJSONUserRepository()
}

// DefaultRepositoryFactory provides a package-level default factory
// This can be overridden for testing or different storage backends
var DefaultRepositoryFactory RepositoryFactory = NewJSONRepositoryFactory()

// GetDefaultRepositoryFactory returns the current default repository factory
func GetDefaultRepositoryFactory() RepositoryFactory {
	return DefaultRepositoryFactory
}

// SetDefaultRepositoryFactory sets a new default repository factory
// This is primarily useful for testing or changing storage backends globally
func SetDefaultRepositoryFactory(factory RepositoryFactory) {
	DefaultRepositoryFactory = factory
}