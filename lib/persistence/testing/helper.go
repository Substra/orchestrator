// Package testing provides helpers and mocks which can be used in tests
package testing

import "github.com/stretchr/testify/mock"

type MockDatabase struct {
	mock.Mock
}

func (m *MockDatabase) PutState(resource string, key string, data []byte) error {
	args := m.Called(resource, key, data)
	return args.Error(0)
}

func (m *MockDatabase) GetState(resource string, key string) ([]byte, error) {
	args := m.Called(resource, key)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockDatabase) GetAll(resource string) ([][]byte, error) {
	args := m.Called(resource)
	return args.Get(0).([][]byte), args.Error(1)
}
