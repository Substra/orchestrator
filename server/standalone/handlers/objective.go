package handlers

import (
	"context"

	"github.com/owkin/orchestrator/lib/asset"
	libCommon "github.com/owkin/orchestrator/lib/common"
	"github.com/owkin/orchestrator/server/common"
	"github.com/owkin/orchestrator/server/common/logger"

	"github.com/owkin/orchestrator/server/standalone/interceptors"
)

// ObjectiveServer is the gRPC facade to Objective manipulation
type ObjectiveServer struct {
	asset.UnimplementedObjectiveServiceServer
}

// NewObjectiveServer creates a grpc server
func NewObjectiveServer() *ObjectiveServer {
	return &ObjectiveServer{}
}

// RegisterObjective will persiste a new objective
func (s *ObjectiveServer) RegisterObjective(ctx context.Context, o *asset.NewObjective) (*asset.Objective, error) {
	logger.Get(ctx).WithField("objective", o).Debug("register objective")

	mspid, err := common.ExtractMSPID(ctx)
	if err != nil {
		return nil, err
	}
	services, err := interceptors.ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}

	return services.GetObjectiveService().RegisterObjective(o, mspid)
}

// GetObjective fetches an objective by its key
func (s *ObjectiveServer) GetObjective(ctx context.Context, params *asset.GetObjectiveParam) (*asset.Objective, error) {
	services, err := interceptors.ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}
	return services.GetObjectiveService().GetObjective(params.Key)
}

// QueryObjectives returns a paginated list of all known objectives
func (s *ObjectiveServer) QueryObjectives(ctx context.Context, params *asset.QueryObjectivesParam) (*asset.QueryObjectivesResponse, error) {
	services, err := interceptors.ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}

	objectives, paginationToken, err := services.GetObjectiveService().QueryObjectives(libCommon.NewPagination(params.PageToken, params.PageSize))
	if err != nil {
		return nil, err
	}

	return &asset.QueryObjectivesResponse{
		Objectives:    objectives,
		NextPageToken: paginationToken,
	}, nil
}

// GetLeaderboard returns for an objective all its certified ComputeTask with ComputeTaskCategory: TEST_TASK with a done status
func (s *ObjectiveServer) GetLeaderboard(ctx context.Context, query *asset.LeaderboardQueryParam) (*asset.Leaderboard, error) {
	services, err := interceptors.ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}
	return services.GetObjectiveService().GetLeaderboard(query)
}
