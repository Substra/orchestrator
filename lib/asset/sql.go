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
	"database/sql/driver"
	"fmt"

	"github.com/owkin/orchestrator/lib/errors"
	"google.golang.org/protobuf/encoding/protojson"
)

// Value implements the driver.Valuer interface.
// Simply returns the JSON-encoded representation of the Objective.
func (o *Objective) Value() (driver.Value, error) {
	return protojson.Marshal(o)
}

// Scan implements the sql.Scanner interface.
// Simply decodes JSON into the Objective.
func (o *Objective) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan objective: %w", errors.ErrByteArray)
	}

	return protojson.Unmarshal(b, o)
}

// Value implements the driver.Valuer interface.
// Returns the JSON-encoded representation of the DataSample.
func (d *DataSample) Value() (driver.Value, error) {
	return protojson.Marshal(d)
}

// Value implements the driver.Valuer interface.
// Returns the JSON-encoded representation of the DataManager.
func (d *DataManager) Value() (driver.Value, error) {
	return protojson.Marshal(d)
}

// Scan implements the sql.Scanner interface.
// Decodes JSON into a DataSample.
func (d *DataSample) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan datasample: %w", errors.ErrByteArray)
	}

	return protojson.Unmarshal(b, d)
}

// Value implements the driver.Valuer interface.
// Simply returns the JSON-encoded representation of the Algo.
func (a *Algo) Value() (driver.Value, error) {
	return protojson.Marshal(a)
}

// Scan implements the sql.Scanner interface.
// Simply decodes JSON into the Algo.
func (a *Algo) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan algo: %w", errors.ErrByteArray)
	}

	return protojson.Unmarshal(b, a)
}

// Scan implements the sql.Scanner interface.
// Decodes JSON into the DataManager
func (d *DataManager) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan datamanager: %w", errors.ErrByteArray)
	}

	return protojson.Unmarshal(b, d)
}

// Value implements the driver.Valuer interface.
// Simply returns the JSON-encoded representation of the ComputeTask.
func (a *ComputeTask) Value() (driver.Value, error) {
	return protojson.Marshal(a)
}

// Scan implements the sql.Scanner interface.
// Simply decodes JSON into the ComputeTask.
func (a *ComputeTask) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan compute task: %w", errors.ErrByteArray)
	}

	return protojson.Unmarshal(b, a)
}

// Value implements the driver.Valuer interface.
// Simply returns the JSON-encoded representation of the Model.
func (a *Model) Value() (driver.Value, error) {
	return protojson.Marshal(a)
}

// Scan implements the sql.Scanner interface.
// Simply decodes JSON into the Model.
func (a *Model) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan model: %w", errors.ErrByteArray)
	}

	return protojson.Unmarshal(b, a)
}

// Value implements the driver.Valuer interface.
// Simply returns the JSON-encoded representation of the ComputePlan.
func (cp *ComputePlan) Value() (driver.Value, error) {
	return protojson.Marshal(cp)
}

// Scan implements the sql.Scanner interface.
// Simply decodes JSON into the ComputePlan.
func (cp *ComputePlan) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan compute plan: %w", errors.ErrByteArray)
	}

	return protojson.Unmarshal(b, cp)
}
