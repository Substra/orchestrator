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
	DataSampleServiceProvider
	AlgoServiceProvider
	PermissionServiceProvider
	DataManagerServiceProvider
	ComputeTaskServiceProvider
}

// Provider is the central part of the dependency injection pattern.
// It is injected into each service, so that they can access their dependencies.
// Services are instanciated as they are required, and reused in subsequent calls.
// Each service should define a ServiceDependencyProvider interface which states what are its requirements.
// Since the Provider implements every Provider interface, it can fit all service dependencies.
type Provider struct {
	dbal        persistence.DBAL
	eventQueue  event.Queue
	node        NodeAPI
	objective   ObjectiveAPI
	permission  PermissionAPI
	datasample  DataSampleAPI
	algo        AlgoAPI
	datamanager DataManagerAPI
	computeTask ComputeTaskAPI
}

// NewProvider return an instance of Provider based on given persistence layer.
func NewProvider(dbal persistence.DBAL, queue event.Queue) *Provider {
	return &Provider{dbal: dbal, eventQueue: queue}
}

// GetNodeDBAL returns the database abstraction layer for Nodes
func (sc *Provider) GetNodeDBAL() persistence.NodeDBAL {
	return sc.dbal
}

// GetObjectiveDBAL returns the database abstraction layer for Objectives
func (sc *Provider) GetObjectiveDBAL() persistence.ObjectiveDBAL {
	return sc.dbal
}

// GetDataSampleDBAL returns the database abstraction layer for DataSamples
func (sc *Provider) GetDataSampleDBAL() persistence.DataSampleDBAL {
	return sc.dbal
}

// GetDataManagerDBAL returns the database abstraction layer for DataSamples
func (sc *Provider) GetDataManagerDBAL() persistence.DataManagerDBAL {
	return sc.dbal
}

// GetAlgoDBAL returns the database abstraction layer for Algos
func (sc *Provider) GetAlgoDBAL() persistence.AlgoDBAL {
	return sc.dbal
}

// GetComputeTaskDBAL returns the database abstraction layer for Tasks
func (sc *Provider) GetComputeTaskDBAL() persistence.ComputeTaskDBAL {
	return sc.dbal
}

// GetEventQueue returns an event.Queue instance
func (sc *Provider) GetEventQueue() event.Queue {
	return sc.eventQueue
}

// GetNodeService returns a NodeAPI instance.
// The service will be instanciated if needed.
func (sc *Provider) GetNodeService() NodeAPI {
	if sc.node == nil {
		sc.node = NewNodeService(sc)
	}
	return sc.node
}

// GetObjectiveService returns an ObjectiveAPI instance.
// The service will be instanciated if needed.
func (sc *Provider) GetObjectiveService() ObjectiveAPI {
	if sc.objective == nil {
		sc.objective = NewObjectiveService(sc)
	}
	return sc.objective
}

// GetDataSampleService returns a DataSampleAPI instance.
// The service will be instanciated if needed.
func (sc *Provider) GetDataSampleService() DataSampleAPI {
	if sc.datasample == nil {
		sc.datasample = NewDataSampleService(sc)
	}
	return sc.datasample
}

// GetDataManagerService returns a DataSampleAPI instance.
// The service will be instanciated if needed.
func (sc *Provider) GetDataManagerService() DataManagerAPI {
	if sc.datamanager == nil {
		sc.datamanager = NewDataManagerService(sc)
	}
	return sc.datamanager
}

// GetAlgoService returns an AlgoAPI instance.
// The service will be instanciated if needed.
func (sc *Provider) GetAlgoService() AlgoAPI {
	if sc.algo == nil {
		sc.algo = NewAlgoService(sc)
	}
	return sc.algo
}

// GetPermissionService returns a PermissionAPI instance.
// The service will be instanciated if needed.
func (sc *Provider) GetPermissionService() PermissionAPI {
	if sc.permission == nil {
		sc.permission = NewPermissionService(sc)
	}
	return sc.permission
}

// GetComputeTaskService returns a ComputeTaskAPI instance.
// The service will be instanciated if needed.
func (sc *Provider) GetComputeTaskService() ComputeTaskAPI {
	if sc.computeTask == nil {
		sc.computeTask = NewComputeTaskService(sc)
	}
	return sc.computeTask
}
