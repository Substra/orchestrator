package common

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const headerChannel = "channel"

// ChannelInterceptor intercepts gRPC requests and makes the channel from headers available to request context.
// It will return an error if a caller attempts to call the orchestrator for a channel it is not part of.
type ChannelInterceptor struct {
	orgChannels map[string][]string
}

// NewChannelInterceptor creates a ChannelInterceptor which will enforce organization & channel consistency.
// ChannelInterceptor MUST come after the mspid interceptor.
func NewChannelInterceptor(config *OrchestratorConfiguration) *ChannelInterceptor {
	orgChannels := make(map[string][]string)

	for channel, orgs := range config.Channels {
		for _, org := range orgs {
			orgChannels[org] = append(orgChannels[org], channel)
		}
	}

	return &ChannelInterceptor{
		orgChannels: orgChannels,
	}
}

// UnaryServerInterceptor will make the channel from headers available from the request context.
func (i *ChannelInterceptor) UnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// Passthrough for ignored methods
	for _, m := range ignoredMethods {
		if strings.Contains(info.FullMethod, m) {
			return handler(ctx, req)
		}
	}

	newCtx, err := i.extractFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return handler(newCtx, req)
}

// StreamServerInterceptor will make the channel from headers available from the request context.
func (i *ChannelInterceptor) StreamServerInterceptor(
	srv interface{},
	stream grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	newCtx, err := i.extractFromContext(stream.Context())
	if err != nil {
		return err
	}
	streamWithContext := BindStreamToContext(newCtx, stream)

	return handler(srv, streamWithContext)
}

func (i *ChannelInterceptor) extractFromContext(ctx context.Context) (context.Context, error) {
	org, err := ExtractMSPID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to extract organization: %w", err)
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("could not extract metadata")
	}

	if len(md.Get(headerChannel)) != 1 {
		return nil, fmt.Errorf("missing or invalid header '%s'", headerChannel)
	}

	channel := md.Get(headerChannel)[0]

	if err := i.checkOrgBelongsToChannel(org, channel); err != nil {
		return nil, err
	}

	return WithChannel(ctx, channel), nil
}

func (i *ChannelInterceptor) checkOrgBelongsToChannel(org, channel string) error {
	channels, ok := i.orgChannels[org]
	if !ok {
		return fmt.Errorf("organization \"%s\" is unknown", org)
	}

	for _, c := range channels {
		if channel == c {
			return nil
		}
	}

	return fmt.Errorf("organization \"%s\" has not access to channel \"%s\"", org, channel)
}

type ctxChannelMarker struct{}

var (
	ctxChannelKey = &ctxChannelMarker{}
)

// WithChannel add channel information to a context
func WithChannel(ctx context.Context, channel string) context.Context {
	return context.WithValue(ctx, ctxChannelKey, channel)
}

// ExtractChannel retrieves channel from request context
// channel is expected to be set by InterceptChannel
func ExtractChannel(ctx context.Context) (string, error) {
	channel, ok := ctx.Value(ctxChannelKey).(string)
	if !ok {
		return "", errors.New("channel not found in context")
	}
	return channel, nil
}
