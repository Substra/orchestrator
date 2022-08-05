package interceptors

import (
	"context"
	"errors"
	"strings"

	orcerrors "github.com/owkin/orchestrator/lib/errors"
	"github.com/owkin/orchestrator/server/common/logger"
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
	for _, m := range IgnoredMethods {
		if strings.Contains(info.FullMethod, m) {
			return res, err
		}
	}

	if err != nil {
		grpcError := fromError(err)

		log := logger.Get(ctx).WithField("method", info.FullMethod).WithError(err)
		log.WithField("grpcCode", status.Code(grpcError)).Error("Error response")

		return res, grpcError
	}

	return res, nil
}

// InterceptDistributedErrors is a gRPC interceptor which converts orchestration errors into nice gRPC ones.
// This allows clients to properly take action based on the returned status.
// In distributed mode, errors returned by the chaincode are generic: our only way to distinguish them is to look at the message.
// This interceptor attempts to set an appropriate error return code by matching the message against known orchestration errors.
func InterceptDistributedErrors(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	res, err := handler(ctx, req)

	// Passthrough for ignored methods
	for _, m := range IgnoredMethods {
		if strings.Contains(info.FullMethod, m) {
			return res, err
		}
	}

	var wrappedErr error
	if err != nil {
		wrappedErr = fromMessage(err.Error())
		log := logger.Get(ctx).WithField("method", info.FullMethod).WithError(err)
		log.WithField("grpcCode", status.Code(wrappedErr)).Error("Error response")
	}

	return res, wrappedErr
}

// fromMessage converts an error to a gRPC status by matching its error message
func fromMessage(msg string) error {
	switch true {
	case strings.Contains(msg, orcerrors.ErrInvalidAsset):
		return status.Error(codes.InvalidArgument, msg)
	case strings.Contains(msg, orcerrors.ErrConflict):
		return status.Error(codes.AlreadyExists, msg)
	case strings.Contains(msg, orcerrors.ErrPermissionDenied):
		return status.Error(codes.PermissionDenied, msg)
	case strings.Contains(msg, orcerrors.ErrNotFound):
		return status.Error(codes.NotFound, msg)
	case strings.Contains(msg, orcerrors.ErrBadRequest):
		return status.Error(codes.FailedPrecondition, msg)
	case strings.Contains(msg, orcerrors.ErrIncompatibleTaskStatus):
		return status.Error(codes.InvalidArgument, msg)
	case strings.Contains(msg, orcerrors.ErrUnimplemented):
		return status.Error(codes.Unimplemented, msg)
	case strings.Contains(msg, orcerrors.ErrCannotDisableModel):
		return status.Error(codes.InvalidArgument, msg)
	case strings.Contains(msg, orcerrors.ErrInternal):
		return status.Error(codes.Internal, msg)
	default:
		return status.Error(codes.Unknown, msg)
	}
}

// fromError converts an error to a gRPC status by matching its error type
func fromError(err error) error {
	if err == nil {
		return nil
	}

	orcError := new(orcerrors.OrcError)
	if !errors.As(err, &orcError) {
		return status.Error(codes.Unknown, err.Error())
	}

	switch true {
	case orcError.Kind == orcerrors.ErrInvalidAsset:
		return status.Error(codes.InvalidArgument, err.Error())
	case orcError.Kind == orcerrors.ErrConflict:
		return status.Error(codes.AlreadyExists, err.Error())
	case orcError.Kind == orcerrors.ErrPermissionDenied:
		return status.Error(codes.PermissionDenied, err.Error())
	case orcError.Kind == orcerrors.ErrNotFound:
		return status.Error(codes.NotFound, err.Error())
	case orcError.Kind == orcerrors.ErrBadRequest:
		return status.Error(codes.FailedPrecondition, err.Error())
	case orcError.Kind == orcerrors.ErrIncompatibleTaskStatus:
		return status.Error(codes.InvalidArgument, err.Error())
	case orcError.Kind == orcerrors.ErrUnimplemented:
		return status.Error(codes.Unimplemented, err.Error())
	case orcError.Kind == orcerrors.ErrCannotDisableModel:
		return status.Error(codes.InvalidArgument, err.Error())
	case orcError.Kind == orcerrors.ErrInternal:
		return status.Error(codes.Internal, err.Error())
	default:
		return status.Error(codes.Unknown, err.Error())
	}
}
