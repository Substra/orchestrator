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

	// ErrInternalError happens when an unexpected error occurs (eg; unreachable code)
	ErrInternalError = "OE0007"

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
	ErrCannotDisableModel = "OE105"
)

// OrcError represents an orchestration error.
// It may wrap another error and will always contain a specific error code.
type OrcError struct {
	Kind     ErrorKind
	msg      string
	internal error
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

// NewError creates an OrcError with given kind and message
func NewError(kind ErrorKind, msg string) *OrcError {
	return &OrcError{
		Kind: kind,
		msg:  msg,
	}
}

// Below are convenience functions to create OrcErrors

// NewNotFound returns an ErrNotFound kind of OrcError with relevant message
func NewNotFound(resource, key string) *OrcError {
	return NewError(ErrNotFound, fmt.Sprintf("%s with key %q not found", resource, key))
}

// NewConflict returns an ErrConflict kind of OrcError with relevant message
func NewConflict(resource, key string) *OrcError {
	return NewError(ErrConflict, fmt.Sprintf("%s with key %q already exists", resource, key))
}

// NewBadRequest returns an ErrBadRequest kind of OrcError with given message
func NewBadRequest(msg string) *OrcError {
	return NewError(ErrBadRequest, msg)
}

// NewInvalidAsset returns an ErrInvalidAsset kind of OrcError with given message
func NewInvalidAsset(msg string) *OrcError {
	return NewError(ErrInvalidAsset, msg)
}

// NewPermissionDenied returns an ErrPermissionDenied kind of OrcError with given message
func NewPermissionDenied(msg string) *OrcError {
	return NewError(ErrPermissionDenied, msg)
}

// NewCannotDisableModel returns an ErrCannotDisableModel kind of OrcError with given message
func NewCannotDisableModel(msg string) *OrcError {
	return NewError(ErrCannotDisableModel, msg)
}

// NewInternal returns an ErrInternalError kind of OrcError with given message
func NewInternal(msg string) *OrcError {
	return NewError(ErrInternalError, msg)
}

// FromValidationError returns an OrcError with ErrInvalidAsset kind wrapping the underlying validation error
func FromValidationError(resource string, err error) *OrcError {
	return NewError(ErrInvalidAsset, fmt.Sprintf("%s is not valid", resource)).Wrap(err)
}
