package interceptors

import (
	"context"
	"strings"
	"time"

	"github.com/substra/orchestrator/server/common/logger"
	"google.golang.org/grpc"
)

type ctxLastErrorMarker struct{}

var ctxLastErrorKey = &ctxLastErrorMarker{}

// RetryInterceptor will retry a failed request according to its checker decision and time budget.
type RetryInterceptor struct {
	shouldRetry RetryChecker
	retryBudget time.Duration
}

func NewRetryInterceptor(retryBudget time.Duration, checker RetryChecker) *RetryInterceptor {
	return &RetryInterceptor{
		retryBudget: retryBudget,
		shouldRetry: checker,
	}
}

// RetryChecker determines if a request should be retried.
// It receives the returned error and elapsed time since first attempt.
type RetryChecker = func(error) bool

func (ri *RetryInterceptor) UnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// Passthrough for ignored methods
	for _, m := range IgnoredMethods {
		if strings.Contains(info.FullMethod, m) {
			return handler(ctx, req)
		}
	}

	attempt := 1
	start := time.Now()
	retryCtx := ctx

	for {
		attemptStart := time.Now()
		res, err := handler(retryCtx, req)
		attemptDuration := time.Since(attemptStart)

		retryCtx = WithLastError(retryCtx, err)

		logger := logger.Get(ctx).WithField("method", info.FullMethod).WithField("attempt_duration", attemptDuration).WithField("attempt", attempt)

		if err == nil {
			return res, nil
		}

		if !ri.budgetAllowRetry(start) {
			logger.Error("retry budget exceeded")
			return nil, err
		}

		if ri.shouldRetry(err) {
			attempt++
			logger.Info("retrying failed transaction")
			continue
		}

		logger.WithError(err).Debug("should not retry on this error")
		return nil, err
	}
}

func (ri *RetryInterceptor) budgetAllowRetry(start time.Time) bool {
	return time.Since(start) <= ri.retryBudget
}

// WithLastError adds last error to the context
func WithLastError(ctx context.Context, err error) context.Context {
	return context.WithValue(ctx, ctxLastErrorKey, err)
}

// GetLastError returns the last error in a retry context
func GetLastError(ctx context.Context) error {
	err, ok := ctx.Value(ctxLastErrorKey).(error)
	if !ok {
		return nil
	}
	return err
}
