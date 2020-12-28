// Package testing provides helpers and mocks which can be used in tests
package testing

import "github.com/stretchr/testify/mock"

// MockDatabase is a convenience mock of the persistence layer interface
type MockDatabase struct {
	mock.Mock
}

// PutState stores data
func (m *MockDatabase) PutState(resource string, key string, data []byte) error {
	args := m.Called(resource, key, data)
	return args.Error(0)
}

// GetState fetches identified data
func (m *MockDatabase) GetState(resource string, key string) ([]byte, error) {
	args := m.Called(resource, key)
	return args.Get(0).([]byte), args.Error(1)
}

// GetAll retrieves all data for a resource kind
func (m *MockDatabase) GetAll(resource string) ([][]byte, error) {
	args := m.Called(resource)
	return args.Get(0).([][]byte), args.Error(1)
}
