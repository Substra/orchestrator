package service

import (
	"github.com/substra/orchestrator/lib/asset"
	"github.com/substra/orchestrator/lib/common"
	orcerrors "github.com/substra/orchestrator/lib/errors"
	"github.com/substra/orchestrator/lib/persistence"
	"github.com/substra/orchestrator/utils"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// AlgoAPI defines the methods to act on Algos
type AlgoAPI interface {
	RegisterAlgo(algo *asset.NewAlgo, owner string) (*asset.Algo, error)
	GetAlgo(string) (*asset.Algo, error)
	QueryAlgos(p *common.Pagination, filter *asset.AlgoQueryFilter) ([]*asset.Algo, common.PaginationToken, error)
	CanDownload(key string, requester string) (bool, error)
	AlgoExists(key string) (bool, error)
	UpdateAlgo(algo *asset.UpdateAlgoParam, requester string) error
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
		Inputs:       a.Inputs,
		Outputs:      a.Outputs,
	}

	algo.Permissions, err = s.GetPermissionService().CreatePermissions(owner, a.NewPermissions)
	if err != nil {
		return nil, err
	}

	event := &asset.Event{
		EventKind: asset.EventKind_EVENT_ASSET_CREATED,
		AssetKey:  a.Key,
		AssetKind: asset.AssetKind_ASSET_ALGO,
		Asset:     &asset.Event_Algo{Algo: algo},
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

// CanDownload checks if the requester can download the algo corresponding to the provided key
func (s *AlgoService) CanDownload(key string, requester string) (bool, error) {
	obj, err := s.GetAlgo(key)

	if err != nil {
		return false, err
	}

	return obj.Permissions.Download.Public || utils.SliceContains(obj.Permissions.Download.AuthorizedIds, requester), nil
}

// AlgoExists returns true if the algo exists
func (s *AlgoService) AlgoExists(key string) (bool, error) {
	return s.GetAlgoDBAL().AlgoExists(key)
}

// UpdateAlgo updates mutable fields of an algo. List of mutable fields : name.
func (s *AlgoService) UpdateAlgo(a *asset.UpdateAlgoParam, requester string) error {
	s.GetLogger().WithField("requester", requester).WithField("algoUpdate", a).Debug("Updating algo")
	err := a.Validate()
	if err != nil {
		return orcerrors.FromValidationError(asset.AlgoKind, err)
	}

	algoKey := a.Key

	algo, err := s.GetAlgoDBAL().GetAlgo(algoKey)
	if err != nil {
		return orcerrors.NewNotFound(asset.AlgoKind, algoKey)
	}

	if algo.GetOwner() != requester {
		return orcerrors.NewPermissionDenied("requester does not own the algo")
	}

	// Update algo name
	algo.Name = a.Name

	event := &asset.Event{
		EventKind: asset.EventKind_EVENT_ASSET_UPDATED,
		AssetKey:  algoKey,
		AssetKind: asset.AssetKind_ASSET_ALGO,
		Asset:     &asset.Event_Algo{Algo: algo},
	}
	err = s.GetEventService().RegisterEvents(event)
	if err != nil {
		return err
	}

	return s.GetAlgoDBAL().UpdateAlgo(algo)
}
