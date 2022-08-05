package interceptors

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestGetRequestID(t *testing.T) {
	ctx := context.Background()

	reqID := GetRequestID(ctx)
	assert.Equal(t, "", reqID, "missing request ID should not fail")

	ctxWithReqID := context.WithValue(ctx, RequestIDMarker, "test")
	assert.Equal(t, "test", GetRequestID(ctxWithReqID))
}

func TestGenerateID(t *testing.T) {
	unaryInfo := &grpc.UnaryServerInfo{
		FullMethod: "TestService.UnaryMethod",
	}
	unaryHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
		reqID := GetRequestID(ctx)
		assert.NotEqual(t, "", reqID)
		return "test", nil
	}

	ctx := metadata.NewIncomingContext(context.Background(), metadata.New(nil))

	_, err := InterceptRequestID(ctx, "test", unaryInfo, unaryHandler)
	assert.NoError(t, err)
}

func TestExtractID(t *testing.T) {
	unaryInfo := &grpc.UnaryServerInfo{
		FullMethod: "TestService.UnaryMethod",
	}
	unaryHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
		reqID := GetRequestID(ctx)
		assert.Equal(t, "testid", reqID)
		return "test", nil
	}

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("reqid", "testid"))

	_, err := InterceptRequestID(ctx, "test", unaryInfo, unaryHandler)
	assert.NoError(t, err)
}
