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

import (
	"github.com/owkin/orchestrator/lib/assets"
	"github.com/stretchr/testify/mock"
)

// MockDBAL is a convenience mock of the persistence layer interface
type MockDBAL struct {
	mock.Mock
}

// AddNode is a mock
func (m *MockDBAL) AddNode(node *assets.Node) error {
	args := m.Called(node)
	return args.Error(0)
}

// NodeExists is a mock
func (m *MockDBAL) NodeExists(id string) (bool, error) {
	args := m.Called(id)
	return args.Bool(0), args.Error(1)
}

// GetNodes is a mock
func (m *MockDBAL) GetNodes() ([]*assets.Node, error) {
	args := m.Called()
	return args.Get(0).([]*assets.Node), args.Error(1)
}

// AddObjective is a mock
func (m *MockDBAL) AddObjective(obj *assets.Objective) error {
	args := m.Called(obj)
	return args.Error(0)
}

// GetObjective is a mock
func (m *MockDBAL) GetObjective(id string) (*assets.Objective, error) {
	args := m.Called(id)
	return args.Get(0).(*assets.Objective), args.Error(1)
}

// GetObjectives is a mock
func (m *MockDBAL) GetObjectives() ([]*assets.Objective, error) {
	args := m.Called()
	return args.Get(0).([]*assets.Objective), args.Error(1)
}
