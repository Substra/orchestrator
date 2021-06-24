package asset

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

// Validate returns an error if the new objective is not valid:
// missing required data, incompatible values, etc.
func (o *NewObjective) Validate() error {
	return validation.ValidateStruct(o,
		validation.Field(&o.Key, validation.Required, is.UUID),
		validation.Field(&o.Name, validation.Required, validation.Length(1, 100)),
		validation.Field(&o.MetricsName, validation.Required, validation.Length(1, 100)),
		validation.Field(&o.Metadata, validation.By(validateMetadata)),
		validation.Field(&o.Description, validation.Required),
		validation.Field(&o.NewPermissions, validation.Required),
		validation.Field(&o.Metrics, validation.Required),
		validation.Field(&o.DataManagerKey, is.UUID),
		validation.Field(&o.DataSampleKeys, validation.When(o.DataManagerKey != "", validation.Each(validation.Required, is.UUID))),
	)
}
