// Copyright 2021 Owkin Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
