package service

import (
	"testing"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/stretchr/testify/assert"
)

func TestNewPermission(t *testing.T) {
	n := asset.NewPermissions{Public: false}

	p := newPermission(&n, "org")

	assert.Contains(t, p.AuthorizedIds, "org", "owner should be added to authorized IDs")
}

func TestValidateAuthorizedIDs(t *testing.T) {
	mockNodeService := new(MockNodeService)
	provider := new(MockServiceProvider)
	provider.On("GetNodeService").Return(mockNodeService)
	service := NewPermissionService(provider)

	nodes := []*asset.Node{
		{Id: "org1"},
		{Id: "org2"},
	}
	mockNodeService.On("GetAllNodes").Return(nodes, nil)

	assert.Error(t, service.validateAuthorizedIDs([]string{"orgA"}), "orgA is not a valid node")
	assert.NoError(t, service.validateAuthorizedIDs([]string{"org1"}), "org1 is a valid node")
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

	provider := new(MockServiceProvider)
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
		"public + node": {
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

	provider := new(MockServiceProvider)
	service := NewPermissionService(provider)

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.outcome, service.MakeIntersection(tc.a, tc.b))
			// merge should be commutative
			assert.Equal(t, tc.outcome, service.MakeIntersection(tc.b, tc.a))
		})
	}
}

func TestMakeUnion(t *testing.T) {
	cases := map[string]struct {
		a       *asset.Permissions
		b       *asset.Permissions
		outcome *asset.Permissions
	}{
		"public + node": {
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
		"two nodes": {
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

	provider := new(MockServiceProvider)
	service := NewPermissionService(provider)

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			ab := service.MakeUnion(tc.a, tc.b)
			ba := service.MakeUnion(tc.b, tc.a)

			assert.Equal(t, ab.Process.Public, ba.Process.Public)
			assert.ElementsMatch(t, ab.Process.AuthorizedIds, ba.Process.AuthorizedIds)
			assert.ElementsMatch(t, ab.Process.AuthorizedIds, tc.outcome.Process.AuthorizedIds)
			assert.ElementsMatch(t, ab.Download.AuthorizedIds, tc.outcome.Download.AuthorizedIds)
		})
	}
}
