package dbal

import (
	"context"

	"github.com/owkin/orchestrator/lib/asset"
)

func (d *DBAL) AddPerformance(perf *asset.Performance) error {
	stmt := `insert into "performances" ("id", "asset", "channel") values ($1, $2, $3)`
	_, err := d.tx.Exec(context.Background(), stmt, perf.ComputeTaskKey, perf, d.channel)
	return err
}

func (d *DBAL) GetComputeTaskPerformance(key string) (*asset.Performance, error) {
	row := d.tx.QueryRow(context.Background(), `select asset from "performances" where id=$1 and channel=$2`, key, d.channel)

	perf := new(asset.Performance)
	err := row.Scan(perf)
	if err != nil {
		return nil, err
	}

	return perf, nil
}
