package interceptors

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestRetryBudget(t *testing.T) {
	interceptor := NewRetryInterceptor(5*time.Second, func(err error) bool { return true })

	assert.True(t, interceptor.budgetAllowRetry(time.Now()))
	assert.False(t, interceptor.budgetAllowRetry(time.Now().Add(-8*time.Second)))
}

func TestRetryOnError(t *testing.T) {
	checked := false
	interceptor := NewRetryInterceptor(2*time.Second, func(err error) bool {
		checked = true
		return err.Error() == "test retry on error"
	})

	unaryInfo := &grpc.UnaryServerInfo{
		FullMethod: "TestService.UnaryMethod",
	}
	unaryHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
		if checked {
			return nil, nil
		}
		return nil, fmt.Errorf("test retry on error")
	}

	_, err := interceptor.UnaryServerInterceptor(context.TODO(), "test", unaryInfo, unaryHandler)
	assert.NoError(t, err)

	assert.True(t, checked)
}

func TestDoNotRetryHardFail(t *testing.T) {
	checked := false
	interceptor := NewRetryInterceptor(2*time.Second, func(err error) bool {
		checked = true
		return false
	})

	unaryInfo := &grpc.UnaryServerInfo{
		FullMethod: "TestService.UnaryMethod",
	}
	unaryHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
		if checked {
			return nil, nil
		}
		return nil, fmt.Errorf("test retry on error")
	}

	_, err := interceptor.UnaryServerInterceptor(context.TODO(), "test", unaryInfo, unaryHandler)
	assert.Error(t, err)

	assert.True(t, checked)
}
