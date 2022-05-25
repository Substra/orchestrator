package persistence

import (
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
)

type ComputeTaskDBAL interface {
	ComputeTaskExists(key string) (bool, error)
	// GetExistingKeys returns a slice with inputs keys existing in storage.
	// The implementer should deal with duplicate keys.
	GetExistingComputeTaskKeys(keys []string) ([]string, error)
	GetComputeTask(key string) (*asset.ComputeTask, error)
	GetComputeTasks(keys []string) ([]*asset.ComputeTask, error)
	AddComputeTasks(task ...*asset.ComputeTask) error
	UpdateComputeTaskStatus(taskKey string, taskStatus asset.ComputeTaskStatus) error
	QueryComputeTasks(p *common.Pagination, filter *asset.TaskQueryFilter) ([]*asset.ComputeTask, common.PaginationToken, error)
	GetComputeTaskChildren(key string) ([]*asset.ComputeTask, error)
	// GetComputePlanTasks returns the tasks of the compute plan identified by the given key
	GetComputePlanTasks(key string) ([]*asset.ComputeTask, error)
	GetComputePlanTasksKeys(key string) ([]string, error)
}

type ComputeTaskDBALProvider interface {
	GetComputeTaskDBAL() ComputeTaskDBAL
}
