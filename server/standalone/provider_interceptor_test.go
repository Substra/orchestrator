// Copyright 2021 Owkin Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package standalone

import (
	"context"
	"errors"
	"testing"

	"github.com/owkin/orchestrator/lib/event"
	persistenceTesting "github.com/owkin/orchestrator/lib/persistence/testing"
	"github.com/owkin/orchestrator/lib/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

type mockedChannel struct {
	mock.Mock
}

func (m *mockedChannel) Publish(data []byte) error {
	args := m.Called(data)
	return args.Error(0)
}

type mockedTransactionDBAL struct {
	persistenceTesting.MockDBAL
}

func (m *mockedTransactionDBAL) Commit() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockedTransactionDBAL) Rollback() error {
	args := m.Called()
	return args.Error(0)
}

type mockTransactionalDBALProvider struct {
	m *mockedTransactionDBAL
}

func (p *mockTransactionalDBALProvider) GetTransactionalDBAL() (TransactionDBAL, error) {
	return p.m, nil
}

func TestExtractProvider(t *testing.T) {
	ctx := context.TODO()

	p := &service.MockServiceProvider{}

	ctxWithProvider := context.WithValue(ctx, ctxProviderKey, p)

	extracted, err := ExtractProvider(ctxWithProvider)
	assert.NoError(t, err, "extraction should not fail")
	assert.Equal(t, p, extracted, "Invocator should be extracted from context")

	_, err = ExtractProvider(ctx)
	assert.Error(t, err, "Extraction should fail on empty context")

}

func TestInjectProvider(t *testing.T) {
	channel := new(mockedChannel)

	db := new(mockedTransactionDBAL)
	db.On("Commit").Once().Return(nil)
	dbProvider := &mockTransactionalDBALProvider{db}

	interceptor := NewProviderInterceptor(dbProvider, channel)

	unaryInfo := &grpc.UnaryServerInfo{
		FullMethod: "TestService.UnaryMethod",
	}
	unaryHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
		_, err := ExtractProvider(ctx)
		assert.NoError(t, err, "Provider extraction should not fail")
		return "test", nil
	}

	interceptor.Intercept(context.TODO(), "test", unaryInfo, unaryHandler)
}

func TestOnSuccess(t *testing.T) {
	channel := new(mockedChannel)
	db := new(mockedTransactionDBAL)
	dbProvider := &mockTransactionalDBALProvider{db}

	db.On("Commit").Once().Return(nil)
	channel.On("Publish", mock.Anything).Once().Return(nil)

	interceptor := NewProviderInterceptor(dbProvider, channel)

	unaryInfo := &grpc.UnaryServerInfo{
		FullMethod: "TestService.UnaryMethod",
	}
	unaryHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
		provider, err := ExtractProvider(ctx)
		require.NoError(t, err)

		provider.GetEventQueue().Enqueue(&event.Event{})

		return "test", nil
	}

	interceptor.Intercept(context.TODO(), "test", unaryInfo, unaryHandler)
}

func TestOnError(t *testing.T) {
	channel := new(mockedChannel)
	db := new(mockedTransactionDBAL)
	dbProvider := &mockTransactionalDBALProvider{db}

	db.On("Rollback").Once().Return(nil)

	interceptor := NewProviderInterceptor(dbProvider, channel)

	unaryInfo := &grpc.UnaryServerInfo{
		FullMethod: "TestService.UnaryMethod",
	}
	unaryHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
		provider, err := ExtractProvider(ctx)
		require.NoError(t, err)

		provider.GetEventQueue().Enqueue(&event.Event{})

		return nil, errors.New("test error")
	}

	res, err := interceptor.Intercept(context.TODO(), "test", unaryInfo, unaryHandler)
	assert.Nil(t, res)
	assert.Error(t, err)
}
