// Copyright 2020 Owkin Inc.
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

package objective

import (
	"github.com/go-ozzo/ozzo-validation/v4"
)

// Validate returns an error if the objective is not valid:
// missing required data, incompatible values, etc.
func (o *Objective) Validate() error {
	return validation.ValidateStruct(o,
		validation.Field(&o.Key, validation.Required, validation.Length(36, 36)),
		validation.Field(&o.Name, validation.Required, validation.Length(1, 100)),
		validation.Field(&o.MetricsName, validation.Required, validation.Length(1, 100)),
		// validate structure against their own rules
		validation.Field(&o.Description),
		validation.Field(&o.Permissions),
		validation.Field(&o.Metrics),
	)
}
