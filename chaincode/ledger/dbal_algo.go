package ledger

import (
	"encoding/json"

	"github.com/go-playground/log/v7"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	"github.com/owkin/orchestrator/lib/errors"
	"google.golang.org/protobuf/encoding/protojson"
)

// AddAlgo stores a new algo
func (db *DB) AddAlgo(algo *asset.Algo) error {
	exists, err := db.hasKey(asset.AlgoKind, algo.GetKey())
	if err != nil {
		return err
	}
	if exists {
		return errors.NewConflict(asset.AlgoKind, algo.Key)
	}

	algoBytes, err := marshaller.Marshal(algo)
	if err != nil {
		return err
	}
	err = db.putState(asset.AlgoKind, algo.GetKey(), algoBytes)
	if err != nil {
		return err
	}

	return nil
}

// GetAlgo retrieves an algo by its key
func (db *DB) GetAlgo(key string) (*asset.Algo, error) {
	a := asset.Algo{}

	b, err := db.getState(asset.AlgoKind, key)
	if err != nil {
		return &a, err
	}

	err = protojson.Unmarshal(b, &a)
	return &a, err
}

// QueryAlgos retrieves all algos
func (db *DB) QueryAlgos(p *common.Pagination, filter *asset.AlgoQueryFilter) ([]*asset.Algo, common.PaginationToken, error) {
	logger := db.logger.WithFields(
		log.F("pagination", p),
		log.F("filter", filter),
	)
	logger.Debug("get algos")

	query := richQuerySelector{
		Selector: couchAssetQuery{
			DocType: asset.AlgoKind,
		},
	}

	if filter != nil {
		assetFilter := map[string]interface{}{}

		if filter.ComputePlanKey != "" {
			algoKeys, err := db.getComputePlanAlgoKeys(filter.ComputePlanKey)
			if err != nil {
				return nil, "", err
			}
			assetFilter["key"] = map[string]interface{}{
				"$in": algoKeys,
			}
		}

		if len(assetFilter) > 0 {
			query.Selector.Asset = assetFilter
		}
	}

	b, err := json.Marshal(query)
	if err != nil {
		return nil, "", err
	}

	queryString := string(b)

	log.WithField("couchQuery", queryString).Debug("mango query")

	resultsIterator, bookmark, err := db.getQueryResultWithPagination(queryString, int32(p.Size), p.Token)
	if err != nil {
		return nil, "", err
	}
	defer resultsIterator.Close()

	algos := make([]*asset.Algo, 0)

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
		algo := &asset.Algo{}
		err = protojson.Unmarshal(storedAsset.Asset, algo)
		if err != nil {
			return nil, "", err
		}

		algos = append(algos, algo)
	}

	return algos, bookmark.Bookmark, nil
}

// AlgoExists implements persistence.AlgoDBAL
func (db *DB) AlgoExists(key string) (bool, error) {
	return db.hasKey(asset.AlgoKind, key)
}

// getComputePlanAlgoKeys returns keys of Algo in use in a ComputePlan
func (db *DB) getComputePlanAlgoKeys(planKey string) ([]string, error) {
	logger := db.logger.WithField("computePlan", planKey)

	query := richQuerySelector{
		Selector: couchAssetQuery{
			DocType: asset.ComputeTaskKind,
			Asset: map[string]interface{}{
				"compute_plan_key": planKey,
			},
		},
		Fields: []string{"asset.algo.key"},
	}

	b, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}
	queryString := string(b)

	logger.WithField("couchQuery", queryString).Debug("mango query")

	resultsIterator, err := db.getQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	// Using a map to deduplicate algo keys
	uniqueKeys := make(map[string]interface{})

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
		task := &asset.ComputeTask{}
		err = protojson.Unmarshal(storedAsset.Asset, task)
		if err != nil {
			return nil, err
		}
		uniqueKeys[task.Algo.Key] = struct{}{}
	}

	keys := make([]string, len(uniqueKeys))
	i := 0
	for k := range uniqueKeys {
		keys[i] = k
		i++
	}

	return keys, nil
}

// UpdateAlgo implements persistence.AlgoDBAL
func (db *DB) UpdateAlgo(algo *asset.Algo) error {
	algoBytes, err := marshaller.Marshal(algo)
	if err != nil {
		return err
	}

	return db.putState(asset.AlgoKind, algo.GetKey(), algoBytes)
}
