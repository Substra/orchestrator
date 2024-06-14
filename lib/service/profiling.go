package service

import (
	"github.com/substra/orchestrator/lib/asset"
	orcerrors "github.com/substra/orchestrator/lib/errors"
)

// ProfilingAPI defines the methods to act on Profiling
type ProfilingAPI interface {
	RegisterProfilingStep(function *asset.ProfilingStep) error
}

// ProfilingServiceProvider defines an object able to provide an FunctionAPI instance
type ProfilingServiceProvider interface {
	GetProfilingService() ProfilingAPI
}

// ProfilingDependencyProvider defines what the ProfilingService needs to perform its duty
type ProfilingDependencyProvider interface {
	LoggerProvider
	EventServiceProvider
}

// ProfilingService is the function manipulation entry point
// it implements the API interface
type ProfilingService struct {
	ProfilingDependencyProvider
}

// NewProfilingService will create a new service with given persistence layer
func NewProfilingService(provider ProfilingDependencyProvider) *ProfilingService {
	return &ProfilingService{provider}
}

// RegisterFunction persist an function
func (s *ProfilingService) RegisterProfilingStep(ps *asset.ProfilingStep) error {
	s.GetLogger().Debug().Msg("Registering profiling step")
	err := ps.Validate()
	if err != nil {
		return orcerrors.FromValidationError(asset.ProfilingStepKind, err)
	}

	event := &asset.Event{
		EventKind: asset.EventKind_EVENT_ASSET_CREATED,
		AssetKey:  ps.AssetKey,
		AssetKind: asset.AssetKind_ASSET_PROFILING_STEP,
		Asset:     &asset.Event_ProfilingStep{ProfilingStep: ps},
	}
	err = s.GetEventService().RegisterEvents(event)

	return err
}
