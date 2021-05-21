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

package interceptors

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/log/v7"
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

// ProviderInterceptor intercepts gRPC requests and assign a request-scoped orchestration.Provider
// to the request context.
type ProviderInterceptor struct {
	amqp         common.AMQPPublisher
	dbalProvider dbal.TransactionalDBALProvider
}

type ctxProviderInterceptorMarker struct{}

var ctxProviderKey = &ctxProviderInterceptorMarker{}

// NewProviderInterceptor returns an instance of ProviderInterceptor
func NewProviderInterceptor(dbalProvider dbal.TransactionalDBALProvider, amqp common.AMQPPublisher) *ProviderInterceptor {
	return &ProviderInterceptor{
		amqp:         amqp,
		dbalProvider: dbalProvider,
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

	tx, err := pi.dbalProvider.GetTransactionalDBAL(channel)
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
		dispatchErr := dispatcher.Dispatch()
		if dispatchErr != nil {
			log.WithError(dispatchErr).
				WithField("events", dispatcher.GetEvents()).
				Error("failed to dispatch events after successful transaction commit")
			return nil, fmt.Errorf("failed to dispatch events: %w", dispatchErr)
		}
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
