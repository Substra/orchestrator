package ledger

import (
	"github.com/substra/orchestrator/lib/asset"
	"github.com/substra/orchestrator/lib/errors"
	"google.golang.org/protobuf/encoding/protojson"
)

func (db *DB) GetModel(key string) (*asset.Model, error) {
	model := new(asset.Model)

	b, err := db.getState(asset.ModelKind, key)
	if err != nil {
		return nil, err
	}

	err = protojson.Unmarshal(b, model)
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

func (db *DB) GetComputeTaskOutputModels(computeTaskKey string) ([]*asset.Model, error) {
	elementKeys, err := db.getIndexKeys(modelTaskKeyIndex, []string{asset.ModelKind, computeTaskKey})
	if err != nil {
		return nil, err
	}

	db.logger.Debug().Int("numChildren", len(elementKeys)).Msg("GetComputeTaskOutputModels")

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

func (db *DB) AddModel(model *asset.Model, identifier string) error {
	exists, err := db.hasKey(asset.ModelKind, model.GetKey())
	if err != nil {
		return err
	}
	if exists {
		return errors.NewConflict(asset.ModelKind, model.Key)
	}
	bytes, err := marshaller.Marshal(model)
	if err != nil {
		return err
	}

	err = db.putState(asset.ModelKind, model.GetKey(), bytes)
	if err != nil {
		return err
	}

	return db.createIndex(modelTaskKeyIndex, []string{asset.ModelKind, model.GetComputeTaskKey(), model.GetKey()})
}

func (db *DB) UpdateModel(model *asset.Model) error {
	bytes, err := marshaller.Marshal(model)
	if err != nil {
		return err
	}

	err = db.putState(asset.ModelKind, model.GetKey(), bytes)
	if err != nil {
		return err
	}

	return nil
}
