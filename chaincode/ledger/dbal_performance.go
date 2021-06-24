package ledger

import (
	"encoding/json"
	"fmt"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/errors"
)

func (db *DB) AddPerformance(perf *asset.Performance) error {
	// use task key since a task can have at most 1 performance report
	exists, err := db.hasKey(asset.PerformanceKind, perf.ComputeTaskKey)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("failed to add performance: %w", errors.ErrConflict)
	}
	bytes, err := json.Marshal(perf)
	if err != nil {
		return err
	}

	err = db.putState(asset.PerformanceKind, perf.ComputeTaskKey, bytes)
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) GetComputeTaskPerformance(key string) (*asset.Performance, error) {
	perf := new(asset.Performance)

	b, err := db.getState(asset.PerformanceKind, key)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, perf)
	if err != nil {
		return nil, err
	}
	return perf, nil
}
