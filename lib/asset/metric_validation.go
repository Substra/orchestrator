package asset

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

// Validate returns an error if the new metric is not valid:
// missing required data, incompatible values, etc.
func (o *NewMetric) Validate() error {
	return validation.ValidateStruct(o,
		validation.Field(&o.Key, validation.Required, is.UUID),
		validation.Field(&o.Name, validation.Required, validation.Length(1, 100)),
		validation.Field(&o.Metadata, validation.By(validateMetadata)),
		validation.Field(&o.Description, validation.Required),
		validation.Field(&o.NewPermissions, validation.Required),
		validation.Field(&o.Address, validation.Required),
	)
}
