package interceptors

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgconn"
	"github.com/owkin/orchestrator/lib/asset"
	persistenceTesting "github.com/owkin/orchestrator/lib/persistence/mocks"
	"github.com/owkin/orchestrator/lib/service"
	"github.com/owkin/orchestrator/server/common"
	"github.com/owkin/orchestrator/server/standalone/dbal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

var piConfig = ProviderInterceptorConfiguration{
	TxRetryBudget: 500 * time.Millisecond,
}

type mockedTransactionDBAL struct {
	persistenceTesting.DBAL
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

func (p *mockTransactionalDBALProvider) GetTransactionalDBAL(_ context.Context, _ string, _ bool) (dbal.TransactionDBAL, error) {
	return p.m, nil
}

func TestExtractProvider(t *testing.T) {
	ctx := context.TODO()

	p := &service.MockDependenciesProvider{}

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

	interceptor := NewProviderInterceptor(dbProvider, publisher, piConfig)

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

	db.AssertExpectations(t)
	publisher.AssertExpectations(t)
}

func TestOnSuccess(t *testing.T) {
	publisher := new(common.MockPublisher)
	db := new(mockedTransactionDBAL)
	dbProvider := &mockTransactionalDBALProvider{db}

	db.On("Commit").Once().Return(nil)

	wg := new(sync.WaitGroup)
	wg.Add(1)

	publisher.On("Publish", "testChannel", mock.Anything).Once().Return(nil).Run(func(args mock.Arguments) {
		wg.Done()
	})

	interceptor := NewProviderInterceptor(dbProvider, publisher, piConfig)

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

	// wait for async event dispatch
	wg.Wait()

	db.AssertExpectations(t)
	publisher.AssertExpectations(t)
}

func TestOnError(t *testing.T) {
	publisher := new(common.MockPublisher)
	db := new(mockedTransactionDBAL)
	dbProvider := &mockTransactionalDBALProvider{db}

	db.On("Rollback").Once().Return(nil)

	interceptor := NewProviderInterceptor(dbProvider, publisher, piConfig)

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

	db.AssertExpectations(t)
	publisher.AssertExpectations(t)
}

func TestRetryOnUnserializableTransaction(t *testing.T) {
	publisher := new(common.MockPublisher)
	db := new(mockedTransactionDBAL)
	dbProvider := &mockTransactionalDBALProvider{db}

	db.On("Rollback").Once().Return(nil)
	db.On("Commit").Once().Return(nil)

	wg := new(sync.WaitGroup)
	wg.Add(1)

	publisher.On("Publish", "testChannel", mock.Anything).Once().Return(nil).Run(func(args mock.Arguments) {
		wg.Done()
	})

	interceptor := NewProviderInterceptor(dbProvider, publisher, piConfig)

	unaryInfo := &grpc.UnaryServerInfo{
		FullMethod: "TestService.UnaryMethod",
	}
	failed := false
	unaryHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
		provider, err := ExtractProvider(ctx)
		require.NoError(t, err)

		err = provider.GetEventQueue().Enqueue(&asset.Event{})
		require.NoError(t, err)

		if !failed {
			failed = true
			return nil, &pgconn.PgError{Code: "40001"}
		}

		return nil, nil
	}

	ctx := common.WithChannel(context.TODO(), "testChannel")
	_, err := interceptor.Intercept(ctx, "test", unaryInfo, unaryHandler)
	assert.NoError(t, err)

	// wait for async event dispatch
	wg.Wait()

	db.AssertExpectations(t)
	publisher.AssertExpectations(t)
}
