package interceptors

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/substra/orchestrator/lib/service"
	"github.com/substra/orchestrator/server/common/interceptors"
	"github.com/substra/orchestrator/server/standalone/dbal"
	"github.com/substra/orchestrator/utils"
	"google.golang.org/grpc"
)

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
	ctx := interceptors.WithChannel(context.TODO(), "testChannel")

	tx := new(utils.MockTx)
	tx.On("Conn").Return(nil)
	tx.On("Commit", utils.AnyContext).Return(nil)
	pool := new(dbal.MockPgPool)
	pool.On("BeginTx", ctx, pgx.TxOptions{IsoLevel: pgx.Serializable}).Return(tx, nil)
	db := &dbal.Database{Pool: pool}

	healthcheck := new(MockHealthReporter)

	interceptor := NewProviderInterceptor(db, healthcheck)

	unaryInfo := &grpc.UnaryServerInfo{
		FullMethod: "TestService.UnaryMethod",
	}
	unaryHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
		_, err := ExtractProvider(ctx)
		assert.NoError(t, err, "Provider extraction should not fail")
		return "test", nil
	}

	_, err := interceptor.UnaryServerInterceptor(ctx, "test", unaryInfo, unaryHandler)
	assert.NoError(t, err)

	tx.AssertExpectations(t)
	pool.AssertExpectations(t)
	healthcheck.AssertExpectations(t)
}

func TestOnSuccess(t *testing.T) {
	ctx := interceptors.WithChannel(context.TODO(), "testChannel")

	tx := new(utils.MockTx)
	tx.On("Conn").Return(nil)
	tx.On("Commit", utils.AnyContext).Return(nil)
	pool := new(dbal.MockPgPool)
	pool.On("BeginTx", ctx, pgx.TxOptions{IsoLevel: pgx.Serializable}).Return(tx, nil)
	db := &dbal.Database{Pool: pool}

	healthcheck := new(MockHealthReporter)

	interceptor := NewProviderInterceptor(db, healthcheck)

	unaryInfo := &grpc.UnaryServerInfo{
		FullMethod: "TestService.UnaryMethod",
	}
	unaryHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
		_, err := ExtractProvider(ctx)
		require.NoError(t, err)

		return "test", nil
	}

	_, err := interceptor.UnaryServerInterceptor(ctx, "test", unaryInfo, unaryHandler)
	assert.NoError(t, err)

	tx.AssertExpectations(t)
	pool.AssertExpectations(t)
	healthcheck.AssertExpectations(t)
}

func TestOnError(t *testing.T) {
	ctx := interceptors.WithChannel(context.TODO(), "testChannel")

	tx := new(utils.MockTx)
	tx.On("Conn").Return(nil)
	tx.On("Rollback", utils.AnyContext).Return(nil)

	pool := new(dbal.MockPgPool)
	pool.On("BeginTx", ctx, pgx.TxOptions{IsoLevel: pgx.Serializable}).Return(tx, nil)
	db := &dbal.Database{Pool: pool}

	healthcheck := new(MockHealthReporter)

	interceptor := NewProviderInterceptor(db, healthcheck)

	unaryInfo := &grpc.UnaryServerInfo{
		FullMethod: "TestService.UnaryMethod",
	}
	unaryHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
		_, err := ExtractProvider(ctx)
		require.NoError(t, err)
		return nil, errors.New("test error")
	}

	res, err := interceptor.UnaryServerInterceptor(ctx, "test", unaryInfo, unaryHandler)
	assert.Nil(t, res)
	assert.Error(t, err)

	tx.AssertExpectations(t)
	pool.AssertExpectations(t)
	healthcheck.AssertExpectations(t)
}
