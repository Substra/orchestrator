// Copyright 2021 Owkin Inc.
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

// AlgoAPI defines the methods to act on Algos
type AlgoAPI interface {
	RegisterAlgo(algo *asset.NewAlgo, owner string) (*asset.Algo, error)
	GetAlgo(string) (*asset.Algo, error)
	GetAlgos(c asset.AlgoCategory, p *common.Pagination) ([]*asset.Algo, common.PaginationToken, error)
}

// AlgoServiceProvider defines an object able to provide an AlgoAPI instance
type AlgoServiceProvider interface {
	GetAlgoService() AlgoAPI
}

// AlgoDependencyProvider defines what the AlgoService needs to perform its duty
type AlgoDependencyProvider interface {
	persistence.AlgoDBALProvider
	event.QueueProvider
	PermissionServiceProvider
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
	log.WithField("owner", owner).WithField("newObj", a).Debug("Registering algo")
	err := a.Validate()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", orchestrationErrors.ErrInvalidAsset, err.Error())
	}

	exists, err := s.GetAlgoDBAL().AlgoExists(a.Key)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("There is already an algo with this key: %w", orchestrationErrors.ErrConflict)
	}

	algo := &asset.Algo{
		Key:         a.Key,
		Name:        a.Name,
		Category:    a.Category,
		Description: a.Description,
		Algorithm:   a.Algorithm,
		Metadata:    a.Metadata,
		Owner:       owner,
	}

	algo.Permissions, err = s.GetPermissionService().CreatePermissions(owner, a.NewPermissions)
	if err != nil {
		return nil, err
	}

	err = s.GetEventQueue().Enqueue(&event.Event{
		EventKind: event.AssetCreated,
		AssetID:   a.Key,
		AssetKind: asset.AlgoKind,
	})
	if err != nil {
		return &asset.Algo{}, err
	}

	err = s.GetAlgoDBAL().AddAlgo(algo)
	return algo, err
}

// GetAlgo retrieves an algo by its ID
func (s *AlgoService) GetAlgo(id string) (*asset.Algo, error) {
	return s.GetAlgoDBAL().GetAlgo(id)
}

// GetAlgos returns all stored algos
func (s *AlgoService) GetAlgos(c asset.AlgoCategory, p *common.Pagination) ([]*asset.Algo, common.PaginationToken, error) {
	return s.GetAlgoDBAL().GetAlgos(c, p)
}
