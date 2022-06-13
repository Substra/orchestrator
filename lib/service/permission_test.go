package service

import (
	"testing"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/stretchr/testify/assert"
)

func TestCreatePermission(t *testing.T) {
	mockOrganizationService := new(MockOrganizationAPI)
	provider := newMockedProvider()
	provider.On("GetOrganizationService").Once().Return(mockOrganizationService)
	service := NewPermissionService(provider)
	organizations := []*asset.Organization{{Id: "org"}}
	mockOrganizationService.On("GetAllOrganizations").Once().Return(organizations, nil)

	n := asset.NewPermissions{Public: false}
	owner := "org"
	permission, err := service.CreatePermission(owner, &n)

	assert.NoError(t, err)
	assert.Equal(t, permission, &asset.Permission{
		Public: false, AuthorizedIds: []string{owner},
	})
	provider.AssertExpectations(t)
}

func TestCreatePermissions(t *testing.T) {
	mockOrganizationService := new(MockOrganizationAPI)
	provider := newMockedProvider()
	provider.On("GetOrganizationService").Once().Return(mockOrganizationService)
	service := NewPermissionService(provider)
	organizations := []*asset.Organization{{Id: "org"}}
	mockOrganizationService.On("GetAllOrganizations").Once().Return(organizations, nil)

	n := asset.NewPermissions{Public: false}
	owner := "org"
	permissions, err := service.CreatePermissions(owner, &n)

	assert.NoError(t, err)
	p := &asset.Permission{
		Public: false, AuthorizedIds: []string{owner},
	}
	assert.Equal(t, permissions, &asset.Permissions{Download: p, Process: p})
	provider.AssertExpectations(t)
}

func TestNewPermission(t *testing.T) {
	n := asset.NewPermissions{Public: false}

	p := newPermission(&n, "org")

	assert.Contains(t, p.AuthorizedIds, "org", "owner should be added to authorized IDs")
}

func TestValidateAuthorizedIDs(t *testing.T) {
	mockOrganizationService := new(MockOrganizationAPI)
	provider := newMockedProvider()
	provider.On("GetOrganizationService").Return(mockOrganizationService)
	service := NewPermissionService(provider)

	organizations := []*asset.Organization{
		{Id: "org1"},
		{Id: "org2"},
	}
	mockOrganizationService.On("GetAllOrganizations").Return(organizations, nil)

	assert.Error(t, service.validateAuthorizedIDs([]string{"orgA"}), "orgA is not a valid organization")
	assert.NoError(t, service.validateAuthorizedIDs([]string{"org1"}), "org1 is a valid organization")
}

func TestCanProcess(t *testing.T) {
	cases := map[string]struct {
		perms     *asset.Permissions
		requester string
		outcome   bool
	}{
		"public": {
			perms:     &asset.Permissions{Process: &asset.Permission{Public: true}},
			requester: "org1",
			outcome:   true,
		},
		"allowed": {
			perms:     &asset.Permissions{Process: &asset.Permission{Public: false, AuthorizedIds: []string{"org1"}}},
			requester: "org1",
			outcome:   true,
		},
		"not allowed": {
			perms:     &asset.Permissions{Process: &asset.Permission{Public: false, AuthorizedIds: []string{"org1"}}},
			requester: "org2",
			outcome:   false,
		},
	}

	provider := newMockedProvider()
	service := NewPermissionService(provider)

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.outcome, service.CanProcess(tc.perms, tc.requester))
		})
	}
}

func TestMakeIntersection(t *testing.T) {
	cases := map[string]struct {
		a       *asset.Permissions
		b       *asset.Permissions
		outcome *asset.Permissions
	}{
		"public + organization": {
			a: &asset.Permissions{
				Process:  &asset.Permission{Public: true},
				Download: &asset.Permission{Public: true},
			},
			b: &asset.Permissions{
				Process:  &asset.Permission{Public: false, AuthorizedIds: []string{"org1"}},
				Download: &asset.Permission{Public: false, AuthorizedIds: []string{"org1"}},
			},
			outcome: &asset.Permissions{
				Process:  &asset.Permission{Public: false, AuthorizedIds: []string{"org1"}},
				Download: &asset.Permission{Public: false, AuthorizedIds: []string{"org1"}},
			},
		},
		"empty": {
			a: &asset.Permissions{
				Process:  &asset.Permission{Public: false, AuthorizedIds: []string{"org1"}},
				Download: &asset.Permission{Public: false, AuthorizedIds: []string{"org1"}},
			},
			b: &asset.Permissions{
				Process:  &asset.Permission{Public: false, AuthorizedIds: []string{"org2"}},
				Download: &asset.Permission{Public: false, AuthorizedIds: []string{"org2"}},
			},
			outcome: &asset.Permissions{
				Process:  &asset.Permission{Public: false, AuthorizedIds: []string{}},
				Download: &asset.Permission{Public: false, AuthorizedIds: []string{}},
			},
		},
		"common id": {
			a: &asset.Permissions{
				Process:  &asset.Permission{Public: false, AuthorizedIds: []string{"org1", "org2"}},
				Download: &asset.Permission{Public: false, AuthorizedIds: []string{"org1", "org2"}},
			},
			b: &asset.Permissions{
				Process:  &asset.Permission{Public: false, AuthorizedIds: []string{"org2"}},
				Download: &asset.Permission{Public: false, AuthorizedIds: []string{"org2"}},
			},
			outcome: &asset.Permissions{
				Process:  &asset.Permission{Public: false, AuthorizedIds: []string{"org2"}},
				Download: &asset.Permission{Public: false, AuthorizedIds: []string{"org2"}},
			},
		},
	}

	provider := newMockedProvider()
	service := NewPermissionService(provider)

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.outcome, service.IntersectPermissions(tc.a, tc.b))
			// merge should be commutative
			assert.Equal(t, tc.outcome, service.IntersectPermissions(tc.b, tc.a))
		})
	}
}

func TestMakeUnion(t *testing.T) {
	cases := map[string]struct {
		a       *asset.Permissions
		b       *asset.Permissions
		outcome *asset.Permissions
	}{
		"public + organization": {
			a: &asset.Permissions{
				Process:  &asset.Permission{Public: true},
				Download: &asset.Permission{Public: true},
			},
			b: &asset.Permissions{
				Process:  &asset.Permission{Public: false, AuthorizedIds: []string{"org1"}},
				Download: &asset.Permission{Public: false, AuthorizedIds: []string{"org1"}},
			},
			outcome: &asset.Permissions{
				Process:  &asset.Permission{Public: true},
				Download: &asset.Permission{Public: true},
			},
		},
		"two organizations": {
			a: &asset.Permissions{
				Process:  &asset.Permission{Public: false, AuthorizedIds: []string{"org1"}},
				Download: &asset.Permission{Public: false, AuthorizedIds: []string{"org1"}},
			},
			b: &asset.Permissions{
				Process:  &asset.Permission{Public: false, AuthorizedIds: []string{"org2"}},
				Download: &asset.Permission{Public: false, AuthorizedIds: []string{"org2"}},
			},
			outcome: &asset.Permissions{
				Process:  &asset.Permission{Public: false, AuthorizedIds: []string{"org1", "org2"}},
				Download: &asset.Permission{Public: false, AuthorizedIds: []string{"org1", "org2"}},
			},
		},
		"duplicates": {
			a: &asset.Permissions{
				Process:  &asset.Permission{Public: false, AuthorizedIds: []string{"org1", "org2"}},
				Download: &asset.Permission{Public: false, AuthorizedIds: []string{"org1", "org2"}},
			},
			b: &asset.Permissions{
				Process:  &asset.Permission{Public: false, AuthorizedIds: []string{"org2"}},
				Download: &asset.Permission{Public: false, AuthorizedIds: []string{"org2"}},
			},
			outcome: &asset.Permissions{
				Process:  &asset.Permission{Public: false, AuthorizedIds: []string{"org1", "org2"}},
				Download: &asset.Permission{Public: false, AuthorizedIds: []string{"org1", "org2"}},
			},
		},
	}

	provider := newMockedProvider()
	service := NewPermissionService(provider)

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			ab := service.UnionPermissions(tc.a, tc.b)
			ba := service.UnionPermissions(tc.b, tc.a)

			assert.Equal(t, ab.Process.Public, ba.Process.Public)
			assert.ElementsMatch(t, ab.Process.AuthorizedIds, ba.Process.AuthorizedIds)
			assert.ElementsMatch(t, ab.Process.AuthorizedIds, tc.outcome.Process.AuthorizedIds)
			assert.ElementsMatch(t, ab.Download.AuthorizedIds, tc.outcome.Download.AuthorizedIds)
		})
	}
}
