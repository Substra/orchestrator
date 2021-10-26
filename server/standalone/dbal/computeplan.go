package dbal

import (
	"errors"
	"strconv"

	"github.com/jackc/pgx/v4"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	orcerrors "github.com/owkin/orchestrator/lib/errors"
)

// ComputePlanExists returns true if a ComputePlan with the given key already exists
func (d *DBAL) ComputePlanExists(key string) (bool, error) {
	row := d.tx.QueryRow(d.ctx, `select count(id) from "compute_plans" where id=$1 and channel=$2`, key, d.channel)

	var count int
	err := row.Scan(&count)

	return count == 1, err
}

// AddComputePlan stores a new ComputePlan in DB
func (d *DBAL) AddComputePlan(plan *asset.ComputePlan) error {
	stmt := `insert into "compute_plans" ("id", "asset", "channel") values ($1, $2, $3)`
	_, err := d.tx.Exec(d.ctx, stmt, plan.GetKey(), plan, d.channel)
	return err
}

// GetComputePlan returns a ComputePlan by its key
func (d *DBAL) GetComputePlan(key string) (*asset.ComputePlan, error) {
	query := `
select cp.asset,
count(t.id),
count(t.id) filter (where t.asset->>'status' = 'STATUS_WAITING'),
count(t.id) filter (where t.asset->>'status' = 'STATUS_TODO'),
count(t.id) filter (where t.asset->>'status' = 'STATUS_DOING'),
count(t.id) filter (where t.asset->>'status' = 'STATUS_CANCELED'),
count(t.id) filter (where t.asset->>'status' = 'STATUS_FAILED'),
count(t.id) filter (where t.asset->>'status' = 'STATUS_DONE')
from "compute_plans" as cp
left join "compute_tasks" as t on (t.asset->>'computePlanKey')::uuid = cp.id and t.channel = cp.channel
where cp.id=$1 and cp.channel=$2
group by cp.asset
`

	row := d.tx.QueryRow(d.ctx, query, key, d.channel)

	plan := new(asset.ComputePlan)
	var total, waiting, todo, doing, canceled, failed, done uint32
	err := row.Scan(plan, &total, &waiting, &todo, &doing, &canceled, &failed, &done)
	println("extracted data from row:")
	println(total, waiting, todo, doing, canceled, failed, done)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, orcerrors.NewNotFound("computeplan", key)
		}
		return nil, err
	}

	plan.TaskCount = total
	plan.WaitingCount = waiting
	plan.TodoCount = todo
	plan.DoingCount = doing
	plan.CanceledCount = canceled
	plan.FailedCount = failed
	plan.DoneCount = done
	plan.Status = getPlanStatus(total, done, doing, waiting, failed, canceled)

	return plan, nil
}

// GetRawComputePlan returns a compute plan without its computed properties.
func (d *DBAL) GetRawComputePlan(key string) (*asset.ComputePlan, error) {
	query := `select asset from compute_plans where id=$1 and channel=$2`

	row := d.tx.QueryRow(d.ctx, query, key, d.channel)

	plan := new(asset.ComputePlan)
	err := row.Scan(plan)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, orcerrors.NewNotFound("computeplan", key)
		}
		return nil, err
	}

	return plan, nil
}

func (d *DBAL) QueryComputePlans(p *common.Pagination) ([]*asset.ComputePlan, common.PaginationToken, error) {
	var rows pgx.Rows
	var err error
	offset, err := getOffset(p.Token)
	if err != nil {
		return nil, "", err
	}

	query := `
select cp.asset,
count(t.id),
count(t.id) filter (where t.asset->>'status' = 'STATUS_WAITING'),
count(t.id) filter (where t.asset->>'status' = 'STATUS_TODO'),
count(t.id) filter (where t.asset->>'status' = 'STATUS_DOING'),
count(t.id) filter (where t.asset->>'status' = 'STATUS_CANCELED'),
count(t.id) filter (where t.asset->>'status' = 'STATUS_FAILED'),
count(t.id) filter (where t.asset->>'status' = 'STATUS_DONE')
from "compute_plans" as cp
left join "compute_tasks" as t on (t.asset->>'computePlanKey')::uuid = cp.id and t.channel = cp.channel
where cp.channel=$3
group by cp.id
order by cp.asset->>'creationDate' asc, cp.id limit $1 offset $2
`

	rows, err = d.tx.Query(d.ctx, query, p.Size+1, offset, d.channel)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var plans []*asset.ComputePlan
	var count int

	for rows.Next() {
		plan := new(asset.ComputePlan)

		var total, waiting, todo, doing, canceled, failed, done uint32

		err = rows.Scan(plan, &total, &waiting, &todo, &doing, &canceled, &failed, &done)
		if err != nil {
			return nil, "", err
		}

		plan.TaskCount = total
		plan.WaitingCount = waiting
		plan.TodoCount = todo
		plan.DoingCount = doing
		plan.CanceledCount = canceled
		plan.FailedCount = failed
		plan.DoneCount = done
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
