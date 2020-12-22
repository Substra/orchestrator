// Package testing provides helpers and mocks which can be used in tests
package testing

import "github.com/stretchr/testify/mock"

type MockDatabase struct {
	mock.Mock
}

func (m *MockDatabase) PutState(key string, data []byte) error {
	args := m.Called(key, data)
	return args.Error(0)
}

func (m *MockDatabase) GetState(key string) ([]byte, error) {
	args := m.Called(key)
	return args.Get(0).([]byte), args.Error(1)
}
