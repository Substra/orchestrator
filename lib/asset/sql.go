package asset

import (
	"database/sql/driver"

	"github.com/substra/orchestrator/lib/errors"
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
// Simply returns the string representation of the ErrorType.
func (e *ErrorType) Value() (driver.Value, error) {
	return e.String(), nil
}

// Scan implements the sql.Scanner interface.
// Simply decodes a string into the ErrorType.
func (e *ErrorType) Scan(value interface{}) error {
	s, ok := value.(string)
	if !ok {
		return errors.NewError(errors.ErrInternal, "cannot scan error type")
	}

	v, ok := ErrorType_value[s]
	if !ok {
		return errors.NewError(errors.ErrInternal, "cannot scan error type")
	}
	*e = ErrorType(v)

	return nil
}

// Value implements the driver.Valuer interface.
// Simply returns the string representation of the AssetKind.
func (k *AssetKind) Value() (driver.Value, error) {
	return k.String(), nil
}

// Scan implements the sql.Scanner interface.
// Simply decodes a string into the AssetKind.
func (k *AssetKind) Scan(value interface{}) error {
	s, ok := value.(string)
	if !ok {
		return errors.NewError(errors.ErrInternal, "cannot scan asset kind")
	}

	v, ok := AssetKind_value[s]
	if !ok {
		return errors.NewError(errors.ErrInternal, "cannot scan asset kind")
	}
	*k = AssetKind(v)

	return nil
}

// Value implements the driver.Valuer interface.
// Simply returns the string representation of the EventKind.
func (k *EventKind) Value() (driver.Value, error) {
	return k.String(), nil
}

// Scan implements the sql.Scanner interface.
// Simply decodes a string into the EventKind.
func (k *EventKind) Scan(value interface{}) error {
	s, ok := value.(string)
	if !ok {
		return errors.NewError(errors.ErrInternal, "cannot scan event kind")
	}

	v, ok := EventKind_value[s]
	if !ok {
		return errors.NewError(errors.ErrInternal, "cannot scan event kind")
	}
	*k = EventKind(v)

	return nil
}

// Value implements the driver.Valuer interface.
// Simply returns the string representation of the FunctionStatus.
func (ts *FunctionStatus) Value() (driver.Value, error) {
	return ts.String(), nil
}

// Scan implements the sql.Scanner interface.
// Simply decodes a string into the FunctionStatus.
func (ts *FunctionStatus) Scan(value interface{}) error {
	s, ok := value.(string)
	if !ok {
		return errors.NewInternal("cannot scan function status: invalid string")
	}

	v, ok := FunctionStatus_value[s]
	if !ok {
		return errors.NewInternal("cannot scan function status: unknown value")
	}
	*ts = FunctionStatus(v)

	return nil
}