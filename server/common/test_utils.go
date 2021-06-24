package common

import "github.com/stretchr/testify/mock"

type MockPublisher struct {
	mock.Mock
}

func (m *MockPublisher) Publish(routingKey string, data []byte) error {
	args := m.Called(routingKey, data)
	return args.Error(0)
}
