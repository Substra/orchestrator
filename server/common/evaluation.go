package common

import (
	"regexp"

	"github.com/substra/orchestrator/utils"
)

// ReadOnlyMethods maps for service "read only" methods.
// This mapping is used to set transactions as "read only"
var ReadOnlyMethods = map[string][]string{
	"Metric":        {"GetMetric", "QueryMetrics"},
	"Organization":  {"GetAllOrganizations"},
	"Function":      {"GetFunction", "QueryFunctions"},
	"Event":         {"QueryEvents"},
	"Model":         {"GetComputeTaskOutputModels", "CanDisableModel", "GetModel"},
	"Dataset":       {"GetDataset"},
	"DataSample":    {"GetDataSample", "QueryDataSamples"},
	"DataManager":   {"GetDataManager", "QueryDataManagers"},
	"ComputeTask":   {"QueryTasks", "GetTask", "GetTaskInputAssets"},
	"ComputePlan":   {"GetPlan", "QueryPlans", "IsPlanRunning"},
	"Performance":   {"QueryPerformances"},
	"Info":          {"QueryVersion"},
	"FailureReport": {"GetFailureReport"},
}

// TransactionChecker is able to characterize a transaction based on the gRPC method.
type TransactionChecker interface {
	// IsEvaluateMethod returns true if the gRPC method has no side effect (read-only).
	IsEvaluateMethod(method string) bool
}

type GrpcMethodChecker struct{}

/*
IsEvaluateMethod maps for each gRPC service its "read only" methods.
Those are methods which should not have any side effect on the storage,
ie: they should not write to the database or ledger.
It is used to initiate read-only transactions when possible.
*/
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

	return utils.SliceContains(methods, serviceMethod[2])
}
