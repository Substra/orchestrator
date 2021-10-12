package asset

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	is "github.com/go-ozzo/ozzo-validation/v4/is"
)

// Validate returns an error if the new DataManager is not valid
func (d *NewDataManager) Validate() error {

	return validation.ValidateStruct(d,
		validation.Field(&d.Key, validation.Required, is.UUID),
		validation.Field(&d.Name, validation.Required, validation.Length(1, 100)),
		validation.Field(&d.NewPermissions, validation.Required),
		validation.Field(&d.Description, validation.Required),
		validation.Field(&d.Opener, validation.Required),
		validation.Field(&d.Metadata, validation.By(validateMetadata)),
		validation.Field(&d.Type, validation.Required, validation.Length(1, 30)),
	)
}
