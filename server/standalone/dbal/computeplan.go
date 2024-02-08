package dbal

import (
	"database/sql"
	"errors"
	"strconv"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4"
	"github.com/substra/orchestrator/lib/asset"
	"github.com/substra/orchestrator/lib/common"
	orcerrors "github.com/substra/orchestrator/lib/errors"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type sqlComputePlan struct {
	Key             string
	Owner           string
	CreationDate    time.Time
	CancelationDate sql.NullTime
	FailureDate     sql.NullTime
	Tag             string
	Name            string
	Metadata        map[string]string
}

// toComputePlan returns a compute plan.
func (cp *sqlComputePlan) toComputePlan() *asset.ComputePlan {
	res := &asset.ComputePlan{
		Key:          cp.Key,
		Owner:        cp.Owner,
		CreationDate: timestamppb.New(cp.CreationDate),
		Tag:          cp.Tag,
		Name:         cp.Name,
		Metadata:     cp.Metadata,
	}

	if cp.CancelationDate.Valid {
		res.CancelationDate = timestamppb.New(cp.CancelationDate.Time)
	} else if cp.FailureDate.Valid {
		res.FailureDate = timestamppb.New(cp.FailureDate.Time)
	}
	return res
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
		Columns("key", "channel", "owner", "creation_date", "tag", "name", "metadata").
		Values(plan.Key, d.channel, plan.Owner, plan.CreationDate.AsTime(), plan.Tag, plan.Name, plan.Metadata)

	return d.exec(stmt)
}

// GetComputePlan fetches a given compute plan
func (d *DBAL) GetComputePlan(key string) (*asset.ComputePlan, error) {
	stmt := getStatementBuilder().
		Select("key", "owner", "creation_date", "cancelation_date", "failure_date", "tag", "name", "metadata").
		From("compute_plans").
		Where(sq.Eq{"key": key, "channel": d.channel})

	row, err := d.queryRow(stmt)
	if err != nil {
		return nil, err
	}

	pl := new(sqlComputePlan)
	err = row.Scan(&pl.Key, &pl.Owner, &pl.CreationDate, &pl.CancelationDate, &pl.FailureDate, &pl.Tag, &pl.Name, &pl.Metadata)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, orcerrors.NewNotFound("computeplan", key)
		}
		return nil, err
	}

	return pl.toComputePlan(), nil
}

func (d *DBAL) QueryComputePlans(p *common.Pagination, filter *asset.PlanQueryFilter) ([]*asset.ComputePlan, common.PaginationToken, error) {
	offset, err := getOffset(p.Token)
	if err != nil {
		return nil, "", err
	}

	stmt := getStatementBuilder().
		Select("key", "owner", "creation_date", "cancelation_date", "failure_date", "tag", "name", "metadata").
		From("compute_plans").
		Where(sq.Eq{"channel": d.channel}).
		OrderBy("creation_date ASC, key ASC").
		Offset(uint64(offset)).
		// Fetch page size + 1 elements to determine whether there is a next page
		Limit(uint64(p.Size + 1))

	if filter != nil && filter.Owner != "" {
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

		err = rows.Scan(&pl.Key, &pl.Owner, &pl.CreationDate, &pl.CancelationDate, &pl.FailureDate, &pl.Tag, &pl.Name, &pl.Metadata)
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

func (d *DBAL) updateComputePlan(key string, column string, value interface{}) error {
	stmt := getStatementBuilder().
		Update("compute_plans").
		Set(column, value).
		Where(sq.Eq{"channel": d.channel, "key": key})

	return d.exec(stmt)
}

func (d *DBAL) SetComputePlanName(plan *asset.ComputePlan, name string) error {
	return d.updateComputePlan(plan.Key, "name", name)
}

func (d *DBAL) CancelComputePlan(plan *asset.ComputePlan, cancelationDate time.Time) error {
	return d.updateComputePlan(plan.Key, "cancelation_date", cancelationDate)
}

func (d *DBAL) FailComputePlan(plan *asset.ComputePlan, failureDate time.Time) error {
	return d.updateComputePlan(plan.Key, "failure_date", failureDate)
}

func (d *DBAL) IsPlanRunning(key string) (bool, error) {
	stmt := getStatementBuilder().
		Select("COUNT(*)").
		From("compute_tasks").
		Where(sq.Eq{
			"status":           []string{"STATUS_WAITING_FOR_PARENT_TASKS", "STATUS_WAITING_FOR_EXECUTOR_SLOT", "STATUS_DOING"},
			"compute_plan_key": key,
			"channel":          d.channel,
		})

	row, err := d.queryRow(stmt)
	if err != nil {
		return false, err
	}

	var count int
	err = row.Scan(&count)

	return count > 0, err
}
