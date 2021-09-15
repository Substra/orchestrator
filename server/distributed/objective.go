package distributed

import (
	"context"
	"strings"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/errors"
)

// ObjectiveAdapter is a grpc server exposing the same node interface,
// but relies on a remote chaincode to actually manage the asset.
type ObjectiveAdapter struct {
	asset.UnimplementedObjectiveServiceServer
}

// NewObjectiveAdapter creates a Server
func NewObjectiveAdapter() *ObjectiveAdapter {
	return &ObjectiveAdapter{}
}

// RegisterObjective will add a new Objective to the network
func (a *ObjectiveAdapter) RegisterObjective(ctx context.Context, in *asset.NewObjective) (*asset.Objective, error) {
	invocator, err := ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.objective:RegisterObjective"

	response := &asset.Objective{}

	err = invocator.Call(ctx, method, in, response)

	if err != nil && isFabricTimeoutRetry(ctx) && strings.Contains(err.Error(), errors.ErrConflict) {
		// In this very specific case we are in a retry context after a timeout.
		// We can assume that the previous request succeeded and created the asset.
		// So we convert the error in a success response.
		err = invocator.Call(ctx, "orchestrator.objective:GetObjective", &asset.GetObjectiveParam{Key: in.Key}, response)
		return response, err
	}

	return response, err
}

// GetObjective returns an objective from its key
func (a *ObjectiveAdapter) GetObjective(ctx context.Context, query *asset.GetObjectiveParam) (*asset.Objective, error) {
	invocator, err := ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.objective:GetObjective"

	response := &asset.Objective{}

	err = invocator.Call(ctx, method, query, response)

	return response, err
}

// QueryObjectives returns all known objectives
func (a *ObjectiveAdapter) QueryObjectives(ctx context.Context, query *asset.QueryObjectivesParam) (*asset.QueryObjectivesResponse, error) {
	invocator, err := ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.objective:QueryObjectives"

	response := &asset.QueryObjectivesResponse{}

	err = invocator.Call(ctx, method, query, response)

	return response, err
}

// GetLeaderboard returns for an objective all its certified ComputeTask with ComputeTaskCategory: TEST_TASK with a done status
func (a *ObjectiveAdapter) GetLeaderboard(ctx context.Context, query *asset.LeaderboardQueryParam) (*asset.Leaderboard, error) {
	invocator, err := ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.objective:GetLeaderboard"

	response := &asset.Leaderboard{}

	err = invocator.Call(ctx, method, query, response)

	return response, err
}
