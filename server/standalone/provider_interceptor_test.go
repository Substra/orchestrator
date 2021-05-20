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

	"github.com/owkin/orchestrator/lib/asset"
	persistenceTesting "github.com/owkin/orchestrator/lib/persistence/testing"
	"github.com/owkin/orchestrator/lib/service"
	"github.com/owkin/orchestrator/server/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

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

func (p *mockTransactionalDBALProvider) GetTransactionalDBAL(_ string) (TransactionDBAL, error) {
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
	publisher := new(common.MockPublisher)

	db := new(mockedTransactionDBAL)
	db.On("Commit").Once().Return(nil)
	dbProvider := &mockTransactionalDBALProvider{db}

	interceptor := NewProviderInterceptor(dbProvider, publisher)

	unaryInfo := &grpc.UnaryServerInfo{
		FullMethod: "TestService.UnaryMethod",
	}
	unaryHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
		_, err := ExtractProvider(ctx)
		assert.NoError(t, err, "Provider extraction should not fail")
		return "test", nil
	}

	ctx := common.WithChannel(context.TODO(), "testChannel")
	_, err := interceptor.Intercept(ctx, "test", unaryInfo, unaryHandler)
	assert.NoError(t, err)
}

func TestOnSuccess(t *testing.T) {
	publisher := new(common.MockPublisher)
	db := new(mockedTransactionDBAL)
	dbProvider := &mockTransactionalDBALProvider{db}

	db.On("Commit").Once().Return(nil)
	publisher.On("Publish", "testChannel", mock.Anything).Once().Return(nil)

	interceptor := NewProviderInterceptor(dbProvider, publisher)

	unaryInfo := &grpc.UnaryServerInfo{
		FullMethod: "TestService.UnaryMethod",
	}
	unaryHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
		provider, err := ExtractProvider(ctx)
		require.NoError(t, err)

		err = provider.GetEventQueue().Enqueue(&asset.Event{})
		require.NoError(t, err)

		return "test", nil
	}

	ctx := common.WithChannel(context.TODO(), "testChannel")
	_, err := interceptor.Intercept(ctx, "test", unaryInfo, unaryHandler)
	assert.NoError(t, err)
}

func TestOnError(t *testing.T) {
	publisher := new(common.MockPublisher)
	db := new(mockedTransactionDBAL)
	dbProvider := &mockTransactionalDBALProvider{db}

	db.On("Rollback").Once().Return(nil)

	interceptor := NewProviderInterceptor(dbProvider, publisher)

	unaryInfo := &grpc.UnaryServerInfo{
		FullMethod: "TestService.UnaryMethod",
	}
	unaryHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
		provider, err := ExtractProvider(ctx)
		require.NoError(t, err)

		err = provider.GetEventQueue().Enqueue(&asset.Event{})
		require.NoError(t, err)

		return nil, errors.New("test error")
	}

	ctx := common.WithChannel(context.TODO(), "testChannel")
	res, err := interceptor.Intercept(ctx, "test", unaryInfo, unaryHandler)
	assert.Nil(t, res)
	assert.Error(t, err)
}
