package ledger

import (
	"encoding/json"

	"github.com/go-playground/log/v7"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	"github.com/owkin/orchestrator/lib/errors"
)

func (db *DB) GetModel(key string) (*asset.Model, error) {
	model := new(asset.Model)

	b, err := db.getState(asset.ModelKind, key)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, model)
	if err != nil {
		return nil, err
	}
	return model, nil
}

func (db *DB) ModelExists(key string) (bool, error) {
	exists, err := db.hasKey(asset.ModelKind, key)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (db *DB) GetComputeTaskOutputModels(key string) ([]*asset.Model, error) {
	elementKeys, err := db.getIndexKeys(modelTaskKeyIndex, []string{asset.ModelKind, key})
	if err != nil {
		return nil, err
	}

	db.logger.WithField("numChildren", len(elementKeys)).Debug("GetComputeTaskOutputModels")

	models := []*asset.Model{}
	for _, modelKey := range elementKeys {
		model, err := db.GetModel(modelKey)
		if err != nil {
			return nil, err
		}
		models = append(models, model)
	}

	return models, nil
}

func (db *DB) AddModel(model *asset.Model) error {
	exists, err := db.hasKey(asset.ModelKind, model.GetKey())
	if err != nil {
		return err
	}
	if exists {
		return errors.NewConflict(asset.ModelKind, model.Key)
	}
	bytes, err := json.Marshal(model)
	if err != nil {
		return err
	}

	err = db.putState(asset.ModelKind, model.GetKey(), bytes)
	if err != nil {
		return err
	}

	return db.createIndex(modelTaskKeyIndex, []string{asset.ModelKind, model.GetComputeTaskKey(), model.GetKey()})
}

func (db *DB) QueryModels(c asset.ModelCategory, p *common.Pagination) ([]*asset.Model, common.PaginationToken, error) {
	logger := db.logger.WithFields(
		log.F("pagination", p),
		log.F("model_category", c.String()),
	)
	logger.Debug("get models")

	query := richQuerySelector{
		Selector: couchAssetQuery{
			DocType: asset.ModelKind,
		},
	}

	assetFilter := map[string]interface{}{}
	if c != asset.ModelCategory_MODEL_UNKNOWN {
		assetFilter["category"] = c.String()
	}
	if len(assetFilter) > 0 {
		query.Selector.Asset = assetFilter
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

	models := make([]*asset.Model, 0)

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
		model := &asset.Model{}
		err = json.Unmarshal(storedAsset.Asset, model)
		if err != nil {
			return nil, "", err
		}

		models = append(models, model)
	}

	return models, bookmark.Bookmark, nil
}

func (db *DB) UpdateModel(model *asset.Model) error {
	bytes, err := json.Marshal(model)
	if err != nil {
		return err
	}

	err = db.putState(asset.ModelKind, model.GetKey(), bytes)
	if err != nil {
		return err
	}

	return nil
}
