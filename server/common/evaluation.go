package common

import (
	"regexp"

	"github.com/substra/orchestrator/utils"
)

// ReadOnlyMethods maps for service "read only" methods.
// This mapping is used in chaincode mode to determine Evaluate transactions
// and in Standalone mode to set transactions as "read only"
var ReadOnlyMethods = map[string][]string{
	"Metric":        {"GetMetric", "QueryMetrics"},
	"Organization":  {"GetAllOrganizations"},
	"Algo":          {"GetAlgo", "QueryAlgos"},
	"Event":         {"QueryEvents"},
	"Model":         {"GetComputeTaskOutputModels", "CanDisableModel", "GetModel"},
	"Dataset":       {"GetDataset"},
	"DataSample":    {"GetDataSample", "QueryDataSamples"},
	"DataManager":   {"GetDataManager", "QueryDataManagers"},
	"ComputeTask":   {"QueryTasks", "GetTask", "GetTaskInputAssets"},
	"ComputePlan":   {"GetPlan", "QueryPlans", "IsComputePlanRunning"},
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
This mapping is used in distributed mode to flag non Evaluate transactions
and prevent the use of non-safe storage primitives in non evaluate context.
In Standalone mode it is used in a similar way to initiate read-only transactions
when possible.
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
