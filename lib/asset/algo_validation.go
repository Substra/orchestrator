package asset

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	is "github.com/go-ozzo/ozzo-validation/v4/is"
)

// Validate returns an error if the new algo is not valid:
// missing required data, incompatible values, etc.
func (a *NewAlgo) Validate() error {
	return validation.ValidateStruct(a,
		validation.Field(&a.Key, validation.Required, is.UUID),
		validation.Field(&a.Name, validation.Required, validation.Length(1, 100)),
		validation.Field(&a.Category, validation.In(AlgoCategory_ALGO_SIMPLE, AlgoCategory_ALGO_COMPOSITE, AlgoCategory_ALGO_AGGREGATE, AlgoCategory_ALGO_METRIC)),
		validation.Field(&a.Description, validation.Required),
		validation.Field(&a.Algorithm, validation.Required),
		validation.Field(&a.Metadata, validation.By(validateMetadata)),
		validation.Field(&a.NewPermissions, validation.Required),
	)
}
