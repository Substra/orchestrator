package dbal

import (
	"errors"

	"github.com/jackc/pgx/v4"
	"github.com/owkin/orchestrator/lib/asset"
	orcerrors "github.com/owkin/orchestrator/lib/errors"
)

func (d *DBAL) GetFailureReport(computeTaskKey string) (*asset.FailureReport, error) {
	row := d.tx.QueryRow(d.ctx, `select asset from "failure_reports" where compute_task_id=$1 and channel=$2`, computeTaskKey, d.channel)

	failureReport := new(asset.FailureReport)
	err := row.Scan(failureReport)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, orcerrors.NewNotFound("failure report", computeTaskKey)
		}
		return nil, err
	}

	return failureReport, nil
}

func (d *DBAL) AddFailureReport(failureReport *asset.FailureReport) error {
	stmt := `insert into "failure_reports" ("compute_task_id", "asset", "channel") values ($1, $2, $3)`
	_, err := d.tx.Exec(d.ctx, stmt, failureReport.GetComputeTaskKey(), failureReport, d.channel)
	return err
}
