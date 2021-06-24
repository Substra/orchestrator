package common

import (
	"regexp"

	"github.com/owkin/orchestrator/utils"
)

// ReadOnlyMethods maps for each service the "read only" methods.
// This mapping is also used in chaincode mode to determine Evaluate transactions
var ReadOnlyMethods = map[string][]string{
	"Objective":   {"GetObjective", "QueryObjectives", "GetLeaderboard"},
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
