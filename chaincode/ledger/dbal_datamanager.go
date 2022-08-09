package ledger

import (
	"encoding/json"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	"github.com/owkin/orchestrator/lib/errors"
	"google.golang.org/protobuf/encoding/protojson"
)

// AddDataManager stores a new DataManager
func (db *DB) AddDataManager(datamanager *asset.DataManager) error {
	exists, err := db.hasKey(asset.DataManagerKind, datamanager.GetKey())
	if err != nil {
		return err
	}
	if exists {
		return errors.NewConflict(asset.DataManagerKind, datamanager.Key)
	}

	dataManagerBytes, err := marshaller.Marshal(datamanager)
	if err != nil {
		return err
	}

	err = db.putState(asset.DataManagerKind, datamanager.GetKey(), dataManagerBytes)
	if err != nil {
		return err
	}

	return nil
}

// DataManagerExists implements persistence.DataManagerDBAL
func (db *DB) DataManagerExists(key string) (bool, error) {
	return db.hasKey(asset.DataManagerKind, key)
}

// GetDataManager implements persistence.DataManagerDBAL
func (db *DB) GetDataManager(key string) (*asset.DataManager, error) {
	d := asset.DataManager{}

	b, err := db.getState(asset.DataManagerKind, key)
	if err != nil {
		return nil, err
	}

	err = protojson.Unmarshal(b, &d)
	return &d, err
}

// QueryDataManagers implements persistence.DataManagerDBAL
func (db *DB) QueryDataManagers(p *common.Pagination) ([]*asset.DataManager, common.PaginationToken, error) {
	query := richQuerySelector{
		Selector: couchAssetQuery{
			DocType: asset.DataManagerKind,
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

	dms := make([]*asset.DataManager, 0)

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
		dm := &asset.DataManager{}
		err = protojson.Unmarshal(storedAsset.Asset, dm)
		if err != nil {
			return nil, "", err
		}

		dms = append(dms, dm)
	}

	return dms, bookmark.Bookmark, nil
}

// UpdateDataManager implements persistence.DataManagerDBAL
func (db *DB) UpdateDataManager(plan *asset.DataManager) error {
	planBytes, err := marshaller.Marshal(plan)
	if err != nil {
		return err
	}

	return db.putState(asset.DataManagerKind, plan.GetKey(), planBytes)
}
