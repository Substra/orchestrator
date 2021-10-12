package dbal

import (
	"strconv"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
)

func (d *DBAL) AddPerformance(perf *asset.Performance) error {
	stmt := `insert into "performances" ("compute_task_id", "metric_id", "asset", "channel") values ($1, $2, $3, $4)`
	_, err := d.tx.Exec(d.ctx, stmt, perf.GetComputeTaskKey(), perf.GetMetricKey(), perf, d.channel)
	return err
}

func (d *DBAL) CountComputeTaskPerformances(computeTaskKey string) (int, error) {
	row := d.tx.QueryRow(d.ctx, `select count(*) from "performances" where compute_task_id=$1 and channel=$2`, computeTaskKey, d.channel)

	var count int
	err := row.Scan(&count)

	return count, err
}

// PerformanceExists implements persistence.PerformanceDBAL
func (d *DBAL) PerformanceExists(perf *asset.Performance) (bool, error) {
	row := d.tx.QueryRow(d.ctx, `select count(*) from "performances" where compute_task_id=$1 and metric_id=$2 and channel=$3`, perf.GetComputeTaskKey(), perf.GetMetricKey(), d.channel)

	var count int
	err := row.Scan(&count)

	return count >= 1, err
}

func (d *DBAL) QueryPerformances(p *common.Pagination, filter *asset.PerformanceQueryFilter) ([]*asset.Performance, common.PaginationToken, error) {
	var rows pgx.Rows
	var err error

	offset, err := getOffset(p.Token)
	if err != nil {
		return nil, "", err
	}

	pgDialect := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	builder := pgDialect.Select("asset").
		From("performances").
		Where(squirrel.Eq{"channel": d.channel}).
		OrderByClause("asset->>'creationDate' ASC, metric_id DESC, compute_task_id DESC").
		Offset(uint64(offset)).
		Limit(uint64(p.Size + 1))

	if filter.ComputeTaskKey != "" {
		builder = builder.Where(squirrel.Eq{"compute_task_id": filter.ComputeTaskKey})
	}

	if filter.MetricKey != "" {
		builder = builder.Where(squirrel.Eq{"metric_id": filter.MetricKey})
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

	var performances []*asset.Performance
	var count int

	for rows.Next() {
		performance := new(asset.Performance)

		err = rows.Scan(&performance)
		if err != nil {
			return nil, "", err
		}

		performances = append(performances, performance)
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

	return performances, bookmark, nil
}
