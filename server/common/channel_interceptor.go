// Copyright 2021 Owkin Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

// InterceptChannel is a gRPC interceptor and will make the channel from headers available to request context.
func (i *ChannelInterceptor) InterceptChannel(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// Passthrough for ignored methods
	for _, m := range ignoredMethods {
		if strings.Contains(info.FullMethod, m) {
			return handler(ctx, req)
		}
	}

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

	newCtx := WithChannel(ctx, channel)
	return handler(newCtx, req)
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
