package service

import (
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	orcerrors "github.com/owkin/orchestrator/lib/errors"
	"github.com/owkin/orchestrator/lib/persistence"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// AlgoAPI defines the methods to act on Algos
type AlgoAPI interface {
	RegisterAlgo(algo *asset.NewAlgo, owner string) (*asset.Algo, error)
	GetAlgo(string) (*asset.Algo, error)
	QueryAlgos(p *common.Pagination, filter *asset.AlgoQueryFilter) ([]*asset.Algo, common.PaginationToken, error)
}

// AlgoServiceProvider defines an object able to provide an AlgoAPI instance
type AlgoServiceProvider interface {
	GetAlgoService() AlgoAPI
}

// AlgoDependencyProvider defines what the AlgoService needs to perform its duty
type AlgoDependencyProvider interface {
	LoggerProvider
	persistence.AlgoDBALProvider
	EventServiceProvider
	PermissionServiceProvider
	TimeServiceProvider
}

// AlgoService is the algo manipulation entry point
// it implements the API interface
type AlgoService struct {
	AlgoDependencyProvider
}

// NewAlgoService will create a new service with given persistence layer
func NewAlgoService(provider AlgoDependencyProvider) *AlgoService {
	return &AlgoService{provider}
}

// RegisterAlgo persist an algo
func (s *AlgoService) RegisterAlgo(a *asset.NewAlgo, owner string) (*asset.Algo, error) {
	s.GetLogger().WithField("owner", owner).WithField("newObj", a).Debug("Registering algo")
	err := a.Validate()
	if err != nil {
		return nil, orcerrors.FromValidationError(asset.AlgoKind, err)
	}

	exists, err := s.GetAlgoDBAL().AlgoExists(a.Key)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, orcerrors.NewConflict(asset.AlgoKind, a.Key)
	}

	algo := &asset.Algo{
		Key:          a.Key,
		Name:         a.Name,
		Category:     a.Category,
		Description:  a.Description,
		Algorithm:    a.Algorithm,
		Metadata:     a.Metadata,
		Owner:        owner,
		CreationDate: timestamppb.New(s.GetTimeService().GetTransactionTime()),
	}

	algo.Permissions, err = s.GetPermissionService().CreatePermissions(owner, a.NewPermissions)
	if err != nil {
		return nil, err
	}

	event := &asset.Event{
		EventKind: asset.EventKind_EVENT_ASSET_CREATED,
		AssetKey:  a.Key,
		AssetKind: asset.AssetKind_ASSET_ALGO,
	}
	err = s.GetEventService().RegisterEvents(event)

	if err != nil {
		return nil, err
	}

	err = s.GetAlgoDBAL().AddAlgo(algo)

	if err != nil {
		return nil, err
	}

	return algo, nil
}

// GetAlgo retrieves an algo by its key
func (s *AlgoService) GetAlgo(key string) (*asset.Algo, error) {
	return s.GetAlgoDBAL().GetAlgo(key)
}

// QueryAlgos returns all stored algos
func (s *AlgoService) QueryAlgos(p *common.Pagination, filter *asset.AlgoQueryFilter) ([]*asset.Algo, common.PaginationToken, error) {
	return s.GetAlgoDBAL().QueryAlgos(p, filter)
}
