package asset

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	is "github.com/go-ozzo/ozzo-validation/v4/is"
)

// Validate returns an error if the profiling step is not valid:
// missing required data, incompatible values, etc.
func (a *ProfilingStep) Validate() error {
	return validation.ValidateStruct(a,
		validation.Field(&a.AssetKey, validation.Required, is.UUID),
		validation.Field(&a.Step, nameValidationRules...),
		validation.Field(&a.Duration, validation.Required),
	)
}
