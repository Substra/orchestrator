package ledger

import (
	"encoding/json"
	"fmt"

	"github.com/go-playground/log/v7"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	"github.com/owkin/orchestrator/lib/errors"
)

const eventResource = "event"

func (db *DB) addSingleEvent(event *asset.Event) error {
	exists, err := db.hasKey(eventResource, event.Id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("failed to add event: %w", errors.ErrConflict)
	}
	protoTime, err := db.ccStub.GetTxTimestamp()
	if err != nil {
		return err
	}
	event.Timestamp = uint64(protoTime.AsTime().Unix())
	bytes, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return db.putState(eventResource, event.Id, bytes)
}

func (db *DB) AddEvents(events ...*asset.Event) error {
	for _, e := range events {
		err := db.addSingleEvent(e)
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *DB) QueryEvents(p *common.Pagination, filter *asset.EventQueryFilter) ([]*asset.Event, common.PaginationToken, error) {
	logger := db.logger.WithFields(
		log.F("pagination", p),
		log.F("filter", filter),
	)
	logger.Debug("query events")

	query := richQuerySelector{
		Selector: couchAssetQuery{
			DocType: eventResource,
		},
	}

	assetFilter := map[string]interface{}{}
	if filter.AssetKey != "" {
		assetFilter["assetKey"] = filter.AssetKey
	}
	if filter.AssetKind != asset.AssetKind_ASSET_UNKNOWN {
		assetFilter["assetKind"] = filter.AssetKind.String()
	}
	if filter.EventKind != asset.EventKind_EVENT_UNKNOWN {
		assetFilter["eventKind"] = filter.EventKind.String()
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

	events := make([]*asset.Event, 0)

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
		event := &asset.Event{}
		err = json.Unmarshal(storedAsset.Asset, event)
		if err != nil {
			return nil, "", err
		}

		events = append(events, event)
	}

	return events, bookmark.Bookmark, nil
}
