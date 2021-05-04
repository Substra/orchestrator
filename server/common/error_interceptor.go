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

// Although we are in common module, this file contains two separate implementations for error interception.
// This is because the two should be kept in sync and share the same tests, which is easier if they live in the same module.

// InterceptStandaloneErrors is a gRPC interceptor which converts orchestration errors into nice gRPC ones.
// This allows clients to properly take action based on the returned status.
func InterceptStandaloneErrors(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	res, err := handler(ctx, req)

	// Passthrough for ignored methods
	for _, m := range ignoredMethods {
		if strings.Contains(info.FullMethod, m) {
			return res, err
		}
	}

	return res, fromError(err)
}

// InterceptDistributedErrors is a gRPC interceptor which converts orchestration errors into nice gRPC ones.
// This allows clients to properly take action based on the returned status.
// In distributed mode, errors returned by the chaincode are generic: our only way to distinguish them is to look at the message.
// This interceptor attempts to set an appropriate error return code by matching the message against known orchestration errors.
func InterceptDistributedErrors(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	res, err := handler(ctx, req)

	// Passthrough for ignored methods
	for _, m := range ignoredMethods {
		if strings.Contains(info.FullMethod, m) {
			return res, err
		}
	}

	var wrappedErr error
	if err != nil {
		wrappedErr = fromMessage(err.Error())
	}

	return res, wrappedErr
}

// fromMessage converts an error to a gRPC status by matching its error message
func fromMessage(msg string) error {
	switch true {
	case strings.Contains(msg, orchestrationErrors.ErrInvalidAsset.Error()):
		return status.Error(codes.InvalidArgument, msg)
	case strings.Contains(msg, orchestrationErrors.ErrConflict.Error()):
		return status.Error(codes.AlreadyExists, msg)
	case strings.Contains(msg, orchestrationErrors.ErrPermissionDenied.Error()):
		return status.Error(codes.PermissionDenied, msg)
	case strings.Contains(msg, orchestrationErrors.ErrReferenceNotFound.Error()):
		return status.Error(codes.InvalidArgument, msg)
	case strings.Contains(msg, orchestrationErrors.ErrNotFound.Error()):
		return status.Error(codes.NotFound, msg)
	case strings.Contains(msg, orchestrationErrors.ErrBadRequest.Error()):
		return status.Error(codes.FailedPrecondition, msg)
	case strings.Contains(msg, orchestrationErrors.ErrIncompatibleTaskStatus.Error()):
		return status.Error(codes.InvalidArgument, msg)
	case strings.Contains(msg, orchestrationErrors.ErrUnimplemented.Error()):
		return status.Error(codes.Unimplemented, msg)
	case strings.Contains(msg, orchestrationErrors.ErrCannotDisableModel.Error()):
		return status.Error(codes.InvalidArgument, msg)
	default:
		return status.Error(codes.Unknown, msg)
	}
}

// fromError converts an error to a gRPC status by matching its error type
func fromError(err error) error {
	switch true {
	case err == nil:
		return nil
	case errors.Is(err, orchestrationErrors.ErrInvalidAsset):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, orchestrationErrors.ErrConflict):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, orchestrationErrors.ErrPermissionDenied):
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, orchestrationErrors.ErrReferenceNotFound):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, orchestrationErrors.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, orchestrationErrors.ErrBadRequest):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, orchestrationErrors.ErrIncompatibleTaskStatus):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, orchestrationErrors.ErrUnimplemented):
		return status.Error(codes.Unimplemented, err.Error())
	case errors.Is(err, orchestrationErrors.ErrCannotDisableModel):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Unknown, err.Error())
	}
}
