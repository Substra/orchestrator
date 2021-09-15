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
	UpdateDataManager(datamanager *asset.DataManagerUpdateParam, requester string) error
	GetDataManager(key string) (*asset.DataManager, error)
	QueryDataManagers(p *common.Pagination) ([]*asset.DataManager, common.PaginationToken, error)
	CheckOwner(keys []string, requester string) error
}

// DataManagerServiceProvider defines an object able to provide an DataManagerAPI instance
type DataManagerServiceProvider interface {
	GetDataManagerService() DataManagerAPI
}

// DataManagerDependencyProvider defines what the DataManagerService needs to perform its duty
type DataManagerDependencyProvider interface {
	LoggerProvider
	persistence.DataManagerDBALProvider
	ObjectiveServiceProvider
	PermissionServiceProvider
	EventServiceProvider
	TimeServiceProvider
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

	// The objective key should be empty or referencing a valid objective
	if d.GetObjectiveKey() != "" {
		ok, err := s.GetObjectiveService().CanDownload(d.GetObjectiveKey(), owner)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, orcerrors.NewPermissionDenied("datamanager owner cannot download the objective")
		}
	}

	datamanager := &asset.DataManager{
		Key:          d.GetKey(),
		Name:         d.GetName(),
		Owner:        owner,
		ObjectiveKey: d.GetObjectiveKey(),
		Description:  d.GetDescription(),
		Opener:       d.GetOpener(),
		Metadata:     d.GetMetadata(),
		Type:         d.GetType(),
		CreationDate: timestamppb.New(s.GetTimeService().GetTransactionTime()),
	}

	datamanager.Permissions, err = s.GetPermissionService().CreatePermissions(owner, d.NewPermissions)
	if err != nil {
		return nil, err
	}

	event := &asset.Event{
		EventKind: asset.EventKind_EVENT_ASSET_CREATED,
		AssetKey:  d.Key,
		AssetKind: asset.AssetKind_ASSET_DATA_MANAGER,
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

// UpdateDataManager updates a DataManager to link an objective
func (s *DataManagerService) UpdateDataManager(d *asset.DataManagerUpdateParam, requester string) error {
	s.GetLogger().WithField("owner", requester).WithField("dataManagerUpdate", d).Debug("updating data manager")
	err := d.Validate()
	if err != nil {
		return orcerrors.FromValidationError(asset.DataManagerKind, err)
	}

	datamanager, err := s.GetDataManagerDBAL().GetDataManager(d.GetKey())
	if err != nil {
		return orcerrors.NewNotFound(asset.DataManagerKind, d.Key).Wrap(err)
	}

	if !s.GetPermissionService().CanProcess(datamanager.Permissions, requester) {
		return orcerrors.NewPermissionDenied("requester does not have the permissions to update the datamanager")
	}

	if datamanager.GetObjectiveKey() != "" {
		return orcerrors.NewBadRequest("datamanager already has an objective")
	}

	// Validation of the asset existence is implicit because to check download permissions the ObjectiveService needs to query the
	// persistence layer
	ok, err := s.GetObjectiveService().CanDownload(d.GetObjectiveKey(), datamanager.Owner)
	if err != nil {
		return err
	}
	if !ok {
		return orcerrors.NewPermissionDenied("the datamanager owner can't download the provided objective")
	}

	datamanager.ObjectiveKey = d.GetObjectiveKey()

	event := &asset.Event{
		EventKind: asset.EventKind_EVENT_ASSET_UPDATED,
		AssetKey:  d.Key,
		AssetKind: asset.AssetKind_ASSET_DATA_MANAGER,
	}
	err = s.GetEventService().RegisterEvents(event)
	if err != nil {
		return err
	}

	return s.GetDataManagerDBAL().UpdateDataManager(datamanager)
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
