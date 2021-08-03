package distributed

import (
	"context"

	"github.com/owkin/orchestrator/lib/asset"
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
