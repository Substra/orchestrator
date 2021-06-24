package distributed

import (
	"context"
	"testing"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/stretchr/testify/assert"
)

func TestObjectiveAdapterImplementServer(t *testing.T) {
	adapter := NewObjectiveAdapter()
	assert.Implementsf(t, (*asset.ObjectiveServiceServer)(nil), adapter, "ObjectiveAdapter should implements ObjectiveServiceServer")
}

func TestRegisterObjective(t *testing.T) {
	adapter := NewObjectiveAdapter()

	newObj := &asset.NewObjective{
		Key: "uuid",
	}

	newCtx := context.TODO()
	invocator := &mockedInvocator{}

	invocator.On("Call", "orchestrator.objective:RegisterObjective", newObj, &asset.Objective{}).Return(nil)

	ctx := context.WithValue(newCtx, ctxInvocatorKey, invocator)

	_, err := adapter.RegisterObjective(ctx, newObj)

	assert.NoError(t, err, "Registration should pass")
}

func TestGetObjective(t *testing.T) {
	adapter := NewObjectiveAdapter()

	newCtx := context.TODO()
	invocator := &mockedInvocator{}

	param := &asset.GetObjectiveParam{Key: "uuid"}

	invocator.On("Call", "orchestrator.objective:GetObjective", param, &asset.Objective{}).Return(nil)

	ctx := context.WithValue(newCtx, ctxInvocatorKey, invocator)

	_, err := adapter.GetObjective(ctx, param)

	assert.NoError(t, err, "Query should pass")
}

func TestQueryObjectives(t *testing.T) {
	adapter := NewObjectiveAdapter()

	newCtx := context.TODO()
	invocator := &mockedInvocator{}

	param := &asset.QueryObjectivesParam{PageToken: "uuid", PageSize: 20}

	invocator.On("Call", "orchestrator.objective:QueryObjectives", param, &asset.QueryObjectivesResponse{}).Return(nil)

	ctx := context.WithValue(newCtx, ctxInvocatorKey, invocator)

	_, err := adapter.QueryObjectives(ctx, param)

	assert.NoError(t, err, "Query should pass")
}

func TestGetLeaderboard(t *testing.T) {
	adapter := NewObjectiveAdapter()

	newCtx := context.TODO()
	invocator := &mockedInvocator{}

	param := &asset.LeaderboardQueryParam{ObjectiveKey: "uuid", SortOrder: 0}

	invocator.On("Call", "orchestrator.objective:GetLeaderboard", param, &asset.Leaderboard{}).Return(nil)

	ctx := context.WithValue(newCtx, ctxInvocatorKey, invocator)

	_, err := adapter.GetLeaderboard(ctx, param)

	assert.NoError(t, err, "Query should pass")
}
