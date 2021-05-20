// Copyright 2020 Owkin Inc.
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
	"fmt"

	"github.com/go-playground/log/v7"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	orcerrors "github.com/owkin/orchestrator/lib/errors"
	"github.com/owkin/orchestrator/lib/persistence"
	"github.com/owkin/orchestrator/utils"
)

// ObjectiveAPI defines the methods to act on Objectives
type ObjectiveAPI interface {
	RegisterObjective(objective *asset.NewObjective, owner string) (*asset.Objective, error)
	GetObjective(string) (*asset.Objective, error)
	QueryObjectives(p *common.Pagination) ([]*asset.Objective, common.PaginationToken, error)
	ObjectiveExists(key string) (bool, error)
	CanDownload(key string, requester string) (bool, error)
}

// ObjectiveServiceProvider defines an object able to provide an ObjectiveAPI instance
type ObjectiveServiceProvider interface {
	GetObjectiveService() ObjectiveAPI
}

// ObjectiveDependencyProvider defines what the ObjectiveService needs to perform its duty
type ObjectiveDependencyProvider interface {
	persistence.ObjectiveDBALProvider
	EventServiceProvider
	PermissionServiceProvider
	DataSampleServiceProvider
	DataManagerServiceProvider
}

// ObjectiveService is the objective manipulation entry point
// it implements the API interface
type ObjectiveService struct {
	ObjectiveDependencyProvider
}

// NewObjectiveService will create a new service with given persistence layer
func NewObjectiveService(provider ObjectiveDependencyProvider) *ObjectiveService {
	return &ObjectiveService{provider}
}

// RegisterObjective persist an objective
func (s *ObjectiveService) RegisterObjective(o *asset.NewObjective, owner string) (*asset.Objective, error) {
	log.WithField("owner", owner).WithField("newObj", o).Debug("Registering objective")
	err := o.Validate()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", orcerrors.ErrInvalidAsset, err.Error())
	}

	objective := &asset.Objective{
		Key:         o.Key,
		Name:        o.Name,
		Description: o.Description,
		MetricsName: o.MetricsName,
		Metrics:     o.Metrics,
		Metadata:    o.Metadata,
		Owner:       owner,
	}

	if o.DataManagerKey != "" {
		err := s.GetDataSampleService().CheckSameManager(o.GetDataManagerKey(), o.GetDataSampleKeys())
		if err != nil {
			return nil, err
		}
		testOnly, err := s.GetDataSampleService().IsTestOnly(o.GetDataSampleKeys())
		if err != nil {
			return nil, err
		}
		if !testOnly {
			return nil, fmt.Errorf("datasamples are not test only: %w", orcerrors.ErrInvalidAsset)
		}

		// Couple manager + samples is valid, let's associate them with the objective
		objective.DataManagerKey = o.GetDataManagerKey()
		objective.DataSampleKeys = o.GetDataSampleKeys()

	}

	objective.Permissions, err = s.GetPermissionService().CreatePermissions(owner, o.NewPermissions)
	if err != nil {
		return nil, err
	}

	event := &asset.Event{
		EventKind: asset.EventKind_EVENT_ASSET_CREATED,
		AssetKey:  o.Key,
		AssetKind: asset.AssetKind_ASSET_OBJECTIVE,
	}
	err = s.GetEventService().RegisterEvent(event)
	if err != nil {
		return nil, err
	}

	err = s.GetObjectiveDBAL().AddObjective(objective)

	if err != nil {
		return nil, err
	}

	if o.DataManagerKey != "" {
		// Associates an objective to a datamanager, more precisely, it adds the objective key to the datamanager
		dataManagerUpdate := &asset.DataManagerUpdateParam{
			Key:          objective.DataManagerKey,
			ObjectiveKey: objective.Key,
		}
		err = s.GetDataManagerService().UpdateDataManager(dataManagerUpdate, owner)
		if err != nil {
			return nil, fmt.Errorf("datamanager cannot be associated with the objective: %w", orcerrors.ErrBadRequest)
		}
	}

	return objective, err
}

// GetObjective retrieves an objective by its key
func (s *ObjectiveService) GetObjective(key string) (*asset.Objective, error) {
	return s.GetObjectiveDBAL().GetObjective(key)
}

// QueryObjectives returns all stored objectives
func (s *ObjectiveService) QueryObjectives(p *common.Pagination) ([]*asset.Objective, common.PaginationToken, error) {
	return s.GetObjectiveDBAL().QueryObjectives(p)
}

// ObjectiveExists checks if an objective with the provided key exists in the persistence layer
func (s *ObjectiveService) ObjectiveExists(key string) (bool, error) {
	return s.GetObjectiveDBAL().ObjectiveExists(key)
}

// CanDownload checks if the requester can download the objective corresponding to the provided key
func (s *ObjectiveService) CanDownload(key string, requester string) (bool, error) {
	obj, err := s.GetObjective(key)

	if err != nil {
		return false, err
	}

	if obj.Permissions.Download.Public || utils.StringInSlice(obj.Permissions.Download.AuthorizedIds, requester) {
		return true, nil
	}

	return false, nil
}
