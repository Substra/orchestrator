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

// GetAllNodes is a mock
func (m *MockDBAL) GetAllNodes() ([]*asset.Node, error) {
	args := m.Called()
	return args.Get(0).([]*asset.Node), args.Error(1)
}

// GetNode is a mock
func (m *MockDBAL) GetNode(id string) (*asset.Node, error) {
	args := m.Called(id)
	return args.Get(0).(*asset.Node), args.Error(1)
}

// AddObjective is a mock
func (m *MockDBAL) AddObjective(obj *asset.Objective) error {
	args := m.Called(obj)
	return args.Error(0)
}

// GetObjective is a mock
func (m *MockDBAL) GetObjective(key string) (*asset.Objective, error) {
	args := m.Called(key)
	return args.Get(0).(*asset.Objective), args.Error(1)
}

// QueryObjectives is a mock
func (m *MockDBAL) QueryObjectives(p *common.Pagination) ([]*asset.Objective, common.PaginationToken, error) {
	args := m.Called(p)
	return args.Get(0).([]*asset.Objective), args.Get(1).(common.PaginationToken), args.Error(2)
}

// ObjectiveExists is a mock
func (m *MockDBAL) ObjectiveExists(key string) (bool, error) {
	args := m.Called(key)
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
func (m *MockDBAL) GetDataSample(key string) (*asset.DataSample, error) {
	args := m.Called(key)
	return args.Get(0).(*asset.DataSample), args.Error(1)
}

// QueryDataSamples is a mock
func (m *MockDBAL) QueryDataSamples(p *common.Pagination) ([]*asset.DataSample, common.PaginationToken, error) {
	args := m.Called(p)
	return args.Get(0).([]*asset.DataSample), args.Get(1).(common.PaginationToken), args.Error(2)
}

// DataSampleExists is a mock
func (m *MockDBAL) DataSampleExists(key string) (bool, error) {
	args := m.Called(key)
	return args.Bool(0), args.Error(1)
}

// AddAlgo is a mock
func (m *MockDBAL) AddAlgo(obj *asset.Algo) error {
	args := m.Called(obj)
	return args.Error(0)
}

// GetAlgo is a mock
func (m *MockDBAL) GetAlgo(key string) (*asset.Algo, error) {
	args := m.Called(key)
	return args.Get(0).(*asset.Algo), args.Error(1)
}

// QueryAlgos is a mock
func (m *MockDBAL) QueryAlgos(c asset.AlgoCategory, p *common.Pagination) ([]*asset.Algo, common.PaginationToken, error) {
	args := m.Called(c, p)
	return args.Get(0).([]*asset.Algo), args.Get(1).(common.PaginationToken), args.Error(2)
}

// AlgoExists is a mock
func (m *MockDBAL) AlgoExists(key string) (bool, error) {
	args := m.Called(key)
	return args.Bool(0), args.Error(1)
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
func (m *MockDBAL) GetDataManager(key string) (*asset.DataManager, error) {
	args := m.Called(key)
	return args.Get(0).(*asset.DataManager), args.Error(1)
}

// QueryDataManagers is a mock
func (m *MockDBAL) QueryDataManagers(p *common.Pagination) ([]*asset.DataManager, common.PaginationToken, error) {
	args := m.Called(p)
	return args.Get(0).([]*asset.DataManager), args.Get(1).(common.PaginationToken), args.Error(2)
}

// DataManagerExists is a mock
func (m *MockDBAL) DataManagerExists(key string) (bool, error) {
	args := m.Called(key)
	return args.Bool(0), args.Error(1)
}

// GetDataset is a mock
func (m *MockDBAL) GetDataset(id string) (*asset.Dataset, error) {
	args := m.Called(id)
	return args.Get(0).(*asset.Dataset), args.Error(1)
}

// ComputeTaskExists is a mock
func (m *MockDBAL) ComputeTaskExists(key string) (bool, error) {
	args := m.Called(key)
	return args.Bool(0), args.Error(1)
}

// GetComputeTask is a mock
func (m *MockDBAL) GetComputeTask(key string) (*asset.ComputeTask, error) {
	args := m.Called(key)
	return args.Get(0).(*asset.ComputeTask), args.Error(1)
}

// GetComputeTaskChildren is a mock
func (m *MockDBAL) GetComputeTaskChildren(key string) ([]*asset.ComputeTask, error) {
	args := m.Called(key)
	return args.Get(0).([]*asset.ComputeTask), args.Error(1)
}

// GetComputePlanTasks is a mock
func (m *MockDBAL) GetComputePlanTasks(key string) ([]*asset.ComputeTask, error) {
	args := m.Called(key)
	return args.Get(0).([]*asset.ComputeTask), args.Error(1)
}

// GetComputeTasks is a mock
func (m *MockDBAL) GetComputeTasks(keys []string) ([]*asset.ComputeTask, error) {
	args := m.Called(keys)
	return args.Get(0).([]*asset.ComputeTask), args.Error(1)
}

func (m *MockDBAL) GetComputePlanTasksKeys(key string) ([]string, error) {
	args := m.Called(key)
	return args.Get(0).([]string), args.Error(1)
}

// AddComputeTask is a mock
func (m *MockDBAL) AddComputeTask(t *asset.ComputeTask) error {
	args := m.Called(t)
	return args.Error(0)
}

// UpdateComputeTask is a mock
func (m *MockDBAL) UpdateComputeTask(t *asset.ComputeTask) error {
	args := m.Called(t)
	return args.Error(0)
}

func (m *MockDBAL) QueryComputeTasks(p *common.Pagination, filter *asset.TaskQueryFilter) ([]*asset.ComputeTask, common.PaginationToken, error) {
	args := m.Called(p, filter)
	return args.Get(0).([]*asset.ComputeTask), args.String(1), args.Error(2)
}

func (m *MockDBAL) ModelExists(key string) (bool, error) {
	args := m.Called(key)
	return args.Bool(0), args.Error(1)
}

func (m *MockDBAL) GetModel(key string) (*asset.Model, error) {
	args := m.Called(key)
	return args.Get(0).(*asset.Model), args.Error(1)
}

func (m *MockDBAL) GetComputeTaskOutputModels(key string) ([]*asset.Model, error) {
	args := m.Called(key)
	return args.Get(0).([]*asset.Model), args.Error(1)
}

func (m *MockDBAL) GetComputeTaskInputModels(key string) ([]*asset.Model, error) {
	args := m.Called(key)
	return args.Get(0).([]*asset.Model), args.Error(1)
}

func (m *MockDBAL) AddModel(model *asset.Model) error {
	args := m.Called(model)
	return args.Error(0)
}

func (m *MockDBAL) ComputePlanExists(key string) (bool, error) {
	args := m.Called(key)
	return args.Bool(0), args.Error(1)
}

func (m *MockDBAL) GetComputePlan(key string) (*asset.ComputePlan, error) {
	args := m.Called(key)
	return args.Get(0).(*asset.ComputePlan), args.Error(1)
}

func (m *MockDBAL) AddComputePlan(plan *asset.ComputePlan) error {
	args := m.Called(plan)
	return args.Error(0)
}

func (m *MockDBAL) QueryComputePlans(p *common.Pagination) ([]*asset.ComputePlan, common.PaginationToken, error) {
	args := m.Called(p)
	return args.Get(0).([]*asset.ComputePlan), args.String(1), args.Error(2)
}

func (m *MockDBAL) UpdateModel(model *asset.Model) error {
	args := m.Called(model)
	return args.Error(0)
}

func (m *MockDBAL) QueryModels(c asset.ModelCategory, p *common.Pagination) ([]*asset.Model, common.PaginationToken, error) {
	args := m.Called(c, p)
	return args.Get(0).([]*asset.Model), args.Get(1).(common.PaginationToken), args.Error(2)
}

func (m *MockDBAL) AddPerformance(perf *asset.Performance) error {
	args := m.Called(perf)
	return args.Error(0)
}

func (m *MockDBAL) GetComputeTaskPerformance(key string) (*asset.Performance, error) {
	args := m.Called(key)
	return args.Get(0).(*asset.Performance), args.Error(1)
}
