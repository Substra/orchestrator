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

// Package objective defines the Objective asset.
// An Objective is the hypothesis against which a model is trained/evaluated.
package objective

import (
	"encoding/json"

	"github.com/substrafoundation/substra-orchestrator/lib/persistence"
)

const resource = "objectives"

// API defines the methods to act on Objectives
type API interface {
	RegisterObjective(*Objective) error
	GetObjective(string) (*Objective, error)
	GetObjectives() ([]*Objective, error)
}

// Service is the objective manipulation entry point
// it implements the API interface
type Service struct {
	db persistence.Database
}

// NewService will create a new service with given persistence layer
func NewService(db persistence.Database) *Service {
	return &Service{db: db}
}

// RegisterObjective persist an objective
func (s *Service) RegisterObjective(o *Objective) error {
	b, err := json.Marshal(o)
	if err != nil {
		return err
	}

	s.db.PutState(resource, o.GetKey(), b)
	return nil
}

// GetObjective retrieves an objective by its ID
func (s *Service) GetObjective(id string) (*Objective, error) {
	o := Objective{}

	b, err := s.db.GetState(resource, id)
	if err != nil {
		return &o, err
	}

	err = json.Unmarshal(b, &o)
	return &o, err
}

// GetObjectives returns all stored objectives
func (s *Service) GetObjectives() ([]*Objective, error) {
	b, err := s.db.GetAll(resource)
	if err != nil {
		return nil, err
	}

	var objectives []*Objective

	for _, nodeBytes := range b {
		o := Objective{}
		err = json.Unmarshal(nodeBytes, &o)
		if err != nil {
			return nil, err
		}
		objectives = append(objectives, &o)
	}

	return objectives, nil
}
