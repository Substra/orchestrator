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
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	"github.com/stretchr/testify/mock"
)

// MockDBAL is a convenience mock of the persistence layer interface
type MockDBAL struct {
	mock.Mock
}

// AddNode is a mock
func (m *MockDBAL) AddNode(node *asset.Node) error {
	args := m.Called(node)
	return args.Error(0)
}

// NodeExists is a mock
func (m *MockDBAL) NodeExists(id string) (bool, error) {
	args := m.Called(id)
	return args.Bool(0), args.Error(1)
}

// GetNodes is a mock
func (m *MockDBAL) GetNodes() ([]*asset.Node, error) {
	args := m.Called()
	return args.Get(0).([]*asset.Node), args.Error(1)
}

// AddObjective is a mock
func (m *MockDBAL) AddObjective(obj *asset.Objective) error {
	args := m.Called(obj)
	return args.Error(0)
}

// GetObjective is a mock
func (m *MockDBAL) GetObjective(id string) (*asset.Objective, error) {
	args := m.Called(id)
	return args.Get(0).(*asset.Objective), args.Error(1)
}

// GetObjectives is a mock
func (m *MockDBAL) GetObjectives(p *common.Pagination) ([]*asset.Objective, common.PaginationToken, error) {
	args := m.Called(p)
	return args.Get(0).([]*asset.Objective), args.Get(1).(common.PaginationToken), args.Error(2)
}

// ObjectiveExists is a mock
func (m *MockDBAL) ObjectiveExists(id string) (bool, error) {
	args := m.Called(id)
	return args.Bool(0), args.Error(1)
}

// AddDataSample is a mock
func (m *MockDBAL) AddDataSample(dataSample *asset.DataSample) error {
	args := m.Called(dataSample)
	return args.Error(0)
}

// UpdateDataSample is a mock
func (m *MockDBAL) UpdateDataSample(dataSample *asset.DataSample) error {
	args := m.Called(dataSample)
	return args.Error(0)
}

// GetDataSample is a mock
func (m *MockDBAL) GetDataSample(id string) (*asset.DataSample, error) {
	args := m.Called(id)
	return args.Get(0).(*asset.DataSample), args.Error(1)
}

// GetDataSamples is a mock
func (m *MockDBAL) GetDataSamples(p *common.Pagination) ([]*asset.DataSample, common.PaginationToken, error) {
	args := m.Called(p)
	return args.Get(0).([]*asset.DataSample), args.Get(1).(common.PaginationToken), args.Error(2)
}

// AddAlgo is a mock
func (m *MockDBAL) AddAlgo(obj *asset.Algo) error {
	args := m.Called(obj)
	return args.Error(0)
}

// GetAlgo is a mock
func (m *MockDBAL) GetAlgo(id string) (*asset.Algo, error) {
	args := m.Called(id)
	return args.Get(0).(*asset.Algo), args.Error(1)
}

// GetAlgos is a mock
func (m *MockDBAL) GetAlgos(p *common.Pagination) ([]*asset.Algo, common.PaginationToken, error) {
	args := m.Called(p)
	return args.Get(0).([]*asset.Algo), args.Get(1).(common.PaginationToken), args.Error(2)
}

// AddDataManager is a mock
func (m *MockDBAL) AddDataManager(datamanager *asset.DataManager) error {
	args := m.Called(datamanager)
	return args.Error(0)
}

// UpdateDataManager is a mock
func (m *MockDBAL) UpdateDataManager(datamanager *asset.DataManager) error {
	args := m.Called(datamanager)
	return args.Error(0)
}

// GetDataManager is a mock
func (m *MockDBAL) GetDataManager(id string) (*asset.DataManager, error) {
	args := m.Called(id)
	return args.Get(0).(*asset.DataManager), args.Error(1)
}

// GetDataManagers is a mock
func (m *MockDBAL) GetDataManagers(p *common.Pagination) ([]*asset.DataManager, common.PaginationToken, error) {
	args := m.Called(p)
	return args.Get(0).([]*asset.DataManager), args.Get(1).(common.PaginationToken), args.Error(2)
}

// DataManagersExists is a mock
func (m *MockDBAL) DataManagersExists(id string) (bool, error) {
	args := m.Called(id)
	return args.Bool(0), args.Error(1)
}
