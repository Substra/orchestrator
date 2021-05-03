// Copyright 2021 Owkin Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package distributed

import (
	"context"

	"github.com/owkin/orchestrator/lib/asset"
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
	invocator, err := ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.computeplan:RegisterPlan"

	response := &asset.ComputePlan{}

	err = invocator.Call(method, in, response)

	return response, err
}

func (a *ComputePlanAdapter) GetPlan(ctx context.Context, param *asset.GetComputePlanParam) (*asset.ComputePlan, error) {
	invocator, err := ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.computeplan:GetPlan"

	response := &asset.ComputePlan{}

	err = invocator.Call(method, param, response)

	return response, err
}

func (a *ComputePlanAdapter) QueryPlans(ctx context.Context, param *asset.QueryPlansParam) (*asset.QueryPlansResponse, error) {
	invocator, err := ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.computeplan:QueryPlans"

	response := &asset.QueryPlansResponse{}

	err = invocator.Call(method, param, response)

	return response, err
}

func (a *ComputePlanAdapter) ApplyPlanAction(ctx context.Context, param *asset.ApplyPlanActionParam) (*asset.ApplyPlanActionResponse, error) {
	invocator, err := ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.computeplan:ApplyPlanAction"

	response := &asset.ApplyPlanActionResponse{}

	err = invocator.Call(method, param, nil)

	return response, err
}
