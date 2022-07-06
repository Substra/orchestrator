package asset

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

// Validate returns an error if the new organization is not valid
func (o *RegisterOrganizationParam) Validate() error {
	return validation.ValidateStruct(o,
		validation.Field(&o.Address, is.RequestURL),
	)
}
