package ledger

import (
	"encoding/json"
	"fmt"

	"github.com/go-playground/log/v7"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	"github.com/owkin/orchestrator/lib/errors"
)

// AddObjective stores a new objective
func (db *DB) AddObjective(obj *asset.Objective) error {
	exists, err := db.hasKey(asset.ObjectiveKind, obj.GetKey())
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("objective already exists: %w", errors.ErrConflict)
	}

	objBytes, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	err = db.putState(asset.ObjectiveKind, obj.GetKey(), objBytes)
	if err != nil {
		return err
	}

	return db.createIndex("objective~owner~key", []string{asset.ObjectiveKind, obj.Owner, obj.Key})
}

// ObjectiveExists implements persistence.ObjectiveDBAL
func (db *DB) ObjectiveExists(key string) (bool, error) {
	return db.hasKey(asset.ObjectiveKind, key)
}

// GetObjective retrieves an objective by its key
func (db *DB) GetObjective(key string) (*asset.Objective, error) {
	o := asset.Objective{}

	b, err := db.getState(asset.ObjectiveKind, key)
	if err != nil {
		return &o, err
	}

	err = json.Unmarshal(b, &o)
	return &o, err
}

// QueryObjectives retrieves all objectives
func (db *DB) QueryObjectives(p *common.Pagination) ([]*asset.Objective, common.PaginationToken, error) {
	elementsKeys, bookmark, err := db.getIndexKeysWithPagination("objective~owner~key", []string{asset.ObjectiveKind}, p.Size, p.Token)
	if err != nil {
		return nil, "", err
	}

	db.logger.WithField("keys", elementsKeys).Debug("QueryObjectives")

	var objectives []*asset.Objective
	for _, key := range elementsKeys {
		objective, err := db.GetObjective(key)
		if err != nil {
			return objectives, bookmark, err
		}
		objectives = append(objectives, objective)
	}

	return objectives, bookmark, nil
}

// GetLeaderboard returns for an objective all its certified ComputeTask with ComputeTaskCategory: TEST_TASK with a done status
func (db *DB) GetLeaderboard(key string) (*asset.Leaderboard, error) {
	o, err := db.GetObjective(key)
	if err != nil {
		return nil, err
	}

	lb := asset.Leaderboard{}
	lb.Objective = o

	selector := couchAssetQuery{
		DocType: asset.ComputeTaskKind,
		Asset: map[string]interface{}{
			"status":   "STATUS_DONE",
			"category": "TASK_TEST",
			"test": map[string]interface{}{
				"certified":     true,
				"objective_key": key,
			},
		},
	}

	b, err := json.Marshal(selector)
	if err != nil {
		return nil, err
	}

	queryString := fmt.Sprintf(`{"selector":%s}}`, string(b))
	log.WithField("couchQuery", queryString).Debug("mango query board items")

	if err != nil {
		return nil, err
	}

	resultsIterator, err := db.getQueryResult(queryString)

	if err != nil {
		return nil, err
	}

	defer resultsIterator.Close()

	var boardItems []*asset.BoardItem

	//build BoardItem
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var storedAsset storedAsset
		err = json.Unmarshal(queryResult.Value, &storedAsset)
		if err != nil {
			return nil, err
		}
		task := &asset.ComputeTask{}
		err = json.Unmarshal(storedAsset.Asset, task)
		if err != nil {
			return nil, err
		}

		perf, err := db.GetComputeTaskPerformance(task.Key)

		if err != nil {
			return nil, err
		}

		boardItem := asset.BoardItem{
			Algo:           task.Algo,
			ObjectiveKey:   key,
			ComputeTaskKey: task.Key,
			Perf:           perf.PerformanceValue,
		}

		boardItems = append(boardItems, &boardItem)
	}

	lb.BoardItems = boardItems
	return &lb, nil
}
