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
	UpdateComputeTask(task *asset.ComputeTask) error
	QueryComputeTasks(p *common.Pagination, filter *asset.TaskQueryFilter) ([]*asset.ComputeTask, common.PaginationToken, error)
	GetComputeTaskChildren(key string) ([]*asset.ComputeTask, error)
	// GetComputePlanTasks returns the tasks of the compute plan identified by the given key
	GetComputePlanTasks(key string) ([]*asset.ComputeTask, error)
	GetComputePlanTasksKeys(key string) ([]string, error)
}

type ComputeTaskDBALProvider interface {
	GetComputeTaskDBAL() ComputeTaskDBAL
}
