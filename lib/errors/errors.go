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

// Package errors defines orchestration errors.
// This is following the patterns from https://blog.golang.org/go1.13-errors
// so it mostly contains sentinel values.
package errors

import "errors"

// ErrByteArray happens when attempting to load invalid data as json byte array
var ErrByteArray = errors.New("not a byte array")

// ErrConflict is a sentinel value to mark conflicting asset errors
var ErrConflict = errors.New("conflict")

// ErrInvalidAsset mark asset validation errors
var ErrInvalidAsset = errors.New("invalid asset")

// ErrPermissionDenied happens when you try to perform an action on an asset
// that you do not own.
var ErrPermissionDenied = errors.New("permission denied")
