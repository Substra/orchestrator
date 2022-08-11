package service

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/substra/orchestrator/lib/event"
	"github.com/substra/orchestrator/lib/persistence"
)

// LoggerProvider describes a provider of logger instance.
type LoggerProvider interface {
	GetLogger() *zerolog.Logger
}

type ChannelProvider interface {
	GetChannel() string
}

// DependenciesProvider describes a Provider exposing all orchestration services.
type DependenciesProvider interface {
	persistence.DBALProvider
	event.QueueProvider
	OrganizationServiceProvider
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
	LoggerProvider
	TimeServiceProvider
	FailureReportServiceProvider
	ChannelProvider
}

// Provider is the central part of the dependency injection pattern.
// It is injected into each service, so that they can access their dependencies.
// Services are instanciated as they are required, and reused in subsequent calls.
// Each service should define a ServiceDependencyProvider interface which states what are its requirements.
// Since the Provider implements every Provider interface, it can fit all service dependencies.
type Provider struct {
	logger        *zerolog.Logger
	channel       string
	dbal          persistence.DBAL
	eventQueue    event.Queue
	organization  OrganizationAPI
	permission    PermissionAPI
	datasample    DataSampleAPI
	algo          AlgoAPI
	datamanager   DataManagerAPI
	dataset       DatasetAPI
	computeTask   ComputeTaskAPI
	model         ModelAPI
	computePlan   ComputePlanAPI
	performance   PerformanceAPI
	event         EventAPI
	time          TimeAPI
	failureReport FailureReportAPI
}

// GetLogger returns a logger instance.
func (sc *Provider) GetLogger() *zerolog.Logger {
	return sc.logger
}

func (sc *Provider) GetTimeService() TimeAPI {
	return sc.time
}

func (sc *Provider) GetChannel() string {
	return sc.channel
}

// NewProvider return an instance of Provider based on given persistence layer.
func NewProvider(ctx context.Context, dbal persistence.DBAL, queue event.Queue, time TimeAPI, channel string) *Provider {
	return &Provider{
		logger:     log.Ctx(ctx),
		dbal:       dbal,
		eventQueue: queue,
		time:       time,
		channel:    channel,
	}
}

// GetOrganizationDBAL returns the database abstraction layer for Organizations
func (sc *Provider) GetOrganizationDBAL() persistence.OrganizationDBAL {
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

func (sc *Provider) GetFailureReportDBAL() persistence.FailureReportDBAL {
	return sc.dbal
}

// GetEventQueue returns an event.Queue instance
func (sc *Provider) GetEventQueue() event.Queue {
	return sc.eventQueue
}

// GetOrganizationService returns a OrganizationAPI instance.
// The service will be instanciated if needed.
func (sc *Provider) GetOrganizationService() OrganizationAPI {
	if sc.organization == nil {
		sc.organization = NewOrganizationService(sc)
	}
	return sc.organization
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

// GetFailureReportService returns a FailureAPI instance.
// The service will be instantiated if needed.
func (sc *Provider) GetFailureReportService() FailureReportAPI {
	if sc.failureReport == nil {
		sc.failureReport = NewFailureReportService(sc)
	}
	return sc.failureReport
}
