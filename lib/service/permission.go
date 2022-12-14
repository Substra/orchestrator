package service

import (
	"github.com/substra/orchestrator/lib/asset"
	orcerrors "github.com/substra/orchestrator/lib/errors"
	"github.com/substra/orchestrator/utils"
)

// PermissionAPI defines the methods to act on Permissions
type PermissionAPI interface {
	CreatePermission(owner string, newPerms *asset.NewPermissions) (*asset.Permission, error)
	CreatePermissions(owner string, newPerms *asset.NewPermissions) (*asset.Permissions, error)
	CanProcess(perms *asset.Permissions, requester string) bool
	IntersectPermissions(x, y *asset.Permissions) *asset.Permissions
	UnionPermissions(x, y *asset.Permissions) *asset.Permissions
	IntersectPermission(x, y *asset.Permission) *asset.Permission
	UnionPermission(x, y *asset.Permission) *asset.Permission
}

// PermissionServiceProvider defines an object able to provide a PermissionAPI instance.
type PermissionServiceProvider interface {
	GetPermissionService() PermissionAPI
}

// PermissionDependencyProvider defines what the PermissionService needs to perform its duty
type PermissionDependencyProvider interface {
	LoggerProvider
	OrganizationServiceProvider
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

// CreatePermission processes a NewPermissions object into a Permission one.
func (s *PermissionService) CreatePermission(owner string, newPerms *asset.NewPermissions) (*asset.Permission, error) {
	if newPerms == nil {
		newPerms = &asset.NewPermissions{
			Public: false,
		}
	}

	if !newPerms.Public {
		// Restricted access, let's check that authorizedIds are valid
		if err := s.validateAuthorizedIDs(newPerms.AuthorizedIds); err != nil {
			return nil, err
		}
	}

	return newPermission(newPerms, owner), nil
}

// CreatePermissions processes a NewPermissions object into a Permissions one.
func (s *PermissionService) CreatePermissions(owner string, newPerms *asset.NewPermissions) (*asset.Permissions, error) {
	defaultPerms, err := s.CreatePermission(owner, newPerms)
	if err != nil {
		return nil, err
	}

	// Download permission is not implemented in the organization server, so let's use the same permissions for process & download
	permissions := &asset.Permissions{
		Process:  defaultPerms,
		Download: defaultPerms,
	}

	return permissions, nil
}

// validateAuthorizedIds checks that given IDs are valid organizations in the network.
// Returns nil if all IDs are valid, an Error otherwise
func (s *PermissionService) validateAuthorizedIDs(ids []string) error {
	organizations, err := s.GetOrganizationService().GetAllOrganizations()
	if err != nil {
		return err
	}

	var organizationIDs []string

	for _, n := range organizations {
		organizationIDs = append(organizationIDs, n.Id)
	}

	for _, authorizedID := range ids {
		if !utils.SliceContains(organizationIDs, authorizedID) {
			return orcerrors.NewBadRequest("invalid permission input values")
		}
	}

	return nil
}

func (s *PermissionService) CanProcess(perms *asset.Permissions, requester string) bool {
	if perms.Process.Public || utils.SliceContains(perms.Process.AuthorizedIds, requester) {
		return true
	}
	s.GetLogger().Debug().Str("requester", requester).Interface("permissions", perms).Msg("Requester can't process the asset")
	return false
}

func (s *PermissionService) IntersectPermissions(x, y *asset.Permissions) *asset.Permissions {
	return &asset.Permissions{
		Process:  s.IntersectPermission(x.Process, y.Process),
		Download: s.IntersectPermission(x.Download, y.Download),
	}
}

func (s *PermissionService) UnionPermissions(x, y *asset.Permissions) *asset.Permissions {
	return &asset.Permissions{
		Process:  s.UnionPermission(x.Process, y.Process),
		Download: s.UnionPermission(x.Download, y.Download),
	}
}

func (s *PermissionService) IntersectPermission(x, y *asset.Permission) *asset.Permission {
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

func (s *PermissionService) UnionPermission(x, y *asset.Permission) *asset.Permission {
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
	if !utils.SliceContains(IDs, owner) {
		IDs = append(IDs, owner)
	}

	return &asset.Permission{
		Public:        newPerms.Public,
		AuthorizedIds: IDs,
	}
}
