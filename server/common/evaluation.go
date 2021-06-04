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

package common

import (
	"regexp"

	"github.com/owkin/orchestrator/utils"
)

// ReadOnlyMethods maps for each service the "read only" methods.
// This mapping is also used in chaincode mode to determine Evaluate transactions
var ReadOnlyMethods = map[string][]string{
	"Objective":   {"GetObjective", "QueryObjectives", "QueryLeaderboard"},
	"Node":        {"GetAllNodes"},
	"Algo":        {"GetAlgo", "QueryAlgos"},
	"Event":       {"QueryEvents"},
	"Model":       {"GetComputeTaskOutputModels", "GetComputeTaskInputModels", "CanDisableModel", "GetModel", "QueryModels"},
	"Dataset":     {"GetDataset"},
	"DataSample":  {"QueryDataSamples"},
	"DataManager": {"GetDataManager", "QueryDataManagers"},
	"ComputeTask": {"QueryTasks", "GetTask"},
	"ComputePlan": {"GetPlan", "QueryPlans"},
	"Performance": {"GetComputeTaskPerformance"},
}

// TransactionChecker is able to characterize a transaction based on the gRPC method.
type TransactionChecker interface {
	// IsEvaluateMethod returns true if the gRPC method has no side effect (read-only).
	IsEvaluateMethod(method string) bool
}

type GrpcMethodChecker struct{}

// IsEvaluateMethod returns true when the grpc method name given as input is read only.
// Input should be a grpc method name such as: "/orchestrator.ComputeTaskService/RegisterTask".
// Any unknown method will be considered as non-evaluate (read-write).
func (c GrpcMethodChecker) IsEvaluateMethod(method string) bool {
	re := regexp.MustCompile(`^/orchestrator\.(\w+)Service/(\w+)$`)
	serviceMethod := re.FindStringSubmatch(method)

	if len(serviceMethod) != 3 {
		// Unexpected format
		return false
	}

	methods, ok := ReadOnlyMethods[serviceMethod[1]]
	if !ok {
		// Service not found
		return false
	}

	return utils.StringInSlice(methods, serviceMethod[2])
}
