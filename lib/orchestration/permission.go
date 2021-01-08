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

package orchestration

import (
	"errors"

	"github.com/owkin/orchestrator/lib/assets"
	"github.com/owkin/orchestrator/utils"
)

// PermissionAPI defines the methods to act on Permissions
type PermissionAPI interface {
	CreatePermissions(owner string, newPerms *assets.NewPermissions) (*assets.Permissions, error)
}

// PermissionServiceProvider defines an object able to provide a PermissionAPI instance.
type PermissionServiceProvider interface {
	GetPermissionService() PermissionAPI
}

// PermissionDependencyProvider defines what the PermissionService needs to perform its duty
type PermissionDependencyProvider interface {
	NodeServiceProvider
}

// PermissionService is the entry point to manipulate permissions.
// it implements the API interface
type PermissionService struct {
	PermissionDependencyProvider
}

// NewPermissionService creates a new service
func NewPermissionService(provider PermissionDependencyProvider) *PermissionService {
	return &PermissionService{provider}
}

// CreatePermissions process a NewPermissions object into a Permissions one.
func (s *PermissionService) CreatePermissions(owner string, newPerms *assets.NewPermissions) (*assets.Permissions, error) {
	if !newPerms.Public {
		// Restricted access, let's check that authorizedIds are valid
		if err := s.validateAuthorizedIDs(newPerms.AuthorizedIds); err != nil {
			return &assets.Permissions{}, err
		}
	}

	defaultPerms := newPermission(newPerms, owner)

	// Download permission is not implemented in the node server, so let's use the same permissions for process & download
	permissions := &assets.Permissions{
		Process:  defaultPerms,
		Download: defaultPerms,
	}

	return permissions, nil
}

// validateAuthorizedIds checks that given IDs are valid nodes in the network.
// Returns nil if all IDs are valid, an Error otherwise
func (s *PermissionService) validateAuthorizedIDs(IDs []string) error {
	nodes, err := s.GetNodeService().GetNodes()
	if err != nil {
		return err
	}

	var nodeIDs []string

	for _, n := range nodes {
		nodeIDs = append(nodeIDs, n.Id)
	}

	for _, authorizedID := range IDs {
		if !utils.StringInSlice(nodeIDs, authorizedID) {
			return errors.New("Invalid permission input values")
		}
	}

	return nil
}

// newPermission processes a NewPermission into a Permission.
// This takes care of adding the owner to the authorized IDs.
func newPermission(newPerms *assets.NewPermissions, owner string) *assets.Permission {
	IDs := newPerms.AuthorizedIds

	// Owner must always be defined in the list of authorizedIDs, if the permission is private,
	// it will ease the merge of private permissions
	if !utils.StringInSlice(IDs, owner) {
		IDs = append(IDs, owner)
	}

	return &assets.Permission{
		Public:        newPerms.Public,
		AuthorizedIds: IDs,
	}
}
