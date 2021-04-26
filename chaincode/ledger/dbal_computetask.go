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

package ledger

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	"github.com/go-playground/log/v7"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	"github.com/owkin/orchestrator/lib/errors"
)

// AddComputeTask stores a new ComputeTask in DB
func (db *DB) AddComputeTask(t *asset.ComputeTask) error {
	exists, err := db.hasKey(asset.ComputeTaskKind, t.GetKey())
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("failed to add compute task: %w", errors.ErrConflict)
	}

	bytes, err := json.Marshal(t)
	if err != nil {
		return err
	}

	err = db.putState(asset.ComputeTaskKind, t.GetKey(), bytes)
	if err != nil {
		return err
	}

	if err := db.createIndex("computeTask~category~worker~status~algo~key", []string{asset.ComputeTaskKind, t.Category.String(), t.Worker, t.Status.String(), t.Algo.Key, t.Key}); err != nil {
		return err
	}
	for _, parentTask := range t.ParentTaskKeys {
		err = db.createIndex("computeTask~parentTask~key", []string{asset.ComputeTaskKind, parentTask, t.Key})
		if err != nil {
			return err
		}
	}
	if t.ComputePlanKey != "" {
		if err := db.createIndex("computePlan~computePlanKey~worker~rank~key", []string{asset.ComputePlanKind, t.ComputePlanKey, t.Worker, strconv.Itoa(int(t.Rank)), t.Key}); err != nil {
			return err
		}
		if err := db.createIndex("algo~computeplankey~key", []string{asset.AlgoKind, t.ComputePlanKey, t.Algo.Key}); err != nil {
			return err
		}
	}

	if t.Category == asset.ComputeTaskCategory_TASK_TEST {
		testData, ok := t.Data.(*asset.ComputeTask_Test)
		if !ok {
			return fmt.Errorf("compute task data does not match task category: %w", errors.ErrInvalidAsset)
		}

		if err := db.createIndex("computeTask~category~objective~certified~key", []string{asset.ComputeTaskKind, t.Category.String(), testData.Test.ObjectiveKey, strconv.FormatBool(testData.Test.Certified), t.Key}); err != nil {
			return err
		}
		for _, parentTask := range t.ParentTaskKeys {
			if err = db.createIndex("computeTask~category~parentTask~certified~key", []string{asset.ComputeTaskKind, t.Category.String(), parentTask, strconv.FormatBool(testData.Test.Certified), t.Key}); err != nil {
				return err
			}
		}
	}

	return nil
}

// UpdateComputeTask updates an existing task.
func (db *DB) UpdateComputeTask(task *asset.ComputeTask) error {
	// We need the current task to be able to update its indexes
	prevTask, err := db.GetComputeTask(task.Key)
	if err != nil {
		return err
	}

	// We only handle status update for now
	prevStatus := prevTask.Status
	// Ignore status in comparison
	prevTask.Status = task.Status
	if !reflect.DeepEqual(prevTask, task) {
		// We only implement status update, so prevent any other update as it would require full index update
		return fmt.Errorf("only task status update is implemented: %w", errors.ErrUnimplemented)
	}
	prevTask.Status = prevStatus

	b, err := json.Marshal(task)
	if err != nil {
		return err
	}

	err = db.putState(asset.ComputeTaskKind, task.Key, b)
	if err != nil {
		return err
	}

	// Update status indexes
	if prevTask.Status != task.Status {
		err = db.updateIndex(
			"computeTask~category~worker~status~algo~key",
			[]string{asset.ComputeTaskKind, prevTask.Category.String(), prevTask.Worker, prevTask.Status.String()},
			[]string{asset.ComputeTaskKind, task.Category.String(), task.Worker, task.Status.String()},
		)
		if err != nil {
			return err
		}
	}
	return nil
}

// ComputeTaskExists returns true if a task with the given key exists
func (db *DB) ComputeTaskExists(key string) (bool, error) {
	exists, err := db.hasKey(asset.ComputeTaskKind, key)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// GetComputeTasks returns all compute tasks identified by the provided keys.
// It should not be used where pagination is expected!
func (db *DB) GetComputeTasks(keys []string) ([]*asset.ComputeTask, error) {
	tasks := make([]*asset.ComputeTask, 0, len(keys))

	for _, key := range keys {
		bytes, err := db.getState(asset.ComputeTaskKind, key)
		if err != nil {
			return nil, err
		}
		task := &asset.ComputeTask{}

		err = json.Unmarshal(bytes, task)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// GetComputeTask returns a ComputeTask by its key
func (db *DB) GetComputeTask(key string) (*asset.ComputeTask, error) {
	task := new(asset.ComputeTask)

	b, err := db.getState(asset.ComputeTaskKind, key)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, task)
	if err != nil {
		return nil, err
	}
	return task, nil
}

// GetComputeTaskChildren returns the children of the task identified by the given key
func (db *DB) GetComputeTaskChildren(key string) ([]*asset.ComputeTask, error) {
	elementKeys, err := db.getIndexKeys("computeTask~parentTask~key", []string{asset.ComputeTaskKind, key})
	if err != nil {
		return nil, err
	}

	db.logger.WithField("numChildren", len(elementKeys)).Debug("GetComputeTaskChildren")

	tasks := []*asset.ComputeTask{}
	for _, childKey := range elementKeys {
		task, err := db.GetComputeTask(childKey)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (db *DB) QueryComputeTasks(p *common.Pagination, filter *asset.TaskQueryFilter) ([]*asset.ComputeTask, common.PaginationToken, error) {
	logger := db.logger.WithFields(
		log.F("pagination", p),
		log.F("filter", filter),
	)
	logger.Debug("query compute task")

	selector := couchTaskQuery{
		DocType: asset.ComputeTaskKind,
	}

	assetFilter := map[string]interface{}{}

	if filter.Category != asset.ComputeTaskCategory_TASK_UNKNOWN {
		assetFilter["category"] = filter.Category.String()
	}
	if filter.Status != asset.ComputeTaskStatus_STATUS_UNKNOWN {
		assetFilter["status"] = filter.Status.String()
	}
	if filter.Worker != "" {
		assetFilter["worker"] = filter.Worker
	}

	if len(assetFilter) > 0 {
		selector.Asset = assetFilter
	}

	b, err := json.Marshal(selector)
	if err != nil {
		return nil, "", err
	}

	// query should be {"selector":{"doc_type":"computetask", "asset":{"status":2}}}
	queryString := fmt.Sprintf(`{"selector":%s}}`, string(b))
	log.WithField("couchQuery", queryString).Debug("mango query")

	resultsIterator, bookmark, err := db.getQueryResultWithPagination(queryString, int32(p.Size), p.Token)
	if err != nil {
		return nil, "", err
	}
	defer resultsIterator.Close()

	tasks := make([]*asset.ComputeTask, 0)

	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, "", err
		}
		var storedAsset storedAsset
		err = json.Unmarshal(queryResult.Value, &storedAsset)
		if err != nil {
			return nil, "", err
		}
		task := &asset.ComputeTask{}
		err = json.Unmarshal(storedAsset.Asset, task)
		if err != nil {
			return nil, "", err
		}

		tasks = append(tasks, task)
	}

	return tasks, bookmark.Bookmark, nil
}

type couchTaskQuery struct {
	DocType string                 `json:"doc_type"`
	Asset   map[string]interface{} `json:"asset,omitempty"`
}
