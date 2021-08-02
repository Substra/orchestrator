package service

import (
	"github.com/owkin/orchestrator/lib/event"
	"github.com/owkin/orchestrator/lib/persistence"
)

// DependenciesProvider describes a Provider exposing all orchestration services.
type DependenciesProvider interface {
	persistence.DBALProvider
	event.QueueProvider
	NodeServiceProvider
	ObjectiveServiceProvider
	DataSampleServiceProvider
	AlgoServiceProvider
	PermissionServiceProvider
	DataManagerServiceProvider
	DatasetServiceProvider
	ComputeTaskServiceProvider
	ModelServiceProvider
	ComputePlanServiceProvider
	PerformanceServiceProvider
	EventServiceProvider
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
	dataset     DatasetAPI
	computeTask ComputeTaskAPI
	model       ModelAPI
	computePlan ComputePlanAPI
	performance PerformanceAPI
	event       EventAPI
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

// GetDataManagerDBAL returns the database abstraction layer for DataManagers
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

// GetModelDBAL returns the database abstraction layer for Tasks
func (sc *Provider) GetModelDBAL() persistence.ModelDBAL {
	return sc.dbal
}

// GetComputePlanDBAL returns the database abstraction layer for Tasks
func (sc *Provider) GetComputePlanDBAL() persistence.ComputePlanDBAL {
	return sc.dbal
}

// GetPerformanceDBAL returns the database abstraction layer for Tasks
func (sc *Provider) GetPerformanceDBAL() persistence.PerformanceDBAL {
	return sc.dbal
}

func (sc *Provider) GetEventDBAL() persistence.EventDBAL {
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

// GetDataManagerService returns a DataManagerAPI instance.
// The service will be instanciated if needed.
func (sc *Provider) GetDataManagerService() DataManagerAPI {
	if sc.datamanager == nil {
		sc.datamanager = NewDataManagerService(sc)
	}
	return sc.datamanager
}

// GetDatasetService returns a DataSampleAPI instance.
// The service will be instanciated if needed.
func (sc *Provider) GetDatasetService() DatasetAPI {
	if sc.dataset == nil {
		sc.dataset = NewDatasetService(sc)
	}
	return sc.dataset
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

// GetModelService returns a ModelAPI instance.
// The service will be instanciated if needed.
func (sc *Provider) GetModelService() ModelAPI {
	if sc.model == nil {
		sc.model = NewModelService(sc)
	}
	return sc.model
}

// GetComputePlanService returns a ComputePlanAPI instance.
// The service will be instanciated if needed.
func (sc *Provider) GetComputePlanService() ComputePlanAPI {
	if sc.computePlan == nil {
		sc.computePlan = NewComputePlanService(sc)
	}
	return sc.computePlan
}

// GetPerformanceService returns a PerformanceAPI instance.
// The service will be instanciated if needed.
func (sc *Provider) GetPerformanceService() PerformanceAPI {
	if sc.performance == nil {
		sc.performance = NewPerformanceService(sc)
	}
	return sc.performance
}

// GetEventService returns an EventAPI instance.
// The service will be instanciated if needed.
func (sc *Provider) GetEventService() EventAPI {
	if sc.event == nil {
		sc.event = NewEventService(sc)
	}
	return sc.event
}
