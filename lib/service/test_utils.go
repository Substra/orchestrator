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

// MockNodeService is a mock implementing NodeAPI
type MockNodeService struct {
	mock.Mock
}

// GetNodes returns whatever value is passed
func (m *MockNodeService) GetNodes() ([]*asset.Node, error) {
	args := m.Called()
	return args.Get(0).([]*asset.Node), args.Error(1)
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
