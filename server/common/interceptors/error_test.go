package interceptors

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/substra/orchestrator/lib/asset"
	"github.com/substra/orchestrator/lib/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestStatusConversion(t *testing.T) {
	cases := map[string]struct {
		err  error
		code codes.Code
	}{
		"conflict":                 {err: errors.NewError(errors.ErrConflict, "test"), code: codes.AlreadyExists},
		"validation":               {err: errors.NewError(errors.ErrInvalidAsset, "test"), code: codes.InvalidArgument},
		"unknown":                  {err: fmt.Errorf("some unknown error"), code: codes.Unknown},
		"unauthorized":             {err: errors.NewError(errors.ErrPermissionDenied, "test"), code: codes.PermissionDenied},
		"notfound":                 {err: errors.NewError(errors.ErrNotFound, "test"), code: codes.NotFound},
		"badrequest":               {err: errors.NewError(errors.ErrBadRequest, "test"), code: codes.FailedPrecondition},
		"incompatible_status":      {err: errors.NewError(errors.ErrIncompatibleTaskStatus, "test"), code: codes.InvalidArgument},
		"unimplemented":            {err: errors.NewError(errors.ErrUnimplemented, "test"), code: codes.Unimplemented},
		"unprocessable model":      {err: errors.NewError(errors.ErrCannotDisableModel, "test"), code: codes.InvalidArgument},
		"internal":                 {err: errors.NewInternal("test"), code: codes.Internal},
		"missing task output":      {err: errors.NewMissingTaskOutput("test", "output"), code: codes.InvalidArgument},
		"incompatible task output": {err: errors.NewIncompatibleTaskOutput("test", "output", asset.AssetKind_ASSET_MODEL.String(), asset.AssetKind_ASSET_PERFORMANCE.String()), code: codes.InvalidArgument},
	}

	for name, tc := range cases {
		t.Run(fmt.Sprintf("fromError(%s)", name), func(t *testing.T) {
			assert.Equal(t, tc.code, status.Convert(fromError(tc.err)).Code())
		})
		err := fmt.Errorf("new error with embedded code: %s in the message", tc.err.Error())
		t.Run(fmt.Sprintf("fromMessage(%s)", name), func(t *testing.T) {
			assert.Equal(t, tc.code, status.Convert(fromMessage(err.Error())).Code())
		})
	}

	assert.Nil(t, fromError(nil), "nil should not be mapped")
}
