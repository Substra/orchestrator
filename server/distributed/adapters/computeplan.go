package adapters

import (
	"context"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/server/distributed/interceptors"
)

// ComputePlanAdapter is a grpc server exposing the same plan interface,
// but relies on a remote chaincode to actually manage the asset.
type ComputePlanAdapter struct {
	asset.UnimplementedComputePlanServiceServer
}

// NewComputePlanAdapter creates a Server
func NewComputePlanAdapter() *ComputePlanAdapter {
	return &ComputePlanAdapter{}
}

func (a *ComputePlanAdapter) RegisterPlan(ctx context.Context, in *asset.NewComputePlan) (*asset.ComputePlan, error) {
	invocator, err := interceptors.ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.computeplan:RegisterPlan"

	response := &asset.ComputePlan{}

	err = invocator.Call(ctx, method, in, response)

	return response, err
}

func (a *ComputePlanAdapter) GetPlan(ctx context.Context, param *asset.GetComputePlanParam) (*asset.ComputePlan, error) {
	invocator, err := interceptors.ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.computeplan:GetPlan"

	response := &asset.ComputePlan{}

	err = invocator.Call(ctx, method, param, response)

	return response, err
}

func (a *ComputePlanAdapter) QueryPlans(ctx context.Context, param *asset.QueryPlansParam) (*asset.QueryPlansResponse, error) {
	invocator, err := interceptors.ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.computeplan:QueryPlans"

	response := &asset.QueryPlansResponse{}

	err = invocator.Call(ctx, method, param, response)

	return response, err
}

func (a *ComputePlanAdapter) ApplyPlanAction(ctx context.Context, param *asset.ApplyPlanActionParam) (*asset.ApplyPlanActionResponse, error) {
	invocator, err := interceptors.ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.computeplan:ApplyPlanAction"

	response := &asset.ApplyPlanActionResponse{}

	err = invocator.Call(ctx, method, param, nil)

	return response, err
}

// UpdatePlan will update a ComputePlan from the state
func (a *ComputePlanAdapter) UpdatePlan(ctx context.Context, param *asset.UpdateComputePlanParam) (*asset.UpdateComputePlanResponse, error) {
	invocator, err := interceptors.ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.computeplan:UpdatePlan"

	response := &asset.UpdateComputePlanResponse{}

	err = invocator.Call(ctx, method, param, nil)

	return response, err
}
