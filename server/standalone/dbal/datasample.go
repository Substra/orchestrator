package dbal

import (
	"context"
	"errors"
	"fmt"
	"strconv"

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
	row := d.tx.QueryRow(context.Background(), `select count(id) from "datasamples" where id=$1 and channel=$2`, key, d.channel)

	var count int
	err := row.Scan(&count)

	return count == 1, err
}

// AddDataSamples insert samples in storage according to the most efficient way.
// Up to 5 samples, they will be inserted one by one.
// For more than 5 samples they will be processed in batch.
func (d *DBAL) AddDataSamples(datasamples ...*asset.DataSample) error {
	if len(datasamples) >= 5 {
		log.WithField("numSamples", len(datasamples)).Debug("dbal: adding multiple datasamples in batch mode")
		return d.addDataSamples(datasamples)
	}

	for _, ds := range datasamples {
		err := d.addDataSample(ds)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *DBAL) addDataSample(dataSample *asset.DataSample) error {
	stmt := `insert into "datasamples" ("id", "asset", "channel") values ($1, $2, $3)`
	_, err := d.tx.Exec(context.Background(), stmt, dataSample.GetKey(), dataSample, d.channel)
	return err
}

func (d *DBAL) addDataSamples(samples []*asset.DataSample) error {
	_, err := d.tx.CopyFrom(
		context.Background(),
		pgx.Identifier{"datasamples"},
		[]string{"id", "asset", "channel"},
		pgx.CopyFromSlice(len(samples), func(i int) ([]interface{}, error) {
			v, err := protojson.Marshal(samples[i])
			if err != nil {
				return nil, err
			}
			// expect binary representation, not string
			id, err := uuid.Parse(samples[i].Key)
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
	_, err := d.tx.Exec(context.Background(), stmt, dataSample.GetKey(), d.channel, dataSample)
	return err
}

// GetDataSample implements persistence.DataSample
func (d *DBAL) GetDataSample(key string) (*asset.DataSample, error) {
	row := d.tx.QueryRow(context.Background(), `select "asset" from "datasamples" where id=$1 and channel=$2`, key, d.channel)

	datasample := new(asset.DataSample)
	err := row.Scan(&datasample)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("datasample not found: %w", orcerrors.ErrNotFound)
		}
		return nil, err
	}

	return datasample, nil
}

// QueryDataSamples implements persistence.DataSample
func (d *DBAL) QueryDataSamples(p *common.Pagination) ([]*asset.DataSample, common.PaginationToken, error) {
	var rows pgx.Rows
	var err error

	offset, err := getOffset(p.Token)
	if err != nil {
		return nil, "", err
	}

	query := `select "asset" from "datasamples" where channel=$3 order by created_at asc limit $1 offset $2`
	rows, err = d.tx.Query(context.Background(), query, p.Size+1, offset, d.channel)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var datasamples []*asset.DataSample
	var count int

	for rows.Next() {
		datasample := new(asset.DataSample)

		err = rows.Scan(&datasample)
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

// GetDataSamplesKeysByDataManager implements persistence.DataSample
func (d *DBAL) GetDataSamplesKeysByDataManager(dataManagerKey string, testOnly bool) ([]string, error) {
	var rows pgx.Rows
	var err error

	testOnlyFilter := `not`
	if testOnly {
		testOnlyFilter = ``
	}

	query := `select "id" from "datasamples" where channel=$1 and (asset->'dataManagerKeys') ? $2 and ` + testOnlyFilter + ` (asset ? 'testOnly' and (asset->'testOnly')::boolean) order by created_at asc`

	rows, err = d.tx.Query(context.Background(), query, d.channel, dataManagerKey)

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
