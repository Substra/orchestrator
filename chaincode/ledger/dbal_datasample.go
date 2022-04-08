package ledger

import (
	"encoding/json"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	"github.com/owkin/orchestrator/lib/errors"
	"google.golang.org/protobuf/encoding/protojson"
)

// DataSampleExists implements persistence.DataSampleDBAL
func (db *DB) DataSampleExists(key string) (bool, error) {
	return db.hasKey(asset.DataSampleKind, key)
}

func (db *DB) AddDataSamples(samples ...*asset.DataSample) error {
	for _, sample := range samples {
		err := db.addDataSample(sample)
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *DB) addDataSample(dataSample *asset.DataSample) error {
	exists, err := db.hasKey(asset.DataSampleKind, dataSample.GetKey())
	if err != nil {
		return err
	}
	if exists {
		return errors.NewConflict(asset.DataSampleKind, dataSample.Key)
	}

	dataSampleBytes, err := marshaller.Marshal(dataSample)
	if err != nil {
		return err
	}

	return db.putState(asset.DataSampleKind, dataSample.GetKey(), dataSampleBytes)
}

// UpdateDataSample implements persistence.DataSampleDBAL
func (db *DB) UpdateDataSample(dataSample *asset.DataSample) error {
	dataSampleBytes, err := marshaller.Marshal(dataSample)
	if err != nil {
		return err
	}

	return db.putState(asset.DataSampleKind, dataSample.GetKey(), dataSampleBytes)
}

// GetDataSample implements persistence.DataSampleDBAL
func (db *DB) GetDataSample(key string) (*asset.DataSample, error) {
	o := asset.DataSample{}

	b, err := db.getState(asset.DataSampleKind, key)

	if err != nil {
		return nil, err
	}

	err = protojson.Unmarshal(b, &o)
	return &o, err
}

// QueryDataSamples implements persistence.DataSampleDBAL
func (db *DB) QueryDataSamples(p *common.Pagination, filter *asset.DataSampleQueryFilter) ([]*asset.DataSample, common.PaginationToken, error) {
	query := richQuerySelector{
		Selector: couchAssetQuery{
			DocType: asset.DataSampleKind,
		},
	}

	if filter != nil && len(filter.Keys) > 0 {
		query.Selector.Asset = map[string]interface{}{"key": map[string]interface{}{"$in": filter.Keys}}
	}

	b, err := json.Marshal(query)
	if err != nil {
		return nil, "", err
	}

	queryString := string(b)

	resultsIterator, bookmark, err := db.getQueryResultWithPagination(queryString, int32(p.Size), p.Token)
	if err != nil {
		return nil, "", err
	}
	defer resultsIterator.Close()

	datasamples := make([]*asset.DataSample, 0)

	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, "", err
		}
		var storedAsset storedAsset
		err = json.Unmarshal(queryResult.Value, &storedAsset)
		if err != nil {
			return nil, "", err
		}
		ds := &asset.DataSample{}
		err = protojson.Unmarshal(storedAsset.Asset, ds)
		if err != nil {
			return nil, "", err
		}

		datasamples = append(datasamples, ds)
	}

	return datasamples, bookmark.Bookmark, nil
}

// GetDataSampleKeysByManager returns sample keys linked to given manager.
func (db *DB) GetDataSampleKeysByManager(dataManagerKey string, testOnly bool) ([]string, error) {
	query := richQuerySelector{
		Selector: couchAssetQuery{
			DocType: asset.DataSampleKind,
			Asset: map[string]interface{}{
				"data_manager_keys": map[string]interface{}{
					"$elemMatch": map[string]interface{}{
						"$eq": dataManagerKey,
					},
				},
				"test_only": testOnly,
			},
		},
		Fields: []string{"asset.key"},
	}

	b, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}

	queryString := string(b)

	resultsIterator, err := db.getQueryResult(queryString)
	if err != nil {
		return nil, err
	}

	defer resultsIterator.Close()

	keys := make([]string, 0)

	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var storedAsset storedAsset
		err = json.Unmarshal(queryResult.Value, &storedAsset)
		if err != nil {
			return nil, err
		}
		ds := &asset.DataSample{}
		err = protojson.Unmarshal(storedAsset.Asset, ds)
		if err != nil {
			return nil, err
		}

		keys = append(keys, ds.Key)
	}

	return keys, nil
}
