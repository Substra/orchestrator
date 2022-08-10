package ledger

import (
	"github.com/substra/orchestrator/lib/asset"
	"github.com/substra/orchestrator/lib/errors"
	"google.golang.org/protobuf/encoding/protojson"
)

func (db *DB) GetFailureReport(computeTaskKey string) (*asset.FailureReport, error) {
	failureReport := new(asset.FailureReport)

	b, err := db.getState(asset.FailureReportKind, computeTaskKey)
	if err != nil {
		return nil, err
	}

	err = protojson.Unmarshal(b, failureReport)
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
	bytes, err := marshaller.Marshal(failureReport)
	if err != nil {
		return err
	}

	return db.putState(asset.FailureReportKind, failureReport.GetComputeTaskKey(), bytes)
}
