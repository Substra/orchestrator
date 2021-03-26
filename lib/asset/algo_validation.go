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

// Validate returns an error if the new algo is not valid:
// missing required data, incompatible values, etc.
func (a *NewAlgo) Validate() error {
	return validation.ValidateStruct(a,
		validation.Field(&a.Key, validation.Required, is.UUID),
		validation.Field(&a.Name, validation.Required, validation.Length(1, 100)),
		validation.Field(&a.Category, validation.In(AlgoCategory_SIMPLE, AlgoCategory_COMPOSITE, AlgoCategory_AGGREGATE)),
		validation.Field(&a.Description, validation.Required),
		validation.Field(&a.Algorithm, validation.Required),
		validation.Field(&a.Metadata, validation.Each(validation.Length(0, 100))),
		validation.Field(&a.NewPermissions, validation.Required),
	)
}
