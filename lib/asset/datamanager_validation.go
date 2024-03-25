package asset

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	is "github.com/go-ozzo/ozzo-validation/v4/is"
)

// Validate returns an error if the new DataManager is not valid
func (d *NewDataManager) Validate() error {

	return validation.ValidateStruct(d,
		validation.Field(&d.Key, validation.Required, is.UUID),
		validation.Field(&d.Name, nameValidationRules...),
		validation.Field(&d.NewPermissions, validation.Required),
		validation.Field(&d.Description, validation.Required),
		validation.Field(&d.Opener, validation.Required),
		validation.Field(&d.Metadata, validation.By(validateMetadata)),
		validation.Field(&d.LogsPermission, validation.Required),
	)
}

// Validate returns an error if the updated function is not valid:
// missing required data, incompatible values, etc.
func (o *UpdateDataManagerParam) Validate() error {
	return validation.ValidateStruct(o,
		validation.Field(&o.Key, validation.Required, is.UUID),
		validation.Field(&o.Name, nameValidationRules...),
	)
}
