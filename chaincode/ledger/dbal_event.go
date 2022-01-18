package ledger

import (
	"encoding/json"
	"time"

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
		return errors.NewConflict("event", event.Id)
	}
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

func (db *DB) QueryEvents(p *common.Pagination, filter *asset.EventQueryFilter, sortOrder asset.SortOrder) ([]*asset.Event, common.PaginationToken, error) {
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

	sort := CouchDBSortAsc
	if sortOrder == asset.SortOrder_DESCENDING {
		sort = CouchDBSortDesc
	}

	query.Sort = []map[string]string{
		{"asset.timestamp": sort},
		{"asset.id": sort},
	}

	assetFilter := map[string]interface{}{}
	if filter.AssetKey != "" {
		assetFilter["asset_key"] = filter.AssetKey
	}
	if filter.AssetKind != asset.AssetKind_ASSET_UNKNOWN {
		assetFilter["asset_kind"] = filter.AssetKind.String()
	}
	if filter.EventKind != asset.EventKind_EVENT_UNKNOWN {
		assetFilter["event_kind"] = filter.EventKind.String()
	}
	if filter.Metadata != nil {
		assetFilter["metadata"] = filter.Metadata
	}
	if filter.Start != nil || filter.End != nil {
		tsFilter := make(map[string]string)
		if filter.Start != nil {
			tsFilter["$gte"] = filter.Start.AsTime().Format(time.RFC3339Nano)
		}
		if filter.End != nil {
			tsFilter["$lte"] = filter.End.AsTime().Format(time.RFC3339Nano)
		}
		assetFilter["timestamp"] = tsFilter
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
		event.Channel = db.ccStub.GetChannelID()

		events = append(events, event)
	}

	return events, bookmark.Bookmark, nil
}
