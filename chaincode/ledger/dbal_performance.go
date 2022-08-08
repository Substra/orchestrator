package ledger

import (
	"encoding/json"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	"github.com/owkin/orchestrator/lib/errors"
	"google.golang.org/protobuf/encoding/protojson"
)

func (db *DB) AddPerformance(perf *asset.Performance, identifier string) error {
	exists, err := db.hasKey(asset.PerformanceKind, perf.GetKey())
	if err != nil {
		return err
	}
	if exists {
		return errors.NewConflict(asset.PerformanceKind, perf.GetKey())
	}
	bytes, err := marshaller.Marshal(perf)
	if err != nil {
		return err
	}

	err = db.putState(asset.PerformanceKind, perf.GetKey(), bytes)
	if err != nil {
		return err
	}

	return db.createIndex(performanceIndex, []string{asset.PerformanceKind, perf.GetComputeTaskKey(), perf.GetMetricKey()})
}

// PerformanceExists implements persistence.PerformanceDBAL
func (db *DB) PerformanceExists(perf *asset.Performance) (bool, error) {
	return db.hasKey(asset.PerformanceKind, perf.GetKey())
}

func (db *DB) QueryPerformances(p *common.Pagination, filter *asset.PerformanceQueryFilter) ([]*asset.Performance, common.PaginationToken, error) {
	query := richQuerySelector{
		Selector: couchAssetQuery{
			DocType: asset.PerformanceKind,
		},
	}

	if filter != nil {
		assetFilter := map[string]interface{}{}
		if filter.ComputeTaskKey != "" {
			assetFilter["compute_task_key"] = filter.ComputeTaskKey
		}
		if filter.MetricKey != "" {
			assetFilter["metric_key"] = filter.MetricKey
		}
		query.Selector.Asset = assetFilter
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

	performances := make([]*asset.Performance, 0)

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
		performance := &asset.Performance{}
		err = protojson.Unmarshal(storedAsset.Asset, performance)
		if err != nil {
			return nil, "", err
		}

		performances = append(performances, performance)
	}
	return performances, bookmark.Bookmark, nil
}
