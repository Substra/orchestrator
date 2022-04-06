package dbal

import (
	"errors"
	"strconv"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	orcerrors "github.com/owkin/orchestrator/lib/errors"
	"github.com/owkin/orchestrator/lib/persistence"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type sqlComputePlan struct {
	Key                      string
	Owner                    string
	DeleteIntermediaryModels bool
	CreationDate             time.Time
	Tag                      string
	Metadata                 map[string]string
	TaskCounts               persistence.ComputePlanTaskCount
}

// toRawComputePlan returns a compute plan without its computed properties.
func (cp *sqlComputePlan) toRawComputePlan() *asset.ComputePlan {
	return &asset.ComputePlan{
		Key:                      cp.Key,
		Owner:                    cp.Owner,
		DeleteIntermediaryModels: cp.DeleteIntermediaryModels,
		CreationDate:             timestamppb.New(cp.CreationDate),
		Tag:                      cp.Tag,
		Metadata:                 cp.Metadata,
	}
}

// toComputePlan returns a compute plan with its computed properties.
func (cp *sqlComputePlan) toComputePlan() *asset.ComputePlan {
	plan := cp.toRawComputePlan()
	plan.TaskCount = cp.TaskCounts.Total
	plan.WaitingCount = cp.TaskCounts.Waiting
	plan.TodoCount = cp.TaskCounts.Todo
	plan.DoingCount = cp.TaskCounts.Doing
	plan.CanceledCount = cp.TaskCounts.Canceled
	plan.FailedCount = cp.TaskCounts.Failed
	plan.DoneCount = cp.TaskCounts.Done
	plan.Status = cp.TaskCounts.GetPlanStatus()
	return plan
}

// ComputePlanExists returns true if a ComputePlan with the given key already exists
func (d *DBAL) ComputePlanExists(key string) (bool, error) {
	stmt := getStatementBuilder().
		Select("COUNT(key)").
		From("compute_plans").
		Where(sq.Eq{"key": key, "channel": d.channel})

	row, err := d.queryRow(stmt)
	if err != nil {
		return false, err
	}

	var count int
	err = row.Scan(&count)

	return count == 1, err
}

// AddComputePlan stores a new ComputePlan in DB
func (d *DBAL) AddComputePlan(plan *asset.ComputePlan) error {
	stmt := getStatementBuilder().
		Insert("compute_plans").
		Columns("key", "channel", "owner", "delete_intermediary_models", "creation_date", "tag", "metadata").
		Values(plan.Key, d.channel, plan.Owner, plan.DeleteIntermediaryModels, plan.CreationDate.AsTime(), plan.Tag, plan.Metadata)

	return d.exec(stmt)
}

// GetComputePlan returns a ComputePlan by its key
func (d *DBAL) GetComputePlan(key string) (*asset.ComputePlan, error) {
	stmt := getStatementBuilder().
		Select("key", "owner", "delete_intermediary_models", "creation_date", "tag", "metadata", "task_count", "waiting_count", "todo_count", "doing_count", "canceled_count", "failed_count", "done_count").
		From("expanded_compute_plans").
		Where(sq.Eq{"key": key, "channel": d.channel})

	row, err := d.queryRow(stmt)
	if err != nil {
		return nil, err
	}

	pl := new(sqlComputePlan)

	err = row.Scan(&pl.Key, &pl.Owner, &pl.DeleteIntermediaryModels, &pl.CreationDate, &pl.Tag, &pl.Metadata, &pl.TaskCounts.Total, &pl.TaskCounts.Waiting, &pl.TaskCounts.Todo, &pl.TaskCounts.Doing, &pl.TaskCounts.Canceled, &pl.TaskCounts.Failed, &pl.TaskCounts.Done)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, orcerrors.NewNotFound("computeplan", key)
		}
		return nil, err
	}

	return pl.toComputePlan(), nil
}

// GetRawComputePlan returns a compute plan without its computed properties.
func (d *DBAL) GetRawComputePlan(key string) (*asset.ComputePlan, error) {
	stmt := getStatementBuilder().
		Select("key", "owner", "delete_intermediary_models", "creation_date", "tag", "metadata").
		From("compute_plans").
		Where(sq.Eq{"key": key, "channel": d.channel})

	row, err := d.queryRow(stmt)
	if err != nil {
		return nil, err
	}

	pl := new(sqlComputePlan)
	err = row.Scan(&pl.Key, &pl.Owner, &pl.DeleteIntermediaryModels, &pl.CreationDate, &pl.Tag, &pl.Metadata)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, orcerrors.NewNotFound("computeplan", key)
		}
		return nil, err
	}

	return pl.toRawComputePlan(), nil
}

func (d *DBAL) QueryComputePlans(p *common.Pagination, filter *asset.PlanQueryFilter) ([]*asset.ComputePlan, common.PaginationToken, error) {
	offset, err := getOffset(p.Token)
	if err != nil {
		return nil, "", err
	}

	stmt := getStatementBuilder().
		Select("key", "owner", "delete_intermediary_models", "creation_date", "tag", "metadata", "task_count", "waiting_count", "todo_count", "doing_count", "canceled_count", "failed_count", "done_count").
		From("expanded_compute_plans").
		Where(sq.Eq{"channel": d.channel}).
		OrderBy("creation_date ASC, key ASC").
		Offset(uint64(offset)).
		// Fetch page size + 1 elements to determine whether there is a next page
		Limit(uint64(p.Size + 1))

	if filter.Owner != "" {
		stmt = stmt.Where(sq.Eq{"owner": filter.Owner})
	}

	rows, err := d.query(stmt)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var plans []*asset.ComputePlan
	var count int

	for rows.Next() {
		pl := new(sqlComputePlan)

		err = rows.Scan(&pl.Key, &pl.Owner, &pl.DeleteIntermediaryModels, &pl.CreationDate, &pl.Tag, &pl.Metadata, &pl.TaskCounts.Total, &pl.TaskCounts.Waiting, &pl.TaskCounts.Todo, &pl.TaskCounts.Doing, &pl.TaskCounts.Canceled, &pl.TaskCounts.Failed, &pl.TaskCounts.Done)
		if err != nil {
			return nil, "", err
		}

		plans = append(plans, pl.toComputePlan())
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
