// Package errors defines orchestration errors.
// This is following the patterns from https://blog.golang.org/go1.13-errors
// Error messages should contain a unique code easily parsable from string representation.
// This limitation comes from the fact that chaincode only returns string representation of errors,
// meaning we loose the grpc return code between chaincode and orchestrator. This error ID is a way to circumvent that.
// Error IDs are 4 digit prefixed by "OE" (for Orchestration Error).
// Each error kind has 100 numbers reserved, here are the assigned ranges:
// - generic errors: 0-99
// - asset related errors: 100-199
package errors

import (
	"fmt"
	"runtime"
)

// ErrorKind is unique per kind of orchestration error.
// It may be used to determine response status.
type ErrorKind = string

var (
	// Generic errors
	// Range 0-99

	// ErrByteArray happens when attempting to load invalid data as json byte array
	ErrByteArray = "OE0001"

	// ErrNotFound happens when the asset is not present in the persistence layer
	ErrNotFound = "OE0002"

	// ErrBadRequest happens when the request can't be fulfilled
	ErrBadRequest = "OE0003"

	// ErrConflict is a sentinel value to mark conflicting asset errors.
	ErrConflict = "OE0006" // value 6 match gRPC AlreadyExists status code

	// ErrInternal happens when an unexpected error occurs (eg; unreachable code)
	ErrInternal = "OE0007"

	// ErrUnimplemented occurs when unimplemented code is triggered
	ErrUnimplemented = "OE0010"

	// Asset specific errors
	// Range 100-199

	// ErrInvalidAsset marks asset validation errors
	ErrInvalidAsset = "OE0101"

	// ErrPermissionDenied happens when you try to perform an action on an asset
	// that you do not own.
	ErrPermissionDenied = "OE0102"

	// ErrReferenceNotFound (OE0103) happened when a sub-asset was not present in the persistence layer

	// ErrIncompatibleTaskStatus occurs when a task cannot be processed due to its status
	ErrIncompatibleTaskStatus = "OE0104"

	// ErrCannotDisableModel occurs when attempting to disable a model that is not eligible
	ErrCannotDisableModel = "OE0105"

	// ErrMissingTaskOutput occurs when attempting to register an asset referencing a non-existing task output
	ErrMissingTaskOutput = "OE0106"

	// ErrIncompatibleKind occurs when attempting to register an asset for a task output of a different kind
	ErrIncompatibleKind = "OE0107"

	// ErrCannotDisableOutput occurs when attempting to disable an output that is not eligible
	ErrCannotDisableOutput = "OE0108"

	// ErrTerminatedComputePlan occurs when attempting to cancel or fail an already terminated compute plan
	ErrTerminatedComputePlan = "OE0109"
)

// OrcError represents an orchestration error.
// It may wrap another error and will always contain a specific error code.
type OrcError struct {
	Kind     ErrorKind
	msg      string
	internal error
	source   string
}

// Error returns the error message
func (e *OrcError) Error() string {
	out := e.Kind + ": " + e.msg
	if e.internal != nil {
		out = fmt.Sprintf("%s: %v", out, e.internal)
	}

	return out
}

// Unwrap returns the wrapped error if any
func (e *OrcError) Unwrap() error {
	return e.internal
}

// Wrap make sure the given error is embedded in the error chain.
// It returns the OrcError for a convenient fluent interface.
func (e *OrcError) Wrap(err error) *OrcError {
	e.internal = err
	return e
}

// Source will return error's source as file:line
func (e *OrcError) Source() string {
	return e.source
}

// NewError creates an OrcError with given kind and message
func NewError(kind ErrorKind, msg string) *OrcError {
	return newErrorWithSource(kind, msg)
}

// newErrorWithSource should be called by public error creation methods.
// It will set the source property from the caller.
func newErrorWithSource(kind ErrorKind, msg string) *OrcError {
	var caller string

	_, file, line, ok := runtime.Caller(2)
	if ok {
		caller = fmt.Sprintf("%s:%d", file, line)
	}

	return &OrcError{
		Kind:   kind,
		msg:    msg,
		source: caller,
	}
}

// Below are convenience functions to create OrcErrors

// NewNotFound returns an ErrNotFound kind of OrcError with relevant message
func NewNotFound(resource, key string) *OrcError {
	return newErrorWithSource(ErrNotFound, fmt.Sprintf("%s with key %q not found", resource, key))
}

// NewConflict returns an ErrConflict kind of OrcError with relevant message
func NewConflict(resource, key string) *OrcError {
	return newErrorWithSource(ErrConflict, fmt.Sprintf("%s with key %q already exists", resource, key))
}

// NewBadRequest returns an ErrBadRequest kind of OrcError with given message
func NewBadRequest(msg string) *OrcError {
	return newErrorWithSource(ErrBadRequest, msg)
}

// NewInvalidAsset returns an ErrInvalidAsset kind of OrcError with given message
func NewInvalidAsset(msg string) *OrcError {
	return newErrorWithSource(ErrInvalidAsset, msg)
}

// NewPermissionDenied returns an ErrPermissionDenied kind of OrcError with given message
func NewPermissionDenied(msg string) *OrcError {
	return newErrorWithSource(ErrPermissionDenied, msg)
}

// NewCannotDisableModel returns an ErrCannotDisableModel kind of OrcError with given message
func NewCannotDisableModel(msg string) *OrcError {
	return newErrorWithSource(ErrCannotDisableModel, msg)
}

// NewCannotDisableAsset returns an ErrCannotDisableModel kind of OrcError with given message
func NewCannotDisableAsset(msg string) *OrcError {
	return newErrorWithSource(ErrCannotDisableOutput, msg)
}

// NewInternal returns an ErrInternalError kind of OrcError with given message
func NewInternal(msg string) *OrcError {
	return newErrorWithSource(ErrInternal, msg)
}

// NewUnimplemented returns an ErrUnimplemented kind of OrcError with given message
func NewUnimplemented(msg string) *OrcError {
	return newErrorWithSource(ErrUnimplemented, msg)
}

func NewMissingTaskOutput(taskKey, identifier string) *OrcError {
	return newErrorWithSource(ErrMissingTaskOutput, fmt.Sprintf("Task %q has no output named %q", taskKey, identifier))
}

// NewTerminatedComputePlan returns an ErrTerminatedComputePlan kind of OrcError with given message
func NewTerminatedComputePlan(msg string) *OrcError {
	return newErrorWithSource(ErrTerminatedComputePlan, msg)
}

func NewIncompatibleTaskOutput(taskKey, identifier, expected, actual string) *OrcError {
	return newErrorWithSource(
		ErrIncompatibleKind,
		fmt.Sprintf(
			"Incompatible kind for task %q: output %q expects %q, received %q",
			taskKey,
			identifier,
			expected,
			actual,
		),
	)
}

// FromValidationError returns an OrcError with ErrInvalidAsset kind wrapping the underlying validation error
func FromValidationError(resource string, err error) *OrcError {
	return newErrorWithSource(ErrInvalidAsset, fmt.Sprintf("%s is not valid", resource)).Wrap(err)
}
