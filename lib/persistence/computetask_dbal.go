package persistence

import (
	"github.com/substra/orchestrator/lib/asset"
	"github.com/substra/orchestrator/lib/common"
)

// ComputeTaskOutputCounter counts registered outputs by identifier
type ComputeTaskOutputCounter = map[string]int

type ComputeTaskDBAL interface {
	// GetExistingComputeTaskKeys returns a slice with inputs keys existing in storage.
	GetExistingComputeTaskKeys(keys []string) ([]string, error)
	GetComputeTask(key string) (*asset.ComputeTask, error)
	GetComputeTasks(keys []string) ([]*asset.ComputeTask, error)
	AddComputeTasks(task ...*asset.ComputeTask) error
	UpdateComputeTaskStatus(taskKey string, taskStatus asset.ComputeTaskStatus) error
	QueryComputeTasks(p *common.Pagination, filter *asset.TaskQueryFilter) ([]*asset.ComputeTask, common.PaginationToken, error)
	GetComputeTaskChildren(key string) ([]*asset.ComputeTask, error)
	GetComputeTaskParents(key string) ([]*asset.ComputeTask, error)
	// GetComputePlanTasks returns the tasks of the compute plan identified by the given key
	GetComputePlanTasks(key string) ([]*asset.ComputeTask, error)
	GetComputePlanTasksKeys(key string) ([]string, error)
	AddComputeTaskOutputAsset(output *asset.ComputeTaskOutputAsset) error
	// CountComputeTaskRegisteredOutputs returns the number of registered outputs by identifier
	CountComputeTaskRegisteredOutputs(key string) (ComputeTaskOutputCounter, error)
	GetComputeTaskOutputAssets(taskKey, identifier string) ([]*asset.ComputeTaskOutputAsset, error)
}

type ComputeTaskDBALProvider interface {
	GetComputeTaskDBAL() ComputeTaskDBAL
}
