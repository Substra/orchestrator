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

package service

import (
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	"github.com/owkin/orchestrator/lib/event"
	"github.com/owkin/orchestrator/lib/persistence"
	"github.com/stretchr/testify/mock"
)

// MockServiceProvider is a mock implementing DatabaseProvider
type MockServiceProvider struct {
	mock.Mock
}

// GetNodeDBAL returns whatever value is passed
func (m *MockServiceProvider) GetNodeDBAL() persistence.NodeDBAL {
	args := m.Called()
	return args.Get(0).(persistence.NodeDBAL)
}

// GetObjectiveDBAL returns whatever value is passed
func (m *MockServiceProvider) GetObjectiveDBAL() persistence.ObjectiveDBAL {
	args := m.Called()
	return args.Get(0).(persistence.ObjectiveDBAL)
}

// GetDataSampleDBAL returns whatever value is passed
func (m *MockServiceProvider) GetDataSampleDBAL() persistence.DataSampleDBAL {
	args := m.Called()
	return args.Get(0).(persistence.DataSampleDBAL)
}

// GetAlgoDBAL returns whatever value is passed
func (m *MockServiceProvider) GetAlgoDBAL() persistence.AlgoDBAL {
	args := m.Called()
	return args.Get(0).(persistence.AlgoDBAL)
}

// GetDataManagerDBAL returns whatever value is passed
func (m *MockServiceProvider) GetDataManagerDBAL() persistence.DataManagerDBAL {
	args := m.Called()
	return args.Get(0).(persistence.DataManagerDBAL)
}

// GetDatasetDBAL returns whatever value is passed
func (m *MockServiceProvider) GetDatasetDBAL() persistence.DatasetDBAL {
	args := m.Called()
	return args.Get(0).(persistence.DatasetDBAL)
}

// GetComputeTaskDBAL returns whatever value is passed
func (m *MockServiceProvider) GetComputeTaskDBAL() persistence.ComputeTaskDBAL {
	args := m.Called()
	return args.Get(0).(persistence.ComputeTaskDBAL)
}

// GetComputePlanDBAL returns whatever value is passed
func (m *MockServiceProvider) GetComputePlanDBAL() persistence.ComputePlanDBAL {
	args := m.Called()
	return args.Get(0).(persistence.ComputePlanDBAL)
}

// GetComputeTaskDBAL returns whatever value is passed
func (m *MockServiceProvider) GetModelDBAL() persistence.ModelDBAL {
	args := m.Called()
	return args.Get(0).(persistence.ModelDBAL)
}

// GetEventQueue returns whatever value is passed
func (m *MockServiceProvider) GetEventQueue() event.Queue {
	args := m.Called()
	return args.Get(0).(event.Queue)
}

// GetNodeService returns whatever value is passed
func (m *MockServiceProvider) GetNodeService() NodeAPI {
	args := m.Called()
	return args.Get(0).(NodeAPI)
}

// GetObjectiveService return whatever value is passed
func (m *MockServiceProvider) GetObjectiveService() ObjectiveAPI {
	args := m.Called()
	return args.Get(0).(ObjectiveAPI)
}

// GetPermissionService returns whatever value is passed
func (m *MockServiceProvider) GetPermissionService() PermissionAPI {
	args := m.Called()
	return args.Get(0).(PermissionAPI)
}

// GetDataSampleService returns whatever value is passed
func (m *MockServiceProvider) GetDataSampleService() DataSampleAPI {
	args := m.Called()
	return args.Get(0).(DataSampleAPI)
}

// GetDataManagerService returns whatever value is passed
func (m *MockServiceProvider) GetDataManagerService() DataManagerAPI {
	args := m.Called()
	return args.Get(0).(DataManagerAPI)
}

// GetDatasetService returns whatever value is passed
func (m *MockServiceProvider) GetDatasetService() DatasetAPI {
	args := m.Called()
	return args.Get(0).(DatasetAPI)
}

// GetAlgoService return whatever value is passed
func (m *MockServiceProvider) GetAlgoService() AlgoAPI {
	args := m.Called()
	return args.Get(0).(AlgoAPI)
}

// GetComputeTaskService return whatever value is passed
func (m *MockServiceProvider) GetComputeTaskService() ComputeTaskAPI {
	args := m.Called()
	return args.Get(0).(ComputeTaskAPI)
}

// GetModelService return whatever value is passed
func (m *MockServiceProvider) GetModelService() ModelAPI {
	args := m.Called()
	return args.Get(0).(ModelAPI)
}

// GetComputePlanService return whatever value is passed
func (m *MockServiceProvider) GetComputePlanService() ComputePlanAPI {
	args := m.Called()
	return args.Get(0).(ComputePlanAPI)
}

// MockNodeService is a mock implementing NodeAPI
type MockNodeService struct {
	mock.Mock
}

// GetNodes returns whatever value is passed
func (m *MockNodeService) GetNodes() ([]*asset.Node, error) {
	args := m.Called()
	return args.Get(0).([]*asset.Node), args.Error(1)
}

// GetNode returns whatever value is passed
func (m *MockNodeService) GetNode(id string) (*asset.Node, error) {
	args := m.Called(id)
	return args.Get(0).(*asset.Node), args.Error(1)
}

// RegisterNode returns whatever value is passed
func (m *MockNodeService) RegisterNode(id string) (*asset.Node, error) {
	args := m.Called(id)
	return args.Get(0).(*asset.Node), args.Error(1)
}

// MockPermissionService is a mock implementing PermissionAPI
type MockPermissionService struct {
	mock.Mock
}

// CreatePermissions returns whatever value is passed
func (m *MockPermissionService) CreatePermissions(owner string, perms *asset.NewPermissions) (*asset.Permissions, error) {
	args := m.Called(owner, perms)
	return args.Get(0).(*asset.Permissions), args.Error(1)
}

// CanProcess returns whatever value is passed
func (m *MockPermissionService) CanProcess(perms *asset.Permissions, requester string) bool {
	args := m.Called(perms, requester)
	return args.Bool(0)
}

// MakeIntersection returns whatever is passed
func (m *MockPermissionService) MakeIntersection(x, y *asset.Permissions) *asset.Permissions {
	args := m.Called(x, y)
	return args.Get(0).(*asset.Permissions)
}

// MakeUnion returns whatever is passed
func (m *MockPermissionService) MakeUnion(x, y *asset.Permissions) *asset.Permissions {
	args := m.Called(x, y)
	return args.Get(0).(*asset.Permissions)
}

// MockObjectiveService is a mock implementing ObjectiveAPI
type MockObjectiveService struct {
	mock.Mock
}

// RegisterObjective returns whatever value is passed
func (m *MockObjectiveService) RegisterObjective(objective *asset.NewObjective, owner string) (*asset.Objective, error) {
	args := m.Called(objective, owner)
	return args.Get(0).(*asset.Objective), args.Error(1)
}

// GetObjective returns whatever value is passed
func (m *MockObjectiveService) GetObjective(key string) (*asset.Objective, error) {
	args := m.Called(key)
	return args.Get(0).(*asset.Objective), args.Error(1)
}

// GetObjectives returns whatever value is passed
func (m *MockObjectiveService) GetObjectives(p *common.Pagination) ([]*asset.Objective, common.PaginationToken, error) {
	args := m.Called(p)
	return args.Get(0).([]*asset.Objective), args.Get(1).(common.PaginationToken), args.Error(2)
}

// ObjectiveExists returns whatever value is passed
func (m *MockObjectiveService) ObjectiveExists(key string) (bool, error) {
	args := m.Called(key)
	return args.Bool(0), args.Error(1)
}

// CanDownload returns whatever value is passed
func (m *MockObjectiveService) CanDownload(key string, requester string) (bool, error) {
	args := m.Called(key, requester)
	return args.Bool(0), args.Error(1)
}

// MockDataSampleService is a mock implementing DataSampleAPI
type MockDataSampleService struct {
	mock.Mock
}

// RegisterDataSample returns whatever value is passed
func (m *MockDataSampleService) RegisterDataSample(datasample *asset.NewDataSample, owner string) error {
	args := m.Called(datasample, owner)
	return args.Error(0)
}

// UpdateDataSample returns whatever value is passed
func (m *MockDataSampleService) UpdateDataSample(datasample *asset.DataSampleUpdateParam, owner string) error {
	args := m.Called(datasample, owner)
	return args.Error(0)
}

// GetDataSamples returns whatever value is passed
func (m *MockDataSampleService) GetDataSamples(p *common.Pagination) ([]*asset.DataSample, common.PaginationToken, error) {
	args := m.Called(p)
	return args.Get(0).([]*asset.DataSample), args.Get(1).(common.PaginationToken), args.Error(2)
}

// ContainsTestSample returns whatever value is passed
func (m *MockDataSampleService) ContainsTestSample(keys []string) (bool, error) {
	args := m.Called(keys)
	return args.Bool(0), args.Error(1)
}

// IsTestOnly returns whatever value is passed
func (m *MockDataSampleService) IsTestOnly(keys []string) (bool, error) {
	args := m.Called(keys)
	return args.Bool(0), args.Error(1)
}

// CheckSameManager returns whatever value is passed
func (m *MockDataSampleService) CheckSameManager(managerKey string, sampleKeys []string) error {
	args := m.Called(managerKey, sampleKeys)
	return args.Error(0)
}

// MockAlgoService is a mock implementing AlgoAPI
type MockAlgoService struct {
	mock.Mock
}

// RegisterAlgo returns whatever value is passed
func (m *MockAlgoService) RegisterAlgo(algo *asset.NewAlgo, owner string) (*asset.Algo, error) {
	args := m.Called(algo, owner)
	return args.Get(0).(*asset.Algo), args.Error(1)
}

// GetAlgo returns whatever value is passed
func (m *MockAlgoService) GetAlgo(key string) (*asset.Algo, error) {
	args := m.Called(key)
	return args.Get(0).(*asset.Algo), args.Error(1)
}

// GetAlgos returns whatever value is passed
func (m *MockAlgoService) GetAlgos(c asset.AlgoCategory, p *common.Pagination) ([]*asset.Algo, common.PaginationToken, error) {
	args := m.Called(c, p)
	return args.Get(0).([]*asset.Algo), args.Get(1).(common.PaginationToken), args.Error(2)
}

// MockDataManagerService is a mock implementing DataManagerAPI
type MockDataManagerService struct {
	mock.Mock
}

// RegisterDataManager returns whatever value is passed
func (m *MockDataManagerService) RegisterDataManager(datamanager *asset.NewDataManager, owner string) error {
	args := m.Called(datamanager, owner)
	return args.Error(0)
}

// UpdateDataManager returns whatever value is passed
func (m *MockDataManagerService) UpdateDataManager(datamanager *asset.DataManagerUpdateParam, requester string) error {
	args := m.Called(datamanager, requester)
	return args.Error(0)
}

// GetDataManager returns whatever value is passed
func (m *MockDataManagerService) GetDataManager(key string) (*asset.DataManager, error) {
	args := m.Called(key)
	return args.Get(0).(*asset.DataManager), args.Error(1)
}

// GetDataManagers returns whatever value is passed
func (m *MockDataManagerService) GetDataManagers(p *common.Pagination) ([]*asset.DataManager, common.PaginationToken, error) {
	args := m.Called(p)
	return args.Get(0).([]*asset.DataManager), args.Get(1).(common.PaginationToken), args.Error(2)
}

// CheckOwner returns whatever value is passed
func (m *MockDataManagerService) CheckOwner(keys []string, requester string) error {
	args := m.Called(keys, requester)
	return args.Error(0)
}

// MockDatasetService is a mock implementing DatasetAPI
type MockDatasetService struct {
	mock.Mock
}

// GetDataset returns whatever value is passed
func (m *MockDatasetService) GetDataset(id string) (*asset.Dataset, error) {
	args := m.Called(id)
	return args.Get(0).(*asset.Dataset), args.Error(1)
}

// MockDispatcher is a mock implenting Dispatcher behavior
type MockDispatcher struct {
	mock.Mock
}

// Enqueue returns whatever value is passed
func (m *MockDispatcher) Enqueue(event *event.Event) error {
	args := m.Called(event)
	return args.Error(0)
}

// GetEvents returns whatever value is passed
func (m *MockDispatcher) GetEvents() []*event.Event {
	args := m.Called()
	return args.Get(0).([]*event.Event)
}

// Len returns whatever value is passed
func (m *MockDispatcher) Len() int {
	args := m.Called()
	return args.Int(0)
}

// Dispatch returns whatever value is passed
func (m *MockDispatcher) Dispatch() error {
	args := m.Called()
	return args.Error(0)
}

type MockComputeTaskService struct {
	mock.Mock
}

func (m *MockComputeTaskService) RegisterTask(task *asset.NewComputeTask, owner string) (*asset.ComputeTask, error) {
	args := m.Called(task, owner)
	return args.Get(0).(*asset.ComputeTask), args.Error(1)
}

func (m *MockComputeTaskService) GetTasks(p *common.Pagination, filter *asset.TaskQueryFilter) ([]*asset.ComputeTask, common.PaginationToken, error) {
	args := m.Called(p, filter)
	return args.Get(0).([]*asset.ComputeTask), args.String(1), args.Error(2)
}

func (m *MockComputeTaskService) ApplyTaskAction(key string, action asset.ComputeTaskAction, reason string, requester string) error {
	args := m.Called(key, action, reason, requester)
	return args.Error(0)
}

type MockModelService struct {
	mock.Mock
}

func (m *MockModelService) RegisterModel(newModel *asset.NewModel, requester string) (*asset.Model, error) {
	args := m.Called(newModel, requester)
	return args.Get(0).(*asset.Model), args.Error(1)
}

func (m *MockModelService) GetTaskModels(key string) ([]*asset.Model, error) {
	args := m.Called(key)
	return args.Get(0).([]*asset.Model), args.Error(1)
}

type MockComputePlanService struct {
	mock.Mock
}

func (m *MockComputePlanService) RegisterPlan(plan *asset.NewComputePlan, owner string) (*asset.ComputePlan, error) {
	args := m.Called(plan, owner)
	return args.Get(0).(*asset.ComputePlan), args.Error(1)
}

func (m *MockComputePlanService) GetPlan(key string) (*asset.ComputePlan, error) {
	args := m.Called(key)
	return args.Get(0).(*asset.ComputePlan), args.Error(1)
}

func (m *MockComputePlanService) GetPlans(p *common.Pagination) ([]*asset.ComputePlan, common.PaginationToken, error) {
	args := m.Called(p)
	return args.Get(0).([]*asset.ComputePlan), args.String(1), args.Error(2)
}

func (m *MockComputePlanService) ApplyPlanAction(key string, action asset.ComputePlanAction, requester string) error {
	args := m.Called(key, action, requester)
	return args.Error(0)
}
