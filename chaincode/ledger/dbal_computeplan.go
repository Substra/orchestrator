package ledger

import (
	"encoding/json"
	"time"

	"github.com/substra/orchestrator/lib/asset"
	"github.com/substra/orchestrator/lib/common"
	"github.com/substra/orchestrator/lib/errors"
	"google.golang.org/protobuf/encoding/protojson"
)

// storableComputePlan is a custom representation of asset.ComputePlan to enforce storing without its computed properties.
type storableComputePlan struct {
	Key                      string            `json:"key"`
	Owner                    string            `json:"owner"`
	DeleteIntermediaryModels bool              `json:"delete_intermediary_models"`
	CreationDate             *time.Time        `json:"creation_date"`
	Tag                      string            `json:"tag"`
	Name                     string            `json:"name"`
	Metadata                 map[string]string `json:"metadata"`
	CancelationDate          *time.Time        `json:"cancelation_date"`
	FailureDate              *time.Time        `json:"failure_date"`
}

// newStorableComputePlan returns a storableComputePlan
func newStorableComputePlan(plan *asset.ComputePlan) *storableComputePlan {
	storablePlan := &storableComputePlan{
		Key:                      plan.Key,
		Owner:                    plan.Owner,
		DeleteIntermediaryModels: plan.DeleteIntermediaryModels,
		Tag:                      plan.Tag,
		Name:                     plan.Name,
		Metadata:                 plan.Metadata,
	}

	if plan.CreationDate != nil {
		creationDate := plan.CreationDate.AsTime()
		storablePlan.CreationDate = &creationDate
	}
	if plan.CancelationDate != nil {
		cancelationDate := plan.CancelationDate.AsTime()
		storablePlan.CancelationDate = &cancelationDate
	}

	return storablePlan
}

// AddComputePlan stores a new ComputePlan in DB
func (db *DB) AddComputePlan(cp *asset.ComputePlan) error {
	exists, err := db.hasKey(asset.ComputePlanKind, cp.GetKey())
	if err != nil {
		return err
	}
	if exists {
		return errors.NewConflict(asset.ComputePlanKind, cp.Key)
	}
	storablePlan := newStorableComputePlan(cp)

	bytes, err := json.Marshal(storablePlan)
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

// GetComputePlan returns a compute plan
func (db *DB) GetComputePlan(key string) (*asset.ComputePlan, error) {
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

		plans = append(plans, plan)
	}

	return plans, bookmark.Bookmark, nil
}

func (db *DB) updateComputePlan(storablePlan *storableComputePlan) error {
	bytes, err := json.Marshal(storablePlan)
	if err != nil {
		return err
	}

	return db.putState(asset.ComputePlanKind, storablePlan.Key, bytes)
}

func (db *DB) SetComputePlanName(plan *asset.ComputePlan, name string) error {
	storablePlan := newStorableComputePlan(plan)
	storablePlan.Name = name
	return db.updateComputePlan(storablePlan)
}

func (db *DB) CancelComputePlan(plan *asset.ComputePlan, cancelationDate time.Time) error {
	storablePlan := newStorableComputePlan(plan)
	storablePlan.CancelationDate = &cancelationDate
	return db.updateComputePlan(storablePlan)
}

func (db *DB) FailComputePlan(plan *asset.ComputePlan, failureDate time.Time) error {
	storablePlan := newStorableComputePlan(plan)
	storablePlan.FailureDate = &failureDate
	return db.updateComputePlan(storablePlan)
}

func (db *DB) IsComputePlanRunning(key string) (bool, error) {
	iterator, err := db.ccStub.GetStateByPartialCompositeKey(computePlanTaskStatusIndex, []string{asset.ComputePlanKind, key})
	if err != nil {
		return false, err
	}
	defer iterator.Close()

	for iterator.HasNext() {
		compositeKey, err := iterator.Next()
		if err != nil {
			return false, err
		}

		_, keyParts, err := db.ccStub.SplitCompositeKey(compositeKey.Key)
		if err != nil {
			return false, err
		}

		taskStatus := keyParts[2]
		if taskStatus == asset.ComputeTaskStatus_STATUS_WAITING.String() ||
			taskStatus == asset.ComputeTaskStatus_STATUS_TODO.String() ||
			taskStatus == asset.ComputeTaskStatus_STATUS_DOING.String() {
			return true, nil
		}
	}

	return false, nil
}
