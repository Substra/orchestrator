package asset

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"golang.org/x/exp/slices"
)

// Validate returns an error if the new FailureReport object is not valid.
func (f *NewFailureReport) Validate() error {
	return validation.ValidateStruct(f,
		validation.Field(&f.ComputeTaskKey, validation.Required, is.UUID),
		validation.Field(&f.ErrorType, validation.In(ErrorType_ERROR_TYPE_BUILD, ErrorType_ERROR_TYPE_EXECUTION, ErrorType_ERROR_TYPE_INTERNAL)),
		validation.Field(&f.LogsAddress, validation.When(slices.Contains([]ErrorType{ErrorType_ERROR_TYPE_EXECUTION, ErrorType_ERROR_TYPE_BUILD}, f.ErrorType), validation.Required).Else(validation.Nil)),
	)
}
