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
	orchestrationErrors "github.com/owkin/orchestrator/lib/errors"
	"github.com/owkin/orchestrator/lib/event"
	"github.com/owkin/orchestrator/lib/persistence"
)

// ObjectiveAPI defines the methods to act on Objectives
type ObjectiveAPI interface {
	RegisterObjective(objective *asset.NewObjective, owner string) (*asset.Objective, error)
	GetObjective(string) (*asset.Objective, error)
	GetObjectives(p *common.Pagination) ([]*asset.Objective, common.PaginationToken, error)
}

// ObjectiveServiceProvider defines an object able to provide an ObjectiveAPI instance
type ObjectiveServiceProvider interface {
	GetObjectiveService() ObjectiveAPI
}

// ObjectiveDependencyProvider defines what the ObjectiveService needs to perform its duty
type ObjectiveDependencyProvider interface {
	persistence.ObjectiveDBALProvider
	event.QueueProvider
	PermissionServiceProvider
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
		return nil, fmt.Errorf("%w: %s", orchestrationErrors.ErrInvalidAsset, err.Error())
	}

	testDataset := o.TestDataset
	if testDataset != nil { // nolint TODO
		// err = datasetService.RegisterDataset(testDataset)
		// if err != nil {
		//	return err
		// }
	}

	objective := &asset.Objective{
		Key:         o.Key,
		Name:        o.Name,
		TestDataset: testDataset,
		Description: o.Description,
		MetricsName: o.MetricsName,
		Metrics:     o.Metrics,
		Metadata:    o.Metadata,
		Owner:       owner,
	}

	objective.Permissions, err = s.GetPermissionService().CreatePermissions(owner, o.NewPermissions)
	if err != nil {
		return &asset.Objective{}, err
	}

	err = s.GetEventQueue().Enqueue(&event.Event{
		EventKind: event.AssetCreated,
		AssetID:   o.Key,
		AssetKind: asset.ObjectiveKind,
	})
	if err != nil {
		return &asset.Objective{}, err
	}

	err = s.GetObjectiveDBAL().AddObjective(objective)
	return objective, err
}

// GetObjective retrieves an objective by its ID
func (s *ObjectiveService) GetObjective(id string) (*asset.Objective, error) {
	return s.GetObjectiveDBAL().GetObjective(id)
}

// GetObjectives returns all stored objectives
func (s *ObjectiveService) GetObjectives(p *common.Pagination) ([]*asset.Objective, common.PaginationToken, error) {
	return s.GetObjectiveDBAL().GetObjectives(p)
}
