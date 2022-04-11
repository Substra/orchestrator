package asset

import (
	"database/sql/driver"

	"github.com/owkin/orchestrator/lib/errors"
	"google.golang.org/protobuf/encoding/protojson"
)

// Value implements the driver.Valuer interface.
// Simply returns the JSON-encoded representation of the Permissions.
func (p *Permissions) Value() (driver.Value, error) {
	return protojson.Marshal(p)
}

// Scan implements the sql.Scanner interface.
// Simply decodes JSON into the Permissions.
func (p *Permissions) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.NewError(errors.ErrByteArray, "cannot scan permissions")
	}

	return protojson.Unmarshal(b, p)
}

// Value implements the driver.Valuer interface.
// Simply returns the JSON-encoded representation of the Permission.
func (p *Permission) Value() (driver.Value, error) {
	return protojson.Marshal(p)
}

// Scan implements the sql.Scanner interface.
// Simply decodes JSON into the Permission.
func (p *Permission) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.NewError(errors.ErrByteArray, "cannot scan permission")
	}

	return protojson.Unmarshal(b, p)
}

// Value implements the driver.Valuer interface.
// Simply returns the string representation of the AlgoCategory.
func (c *AlgoCategory) Value() (driver.Value, error) {
	return c.String(), nil
}

// Scan implements the sql.Scanner interface.
// Simply decodes a string into the AlgoCategory.
func (c *AlgoCategory) Scan(value interface{}) error {
	s, ok := value.(string)
	if !ok {
		return errors.NewError(errors.ErrInternal, "cannot scan algo category")
	}

	v, ok := AlgoCategory_value[s]
	if !ok {
		return errors.NewError(errors.ErrInternal, "cannot scan algo category")
	}
	*c = AlgoCategory(v)

	return nil
}

// Value implements the driver.Valuer interface.
// Simply returns the string representation of the ComputeTaskStatus.
func (ts *ComputeTaskStatus) Value() (driver.Value, error) {
	return ts.String(), nil
}

// Scan implements the sql.Scanner interface.
// Simply decodes a string into the ComputeTaskStatus.
func (ts *ComputeTaskStatus) Scan(value interface{}) error {
	s, ok := value.(string)
	if !ok {
		return errors.NewInternal("cannot scan task status: invalid string")
	}

	v, ok := ComputeTaskStatus_value[s]
	if !ok {
		return errors.NewInternal("cannot scan task status: unknown value")
	}
	*ts = ComputeTaskStatus(v)

	return nil
}

// Value implements the driver.Valuer interface.
// Simply returns the string representation of the ComputeTaskCategory.
func (c *ComputeTaskCategory) Value() (driver.Value, error) {
	return c.String(), nil
}

// Scan implements the sql.Scanner interface.
// Simply decodes a string into the ComputeTaskCategory.
func (c *ComputeTaskCategory) Scan(value interface{}) error {
	s, ok := value.(string)
	if !ok {
		return errors.NewInternal("cannot scan task category: invalid string")
	}

	v, ok := ComputeTaskCategory_value[s]
	if !ok {
		return errors.NewInternal("cannot scan task category: unknown value")
	}
	*c = ComputeTaskCategory(v)

	return nil
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
		return errors.NewError(errors.ErrByteArray, "cannot scan datasample")
	}

	return protojson.Unmarshal(b, d)
}

// Scan implements the sql.Scanner interface.
// Decodes JSON into the DataManager
func (d *DataManager) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.NewError(errors.ErrByteArray, "cannot scan datamanager")
	}

	return protojson.Unmarshal(b, d)
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
		return errors.NewError(errors.ErrByteArray, "cannot scan model")
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
		return errors.NewError(errors.ErrByteArray, "cannot scan compute plan")
	}

	return protojson.Unmarshal(b, cp)
}

// Value implements the driver.Valuer interface.
// Simply returns the JSON-encoded representation of the Performance.
func (p *Performance) Value() (driver.Value, error) {
	return protojson.Marshal(p)
}

// Scan implements the sql.Scanner interface.
// Simply decodes JSON into the Performance.
func (p *Performance) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.NewError(errors.ErrByteArray, "cannot scan performance")
	}

	return protojson.Unmarshal(b, p)
}

// Value simply returns the JSON-encoded representation of the Event.
func (e *Event) Value() (driver.Value, error) {
	return protojson.Marshal(e)
}

// Scan implements the sql.Scanner interface.
// Simply decodes JSON into the Event.
func (e *Event) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.NewError(errors.ErrByteArray, "cannot scan event")
	}

	return protojson.Unmarshal(b, e)
}

// Value implements the driver.Valuer interface.
// Simply returns the JSON-encoded representation of the FailureReport.
func (f *FailureReport) Value() (driver.Value, error) {
	return protojson.Marshal(f)
}

// Scan implements the sql.Scanner interface.
// Simply decodes JSON into the FailureReport.
func (f *FailureReport) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.NewError(errors.ErrByteArray, "cannot scan failure report")
	}

	return protojson.Unmarshal(b, f)
}
