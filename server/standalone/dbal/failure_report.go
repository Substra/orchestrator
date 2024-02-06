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
	AssetKey     string
	AssetType    asset.FailedAssetKind
	ErrorType    asset.ErrorType
	CreationDate time.Time
	Owner        string
	LogsChecksum pgtype.Text
	LogsAddress  pgtype.Text
}

func (r sqlFailureReport) toFailureReport() *asset.FailureReport {
	failureReport := &asset.FailureReport{
		AssetKey:     r.AssetKey,
		AssetType:    r.AssetType,
		ErrorType:    r.ErrorType,
		CreationDate: timestamppb.New(r.CreationDate),
		Owner:        r.Owner,
	}

	if r.LogsAddress.Status == pgtype.Present {
		failureReport.LogsAddress = &asset.Addressable{
			StorageAddress: r.LogsAddress.String,
			Checksum:       r.LogsChecksum.String,
		}
	}

	return failureReport
}

func (d *DBAL) GetFailureReport(assetKey string) (*asset.FailureReport, error) {
	stmt := getStatementBuilder().
		Select("asset_key", "asset_type", "error_type", "creation_date", "owner", "logs_address", "logs_checksum").
		From("expanded_failure_reports").
		Where(sq.Eq{"channel": d.channel, "asset_key": assetKey})

	row, err := d.queryRow(stmt)
	if err != nil {
		return nil, err
	}

	r := new(sqlFailureReport)
	err = row.Scan(&r.AssetKey, &r.AssetType, &r.ErrorType, &r.CreationDate, &r.Owner, &r.LogsAddress, &r.LogsChecksum)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, orcerrors.NewNotFound("failure report", assetKey)
		}
		return nil, err
	}

	return r.toFailureReport(), nil
}

func (d *DBAL) AddFailureReport(failureReport *asset.FailureReport) error {
	var logsAddress pgtype.Text
	if failureReport.LogsAddress != nil {
		err := d.addAddressable(failureReport.LogsAddress, false)
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
		Columns("asset_key", "asset_type", "channel", "error_type", "creation_date", "owner", "logs_address").
		Values(failureReport.AssetKey, failureReport.AssetType.String(), d.channel, failureReport.ErrorType, failureReport.CreationDate.AsTime(), failureReport.Owner, logsAddress)

	return d.exec(stmt)
}
