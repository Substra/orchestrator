// Package errors defines orchestration errors.
// This is following the patterns from https://blog.golang.org/go1.13-errors
// so it mostly contains sentinel values.
// Sentinel values should contain a unique code easily parsable from string representation.
// This limitation comes from the fact that chaincode only returns string representation of errors,
// meaning we loose the grpc return code between chaincode and orchestrator. This error ID is a way to circumvent that.
// Error IDs are 4 digit prefixed by "OE" (for Orchestration Error).
// Each error kind has 100 numbers reserved, here are the assigned ranges:
// - generic errors: 0-99
// - asset related errors: 100-199
package errors

import "errors"

// Generic errors
// Range 0-99

// ErrByteArray happens when attempting to load invalid data as json byte array
var ErrByteArray = errors.New("OE0001")

// ErrNotFound happens when the asset is not present in the persistence layer
var ErrNotFound = errors.New("OE0002")

// ErrBadRequest happens when the request can't be fulfilled
var ErrBadRequest = errors.New("OE0003")

// ErrConflict is a sentinel value to mark conflicting asset errors.
var ErrConflict = errors.New("OE0006") // value 6 match gRPC AlreadyExists status code

// ErrInternalError happens when an unexpected error occurs (eg; unreachable code)
var ErrInternalError = errors.New("OE0007")

// ErrUnimplemented occurs when unimplemented code is triggered
var ErrUnimplemented = errors.New("OE0010")

// Asset specific errors
// Range 100-199

// ErrInvalidAsset mark asset validation errors
var ErrInvalidAsset = errors.New("OE0101")

// ErrPermissionDenied happens when you try to perform an action on an asset
// that you do not own.
var ErrPermissionDenied = errors.New("OE0102")

// ErrReferenceNotFound happens when an asset is not present in the persistence layer
var ErrReferenceNotFound = errors.New("OE0103")

// ErrIncompatibleTaskStatus occurs when a task cannot be processed due to its status
var ErrIncompatibleTaskStatus = errors.New("OE0104")

// ErrCannotDisableModel occurs when attempting to disable a model that is not eligible
var ErrCannotDisableModel = errors.New("OE105")
