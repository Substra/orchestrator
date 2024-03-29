package interceptors

import (
	"context"
	"errors"
	"strings"

	"github.com/rs/zerolog/log"
	orcerrors "github.com/substra/orchestrator/lib/errors"
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

		log.Ctx(ctx).Error().
			Str("method", info.FullMethod).
			Stack().
			Err(err).
			Str("grpcCode", status.Code(grpcError).String()).
			Msg("Error response")

		return res, grpcError
	}

	return res, nil
}

// fromMessage converts an error to a gRPC status by matching its error message
func fromMessage(msg string) error {
	switch {
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
	case strings.Contains(msg, orcerrors.ErrMissingTaskOutput):
		return status.Error(codes.InvalidArgument, msg)
	case strings.Contains(msg, orcerrors.ErrIncompatibleKind):
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

	switch {
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
	case orcError.Kind == orcerrors.ErrMissingTaskOutput:
		return status.Error(codes.InvalidArgument, err.Error())
	case orcError.Kind == orcerrors.ErrIncompatibleKind:
		return status.Error(codes.InvalidArgument, err.Error())
	case orcError.Kind == orcerrors.ErrInternal:
		return status.Error(codes.Internal, err.Error())
	default:
		return status.Error(codes.Unknown, err.Error())
	}
}
