package service

import (
	"github.com/substra/orchestrator/lib/asset"
	"github.com/substra/orchestrator/lib/common"
	orcerrors "github.com/substra/orchestrator/lib/errors"
	"github.com/substra/orchestrator/lib/persistence"
	"github.com/substra/orchestrator/utils"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// FunctionAPI defines the methods to act on Functions
type FunctionAPI interface {
	RegisterFunction(function *asset.NewFunction, owner string) (*asset.Function, error)
	GetFunction(string) (*asset.Function, error)
	QueryFunctions(p *common.Pagination, filter *asset.FunctionQueryFilter) ([]*asset.Function, common.PaginationToken, error)
	CanDownload(key string, requester string) (bool, error)
	FunctionExists(key string) (bool, error)
	UpdateFunction(function *asset.UpdateFunctionParam, requester string) error
}

// FunctionServiceProvider defines an object able to provide an FunctionAPI instance
type FunctionServiceProvider interface {
	GetFunctionService() FunctionAPI
}

// FunctionDependencyProvider defines what the FunctionService needs to perform its duty
type FunctionDependencyProvider interface {
	LoggerProvider
	persistence.FunctionDBALProvider
	EventServiceProvider
	PermissionServiceProvider
	TimeServiceProvider
}

// FunctionService is the function manipulation entry point
// it implements the API interface
type FunctionService struct {
	FunctionDependencyProvider
}

// NewFunctionService will create a new service with given persistence layer
func NewFunctionService(provider FunctionDependencyProvider) *FunctionService {
	return &FunctionService{provider}
}

// RegisterFunction persist an function
func (s *FunctionService) RegisterFunction(a *asset.NewFunction, owner string) (*asset.Function, error) {
	s.GetLogger().Debug().Str("owner", owner).Interface("newFunction", a).Msg("Registering function")
	err := a.Validate()
	if err != nil {
		return nil, orcerrors.FromValidationError(asset.FunctionKind, err)
	}

	exists, err := s.GetFunctionDBAL().FunctionExists(a.Key)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, orcerrors.NewConflict(asset.FunctionKind, a.Key)
	}

	function := &asset.Function{
		Key:          a.Key,
		Name:         a.Name,
		Description:  a.Description,
		Function:    a.Function,
		Metadata:     a.Metadata,
		Owner:        owner,
		CreationDate: timestamppb.New(s.GetTimeService().GetTransactionTime()),
		Inputs:       a.Inputs,
		Outputs:      a.Outputs,
	}

	function.Permissions, err = s.GetPermissionService().CreatePermissions(owner, a.NewPermissions)
	if err != nil {
		return nil, err
	}

	event := &asset.Event{
		EventKind: asset.EventKind_EVENT_ASSET_CREATED,
		AssetKey:  a.Key,
		AssetKind: asset.AssetKind_ASSET_ALGO,
		Asset:     &asset.Event_Function{Function: function},
	}
	err = s.GetEventService().RegisterEvents(event)

	if err != nil {
		return nil, err
	}

	err = s.GetFunctionDBAL().AddFunction(function)

	if err != nil {
		return nil, err
	}

	return function, nil
}

// GetFunction retrieves an function by its key
func (s *FunctionService) GetFunction(key string) (*asset.Function, error) {
	return s.GetFunctionDBAL().GetFunction(key)
}

// QueryFunctions returns all stored functions
func (s *FunctionService) QueryFunctions(p *common.Pagination, filter *asset.FunctionQueryFilter) ([]*asset.Function, common.PaginationToken, error) {
	return s.GetFunctionDBAL().QueryFunctions(p, filter)
}

// CanDownload checks if the requester can download the function corresponding to the provided key
func (s *FunctionService) CanDownload(key string, requester string) (bool, error) {
	obj, err := s.GetFunction(key)

	if err != nil {
		return false, err
	}

	return obj.Permissions.Download.Public || utils.SliceContains(obj.Permissions.Download.AuthorizedIds, requester), nil
}

// FunctionExists returns true if the function exists
func (s *FunctionService) FunctionExists(key string) (bool, error) {
	return s.GetFunctionDBAL().FunctionExists(key)
}

// UpdateFunction updates mutable fields of an function. List of mutable fields : name.
func (s *FunctionService) UpdateFunction(a *asset.UpdateFunctionParam, requester string) error {
	s.GetLogger().Debug().Str("requester", requester).Interface("functionUpdate", a).Msg("Updating function")
	err := a.Validate()
	if err != nil {
		return orcerrors.FromValidationError(asset.FunctionKind, err)
	}

	functionKey := a.Key

	function, err := s.GetFunctionDBAL().GetFunction(functionKey)
	if err != nil {
		return orcerrors.NewNotFound(asset.FunctionKind, functionKey)
	}

	if function.GetOwner() != requester {
		return orcerrors.NewPermissionDenied("requester does not own the function")
	}

	// Update function name
	function.Name = a.Name

	event := &asset.Event{
		EventKind: asset.EventKind_EVENT_ASSET_UPDATED,
		AssetKey:  functionKey,
		AssetKind: asset.AssetKind_ASSET_ALGO,
		Asset:     &asset.Event_Function{Function: function},
	}
	err = s.GetEventService().RegisterEvents(event)
	if err != nil {
		return err
	}

	return s.GetFunctionDBAL().UpdateFunction(function)
}
