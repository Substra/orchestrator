// Package lib defines structures and business logic related to substra orchestration platform.
// This package does not rely on a concrete storage backend, but rather defines a persistence interface.
// Business logic should be agnostic of the backend as well, as it may be called from either a gRPC server
// or a smart contract.
package lib
