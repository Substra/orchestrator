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

package orchestration

import (
	"encoding/json"

	"github.com/owkin/orchestrator/lib/assets"
	"github.com/owkin/orchestrator/lib/persistence"
)

const objectiveResource = "objectives"

// ObjectiveAPI defines the methods to act on Objectives
type ObjectiveAPI interface {
	RegisterObjective(*assets.Objective) error
	GetObjective(string) (*assets.Objective, error)
	GetObjectives() ([]*assets.Objective, error)
}

// ObjectiveServiceProvider defines an object able to provide an ObjectiveAPI instance
type ObjectiveServiceProvider interface {
	GetObjectiveService() ObjectiveAPI
}

// ObjectiveDependencyProvider defines what the ObjectiveService needs to perform its duty
type ObjectiveDependencyProvider interface {
	persistence.DatabaseProvider
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
func (s *ObjectiveService) RegisterObjective(o *assets.Objective) error {
	b, err := json.Marshal(o)
	if err != nil {
		return err
	}

	err = o.Validate()
	if err != nil {
		return err
	}

	testDataset := o.TestDataset
	if testDataset != nil {
		// err = datasetService.RegisterDataset(testDataset)
		// if err != nil {
		//	return err
		// }
	}

	// This will use known nodes and tx creator
	// o.Permissions := NewPermissions()

	s.GetDatabase().PutState(objectiveResource, o.GetKey(), b)
	return nil
}

// GetObjective retrieves an objective by its ID
func (s *ObjectiveService) GetObjective(id string) (*assets.Objective, error) {
	o := assets.Objective{}

	b, err := s.GetDatabase().GetState(objectiveResource, id)
	if err != nil {
		return &o, err
	}

	err = json.Unmarshal(b, &o)
	return &o, err
}

// GetObjectives returns all stored objectives
func (s *ObjectiveService) GetObjectives() ([]*assets.Objective, error) {
	b, err := s.GetDatabase().GetAll(objectiveResource)
	if err != nil {
		return nil, err
	}

	var objectives []*assets.Objective

	for _, nodeBytes := range b {
		o := assets.Objective{}
		err = json.Unmarshal(nodeBytes, &o)
		if err != nil {
			return nil, err
		}
		objectives = append(objectives, &o)
	}

	return objectives, nil
}
