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
	"errors"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/utils"
)

// PermissionAPI defines the methods to act on Permissions
type PermissionAPI interface {
	CreatePermissions(owner string, newPerms *asset.NewPermissions) (*asset.Permissions, error)
	CanProcess(perms *asset.Permissions, requester string) bool
	MergePermissions(x, y *asset.Permissions) *asset.Permissions
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
func (s *PermissionService) CreatePermissions(owner string, newPerms *asset.NewPermissions) (*asset.Permissions, error) {
	if newPerms == nil {
		newPerms = &asset.NewPermissions{
			Public: false,
		}
	}

	if !newPerms.Public {
		// Restricted access, let's check that authorizedIds are valid
		if err := s.validateAuthorizedIDs(newPerms.AuthorizedIds); err != nil {
			return &asset.Permissions{}, err
		}
	}

	defaultPerms := newPermission(newPerms, owner)

	// Download permission is not implemented in the node server, so let's use the same permissions for process & download
	permissions := &asset.Permissions{
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

func (s *PermissionService) CanProcess(perms *asset.Permissions, requester string) bool {
	if perms.Process.Public || utils.StringInSlice(perms.Process.AuthorizedIds, requester) {
		return true
	}

	return false
}

func (s *PermissionService) MergePermissions(x, y *asset.Permissions) *asset.Permissions {
	return &asset.Permissions{
		Process:  mergePermission(x.Process, y.Process),
		Download: mergePermission(x.Download, y.Download),
	}
}

func mergePermission(x, y *asset.Permission) *asset.Permission {
	priv := &asset.Permission{}
	priv.Public = x.Public && y.Public

	switch {
	case !x.Public && y.Public:
		priv.AuthorizedIds = x.AuthorizedIds
	case x.Public && !y.Public:
		priv.AuthorizedIds = y.AuthorizedIds
	default:
		priv.AuthorizedIds = utils.Intersection(x.AuthorizedIds, y.AuthorizedIds)
	}
	return priv
}

// newPermission processes a NewPermission into a Permission.
// This takes care of adding the owner to the authorized IDs.
func newPermission(newPerms *asset.NewPermissions, owner string) *asset.Permission {
	IDs := newPerms.AuthorizedIds

	// Owner must always be defined in the list of authorizedIDs, if the permission is private,
	// it will ease the merge of private permissions
	if !utils.StringInSlice(IDs, owner) {
		IDs = append(IDs, owner)
	}

	return &asset.Permission{
		Public:        newPerms.Public,
		AuthorizedIds: IDs,
	}
}
