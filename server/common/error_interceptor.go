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
	"strings"

	orchestrationErrors "github.com/owkin/orchestrator/lib/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// InterceptErrors is a gRPC interceptor which converts orchestration errors into nice gRPC ones.
// This allows clients to properly take action based on the returned status.
func InterceptErrors(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	res, err := handler(ctx, req)

	// Passthrough for ignored methods
	for _, m := range ignoredMethods {
		if strings.Contains(info.FullMethod, m) {
			return res, err
		}
	}

	return res, toStatus(err)
}

// toStatus converts an error to a gRPC status
func toStatus(err error) error {
	switch true {
	case err == nil:
		return nil
	case errors.Is(err, orchestrationErrors.ErrInvalidAsset):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, orchestrationErrors.ErrConflict):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, orchestrationErrors.ErrPermissionDenied):
		return status.Error(codes.PermissionDenied, err.Error())
	default:
		return status.Error(codes.Unknown, err.Error())
	}
}
