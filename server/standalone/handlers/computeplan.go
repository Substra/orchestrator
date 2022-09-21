package handlers

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/substra/orchestrator/lib/asset"
	libCommon "github.com/substra/orchestrator/lib/common"
	commonInterceptors "github.com/substra/orchestrator/server/common/interceptors"
	"github.com/substra/orchestrator/server/standalone/interceptors"
)

type ComputePlanServer struct {
	asset.UnimplementedComputePlanServiceServer
}

func NewComputePlanServer() *ComputePlanServer {
	return &ComputePlanServer{}
}

func (s *ComputePlanServer) RegisterPlan(ctx context.Context, in *asset.NewComputePlan) (*asset.ComputePlan, error) {
	owner, err := commonInterceptors.ExtractMSPID(ctx)
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

	plans, nextPage, err := provider.GetComputePlanService().QueryPlans(libCommon.NewPagination(param.PageToken, param.PageSize), param.Filter)
	if err != nil {
		return nil, err
	}

	return &asset.QueryPlansResponse{
		Plans:         plans,
		NextPageToken: nextPage,
	}, nil
}

func (s *ComputePlanServer) ApplyPlanAction(ctx context.Context, param *asset.ApplyPlanActionParam) (*asset.ApplyPlanActionResponse, error) {
	requester, err := commonInterceptors.ExtractMSPID(ctx)
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

// UpdatePlan will update mutable fields of the existing ComputePlan. List of mutable fields: name.
func (s *ComputePlanServer) UpdatePlan(ctx context.Context, params *asset.UpdateComputePlanParam) (*asset.UpdateComputePlanResponse, error) {
	log.Ctx(ctx).Debug().Interface("computeplan", params).Msg("Update Compute Plan")

	requester, err := commonInterceptors.ExtractMSPID(ctx)
	if err != nil {
		return nil, err
	}
	services, err := interceptors.ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}

	err = services.GetComputePlanService().UpdatePlan(params, requester)
	if err != nil {
		return nil, err
	}

	return &asset.UpdateComputePlanResponse{}, nil
}

func (s *ComputePlanServer) IsComputePlanRunning(ctx context.Context, param *asset.IsComputePlanRunningParam) (*asset.IsComputePlanRunningResponse, error) {
	provider, err := interceptors.ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}

	isRunning, err := provider.GetComputePlanService().IsComputePlanRunning(param.Key)
	if err != nil {
		return nil, err
	}

	return &asset.IsComputePlanRunningResponse{IsRunning: isRunning}, nil
}
