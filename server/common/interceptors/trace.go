package interceptors

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const headerRequestID = "reqid"

type requestIDMarker struct{}

// RequestIDMarker is the identifier of the RequestID in context.
var RequestIDMarker = &requestIDMarker{}

// InterceptRequestID adds a unique identifier to the context.
// This identifier is retrieved from query header "reqid" and is generated if there is no such header.
func InterceptRequestID(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("could not extract metadata")
	}

	var requestID string

	if len(md.Get(headerRequestID)) == 1 {
		requestID = md.Get(headerRequestID)[0]
	} else {
		u, err := uuid.NewRandom()
		if err != nil {
			return nil, err
		}
		requestID = fmt.Sprintf("%v", u)[:8]
	}

	newCtx := context.WithValue(ctx, RequestIDMarker, requestID)
	return handler(newCtx, req)
}

// GetRequestID extracts the request ID from context.
// Returns an empty string if context does not contain an ID.
func GetRequestID(ctx context.Context) string {
	reqID, ok := ctx.Value(RequestIDMarker).(string)
	if !ok {
		reqID = ""
	}

	return reqID
}
