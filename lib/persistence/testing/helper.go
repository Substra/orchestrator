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

// Package testing provides helpers and mocks which can be used in tests
package testing

import "github.com/stretchr/testify/mock"

// MockDatabase is a convenience mock of the persistence layer interface
type MockDatabase struct {
	mock.Mock
}

// PutState stores data
func (m *MockDatabase) PutState(resource string, key string, data []byte) error {
	args := m.Called(resource, key, data)
	return args.Error(0)
}

// GetState fetches identified data
func (m *MockDatabase) GetState(resource string, key string) ([]byte, error) {
	args := m.Called(resource, key)
	return args.Get(0).([]byte), args.Error(1)
}

// GetAll retrieves all data for a resource kind
func (m *MockDatabase) GetAll(resource string) ([][]byte, error) {
	args := m.Called(resource)
	return args.Get(0).([][]byte), args.Error(1)
}
