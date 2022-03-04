package ledger

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/go-playground/log/v7"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	"github.com/owkin/orchestrator/lib/errors"
	"github.com/owkin/orchestrator/utils"
	"google.golang.org/protobuf/encoding/protojson"
)

// addComputeTask stores a new ComputeTask in DB
func (db *DB) addComputeTask(t *asset.ComputeTask) error {
	exists, err := db.hasKey(asset.ComputeTaskKind, t.GetKey())
	if err != nil {
		return err
	}
	if exists {
		return errors.NewConflict(asset.ComputeTaskKind, t.Key)
	}

	bytes, err := marshaller.Marshal(t)
	if err != nil {
		return err
	}

	err = db.putState(asset.ComputeTaskKind, t.GetKey(), bytes)
	if err != nil {
		return err
	}

	err = db.createIndex(computePlanTaskStatusIndex, []string{asset.ComputePlanKind, t.ComputePlanKey, t.Status.String(), t.Key})
	if err != nil {
		return err
	}
	for _, parentTask := range t.ParentTaskKeys {
		err = db.createIndex(computeTaskParentIndex, []string{asset.ComputeTaskKind, parentTask, t.Key})
		if err != nil {
			return err
		}
	}
	return nil
}

func (db *DB) AddComputeTasks(tasks ...*asset.ComputeTask) error {
	for _, task := range tasks {
		err := db.addComputeTask(task)
		if err != nil {
			return err
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
		return errors.NewError(errors.ErrUnimplemented, "only task status update is implemented")
	}
	prevTask.Status = prevStatus

	b, err := marshaller.Marshal(task)
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
			computePlanTaskStatusIndex,
			[]string{asset.ComputePlanKind, prevTask.ComputePlanKey, prevTask.Status.String(), prevTask.Key},
			[]string{asset.ComputePlanKind, task.ComputePlanKey, task.Status.String(), task.Key},
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

func (db *DB) GetExistingComputeTaskKeys(keys []string) ([]string, error) {
	uniqueKeys := utils.UniqueString(keys)
	existingKeys := []string{}

	for _, key := range uniqueKeys {
		exist, err := db.ComputeTaskExists(key)
		if err != nil {
			return nil, err
		}
		if exist {
			existingKeys = append(existingKeys, key)
		}
	}

	return existingKeys, nil
}

// GetComputeTasks returns the list of unique compute tasks identified by the provided keys.
// It should not be used where pagination is expected!
func (db *DB) GetComputeTasks(keys []string) ([]*asset.ComputeTask, error) {
	keys = utils.UniqueString(keys)
	tasks := make([]*asset.ComputeTask, 0, len(keys))

	for _, key := range keys {
		bytes, err := db.getState(asset.ComputeTaskKind, key)
		if err != nil {
			return nil, err
		}
		task := &asset.ComputeTask{}

		err = protojson.Unmarshal(bytes, task)
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

	err = protojson.Unmarshal(b, task)
	if err != nil {
		return nil, err
	}
	return task, nil
}

// GetComputeTaskChildren returns the children of the task identified by the given key
func (db *DB) GetComputeTaskChildren(key string) ([]*asset.ComputeTask, error) {
	elementKeys, err := db.getIndexKeys(computeTaskParentIndex, []string{asset.ComputeTaskKind, key})
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

// GetComputePlanTasksKeys returns the list of task keys from the provided compute plan
func (db *DB) GetComputePlanTasksKeys(key string) ([]string, error) {
	keys, err := db.getIndexKeys(computePlanTaskStatusIndex, []string{asset.ComputePlanKind, key})
	if err != nil {
		return nil, err
	}

	return keys, nil
}

// GetComputePlanTasks returns the tasks of the compute plan identified by the given key
func (db *DB) GetComputePlanTasks(key string) ([]*asset.ComputeTask, error) {
	elementKeys, err := db.GetComputePlanTasksKeys(key)
	if err != nil {
		return nil, err
	}

	db.logger.WithField("numChildren", len(elementKeys)).Debug("GetComputePlanTasks")

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

	query := richQuerySelector{
		Selector: couchAssetQuery{
			DocType: asset.ComputeTaskKind,
		},
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
	if filter.ComputePlanKey != "" {
		assetFilter["compute_plan_key"] = filter.ComputePlanKey
	}
	if filter.AlgoKey != "" {
		assetFilter["algo"] = json.RawMessage(fmt.Sprintf(`{"key": "%s"}`, filter.AlgoKey))
	}

	if len(assetFilter) > 0 {
		query.Selector.Asset = assetFilter
	}

	b, err := json.Marshal(query)
	if err != nil {
		return nil, "", err
	}

	queryString := string(b)

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
		err = protojson.Unmarshal(storedAsset.Asset, task)
		if err != nil {
			return nil, "", err
		}

		tasks = append(tasks, task)
	}
	return tasks, bookmark.Bookmark, nil
}
