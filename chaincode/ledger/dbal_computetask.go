package ledger

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/substra/orchestrator/lib/asset"
	"github.com/substra/orchestrator/lib/common"
	orcerrors "github.com/substra/orchestrator/lib/errors"
	"github.com/substra/orchestrator/lib/persistence"
	"github.com/substra/orchestrator/lib/service"
	"github.com/substra/orchestrator/utils"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// addComputeTask stores a new ComputeTask in DB
func (db *DB) addComputeTask(t *asset.ComputeTask) error {
	exists, err := db.hasKey(asset.ComputeTaskKind, t.GetKey())
	if err != nil {
		return err
	}
	if exists {
		return orcerrors.NewConflict(asset.ComputeTaskKind, t.Key)
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

	err = db.createIndex(computeTaskFunctionStatusIndex, []string{asset.ComputeTaskKind, t.FunctionKey, t.Status.String(), t.Key})
	if err != nil {
		return err
	}
	for _, parentTask := range service.GetParentTaskKeys(t.Inputs) {
		err = db.createIndex(computeTaskParentIndex, []string{asset.ComputeTaskKind, parentTask, t.Key})
		if err != nil {
			return err
		}
		err = db.createIndex(computeTaskChildIndex, []string{asset.ComputeTaskKind, t.Key, parentTask})
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

// UpdateComputeTaskStatus updates the status of an existing task.
// We only implement status update, as any other update would require full index update.
func (db *DB) UpdateComputeTaskStatus(taskKey string, taskStatus asset.ComputeTaskStatus) error {
	// We need the current task to be able to update its indexes
	prevTask, err := db.GetComputeTask(taskKey)
	if err != nil {
		return err
	}

	updatedTask := proto.Clone(prevTask).(*asset.ComputeTask)
	updatedTask.Status = taskStatus

	b, err := marshaller.Marshal(updatedTask)
	if err != nil {
		return err
	}

	err = db.putState(asset.ComputeTaskKind, taskKey, b)
	if err != nil {
		return err
	}

	// Update status indexes
	if prevTask.Status != updatedTask.Status {
		err = db.updateIndex(
			computePlanTaskStatusIndex,
			[]string{asset.ComputePlanKind, prevTask.ComputePlanKey, prevTask.Status.String(), prevTask.Key},
			[]string{asset.ComputePlanKind, updatedTask.ComputePlanKey, updatedTask.Status.String(), taskKey},
		)
		if err != nil {
			return err
		}
		err = db.updateIndex(
			computeTaskFunctionStatusIndex,
			[]string{asset.ComputeTaskKind, prevTask.FunctionKey, prevTask.Status.String(), prevTask.Key},
			[]string{asset.ComputeTaskKind, updatedTask.FunctionKey, updatedTask.Status.String(), updatedTask.Key},
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (db *DB) GetExistingComputeTaskKeys(keys []string) ([]string, error) {
	uniqueKeys := utils.Unique(keys)
	existingKeys := []string{}

	for _, key := range uniqueKeys {
		exist, err := db.hasKey(asset.ComputeTaskKind, key)
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
	keys = utils.Unique(keys)
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

	db.logger.Debug().Int("numChildren", len(elementKeys)).Msg("GetComputeTaskChildren")

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

// GetComputeTaskParents returns the children of the task identified by the given key
func (db *DB) GetComputeTaskParents(key string) ([]*asset.ComputeTask, error) {
	elementKeys, err := db.getIndexKeys(computeTaskChildIndex, []string{asset.ComputeTaskKind, key})
	if err != nil {
		return nil, err
	}

	db.logger.Debug().Int("numParents", len(elementKeys)).Msg("GetComputeTaskParents")

	tasks := []*asset.ComputeTask{}
	for _, parentKey := range elementKeys {
		task, err := db.GetComputeTask(parentKey)
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

	db.logger.Debug().Int("numChildren", len(elementKeys)).Msg("GetComputePlanTasks")

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
	logger := db.logger.With().
		Interface("pagination", p).
		Interface("filter", filter).
		Logger()
	logger.Debug().Msg("query compute task")

	query := richQuerySelector{
		Selector: couchAssetQuery{
			DocType: asset.ComputeTaskKind,
		},
	}

	if filter != nil {
		assetFilter := map[string]interface{}{}

		if filter.Status != asset.ComputeTaskStatus_STATUS_UNKNOWN {
			assetFilter["status"] = filter.Status.String()
		}
		if filter.Worker != "" {
			assetFilter["worker"] = filter.Worker
		}
		if filter.ComputePlanKey != "" {
			assetFilter["compute_plan_key"] = filter.ComputePlanKey
		}
		if filter.FunctionKey != "" {
			assetFilter["function_key"] = json.RawMessage(fmt.Sprintf(`{"key": "%s"}`, filter.FunctionKey))
		}

		if len(assetFilter) > 0 {
			query.Selector.Asset = assetFilter
		}
	}

	b, err := json.Marshal(query)
	if err != nil {
		return nil, "", err
	}

	queryString := string(b)

	logger.Debug().Str("couchQuery", queryString).Msg("mango query")

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

type storableComputeTaskOutputAsset struct {
	ComputeTaskKey              string `json:"compute_task_key"`
	ComputeTaskOutputIdentifier string `json:"compute_task_output_identifier"`
	AssetKind                   string `json:"asset_kind"`
	AssetKey                    string `json:"asset_key"`
}

func newStorableTaskOutputAsset(o *asset.ComputeTaskOutputAsset) *storableComputeTaskOutputAsset {
	return &storableComputeTaskOutputAsset{
		ComputeTaskKey:              o.ComputeTaskKey,
		ComputeTaskOutputIdentifier: o.ComputeTaskOutputIdentifier,
		AssetKind:                   o.AssetKind.String(),
		AssetKey:                    o.AssetKey,
	}
}

func (s *storableComputeTaskOutputAsset) newComputeTaskOutputAsset() (*asset.ComputeTaskOutputAsset, error) {
	assetKind, ok := asset.AssetKind_value[s.AssetKind]
	if !ok {
		return nil, orcerrors.NewUnimplemented(fmt.Sprintf("Unsupported asset kind: %q", s.AssetKind))
	}

	return &asset.ComputeTaskOutputAsset{
		ComputeTaskKey:              s.ComputeTaskKey,
		ComputeTaskOutputIdentifier: s.ComputeTaskOutputIdentifier,
		AssetKind:                   asset.AssetKind(assetKind),
		AssetKey:                    s.AssetKey,
	}, nil
}

func (db *DB) getTaskOutputAssets(taskKey string) ([]*storableComputeTaskOutputAsset, error) {
	outputAssets := make([]*storableComputeTaskOutputAsset, 0)

	exist, err := db.hasKey(asset.ComputeTaskOutputAssetKind, taskKey)
	if err != nil {
		return nil, err
	}
	if exist {
		bytes, err := db.getState(asset.ComputeTaskOutputAssetKind, taskKey)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(bytes, &outputAssets)
		if err != nil {
			return nil, err
		}
	}

	return outputAssets, nil
}

func (db *DB) AddComputeTaskOutputAsset(output *asset.ComputeTaskOutputAsset) error {
	outputs, err := db.getTaskOutputAssets(output.ComputeTaskKey)
	if err != nil {
		return err
	}
	db.logger.Debug().
		Interface("output", output).
		Interface("existingOutputs", outputs).
		Msg("adding compute task output asset")

	outputs = append(outputs, newStorableTaskOutputAsset(output))

	bytes, err := json.Marshal(outputs)
	if err != nil {
		return err
	}
	return db.putState(asset.ComputeTaskOutputAssetKind, output.ComputeTaskKey, bytes)
}

func (db *DB) CountComputeTaskRegisteredOutputs(key string) (persistence.ComputeTaskOutputCounter, error) {
	counter := make(persistence.ComputeTaskOutputCounter)

	assets, err := db.getTaskOutputAssets(key)
	if err != nil {
		// If there is no outputs registered, getTaskOutputAssets will return an error
		orcErr := new(orcerrors.OrcError)
		if errors.As(err, &orcErr) && orcErr.Kind == orcerrors.ErrNotFound {
			return counter, nil
		}
		return counter, err
	}

	for _, asset := range assets {
		counter[asset.ComputeTaskOutputIdentifier]++
	}

	return counter, nil
}

func (db *DB) GetComputeTaskOutputAssets(taskKey, identifier string) ([]*asset.ComputeTaskOutputAsset, error) {
	storedAssets, err := db.getTaskOutputAssets(taskKey)
	if err != nil {
		return nil, err
	}
	outputAssets := make([]*asset.ComputeTaskOutputAsset, 0)

	for _, storedAsset := range storedAssets {
		if storedAsset.ComputeTaskOutputIdentifier == identifier {
			output, err := storedAsset.newComputeTaskOutputAsset()
			if err != nil {
				return nil, err
			}
			outputAssets = append(outputAssets, output)
		}
	}

	return outputAssets, nil
}


func (db *DB) GetFunctionRunnableTasksKeys(key string) ([]string, error) {
	keysTodo, err := db.getIndexKeys(computeTaskFunctionStatusIndex, []string{asset.ComputeTaskKind, asset.ComputeTaskStatus_STATUS_TODO.String(), key})
	if err != nil {
		return nil, err
	}

	keysDoing, err := db.getIndexKeys(computeTaskFunctionStatusIndex, []string{asset.ComputeTaskKind, asset.ComputeTaskAction_TASK_ACTION_DOING.String(), key})
	if err != nil {
		return nil, err
	}

	return append(keysTodo, keysDoing...) , nil
}