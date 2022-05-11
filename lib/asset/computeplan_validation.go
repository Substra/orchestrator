package asset

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

// Validate returns an error if the NewComputePlan is not valid:
// missing required data, incompatible values, etc.
func (t *NewComputePlan) Validate() error {
	return validation.ValidateStruct(t,
		validation.Field(&t.Key, validation.Required, is.UUID),
		validation.Field(&t.Tag, validation.Length(0, 100)),
		validation.Field(&t.Name, validation.Required, validation.Length(1, 100)),
		validation.Field(&t.Metadata, validation.By(validateMetadata)),
	)
}
