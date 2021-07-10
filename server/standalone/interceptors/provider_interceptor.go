package interceptors

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-playground/log/v7"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/owkin/orchestrator/lib/service"
	"github.com/owkin/orchestrator/server/common"
	"github.com/owkin/orchestrator/server/standalone/dbal"
	"github.com/owkin/orchestrator/server/standalone/event"
	"google.golang.org/grpc"
)

// gRPC methods for which we won't inject a service provider
var ignoredMethods = [...]string{
	"grpc.health",
}

type ProviderInterceptorConfiguration struct {
	TxRetryBudget time.Duration
}

// ProviderInterceptor intercepts gRPC requests and assign a request-scoped orchestration.Provider
// to the request context.
type ProviderInterceptor struct {
	amqp         common.AMQPPublisher
	dbalProvider dbal.TransactionalDBALProvider
	txChecker    common.TransactionChecker
	config       ProviderInterceptorConfiguration
}

type ctxProviderInterceptorMarker struct{}

var ctxProviderKey = &ctxProviderInterceptorMarker{}

// NewProviderInterceptor returns an instance of ProviderInterceptor
func NewProviderInterceptor(dbalProvider dbal.TransactionalDBALProvider, amqp common.AMQPPublisher, config ProviderInterceptorConfiguration) *ProviderInterceptor {
	return &ProviderInterceptor{
		amqp:         amqp,
		dbalProvider: dbalProvider,
		txChecker:    new(common.GrpcMethodChecker),
		config:       config,
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

	attempt := 1
	start := time.Now()

	for {
		attemptStart := time.Now()
		res, err := pi.handleIsolated(ctx, req, info, handler)
		attemptDuration := time.Since(attemptStart)

		if err == nil {
			return res, nil
		}

		if pi.shouldRetry(start, err) {
			attempt++
			log.
				WithField("method", info.FullMethod).
				WithField("previous attempt duration", attemptDuration).
				WithField("next attempt number", attempt).
				Info("retrying conflicting transaction")
			continue
		}
		return nil, err
	}
}

// shouldRetry returns true if the retry budget is not exhausted and the error is a transaction serialization failure.
func (pi *ProviderInterceptor) shouldRetry(start time.Time, err error) bool {
	retryBudgetExhausted := time.Since(start) > pi.config.TxRetryBudget
	var pgErr *pgconn.PgError

	return errors.As(err, &pgErr) && pgErr.Code == pgerrcode.SerializationFailure && !retryBudgetExhausted
}

// handleIsolated creates a transaction both at the SQL and event level.
func (pi *ProviderInterceptor) handleIsolated(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
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

	provider := service.NewProvider(tx, dispatcher)

	newCtx := WithProvider(ctx, provider)
	res, err := handler(newCtx, req)

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
			dispatchErr := dispatcher.Dispatch()
			if dispatchErr != nil {
				// simply log since we cannot rollback the DB transaction anyway
				log.WithError(dispatchErr).
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
