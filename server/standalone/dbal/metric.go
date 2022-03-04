package dbal

import (
	"errors"
	"strconv"

	"github.com/jackc/pgx/v4"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	orcerrors "github.com/owkin/orchestrator/lib/errors"
)

// AddMetric implements persistence.MetricDBAL
func (d *DBAL) AddMetric(obj *asset.Metric) error {
	stmt := `insert into "metrics" ("id", "asset", "channel") values ($1, $2, $3)`
	_, err := d.tx.Exec(d.ctx, stmt, obj.GetKey(), obj, d.channel)

	return err
}

// GetMetric implements persistence.MetricDBAL
func (d *DBAL) GetMetric(key string) (*asset.Metric, error) {
	row := d.tx.QueryRow(d.ctx, `select "asset" from "metrics" where id=$1 and channel=$2`, key, d.channel)

	metric := new(asset.Metric)
	err := row.Scan(metric)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, orcerrors.NewNotFound("metric", key)
		}
		return nil, err
	}

	return metric, nil
}

// QueryMetrics implements persistence.MetricDBAL
func (d *DBAL) QueryMetrics(p *common.Pagination) ([]*asset.Metric, common.PaginationToken, error) {
	offset, err := getOffset(p.Token)
	if err != nil {
		return nil, "", err
	}

	query := `select "asset" from "metrics" where channel=$3 order by asset->>'creationDate' asc, id limit $1 offset $2`
	rows, err := d.tx.Query(d.ctx, query, p.Size+1, offset, d.channel)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var metrics []*asset.Metric
	var count int

	for rows.Next() {
		metric := new(asset.Metric)

		err = rows.Scan(metric)
		if err != nil {
			return nil, "", err
		}

		metrics = append(metrics, metric)
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

	return metrics, bookmark, nil
}

// MetricExists implements persistence.MetricDBAL
func (d *DBAL) MetricExists(key string) (bool, error) {
	row := d.tx.QueryRow(d.ctx, `select count(id) from "metrics" where id=$1 and channel=$2`, key, d.channel)

	var count int
	err := row.Scan(&count)

	return count == 1, err
}
