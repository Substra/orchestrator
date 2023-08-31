package asset

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/substra/orchestrator/utils"
)

// Validate returns an error if the new FailureReport object is not valid.
func (f *NewFailureReport) Validate() error {
	return validation.ValidateStruct(f,
		validation.Field(&f.AssetKey, validation.Required, is.UUID),
		validation.Field(&f.ErrorType, validation.In(ErrorType_ERROR_TYPE_BUILD, ErrorType_ERROR_TYPE_EXECUTION, ErrorType_ERROR_TYPE_INTERNAL)),
		validation.Field(&f.AssetType, validation.In(FailedAssetKind_FAILED_ASSET_UNKNOWN, FailedAssetKind_FAILED_ASSET_FUNCTION, FailedAssetKind_FAILED_ASSET_COMPUTE_TASK)),
		validation.Field(&f.LogsAddress, validation.When(utils.SliceContains([]ErrorType{ErrorType_ERROR_TYPE_EXECUTION, ErrorType_ERROR_TYPE_BUILD}, f.ErrorType), validation.Required).Else(validation.Nil)),
	)
}
