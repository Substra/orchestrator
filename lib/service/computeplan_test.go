package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/substra/orchestrator/lib/asset"
	"github.com/substra/orchestrator/lib/common"
	orcerrors "github.com/substra/orchestrator/lib/errors"
	"github.com/substra/orchestrator/lib/persistence"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestGetPlan(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	provider := newMockedProvider()

	provider.On("GetComputePlanDBAL").Return(dbal)

	service := NewComputePlanService(provider)

	cp := &asset.ComputePlan{Key: "uuid", Owner: "org1", Tag: "test", Name: "My Test"}

	dbal.On("GetComputePlan", "uuid").Once().Return(cp, nil)

	plan, err := service.GetPlan("uuid")
	assert.NoError(t, err)
	assert.Equal(t, cp, plan, "service should set task counters")

	dbal.AssertExpectations(t)
}

func TestRegisterPlan(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	es := new(MockEventAPI)
	ts := new(MockTimeAPI)
	provider := newMockedProvider()

	provider.On("GetEventService").Return(es)
	provider.On("GetComputePlanDBAL").Return(dbal)
	provider.On("GetTimeService").Return(ts)

	ts.On("GetTransactionTime").Once().Return(time.Unix(1337, 0))

	service := NewComputePlanService(provider)

	newPlan := &asset.NewComputePlan{Key: "b9b3ecda-0a90-41da-a2e3-945eeafb06d8", Tag: "test", Name: "My Test"}

	expected := &asset.ComputePlan{
		Key:          "b9b3ecda-0a90-41da-a2e3-945eeafb06d8",
		Tag:          "test",
		Name:         "My Test",
		Owner:        "org1",
		CreationDate: timestamppb.New(time.Unix(1337, 0)),
	}

	dbal.On("AddComputePlan", expected).Once().Return(nil)
	dbal.On("ComputePlanExists", "b9b3ecda-0a90-41da-a2e3-945eeafb06d8").Once().Return(false, nil)

	expectedEvent := &asset.Event{
		AssetKind: asset.AssetKind_ASSET_COMPUTE_PLAN,
		AssetKey:  newPlan.Key,
		EventKind: asset.EventKind_EVENT_ASSET_CREATED,
		Asset:     &asset.Event_ComputePlan{ComputePlan: expected},
	}
	es.On("RegisterEvents", expectedEvent).Once().Return(nil)

	plan, err := service.RegisterPlan(newPlan, "org1")
	assert.NoError(t, err)
	assert.Equal(t, expected, plan)

	es.AssertExpectations(t)
	dbal.AssertExpectations(t)
	ts.AssertExpectations(t)
}

func TestCancelPlan(t *testing.T) {
	ts := new(MockTimeAPI)
	dbal := new(persistence.MockDBAL)
	es := new(MockEventAPI)
	provider := newMockedProvider()

	provider.On("GetTimeService").Return(ts)
	provider.On("GetComputePlanDBAL").Return(dbal)
	provider.On("GetEventService").Return(es)

	service := NewComputePlanService(provider)

	plan := &asset.ComputePlan{
		Key:   "b9b3ecda-0a90-41da-a2e3-945eeafb06d8",
		Tag:   "test",
		Name:  "My Test",
		Owner: "owner",
	}

	ts.On("GetTransactionTime").Return(time.Unix(1337, 0))

	dbal.On("CancelComputePlan", plan, mock.AnythingOfType("time.Time")).Return(nil)
	dbal.On("CancelComputePlan", plan, mock.AnythingOfType("time.Time")).Return(orcerrors.NewBadRequest("already canceled"))

	expectedEvent := &asset.Event{
		AssetKind: asset.AssetKind_ASSET_COMPUTE_PLAN,
		AssetKey:  plan.Key,
		EventKind: asset.EventKind_EVENT_ASSET_UPDATED,
		Asset: &asset.Event_ComputePlan{ComputePlan: &asset.ComputePlan{
			Key:             plan.Key,
			Tag:             plan.Tag,
			Name:            plan.Name,
			Owner:           plan.Owner,
			CancelationDate: timestamppb.New(time.Unix(1337, 0)),
		}},
	}
	es.On("RegisterEvents", expectedEvent).Once().Return(nil)
	err := service.cancelPlan(plan)

	assert.NoError(t, err)

	plan.CancelationDate = timestamppb.Now()
	err = service.cancelPlan(plan)
	assert.ErrorContains(t, err, "compute plan "+plan.Key+" is already terminated")

	ts.AssertExpectations(t)
	dbal.AssertExpectations(t)
	es.AssertExpectations(t)
}

func TestComputePlanAllowIntermediaryModelDeletion(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	provider := newMockedProvider()

	provider.On("GetComputePlanDBAL").Return(dbal)

	service := NewComputePlanService(provider)

	cp := &asset.ComputePlan{
		Key:                      "uuid",
		DeleteIntermediaryModels: true,
	}

	dbal.On("GetComputePlan", "uuid").Once().Return(cp, nil)

	canDelete, err := service.canDeleteModels("uuid")
	assert.NoError(t, err)

	assert.True(t, canDelete)

	dbal.AssertExpectations(t)
}

func TestQueryPlans(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	provider := newMockedProvider()

	provider.On("GetComputePlanDBAL").Return(dbal)

	service := NewComputePlanService(provider)

	pagination := common.NewPagination("", 2)
	filter := &asset.PlanQueryFilter{
		Owner: "owner",
	}

	returnedPlans := []*asset.ComputePlan{{}, {}}

	dbal.On("QueryComputePlans", pagination, filter).Once().Return(returnedPlans, "", nil)

	plans, _, err := service.QueryPlans(pagination, filter)
	assert.NoError(t, err)

	assert.Len(t, plans, 2)
}

func TestComputePlanExists(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	provider := newMockedProvider()

	provider.On("GetComputePlanDBAL").Return(dbal)

	service := NewComputePlanService(provider)

	dbal.On("ComputePlanExists", "uuid").Once().Return(false, nil)

	exist, err := service.computePlanExists("uuid")
	assert.NoError(t, err)
	assert.False(t, exist)
}

func TestUpdateSingleExistingComputePlan(t *testing.T) {
	dbal := new(persistence.MockDBAL)
	provider := newMockedProvider()
	es := new(MockEventAPI)
	provider.On("GetEventService").Return(es)
	provider.On("GetComputePlanDBAL").Return(dbal)
	service := NewComputePlanService(provider)

	existingComputePlan := &asset.ComputePlan{
		Key:   "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83",
		Name:  "compute plan name",
		Owner: "owner",
	}

	updateComputePlanParam := &asset.UpdateComputePlanParam{
		Key:  "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83",
		Name: "Updated compute plan name",
	}

	storedComputePlan := &asset.ComputePlan{
		Key:   "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83",
		Name:  "Updated compute plan name",
		Owner: "owner",
	}

	e := &asset.Event{
		EventKind: asset.EventKind_EVENT_ASSET_UPDATED,
		AssetKind: asset.AssetKind_ASSET_COMPUTE_PLAN,
		AssetKey:  storedComputePlan.Key,
		Asset:     &asset.Event_ComputePlan{ComputePlan: storedComputePlan},
	}

	cases := map[string]struct {
		requester string
		valid     bool
	}{
		"update successful": {
			requester: "owner",
			valid:     true,
		},
		"update rejected: requester is not owner": {
			requester: "user",
			valid:     false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			dbal.On("GetComputePlan", existingComputePlan.GetKey()).Return(existingComputePlan, nil).Once()

			if tc.valid {
				dbal.On("SetComputePlanName", storedComputePlan, storedComputePlan.Name).Return(nil).Once()
				es.On("RegisterEvents", e).Once().Return(nil)
			}

			err := service.UpdatePlan(updateComputePlanParam, tc.requester)

			if tc.valid {
				assert.NoError(t, err, "Update of compute plan should not fail")
			} else {
				assert.Error(t, err, "Update of compute plan should fail")
			}

			dbal.AssertExpectations(t)
			es.AssertExpectations(t)
		})
	}

}
