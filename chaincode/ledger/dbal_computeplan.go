package ledger

import (
	"encoding/json"
	"fmt"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	"github.com/owkin/orchestrator/lib/errors"
)

// AddComputePlan stores a new ComputePlan in DB
func (db *DB) AddComputePlan(cp *asset.ComputePlan) error {
	exists, err := db.hasKey(asset.ComputePlanKind, cp.GetKey())
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("failed to add compute plan: %w", errors.ErrConflict)
	}

	bytes, err := json.Marshal(cp)
	if err != nil {
		return err
	}

	return db.putState(asset.ComputePlanKind, cp.GetKey(), bytes)
}

// ComputePlanExists returns true if a plan with the given key exists
func (db *DB) ComputePlanExists(key string) (bool, error) {
	exists, err := db.hasKey(asset.ComputePlanKind, key)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// GetComputePlan returns a ComputePlan by its key
func (db *DB) GetComputePlan(key string) (*asset.ComputePlan, error) {
	plan := new(asset.ComputePlan)

	b, err := db.getState(asset.ComputePlanKind, key)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, plan)
	if err != nil {
		return nil, err
	}

	allTasks, err := db.getIndexKeys(computePlanTaskStatusIndex, []string{asset.ComputePlanKind, key})
	if err != nil {
		return nil, err
	}
	plan.TaskCount = uint32(len(allTasks))

	doneTasks, err := db.getIndexKeys(computePlanTaskStatusIndex, []string{asset.ComputePlanKind, key, asset.ComputeTaskStatus_STATUS_DONE.String()})
	if err != nil {
		return nil, err
	}
	plan.DoneCount = uint32(len(doneTasks))

	plan.Status, err = db.getPlanStatus(key, len(allTasks), len(doneTasks))
	if err != nil {
		return nil, err
	}

	return plan, nil
}

// QueryComputePlans retrieves all ComputePlans
func (db *DB) QueryComputePlans(p *common.Pagination) ([]*asset.ComputePlan, common.PaginationToken, error) {
	query := richQuerySelector{
		Selector: couchAssetQuery{
			DocType: asset.ComputePlanKind,
		},
		Fields: []string{"asset.key"},
	}

	b, err := json.Marshal(query)
	if err != nil {
		return nil, "", err
	}

	queryString := string(b)

	resultsIterator, bookmark, err := db.getQueryResultWithPagination(queryString, int32(p.Size), p.Token)
	if err != nil {
		return nil, "", err
	}
	defer resultsIterator.Close()

	plans := make([]*asset.ComputePlan, 0)

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
		identifiedPlan := &asset.ComputePlan{} // This is an empty plan which will only contains its key
		err = json.Unmarshal(storedAsset.Asset, identifiedPlan)
		if err != nil {
			return nil, "", err
		}

		plan, err := db.GetComputePlan(identifiedPlan.Key)
		if err != nil {
			return nil, "", err
		}

		plans = append(plans, plan)
	}

	return plans, bookmark.Bookmark, nil
}

// getPlanStatus returns the plan status from its tasks.
// It attempts to limit the amount of read operation.
func (db *DB) getPlanStatus(key string, total, done int) (asset.ComputePlanStatus, error) {
	if done == total {
		return asset.ComputePlanStatus_PLAN_STATUS_DONE, nil
	}

	failedTasks, err := db.getIndexKeys(computePlanTaskStatusIndex, []string{asset.ComputePlanKind, key, asset.ComputeTaskStatus_STATUS_FAILED.String()})
	if err != nil {
		return asset.ComputePlanStatus_PLAN_STATUS_UNKNOWN, err
	}
	if len(failedTasks) > 0 {
		return asset.ComputePlanStatus_PLAN_STATUS_FAILED, nil
	}

	canceledTasks, err := db.getIndexKeys(computePlanTaskStatusIndex, []string{asset.ComputePlanKind, key, asset.ComputeTaskStatus_STATUS_CANCELED.String()})
	if err != nil {
		return asset.ComputePlanStatus_PLAN_STATUS_UNKNOWN, err
	}
	if len(canceledTasks) > 0 {
		return asset.ComputePlanStatus_PLAN_STATUS_CANCELED, nil
	}

	waitingTasks, err := db.getIndexKeys(computePlanTaskStatusIndex, []string{asset.ComputePlanKind, key, asset.ComputeTaskStatus_STATUS_WAITING.String()})
	if err != nil {
		return asset.ComputePlanStatus_PLAN_STATUS_UNKNOWN, err
	}
	if len(waitingTasks) == total {
		return asset.ComputePlanStatus_PLAN_STATUS_WAITING, nil
	}

	if len(waitingTasks) < total && done == 0 {
		doingTasks, err := db.getIndexKeys(computePlanTaskStatusIndex, []string{asset.ComputePlanKind, key, asset.ComputeTaskStatus_STATUS_DOING.String()})
		if err != nil {
			return asset.ComputePlanStatus_PLAN_STATUS_UNKNOWN, err
		}
		// len(waitingTasks) and done == 0 are redundant with upper condition but make the condition more readable.
		// see asset documentation for inference rules.
		if len(waitingTasks) < total && len(doingTasks) == 0 && done == 0 {
			return asset.ComputePlanStatus_PLAN_STATUS_TODO, nil
		}
	}

	return asset.ComputePlanStatus_PLAN_STATUS_DOING, nil
}
