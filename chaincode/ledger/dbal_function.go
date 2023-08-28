package ledger

import (
	"encoding/json"

	"github.com/substra/orchestrator/lib/asset"
	"github.com/substra/orchestrator/lib/common"
	"github.com/substra/orchestrator/lib/errors"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// AddFunction stores a new function
func (db *DB) AddFunction(function *asset.Function) error {
	exists, err := db.hasKey(asset.FunctionKind, function.GetKey())
	if err != nil {
		return err
	}
	if exists {
		return errors.NewConflict(asset.FunctionKind, function.Key)
	}

	functionBytes, err := marshaller.Marshal(function)
	if err != nil {
		return err
	}
	err = db.putState(asset.FunctionKind, function.GetKey(), functionBytes)
	if err != nil {
		return err
	}

	return nil
}

// GetFunction retrieves an function by its key
func (db *DB) GetFunction(key string) (*asset.Function, error) {
	a := asset.Function{}

	b, err := db.getState(asset.FunctionKind, key)
	if err != nil {
		return &a, err
	}

	err = protojson.Unmarshal(b, &a)
	return &a, err
}

// QueryFunctions retrieves all functions
func (db *DB) QueryFunctions(p *common.Pagination, filter *asset.FunctionQueryFilter) ([]*asset.Function, common.PaginationToken, error) {
	logger := db.logger.With().
		Interface("pagination", p).
		Interface("filter", filter).
		Logger()

	logger.Debug().Msg("get functions")

	query := richQuerySelector{
		Selector: couchAssetQuery{
			DocType: asset.FunctionKind,
		},
	}

	if filter != nil {
		assetFilter := map[string]interface{}{}

		if filter.ComputePlanKey != "" {
			functionKeys, err := db.getComputePlanFunctionKeys(filter.ComputePlanKey)
			if err != nil {
				return nil, "", err
			}
			assetFilter["key"] = map[string]interface{}{
				"$in": functionKeys,
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

	logger.Debug().Str("couchQuery", queryString).Msg("mango query")

	resultsIterator, bookmark, err := db.getQueryResultWithPagination(queryString, int32(p.Size), p.Token)
	if err != nil {
		return nil, "", err
	}
	defer resultsIterator.Close()

	functions := make([]*asset.Function, 0)

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
		function := &asset.Function{}
		err = protojson.Unmarshal(storedAsset.Asset, function)
		if err != nil {
			return nil, "", err
		}

		functions = append(functions, function)
	}

	return functions, bookmark.Bookmark, nil
}

// FunctionExists implements persistence.FunctionDBAL
func (db *DB) FunctionExists(key string) (bool, error) {
	return db.hasKey(asset.FunctionKind, key)
}

// getComputePlanFunctionKeys returns keys of Function in use in a ComputePlan
func (db *DB) getComputePlanFunctionKeys(planKey string) ([]string, error) {
	logger := db.logger.With().Str("computePlan", planKey).Logger()

	query := richQuerySelector{
		Selector: couchAssetQuery{
			DocType: asset.ComputeTaskKind,
			Asset: map[string]interface{}{
				"compute_plan_key": planKey,
			},
		},
		Fields: []string{"asset.function.key"},
	}

	b, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}
	queryString := string(b)

	logger.Debug().Str("couchQuery", queryString).Msg("mango query")

	resultsIterator, err := db.getQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	// Using a map to deduplicate function keys
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
		uniqueKeys[task.FunctionKey] = struct{}{}
	}

	keys := make([]string, len(uniqueKeys))
	i := 0
	for k := range uniqueKeys {
		keys[i] = k
		i++
	}

	return keys, nil
}

// UpdateFunction implements persistence.FunctionDBAL
func (db *DB) UpdateFunction(function *asset.Function) error {
	functionBytes, err := marshaller.Marshal(function)
	if err != nil {
		return err
	}

	return db.putState(asset.FunctionKind, function.GetKey(), functionBytes)
}
func (db *DB) UpdateFunctionStatus(functionKey string, functionStatus asset.FunctionStatus) error {
	// We need the current function to be able to update its indexes
	prevFunction, err := db.GetFunction(functionKey)
	if err != nil {
		return err
	}

	updatedFunction := proto.Clone(prevFunction).(*asset.Function)
	updatedFunction.Status = functionStatus

	b, err := marshaller.Marshal(updatedFunction)
	if err != nil {
		return err
	}

	err = db.putState(asset.FunctionKind, functionKey, b)
	if err != nil {
		return err
	}

	return nil
}
