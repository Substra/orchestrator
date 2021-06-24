package asset

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

// Validate returns an error if the new model is not valid:
// missing required data, incompatible values, etc.
func (m *NewModel) Validate() error {
	return validation.ValidateStruct(m,
		validation.Field(&m.Key, validation.Required, is.UUID),
		validation.Field(&m.ComputeTaskKey, validation.Required, is.UUID),
		validation.Field(&m.Category, validation.Required, validation.In(ModelCategory_MODEL_SIMPLE, ModelCategory_MODEL_HEAD, ModelCategory_MODEL_TRUNK)),
		validation.Field(&m.Address, validation.Required),
	)
}
