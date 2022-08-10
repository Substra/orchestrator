package dbal

import (
	"errors"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/substra/orchestrator/lib/asset"
	orcerrors "github.com/substra/orchestrator/lib/errors"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type sqlFailureReport struct {
	ComputeTaskKey string
	ErrorType      asset.ErrorType
	CreationDate   time.Time
	Owner          string
	LogsChecksum   pgtype.Text
	LogsAddress    pgtype.Text
}

func (r sqlFailureReport) toFailureReport() *asset.FailureReport {
	failureReport := &asset.FailureReport{
		ComputeTaskKey: r.ComputeTaskKey,
		ErrorType:      r.ErrorType,
		CreationDate:   timestamppb.New(r.CreationDate),
		Owner:          r.Owner,
	}

	if r.LogsAddress.Status == pgtype.Present {
		failureReport.LogsAddress = &asset.Addressable{
			StorageAddress: r.LogsAddress.String,
			Checksum:       r.LogsChecksum.String,
		}
	}

	return failureReport
}

func (d *DBAL) GetFailureReport(computeTaskKey string) (*asset.FailureReport, error) {
	stmt := getStatementBuilder().
		Select("compute_task_key", "error_type", "creation_date", "owner", "logs_address", "logs_checksum").
		From("expanded_failure_reports").
		Where(sq.Eq{"channel": d.channel, "compute_task_key": computeTaskKey})

	row, err := d.queryRow(stmt)
	if err != nil {
		return nil, err
	}

	r := new(sqlFailureReport)
	err = row.Scan(&r.ComputeTaskKey, &r.ErrorType, &r.CreationDate, &r.Owner, &r.LogsAddress, &r.LogsChecksum)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, orcerrors.NewNotFound("failure report", computeTaskKey)
		}
		return nil, err
	}

	return r.toFailureReport(), nil
}

func (d *DBAL) AddFailureReport(failureReport *asset.FailureReport) error {
	var logsAddress pgtype.Text
	if failureReport.LogsAddress != nil {
		err := d.addAddressable(failureReport.LogsAddress)
		if err != nil {
			return err
		}

		logsAddress = pgtype.Text{
			String: failureReport.LogsAddress.StorageAddress,
			Status: pgtype.Present,
		}
	} else {
		logsAddress = pgtype.Text{Status: pgtype.Null}
	}

	stmt := getStatementBuilder().
		Insert("failure_reports").
		Columns("compute_task_key", "channel", "error_type", "creation_date", "owner", "logs_address").
		Values(failureReport.ComputeTaskKey, d.channel, failureReport.ErrorType, failureReport.CreationDate.AsTime(), failureReport.Owner, logsAddress)

	return d.exec(stmt)
}
