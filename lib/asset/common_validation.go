package asset

import (
	"fmt"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/substra/orchestrator/lib/errors"
)

var nameValidationRules = []validation.Rule{
	validation.Required,
	validation.Length(1, 100),
}

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
		return errors.NewInvalidAsset("metadata is not a proper map")
	}

	for k, v := range metadata {
		if len(k) > 100 {
			return errors.NewInvalidAsset(fmt.Sprintf("metadata key %q is too long", k))
		}
		if len(v) > 100 {
			return errors.NewInvalidAsset(fmt.Sprintf("metadata value for key %q is too long", k))
		}
		if strings.Contains(k, "__") {
			return errors.NewInvalidAsset(fmt.Sprintf("'__' cannot be used in a metadata key, please use simple underscore instead for key %q", k))
		}
	}

	return nil
}
