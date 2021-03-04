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

package orchestration

import (
	"github.com/owkin/orchestrator/lib/event"
	"github.com/owkin/orchestrator/lib/persistence"
)

// DependenciesProvider describes a Provider exposing all orchestration services.
type DependenciesProvider interface {
	persistence.NodeDBALProvider
	persistence.ObjectiveDBALProvider
	event.QueueProvider
	NodeServiceProvider
	ObjectiveServiceProvider
	PermissionServiceProvider
}

// ServiceProvider is the central part of the dependency injection pattern.
// It is injected into each service, so that they can access their dependencies.
// Services are instanciated as they are required, and reused in subsequent calls.
// Each service should define a ServiceDependencyProvider interface which states what are its requirements.
// Since the ServiceProvider implements every ServiceProvider interface, it can fit all service dependencies.
type ServiceProvider struct {
	dbal       persistence.DBAL
	eventQueue event.Queue
	node       NodeAPI
	objective  ObjectiveAPI
	permission PermissionAPI
}

// NewServiceProvider return an instance of ServiceProvider based on given persistence layer.
func NewServiceProvider(dbal persistence.DBAL, queue event.Queue) *ServiceProvider {
	return &ServiceProvider{dbal: dbal, eventQueue: queue}
}

// GetNodeDBAL returns the database abstraction layer for Nodes
func (sc *ServiceProvider) GetNodeDBAL() persistence.NodeDBAL {
	return sc.dbal
}

// GetObjectiveDBAL returns the database abstraction layer for Objectives
func (sc *ServiceProvider) GetObjectiveDBAL() persistence.ObjectiveDBAL {
	return sc.dbal
}

// GetEventQueue returns an event.Queue instance
func (sc *ServiceProvider) GetEventQueue() event.Queue {
	return sc.eventQueue
}

// GetNodeService returns a NodeAPI instance.
// The service will be instanciated if needed.
func (sc *ServiceProvider) GetNodeService() NodeAPI {
	if sc.node == nil {
		sc.node = NewNodeService(sc)
	}
	return sc.node
}

// GetObjectiveService returns an ObjectiveAPI instance.
// The service will be instanciated if needed.
func (sc *ServiceProvider) GetObjectiveService() ObjectiveAPI {
	if sc.objective == nil {
		sc.objective = NewObjectiveService(sc)
	}
	return sc.objective
}

// GetPermissionService returns a PermissionAPI instance.
// The service will be instanciated if needed.
func (sc *ServiceProvider) GetPermissionService() PermissionAPI {
	if sc.permission == nil {
		sc.permission = NewPermissionService(sc)
	}
	return sc.permission
}
