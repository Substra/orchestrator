package common

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockPublisher struct {
	mock.Mock
}

func (m *MockPublisher) Publish(ctx context.Context, routingKey string, data []byte) error {
	args := m.Called(ctx, routingKey, data)
	return args.Error(0)
}

func (m *MockPublisher) IsReady() bool {
	args := m.Called()
	return args.Bool(0)
}
