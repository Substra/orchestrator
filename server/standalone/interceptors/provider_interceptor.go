package interceptors

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/owkin/orchestrator/lib/service"
	"github.com/owkin/orchestrator/server/common"
	"github.com/owkin/orchestrator/server/common/logger"
	"github.com/owkin/orchestrator/server/standalone/dbal"
	"github.com/owkin/orchestrator/server/standalone/event"
	"google.golang.org/grpc"
)

// gRPC methods for which we won't inject a service provider
var ignoredMethods = [...]string{
	"grpc.health",
}

// ProviderInterceptor intercepts gRPC requests and assign a request-scoped orchestration.Provider
// to the request context.
type ProviderInterceptor struct {
	amqp         common.AMQPPublisher
	dbalProvider dbal.TransactionalDBALProvider
	txChecker    common.TransactionChecker
}

type ctxProviderInterceptorMarker struct{}

var ctxProviderKey = &ctxProviderInterceptorMarker{}

// NewProviderInterceptor returns an instance of ProviderInterceptor
func NewProviderInterceptor(dbalProvider dbal.TransactionalDBALProvider, amqp common.AMQPPublisher) *ProviderInterceptor {
	return &ProviderInterceptor{
		amqp:         amqp,
		dbalProvider: dbalProvider,
		txChecker:    new(common.GrpcMethodChecker),
	}
}

func WithProvider(ctx context.Context, provider service.DependenciesProvider) context.Context {
	return context.WithValue(ctx, ctxProviderKey, provider)
}

// Intercept a gRPC request and inject the dependency injection orchestration.Provider into the context.
// The provider can be retrieved from context with ExtractProvider function.
func (pi *ProviderInterceptor) Intercept(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// Passthrough for ignored methods
	for _, m := range ignoredMethods {
		if strings.Contains(info.FullMethod, m) {
			return handler(ctx, req)
		}
	}

	channel, err := common.ExtractChannel(ctx)
	if err != nil {
		return nil, err
	}

	// This dispatcher should stay scoped per request since there is a single event queue
	dispatcher := event.NewAMQPDispatcher(pi.amqp, channel)

	readOnly := pi.txChecker.IsEvaluateMethod(info.FullMethod)

	tx, err := pi.dbalProvider.GetTransactionalDBAL(ctx, channel, readOnly)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	provider := service.NewProvider(logger.Get(ctx), tx, dispatcher)

	ctx = WithProvider(ctx, provider)
	res, err := handler(ctx, req)

	// Events should be dispatched only on successful transactions
	if err != nil {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			return nil, fmt.Errorf("failed to rollback transaction: %w", rollbackErr)
		}
	} else {
		commitErr := tx.Commit()
		if commitErr != nil {
			return nil, fmt.Errorf("failed to commit transaction: %w", commitErr)
		}
		go func() {
			dispatchErr := dispatcher.Dispatch(ctx)
			if dispatchErr != nil {
				// simply log since we cannot rollback the DB transaction anyway
				logger.Get(ctx).WithError(dispatchErr).
					WithField("events", dispatcher.GetEvents()).
					Error("failed to dispatch events after successful transaction commit")
			}
		}()

	}

	return res, err
}

// ExtractProvider will return the orchestration.ServiceProvider injected in context
func ExtractProvider(ctx context.Context) (service.DependenciesProvider, error) {
	provider, ok := ctx.Value(ctxProviderKey).(service.DependenciesProvider)
	if !ok {
		return nil, errors.New("provider not found in context")
	}
	return provider, nil
}
