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
