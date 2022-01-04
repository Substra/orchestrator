package ledger

import (
	"encoding/json"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/errors"
)

func (db *DB) GetFailureReport(computeTaskKey string) (*asset.FailureReport, error) {
	failureReport := new(asset.FailureReport)

	b, err := db.getState(asset.FailureReportKind, computeTaskKey)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, failureReport)
	if err != nil {
		return nil, err
	}
	return failureReport, nil
}

func (db *DB) AddFailureReport(failureReport *asset.FailureReport) error {
	exists, err := db.hasKey(asset.FailureReportKind, failureReport.GetComputeTaskKey())
	if err != nil {
		return err
	}
	if exists {
		return errors.NewConflict(asset.FailureReportKind, failureReport.GetComputeTaskKey())
	}
	bytes, err := json.Marshal(failureReport)
	if err != nil {
		return err
	}

	return db.putState(asset.FailureReportKind, failureReport.GetComputeTaskKey(), bytes)
}
