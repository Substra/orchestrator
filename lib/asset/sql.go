package asset

import (
	"database/sql/driver"

	"github.com/owkin/orchestrator/lib/errors"
	"google.golang.org/protobuf/encoding/protojson"
)

// Value implements the driver.Valuer interface.
// Simply returns the JSON-encoded representation of the Node.
func (n *Node) Value() (driver.Value, error) {
	return protojson.Marshal(n)
}

// Scan implements the sql.Scanner interface.
// Simply decodes JSON into Node.
func (n *Node) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.NewError(errors.ErrByteArray, "cannot scan node")
	}

	return protojson.Unmarshal(b, n)
}

// Value implements the driver.Valuer interface.
// Simply returns the JSON-encoded representation of the Metric.
func (o *Metric) Value() (driver.Value, error) {
	return protojson.Marshal(o)
}

// Scan implements the sql.Scanner interface.
// Simply decodes JSON into the Metric.
func (o *Metric) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.NewError(errors.ErrByteArray, "cannot scan metric")
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
		return errors.NewError(errors.ErrByteArray, "cannot scan datasample")
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
		return errors.NewError(errors.ErrByteArray, "cannot scan algo")
	}

	return protojson.Unmarshal(b, a)
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
// Simply returns the JSON-encoded representation of the ComputeTask.
func (a *ComputeTask) Value() (driver.Value, error) {
	return protojson.Marshal(a)
}

// Scan implements the sql.Scanner interface.
// Simply decodes JSON into the ComputeTask.
func (a *ComputeTask) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.NewError(errors.ErrByteArray, "cannot scan compute task")
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
