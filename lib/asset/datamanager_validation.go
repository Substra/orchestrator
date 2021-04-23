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
	validation "github.com/go-ozzo/ozzo-validation/v4"
	is "github.com/go-ozzo/ozzo-validation/v4/is"
)

// Validate returns an error if the new DataManager is not valid
func (d *NewDataManager) Validate() error {

	return validation.ValidateStruct(d,
		validation.Field(&d.Key, validation.Required, is.UUID),
		validation.Field(&d.Name, validation.Required, validation.Length(1, 100)),
		validation.Field(&d.NewPermissions, validation.Required),
		validation.Field(&d.ObjectiveKey, validation.When(d.GetObjectiveKey() != ""), is.UUID),
		validation.Field(&d.Description, validation.Required),
		validation.Field(&d.Opener, validation.Required),
		validation.Field(&d.Metadata, validation.By(validateMetadata)),
		validation.Field(&d.Type, validation.Required, validation.Length(1, 30)),
	)
}

// Validate returns an error if the DataManagerUpdate is not valid
func (d *DataManagerUpdateParam) Validate() error {
	return validation.ValidateStruct(d,
		validation.Field(&d.Key, validation.Required, is.UUID),
		validation.Field(&d.ObjectiveKey, validation.Required, is.UUID),
	)
}
