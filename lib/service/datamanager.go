package service

import (
	"fmt"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	orcerrors "github.com/owkin/orchestrator/lib/errors"
	"github.com/owkin/orchestrator/lib/persistence"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// DataManagerAPI defines the methods to act on DataManagers
type DataManagerAPI interface {
	RegisterDataManager(datamanager *asset.NewDataManager, owner string) (*asset.DataManager, error)
	GetDataManager(key string) (*asset.DataManager, error)
	QueryDataManagers(p *common.Pagination) ([]*asset.DataManager, common.PaginationToken, error)
	CheckOwner(keys []string, requester string) error
	GetCheckedDataManager(key string, dataSampleKeys []string, owner string) (*asset.DataManager, error)
}

// DataManagerServiceProvider defines an object able to provide an DataManagerAPI instance
type DataManagerServiceProvider interface {
	GetDataManagerService() DataManagerAPI
}

// DataManagerDependencyProvider defines what the DataManagerService needs to perform its duty
type DataManagerDependencyProvider interface {
	LoggerProvider
	persistence.DataManagerDBALProvider
	PermissionServiceProvider
	EventServiceProvider
	TimeServiceProvider
	DataSampleServiceProvider
}

// DataManagerService is the DataManager manipulation entry point
// it implements the API interface
type DataManagerService struct {
	DataManagerDependencyProvider
}

// NewDataManagerService will create a new service with given persistence layer
func NewDataManagerService(provider DataManagerDependencyProvider) *DataManagerService {
	return &DataManagerService{provider}
}

type DataManagerPermissions struct {
	asset.Permission
}

// RegisterDataManager persists a DataManager
func (s *DataManagerService) RegisterDataManager(d *asset.NewDataManager, owner string) (*asset.DataManager, error) {
	s.GetLogger().WithField("owner", owner).WithField("newDataManager", d).Debug("Registering DataManager")
	err := d.Validate()
	if err != nil {
		return nil, orcerrors.FromValidationError(asset.DataManagerKind, err)
	}

	exists, err := s.GetDataManagerDBAL().DataManagerExists(d.GetKey())

	if err != nil {
		return nil, err
	}
	if exists {
		return nil, orcerrors.NewConflict(asset.DataManagerKind, d.Key)
	}

	datamanager := &asset.DataManager{
		Key:          d.GetKey(),
		Name:         d.GetName(),
		Owner:        owner,
		Description:  d.GetDescription(),
		Opener:       d.GetOpener(),
		Metadata:     d.GetMetadata(),
		Type:         d.GetType(),
		CreationDate: timestamppb.New(s.GetTimeService().GetTransactionTime()),
	}

	datamanager.LogsPermission, err = s.GetPermissionService().CreatePermission(owner, d.LogsPermission)
	if err != nil {
		return nil, err
	}

	datamanager.Permissions, err = s.GetPermissionService().CreatePermissions(owner, d.NewPermissions)
	if err != nil {
		return nil, err
	}

	event := &asset.Event{
		EventKind: asset.EventKind_EVENT_ASSET_CREATED,
		AssetKey:  datamanager.Key,
		AssetKind: asset.AssetKind_ASSET_DATA_MANAGER,
		Asset:     &asset.Event_DataManager{DataManager: datamanager},
	}
	err = s.GetEventService().RegisterEvents(event)
	if err != nil {
		return nil, err
	}

	err = s.GetDataManagerDBAL().AddDataManager(datamanager)

	if err != nil {
		return nil, err
	}

	return datamanager, nil
}

// GetDataManager retrieves a single DataManager by its key
func (s *DataManagerService) GetDataManager(key string) (*asset.DataManager, error) {
	return s.GetDataManagerDBAL().GetDataManager(key)
}

// QueryDataManagers returns all stored DataManagers
func (s *DataManagerService) QueryDataManagers(p *common.Pagination) ([]*asset.DataManager, common.PaginationToken, error) {
	return s.GetDataManagerDBAL().QueryDataManagers(p)
}

// CheckOwner validates that the DataManagerKeys are owned by the requester and return an error if that's not the case.
func (s *DataManagerService) CheckOwner(keys []string, requester string) error {
	for _, dataManagerKey := range keys {
		ok, err := s.isOwner(dataManagerKey, requester)
		if err != nil {
			return err
		}
		if !ok {
			return orcerrors.NewPermissionDenied(fmt.Sprintf("requester does not own the datamanager %q", dataManagerKey))
		}
	}
	return nil
}

// isOwner validates that the requester owns the asset
func (s *DataManagerService) isOwner(key string, requester string) (bool, error) {
	dm, err := s.GetDataManagerDBAL().GetDataManager(key)
	if err != nil {
		return false, orcerrors.NewNotFound(asset.DataManagerKind, key).Wrap(err)
	}

	return dm.GetOwner() == requester, nil
}

// GetCheckedDataManager returns the DataManager identified by the given key,
// it will return an error if the DataManager is not processable by owner or DataSamples don't share the common manager.
func (s *DataManagerService) GetCheckedDataManager(key string, dataSampleKeys []string, owner string) (*asset.DataManager, error) {
	datamanager, err := s.GetDataManager(key)
	if err != nil {
		return nil, err
	}
	canProcess := s.GetPermissionService().CanProcess(datamanager.Permissions, owner)
	if !canProcess {
		return nil, orcerrors.NewPermissionDenied(fmt.Sprintf("not authorized to process datamanager %q", datamanager.Key))
	}
	err = s.GetDataSampleService().CheckSameManager(key, dataSampleKeys)
	if err != nil {
		return nil, err
	}

	return datamanager, err
}
