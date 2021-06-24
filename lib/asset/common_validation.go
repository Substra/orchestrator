package asset

import (
	"fmt"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/owkin/orchestrator/lib/errors"
)

// Validate makes sure the Addressable object is valid
func (a *Addressable) Validate() error {
	return validation.ValidateStruct(a,
		validation.Field(&a.Checksum, validation.Required, validation.Length(64, 64), is.Hexadecimal),
		validation.Field(&a.StorageAddress, validation.Required, is.URL),
	)
}

// Validate makes sure the Permissions object is valid
func (p *Permissions) Validate() error {
	return validation.ValidateStruct(p,
		validation.Field(&p.Process, validation.Required),
	)
}

func validateMetadata(input interface{}) error {
	metadata, ok := input.(map[string]string)
	if !ok {
		return fmt.Errorf("metadata is not a proper map %w", errors.ErrInvalidAsset)
	}

	for k, v := range metadata {
		if len(k) > 100 {
			return fmt.Errorf("metadata key '%v' is too long, %w", k, errors.ErrInvalidAsset)
		}
		if len(v) > 100 {
			return fmt.Errorf("metadata value for key '%v' is too long, %w", k, errors.ErrInvalidAsset)
		}
	}

	return nil
}
