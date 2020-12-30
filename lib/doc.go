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

// Package lib defines structures and business logic related to substra orchestration platform.
// This package does not rely on a concrete storage backend, but rather defines a persistence interface.
// Business logic should be agnostic of the backend as well, as it may be called from either a gRPC server
// or a smart contract.
package lib
