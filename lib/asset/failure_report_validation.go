package asset

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

// Validate returns an error if the new FailureReport object is not valid.
func (f *NewFailureReport) Validate() error {
	return validation.ValidateStruct(f,
		validation.Field(&f.ComputeTaskKey, validation.Required, is.UUID),
		validation.Field(&f.LogsAddress, validation.Required),
	)
}
