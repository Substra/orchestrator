package dbal

import (
	"strconv"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type sqlPerformance struct {
	ComputeTaskKey   string
	AlgoKey          string
	PerformanceValue float32
	CreationDate     time.Time
}

func (p *sqlPerformance) toPerformance() *asset.Performance {
	return &asset.Performance{
		ComputeTaskKey:   p.ComputeTaskKey,
		MetricKey:        p.AlgoKey,
		PerformanceValue: p.PerformanceValue,
		CreationDate:     timestamppb.New(p.CreationDate),
	}
}

func (d *DBAL) AddPerformance(perf *asset.Performance) error {
	stmt := getStatementBuilder().
		Insert("performances").
		Columns("channel", "compute_task_key", "algo_key", "performance_value", "creation_date").
		Values(d.channel, perf.ComputeTaskKey, perf.MetricKey, perf.PerformanceValue, perf.CreationDate.AsTime())

	return d.exec(stmt)
}

// PerformanceExists implements persistence.PerformanceDBAL
func (d *DBAL) PerformanceExists(perf *asset.Performance) (bool, error) {
	stmt := getStatementBuilder().
		Select("COUNT(*)").
		From("performances").
		Where(sq.Eq{"channel": d.channel, "compute_task_key": perf.ComputeTaskKey, "algo_key": perf.MetricKey})

	row, err := d.queryRow(stmt)
	if err != nil {
		return false, err
	}

	var count int
	err = row.Scan(&count)

	return count >= 1, err
}

func (d *DBAL) QueryPerformances(p *common.Pagination, filter *asset.PerformanceQueryFilter) ([]*asset.Performance, common.PaginationToken, error) {
	offset, err := getOffset(p.Token)
	if err != nil {
		return nil, "", err
	}

	stmt := getStatementBuilder().
		Select("compute_task_key", "algo_key", "performance_value", "creation_date").
		From("performances").
		Where(sq.Eq{"channel": d.channel}).
		OrderByClause("creation_date ASC, algo_key DESC, compute_task_key DESC").
		Offset(uint64(offset)).
		// Fetch page size + 1 elements to determine whether there is a next page
		Limit(uint64(p.Size + 1))

	if filter != nil {
		if filter.ComputeTaskKey != "" {
			stmt = stmt.Where(sq.Eq{"compute_task_key": filter.ComputeTaskKey})
		}

		if filter.MetricKey != "" {
			stmt = stmt.Where(sq.Eq{"algo_key": filter.MetricKey})
		}
	}

	rows, err := d.query(stmt)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var performances []*asset.Performance
	var count int

	for rows.Next() {
		perf := new(sqlPerformance)

		err = rows.Scan(&perf.ComputeTaskKey, &perf.AlgoKey, &perf.PerformanceValue, &perf.CreationDate)
		if err != nil {
			return nil, "", err
		}

		performances = append(performances, perf.toPerformance())
		count++

		if count == int(p.Size) {
			break
		}
	}
	if err = rows.Err(); err != nil {
		return nil, "", err
	}

	bookmark := ""
	if count == int(p.Size) && rows.Next() {
		// there is more to fetch
		bookmark = strconv.Itoa(offset + count)
	}

	return performances, bookmark, nil
}
