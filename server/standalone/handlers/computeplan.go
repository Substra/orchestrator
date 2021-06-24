package handlers

import (
	"context"

	"github.com/owkin/orchestrator/lib/asset"
	libCommon "github.com/owkin/orchestrator/lib/common"
	"github.com/owkin/orchestrator/server/common"

	"github.com/owkin/orchestrator/server/standalone/interceptors"
)

type ComputePlanServer struct {
	asset.UnimplementedComputePlanServiceServer
}

func NewComputePlanServer() *ComputePlanServer {
	return &ComputePlanServer{}
}

func (s *ComputePlanServer) RegisterPlan(ctx context.Context, in *asset.NewComputePlan) (*asset.ComputePlan, error) {
	owner, err := common.ExtractMSPID(ctx)
	if err != nil {
		return nil, err
	}
	provider, err := interceptors.ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}

	plan, err := provider.GetComputePlanService().RegisterPlan(in, owner)

	if err != nil {
		return nil, err
	}

	return plan, nil
}

func (s *ComputePlanServer) GetPlan(ctx context.Context, param *asset.GetComputePlanParam) (*asset.ComputePlan, error) {
	provider, err := interceptors.ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}

	plan, err := provider.GetComputePlanService().GetPlan(param.Key)

	if err != nil {
		return nil, err
	}

	return plan, nil
}

func (s *ComputePlanServer) QueryPlans(ctx context.Context, param *asset.QueryPlansParam) (*asset.QueryPlansResponse, error) {
	provider, err := interceptors.ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}

	plans, nextPage, err := provider.GetComputePlanService().GetPlans(libCommon.NewPagination(param.PageToken, param.PageSize))
	if err != nil {
		return nil, err
	}

	return &asset.QueryPlansResponse{
		Plans:         plans,
		NextPageToken: nextPage,
	}, nil
}

func (s *ComputePlanServer) ApplyPlanAction(ctx context.Context, param *asset.ApplyPlanActionParam) (*asset.ApplyPlanActionResponse, error) {
	requester, err := common.ExtractMSPID(ctx)
	if err != nil {
		return nil, err
	}
	provider, err := interceptors.ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}

	err = provider.GetComputePlanService().ApplyPlanAction(param.Key, param.Action, requester)
	if err != nil {
		return nil, err
	}

	return &asset.ApplyPlanActionResponse{}, nil
}
