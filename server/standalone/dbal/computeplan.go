package dbal

import (
	"errors"
	"strconv"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	orcerrors "github.com/owkin/orchestrator/lib/errors"
	"github.com/owkin/orchestrator/lib/persistence"
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
	pgDialect := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	builder := pgDialect.Select().
		Column("cp.asset").
		Column(squirrel.Expr("COUNT(1)")).
		Column(squirrel.Expr("COUNT(1) FILTER (WHERE t.status = 'STATUS_WAITING')")).
		Column(squirrel.Expr("COUNT(1) FILTER (WHERE t.status = 'STATUS_TODO')")).
		Column(squirrel.Expr("COUNT(1) FILTER (WHERE t.status = 'STATUS_DOING')")).
		Column(squirrel.Expr("COUNT(1) FILTER (WHERE t.status = 'STATUS_CANCELED')")).
		Column(squirrel.Expr("COUNT(1) FILTER (WHERE t.status = 'STATUS_FAILED')")).
		Column(squirrel.Expr("COUNT(1) FILTER (WHERE t.status = 'STATUS_DONE')")).
		From("compute_plans AS cp").
		LeftJoin("compute_tasks AS t ON t.compute_plan_key = cp.id").
		Where(squirrel.Eq{"cp.id": key}).
		Where(squirrel.Eq{"cp.channel": d.channel}).
		GroupBy("cp.id")

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	row := d.tx.QueryRow(d.ctx, query, args...)

	plan := new(asset.ComputePlan)
	var total, waiting, todo, doing, canceled, failed, done uint32
	err = row.Scan(plan, &total, &waiting, &todo, &doing, &canceled, &failed, &done)
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

func (d *DBAL) QueryComputePlans(p *common.Pagination, filter *asset.PlanQueryFilter) ([]*asset.ComputePlan, common.PaginationToken, error) {
	var rows pgx.Rows
	var err error
	offset, err := getOffset(p.Token)
	if err != nil {
		return nil, "", err
	}

	pgDialect := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	builder := pgDialect.Select().
		Column("cp.asset").
		Column(squirrel.Expr("COUNT(1)")).
		Column(squirrel.Expr("COUNT(1) FILTER (WHERE t.status = 'STATUS_WAITING')")).
		Column(squirrel.Expr("COUNT(1) FILTER (WHERE t.status = 'STATUS_TODO')")).
		Column(squirrel.Expr("COUNT(1) FILTER (WHERE t.status = 'STATUS_DOING')")).
		Column(squirrel.Expr("COUNT(1) FILTER (WHERE t.status = 'STATUS_CANCELED')")).
		Column(squirrel.Expr("COUNT(1) FILTER (WHERE t.status = 'STATUS_FAILED')")).
		Column(squirrel.Expr("COUNT(1) FILTER (WHERE t.status = 'STATUS_DONE')")).
		From("compute_plans AS cp").
		LeftJoin("compute_tasks AS t ON t.compute_plan_key = cp.id").
		Where(squirrel.Eq{"cp.channel": d.channel}).
		GroupBy("cp.id").
		OrderBy("cp.asset->>'creationDate' ASC", "cp.id ASC").
		Offset(uint64(offset)).
		// Fetch page size + 1 elements to determine whether there is a next page
		Limit(uint64(p.Size + 1))

	if filter.Owner != "" {
		builder = builder.Where(squirrel.Eq{"asset->>'owner'": filter.Owner})
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, "", err
	}

	rows, err = d.tx.Query(d.ctx, query, args...)
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
	count := persistence.ComputePlanTaskCount{
		Total:    (int)(total),
		Waiting:  (int)(waiting),
		Doing:    (int)(doing),
		Canceled: (int)(canceled),
		Failed:   (int)(failed),
		Done:     (int)(done),
	}

	return count.GetPlanStatus()
}
