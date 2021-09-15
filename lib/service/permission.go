package service

import (
	"github.com/owkin/orchestrator/lib/asset"
	orcerrors "github.com/owkin/orchestrator/lib/errors"
	"github.com/owkin/orchestrator/utils"
)

// PermissionAPI defines the methods to act on Permissions
type PermissionAPI interface {
	CreatePermissions(owner string, newPerms *asset.NewPermissions) (*asset.Permissions, error)
	CanProcess(perms *asset.Permissions, requester string) bool
	MakeIntersection(x, y *asset.Permissions) *asset.Permissions
	MakeUnion(x, y *asset.Permissions) *asset.Permissions
}

// PermissionServiceProvider defines an object able to provide a PermissionAPI instance.
type PermissionServiceProvider interface {
	GetPermissionService() PermissionAPI
}

// PermissionDependencyProvider defines what the PermissionService needs to perform its duty
type PermissionDependencyProvider interface {
	LoggerProvider
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
	nodes, err := s.GetNodeService().GetAllNodes()
	if err != nil {
		return err
	}

	var nodeIDs []string

	for _, n := range nodes {
		nodeIDs = append(nodeIDs, n.Id)
	}

	for _, authorizedID := range IDs {
		if !utils.StringInSlice(nodeIDs, authorizedID) {
			return orcerrors.NewBadRequest("invalid permission input values")
		}
	}

	return nil
}

func (s *PermissionService) CanProcess(perms *asset.Permissions, requester string) bool {
	if perms.Process.Public || utils.StringInSlice(perms.Process.AuthorizedIds, requester) {
		return true
	}
	s.GetLogger().WithField("requester", requester).WithField("permissions", perms).Debug("Requester can't process the asset")
	return false
}

func (s *PermissionService) MakeIntersection(x, y *asset.Permissions) *asset.Permissions {
	return &asset.Permissions{
		Process:  intersect(x.Process, y.Process),
		Download: intersect(x.Download, y.Download),
	}
}

func (s *PermissionService) MakeUnion(x, y *asset.Permissions) *asset.Permissions {
	return &asset.Permissions{
		Process:  union(x.Process, y.Process),
		Download: union(x.Download, y.Download),
	}
}

func intersect(x, y *asset.Permission) *asset.Permission {
	result := &asset.Permission{}
	result.Public = x.Public && y.Public

	switch {
	case !x.Public && y.Public:
		result.AuthorizedIds = x.AuthorizedIds
	case x.Public && !y.Public:
		result.AuthorizedIds = y.AuthorizedIds
	default:
		result.AuthorizedIds = utils.Intersection(x.AuthorizedIds, y.AuthorizedIds)
	}
	return result
}

func union(x, y *asset.Permission) *asset.Permission {
	result := &asset.Permission{}
	result.Public = x.Public || y.Public

	if !result.Public {
		uniqueIds := make(map[string]struct{})

		for _, id := range append(x.AuthorizedIds, y.AuthorizedIds...) {
			uniqueIds[id] = struct{}{}
		}

		authorizedIds := make([]string, len(uniqueIds))

		i := 0
		for id := range uniqueIds {
			authorizedIds[i] = id
			i++
		}

		result.AuthorizedIds = authorizedIds
	}

	return result
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
