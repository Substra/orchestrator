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

package dbal

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	orcerrors "github.com/owkin/orchestrator/lib/errors"
)

// ComputePlanExists returns true if a ComputePlan with the given key already exists
func (d *DBAL) ComputePlanExists(key string) (bool, error) {
	row := d.tx.QueryRow(`select count(id) from "compute_plans" where id=$1 and channel=$2`, key, d.channel)

	var count int
	err := row.Scan(&count)

	return count == 1, err
}

// AddComputePlan stores a new ComputePlan in DB
func (d *DBAL) AddComputePlan(plan *asset.ComputePlan) error {
	stmt := `insert into "compute_plans" ("id", "asset", "channel") values ($1, $2, $3)`
	_, err := d.tx.Exec(stmt, plan.GetKey(), plan, d.channel)
	return err
}

// GetComputePlan returns a ComputePlan by its key
func (d *DBAL) GetComputePlan(key string) (*asset.ComputePlan, error) {
	query := `
select cp.asset,
count(t.id),
count(t.id) filter (where t.asset->>'status' = 'STATUS_DONE'),
count(t.id) filter (where t.asset->>'status' = 'STATUS_DOING'),
count(t.id) filter (where t.asset->>'status' = 'STATUS_WAITING'),
count(t.id) filter (where t.asset->>'status' = 'STATUS_FAILED'),
count(t.id) filter (where t.asset->>'status' = 'STATUS_CANCELED')
from "compute_plans" as cp
left join "compute_tasks" as t on (t.asset->>'computePlanKey')::uuid = cp.id and t.channel = cp.channel
where cp.id=$1 and cp.channel=$2
group by cp.asset
`

	row := d.tx.QueryRow(query, key, d.channel)

	plan := new(asset.ComputePlan)
	var total, done, doing, waiting, failed, canceled uint32
	err := row.Scan(plan, &total, &done, &doing, &waiting, &failed, &canceled)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("computeplan not found: %w", orcerrors.ErrNotFound)
		}
		return nil, err
	}

	plan.DoneCount = done
	plan.TaskCount = total
	plan.Status = getPlanStatus(total, done, doing, waiting, failed, canceled)

	return plan, nil
}

func (d *DBAL) QueryComputePlans(p *common.Pagination) ([]*asset.ComputePlan, common.PaginationToken, error) {
	var rows *sql.Rows
	var err error
	offset, err := getOffset(p.Token)
	if err != nil {
		return nil, "", err
	}

	query := `
select cp.asset,
count(t.id),
count(t.id) filter (where t.asset->>'status' = 'STATUS_DONE'),
count(t.id) filter (where t.asset->>'status' = 'STATUS_DOING'),
count(t.id) filter (where t.asset->>'status' = 'STATUS_WAITING'),
count(t.id) filter (where t.asset->>'status' = 'STATUS_FAILED'),
count(t.id) filter (where t.asset->>'status' = 'STATUS_CANCELED')
from "compute_plans" as cp
left join "compute_tasks" as t on (t.asset->>'computePlanKey')::uuid = cp.id and t.channel = cp.channel
where cp.channel=$3
group by cp.asset, cp.created_at
order by cp.created_at asc limit $1 offset $2
`

	rows, err = d.tx.Query(query, p.Size+1, offset, d.channel)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var plans []*asset.ComputePlan
	var count int

	for rows.Next() {
		plan := new(asset.ComputePlan)
		var total, done, doing, waiting, failed, canceled uint32

		err = rows.Scan(plan, &total, &done, &doing, &waiting, &failed, &canceled)
		if err != nil {
			return nil, "", err
		}

		plan.DoneCount = done
		plan.TaskCount = total
		plan.Status = getPlanStatus(total, done, doing, waiting, failed, canceled)

		plans = append(plans, plan)
		count++

		if count == int(p.Size) {
			break
		}
	}
	if err := rows.Err(); err != nil {
		return nil, "", err
	}

	bookmark := ""
	if count == int(p.Size) && rows.Next() {
		// there is more to fetch
		bookmark = strconv.Itoa(offset + count)
	}

	return plans, bookmark, nil
}

// getPlanStatus determines ComputePlan status from its task counts
func getPlanStatus(total, done, doing, waiting, failed, canceled uint32) asset.ComputePlanStatus {
	switch true {
	case failed > 0:
		return asset.ComputePlanStatus_PLAN_STATUS_FAILED
	case canceled > 0:
		return asset.ComputePlanStatus_PLAN_STATUS_CANCELED
	case total == done:
		return asset.ComputePlanStatus_PLAN_STATUS_DONE
	case total == waiting:
		return asset.ComputePlanStatus_PLAN_STATUS_WAITING
	case waiting < total && done == 0 && doing == 0:
		return asset.ComputePlanStatus_PLAN_STATUS_TODO
	default:
		return asset.ComputePlanStatus_PLAN_STATUS_DOING
	}
}
