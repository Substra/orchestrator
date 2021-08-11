package common

import (
	"context"
	"strings"
	"time"

	"github.com/owkin/orchestrator/server/common/logger"
	"google.golang.org/grpc"
)

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

func (ri *RetryInterceptor) Intercept(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
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
		res, err := handler(ctx, req)
		attemptDuration := time.Since(attemptStart)

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