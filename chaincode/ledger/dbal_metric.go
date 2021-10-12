package ledger

import (
	"encoding/json"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	"github.com/owkin/orchestrator/lib/errors"
)

// AddMetric stores a new metric
func (db *DB) AddMetric(obj *asset.Metric) error {
	exists, err := db.hasKey(asset.MetricKind, obj.GetKey())
	if err != nil {
		return err
	}
	if exists {
		return errors.NewConflict(asset.MetricKind, obj.Key)
	}

	objBytes, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	return db.putState(asset.MetricKind, obj.GetKey(), objBytes)
}

// GetMetric retrieves an metric by its key
func (db *DB) GetMetric(key string) (*asset.Metric, error) {
	o := asset.Metric{}

	b, err := db.getState(asset.MetricKind, key)
	if err != nil {
		return &o, err
	}

	err = json.Unmarshal(b, &o)
	return &o, err
}

// MetricExists implements persistence.MetricDBAL
func (db *DB) MetricExists(key string) (bool, error) {
	return db.hasKey(asset.MetricKind, key)
}

// QueryMetrics retrieves all metrics
func (db *DB) QueryMetrics(p *common.Pagination) ([]*asset.Metric, common.PaginationToken, error) {
	query := richQuerySelector{
		Selector: couchAssetQuery{
			DocType: asset.MetricKind,
		},
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

	metrics := make([]*asset.Metric, 0)

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
		metric := &asset.Metric{}
		err = json.Unmarshal(storedAsset.Asset, metric)
		if err != nil {
			return nil, "", err
		}

		metrics = append(metrics, metric)
	}

	return metrics, bookmark.Bookmark, nil
}
