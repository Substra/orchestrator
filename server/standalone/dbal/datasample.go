package dbal

import (
	"errors"
	"strconv"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/rs/zerolog/log"
	"github.com/substra/orchestrator/lib/asset"
	"github.com/substra/orchestrator/lib/common"
	orcerrors "github.com/substra/orchestrator/lib/errors"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type sqlDataSample struct {
	Key             string
	Owner           string
	Checksum        string
	CreationDate    time.Time
	DataManagerKeys []string
}

func (ds *sqlDataSample) toDataSample() *asset.DataSample {
	return &asset.DataSample{
		Key:             ds.Key,
		DataManagerKeys: ds.DataManagerKeys,
		Owner:           ds.Owner,
		Checksum:        ds.Checksum,
		CreationDate:    timestamppb.New(ds.CreationDate),
	}
}

// DataSampleExists implements persistence.DataSampleDBAL
func (d *DBAL) DataSampleExists(key string) (bool, error) {
	stmt := getStatementBuilder().
		Select("COUNT(key)").
		From("datasamples").
		Where(sq.Eq{"channel": d.channel, "key": key})

	row, err := d.queryRow(stmt)
	if err != nil {
		return false, err
	}

	var count int
	err = row.Scan(&count)

	return count == 1, err
}

// AddDataSamples add one or multiple data samples to storage.
func (d *DBAL) AddDataSamples(datasamples ...*asset.DataSample) error {
	log.Ctx(d.ctx).Debug().Int("numSamples", len(datasamples)).Msg("dbal: adding multiple datasamples in batch mode")
	err := d.insertDataSamples(datasamples)
	if err != nil {
		return err
	}

	return d.insertDataSampleDataManagers(datasamples...)
}

// insertDataSamples insert data samples in database in batch mode.
func (d *DBAL) insertDataSamples(datasamples []*asset.DataSample) error {
	_, err := d.tx.CopyFrom(
		d.ctx,
		pgx.Identifier{"datasamples"},
		[]string{"key", "channel", "owner", "checksum", "creation_date"},
		pgx.CopyFromSlice(len(datasamples), func(i int) ([]interface{}, error) {
			ds := datasamples[i]

			// expect binary representation, not string
			key, err := uuid.Parse(ds.Key)
			if err != nil {
				return nil, err
			}

			return []interface{}{key, d.channel, ds.Owner, ds.Checksum, ds.CreationDate.AsTime()}, nil
		}),
	)

	return err
}

// insertDataSampleDataManagers insert the datasample-datamanager relations in database in batch mode.
func (d *DBAL) insertDataSampleDataManagers(datasamples ...*asset.DataSample) error {
	rows := make([][]interface{}, 0)

	for _, ds := range datasamples {
		if ds.DataManagerKeys != nil {
			dataSampleKey, err := uuid.Parse(ds.Key)
			if err != nil {
				return err
			}

			for _, dmKey := range ds.DataManagerKeys {
				dataManagerKey, err := uuid.Parse(dmKey)
				if err != nil {
					return err
				}
				rows = append(rows, []interface{}{dataSampleKey, dataManagerKey})
			}
		}
	}

	_, err := d.tx.CopyFrom(
		d.ctx,
		pgx.Identifier{"datasample_datamanagers"},
		[]string{"datasample_key", "datamanager_key"},
		pgx.CopyFromRows(rows),
	)

	return err
}

func (d *DBAL) deleteDataSampleDataManagers(dataSampleKey string) error {
	stmt := getStatementBuilder().
		Delete("datasample_datamanagers").
		Where(sq.Eq{"datasample_key": dataSampleKey})

	return d.exec(stmt)
}

// UpdateDataSample implements persistence.DataSampleDBAL
func (d *DBAL) UpdateDataSample(dataSample *asset.DataSample) error {
	err := d.deleteDataSampleDataManagers(dataSample.Key)
	if err != nil {
		return err
	}

	err = d.insertDataSampleDataManagers(dataSample)
	if err != nil {
		return err
	}

	stmt := getStatementBuilder().
		Update("datasamples").
		Set("owner", dataSample.Owner).
		Set("checksum", dataSample.Checksum).
		Where(sq.Eq{"channel": d.channel, "key": dataSample.Key})

	return d.exec(stmt)
}

// GetDataSample implements persistence.DataSample
func (d *DBAL) GetDataSample(key string) (*asset.DataSample, error) {
	stmt := getStatementBuilder().
		Select("key", "owner", "checksum", "creation_date", "datamanager_keys").
		From("expanded_datasamples").
		Where(sq.Eq{"channel": d.channel, "key": key})

	row, err := d.queryRow(stmt)
	if err != nil {
		return nil, err
	}

	ds := new(sqlDataSample)

	err = row.Scan(&ds.Key, &ds.Owner, &ds.Checksum, &ds.CreationDate, &ds.DataManagerKeys)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, orcerrors.NewNotFound("datasample", key)
		}
		return nil, err
	}

	return ds.toDataSample(), nil
}

// QueryDataSamples implements persistence.DataSample
func (d *DBAL) QueryDataSamples(p *common.Pagination, filter *asset.DataSampleQueryFilter) ([]*asset.DataSample, common.PaginationToken, error) {
	offset, err := getOffset(p.Token)
	if err != nil {
		return nil, "", err
	}

	stmt := getStatementBuilder().
		Select("key", "owner", "checksum", "creation_date", "datamanager_keys").
		From("expanded_datasamples").
		Where(sq.Eq{"channel": d.channel}).
		OrderByClause("creation_date ASC, key").
		Offset(uint64(offset)).
		// Fetch page size + 1 elements to determine whether there is a next page
		Limit(uint64(p.Size + 1))

	if filter != nil && len(filter.Keys) > 0 {
		stmt = stmt.Where(sq.Eq{"key": filter.Keys})
	}

	rows, err := d.query(stmt)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var datasamples []*asset.DataSample
	var count int

	for rows.Next() {
		ds := new(sqlDataSample)

		err = rows.Scan(&ds.Key, &ds.Owner, &ds.Checksum, &ds.CreationDate, &ds.DataManagerKeys)
		if err != nil {
			return nil, "", err
		}

		datasamples = append(datasamples, ds.toDataSample())
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

	return datasamples, bookmark, nil
}

// GetDataSampleKeysByManager returns sample keys linked to a given manager.
func (d *DBAL) GetDataSampleKeysByManager(dataManagerKey string) ([]string, error) {
	stmt := getStatementBuilder().
		Select("datasample_key").
		From("datasample_datamanagers").
		Join("datasamples ds ON ds.key = datasample_datamanagers.datasample_key").
		Where(sq.Eq{"datamanager_key": dataManagerKey}).
		OrderByClause("creation_date ASC, key")

	rows, err := d.query(stmt)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var datasampleKeys []string

	for rows.Next() {
		var datasampleKey string

		err = rows.Scan(&datasampleKey)
		if err != nil {
			return nil, err
		}
		datasampleKeys = append(datasampleKeys, datasampleKey)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return datasampleKeys, nil
}
