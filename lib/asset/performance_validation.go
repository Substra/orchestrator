package asset

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	is "github.com/go-ozzo/ozzo-validation/v4/is"
)

// Validate returns an error if the new function is not valid:
// missing required data, incompatible values, etc.
func (a *NewPerformance) Validate() error {
	return validation.ValidateStruct(a,
		validation.Field(&a.ComputeTaskKey, validation.Required, is.UUID),
		validation.Field(&a.ComputeTaskOutputIdentifier, validation.Required),
	)
}
