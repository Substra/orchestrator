package ledger

import (
	"encoding/json"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	"github.com/owkin/orchestrator/lib/errors"
	"github.com/owkin/orchestrator/lib/persistence"
	"google.golang.org/protobuf/encoding/protojson"
)

// AddComputePlan stores a new ComputePlan in DB
func (db *DB) AddComputePlan(cp *asset.ComputePlan) error {
	exists, err := db.hasKey(asset.ComputePlanKind, cp.GetKey())
	if err != nil {
		return err
	}
	if exists {
		return errors.NewConflict(asset.ComputePlanKind, cp.Key)
	}

	bytes, err := marshaller.Marshal(cp)
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
	plan, err := db.GetRawComputePlan(key)
	if err != nil {
		return nil, err
	}

	err = db.computePlanProperties(plan)
	if err != nil {
		return nil, err
	}

	return plan, nil
}

func (db *DB) computePlanProperties(plan *asset.ComputePlan) error {
	count, err := db.getTaskCounts(plan.Key)
	if err != nil {
		return err
	}

	plan.TaskCount = count.Total
	plan.WaitingCount = count.Waiting
	plan.TodoCount = count.Todo
	plan.DoingCount = count.Doing
	plan.CanceledCount = count.Canceled
	plan.FailedCount = count.Failed
	plan.DoneCount = count.Done
	plan.Status = count.GetPlanStatus()

	return nil
}

// GetRawComputePlan returns a compute plan without its computed properties
func (db *DB) GetRawComputePlan(key string) (*asset.ComputePlan, error) {
	plan := new(asset.ComputePlan)

	b, err := db.getState(asset.ComputePlanKind, key)
	if err != nil {
		return nil, err
	}

	err = protojson.Unmarshal(b, plan)
	if err != nil {
		return nil, err
	}

	return plan, nil
}

// QueryComputePlans retrieves all ComputePlans
func (db *DB) QueryComputePlans(p *common.Pagination, filter *asset.PlanQueryFilter) ([]*asset.ComputePlan, common.PaginationToken, error) {
	query := richQuerySelector{
		Selector: couchAssetQuery{
			DocType: asset.ComputePlanKind,
		},
	}

	assetFilter := map[string]interface{}{}

	if filter != nil && filter.Owner != "" {
		assetFilter["owner"] = filter.Owner
	}

	if len(assetFilter) > 0 {
		query.Selector.Asset = assetFilter
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
		plan := &asset.ComputePlan{}
		err = protojson.Unmarshal(storedAsset.Asset, plan)
		if err != nil {
			return nil, "", err
		}

		err = db.computePlanProperties(plan)
		if err != nil {
			return nil, "", err
		}

		plans = append(plans, plan)
	}

	return plans, bookmark.Bookmark, nil
}

// getTaskCounts returns the count of plan's tasks by status.
func (db *DB) getTaskCounts(key string) (persistence.ComputePlanTaskCount, error) {
	count := persistence.ComputePlanTaskCount{}

	iterator, err := db.ccStub.GetStateByPartialCompositeKey(computePlanTaskStatusIndex, []string{asset.ComputePlanKind, key})
	if err != nil {
		return count, err
	}
	defer iterator.Close()

	for iterator.HasNext() {
		compositeKey, err := iterator.Next()
		if err != nil {
			return count, err
		}
		_, keyParts, err := db.ccStub.SplitCompositeKey(compositeKey.Key)
		if err != nil {
			return count, err
		}
		switch keyParts[2] {
		case asset.ComputeTaskStatus_STATUS_CANCELED.String():
			count.Canceled++
		case asset.ComputeTaskStatus_STATUS_DONE.String():
			count.Done++
		case asset.ComputeTaskStatus_STATUS_FAILED.String():
			count.Failed++
		case asset.ComputeTaskStatus_STATUS_WAITING.String():
			count.Waiting++
		case asset.ComputeTaskStatus_STATUS_DOING.String():
			count.Doing++
		case asset.ComputeTaskStatus_STATUS_TODO.String():
			count.Todo++
		}

		count.Total++
	}

	return count, nil
}
