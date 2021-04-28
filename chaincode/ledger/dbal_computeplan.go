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

	err = db.putState(asset.ComputePlanKind, cp.GetKey(), bytes)
	if err != nil {
		return err
	}

	if err = db.createIndex("computePlan~owner~key", []string{asset.ComputePlanKind, cp.Owner, cp.Key}); err != nil {
		return err
	}

	return nil
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

	allTasks, err := db.getIndexKeys(IndexPlanTaskStatus, []string{asset.ComputePlanKind, key})
	if err != nil {
		return nil, err
	}
	plan.TaskCount = uint32(len(allTasks))

	doneTasks, err := db.getIndexKeys(IndexPlanTaskStatus, []string{asset.ComputePlanKind, key, asset.ComputeTaskStatus_STATUS_DONE.String()})
	if err != nil {
		return nil, err
	}
	plan.DoneCount = uint32(len(doneTasks))

	return plan, nil
}

// QueryComputePlans retrieves all ComputePlans
func (db *DB) QueryComputePlans(p *common.Pagination) ([]*asset.ComputePlan, common.PaginationToken, error) {
	elementsKeys, bookmark, err := db.getIndexKeysWithPagination("computePlan~owner~key", []string{asset.ComputePlanKind}, p.Size, p.Token)
	if err != nil {
		return nil, "", err
	}

	db.logger.WithField("numItems", len(elementsKeys)).Debug("QueryComputePlans")

	var plans []*asset.ComputePlan
	for _, key := range elementsKeys {
		plan, err := db.GetComputePlan(key)
		if err != nil {
			return plans, bookmark, err
		}
		plans = append(plans, plan)
	}

	return plans, bookmark, nil
}
