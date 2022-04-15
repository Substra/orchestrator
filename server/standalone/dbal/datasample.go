package dbal

import (
	"errors"
	"strconv"

	sq "github.com/Masterminds/squirrel"
	"github.com/go-playground/log/v7"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	orcerrors "github.com/owkin/orchestrator/lib/errors"
	"google.golang.org/protobuf/encoding/protojson"
)

// DataSampleExists implements persistence.DataSampleDBAL
func (d *DBAL) DataSampleExists(key string) (bool, error) {
	row := d.tx.QueryRow(d.ctx, `select count(id) from "datasamples" where id=$1 and channel=$2`, key, d.channel)

	var count int
	err := row.Scan(&count)

	return count == 1, err
}

// AddDataSamples insert samples in storage in batch mode.
func (d *DBAL) AddDataSamples(datasamples ...*asset.DataSample) error {
	log.WithField("numSamples", len(datasamples)).Debug("dbal: adding multiple datasamples in batch mode")

	_, err := d.tx.CopyFrom(
		d.ctx,
		pgx.Identifier{"datasamples"},
		[]string{"id", "asset", "channel"},
		pgx.CopyFromSlice(len(datasamples), func(i int) ([]interface{}, error) {
			v, err := protojson.Marshal(datasamples[i])
			if err != nil {
				return nil, err
			}
			// expect binary representation, not string
			id, err := uuid.Parse(datasamples[i].Key)
			if err != nil {
				return nil, err
			}
			return []interface{}{id, v, d.channel}, nil
		}),
	)

	return err
}

// UpdateDataSample implements persistence.DataSampleDBAL
func (d *DBAL) UpdateDataSample(dataSample *asset.DataSample) error {
	stmt := `update "datasamples" set asset=$3 where id=$1 and channel=$2`
	_, err := d.tx.Exec(d.ctx, stmt, dataSample.GetKey(), d.channel, dataSample)
	return err
}

// GetDataSample implements persistence.DataSample
func (d *DBAL) GetDataSample(key string) (*asset.DataSample, error) {
	row := d.tx.QueryRow(d.ctx, `select "asset" from "datasamples" where id=$1 and channel=$2`, key, d.channel)

	datasample := new(asset.DataSample)
	err := row.Scan(datasample)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, orcerrors.NewNotFound("datasample", key)
		}
		return nil, err
	}

	return datasample, nil
}

// QueryDataSamples implements persistence.DataSample
func (d *DBAL) QueryDataSamples(p *common.Pagination, filter *asset.DataSampleQueryFilter) ([]*asset.DataSample, common.PaginationToken, error) {
	offset, err := getOffset(p.Token)
	if err != nil {
		return nil, "", err
	}

	stmt := getStatementBuilder().
		Select("asset").
		From("datasamples").
		Where(sq.Eq{"channel": d.channel}).
		OrderByClause("asset->>'creationDate' asc, id").
		Offset(uint64(offset)).
		// Fetch page size + 1 elements to determine whether there is a next page
		Limit(uint64(p.Size + 1))

	if filter != nil && len(filter.Keys) > 0 {
		stmt = stmt.Where(sq.Eq{"id": filter.Keys})
	}

	rows, err := d.query(stmt)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var datasamples []*asset.DataSample
	var count int

	for rows.Next() {
		datasample := new(asset.DataSample)

		err = rows.Scan(datasample)
		if err != nil {
			return nil, "", err
		}

		datasamples = append(datasamples, datasample)
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

	return datasamples, bookmark, nil
}

// GetDataSampleKeysByManager returns sample keys linked to given manager.
func (d *DBAL) GetDataSampleKeysByManager(dataManagerKey string, testOnly bool) ([]string, error) {
	testOnlyFilter := `not`
	if testOnly {
		testOnlyFilter = ``
	}

	query := `select "id" from "datasamples" where channel=$1 and (asset->'dataManagerKeys') ? $2 and ` +
		testOnlyFilter +
		` (asset ? 'testOnly' and (asset->'testOnly')::boolean) order by asset->>'creationDate' asc, id`

	rows, err := d.tx.Query(d.ctx, query, d.channel, dataManagerKey)

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
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return datasampleKeys, nil
}
