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

// Package persistence holds everything related to data persistence
package persistence

// Database is the main interface to act on the persistence layer
// This covers all CRUD operations
type Database interface {
	DBWriter
	DBReader
}

// DBWriter handles persisting and updating data
type DBWriter interface {
	PutState(resource string, key string, data []byte) error
}

// DBReader handles data retrieval
type DBReader interface {
	GetState(resource string, key string) ([]byte, error)
	GetAll(resource string) ([][]byte, error)
}

// DatabaseProvider defines an object able to provide a Database instance
type DatabaseProvider interface {
	GetDatabase() Database
}
