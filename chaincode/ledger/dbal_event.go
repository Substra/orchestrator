package ledger

import (
	"encoding/json"
	"time"

	"github.com/go-playground/log/v7"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	"github.com/owkin/orchestrator/lib/errors"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const eventResource = "event"

// storableEvent is a custom representation of asset.Event to enforce storing timestamp as the number of seconds since the Epoch.
type storableEvent struct {
	ID        string            `json:"id"`
	AssetKey  string            `json:"asset_key"`
	AssetKind string            `json:"asset_kind"`
	EventKind string            `json:"event_kind"`
	Channel   string            `json:"channel"`
	Timestamp int64             `json:"timestamp"`
	Metadata  map[string]string `json:"metadata"`
	Asset     json.RawMessage   `json:"asset"`
}

// newStorableEvent creates a storableEvent object from asset.Event
func newStorableEvent(e *asset.Event) (*storableEvent, error) {
	eventAsset, err := asset.MarshalEventAsset(e)
	if err != nil {
		return nil, err
	}

	proxy := &storableEvent{
		ID:        e.Id,
		AssetKey:  e.AssetKey,
		AssetKind: e.AssetKind.String(),
		EventKind: e.EventKind.String(),
		Channel:   e.Channel,
		Timestamp: e.Timestamp.AsTime().UnixNano(),
		Metadata:  e.Metadata,
		Asset:     eventAsset,
	}

	if e.Metadata == nil {
		proxy.Metadata = map[string]string{}
	}

	return proxy, nil
}

// newEventFromStorable converts a storableEvent back into an asset.Event
func (db *DB) newEventFromStorable(s *storableEvent) (*asset.Event, error) {
	logger := db.logger.WithField("event", s.ID)

	eventKind, ok := asset.EventKind_value[s.EventKind]
	if !ok {
		logger.WithField("eventKind", s.EventKind).Warn("unable to convert event kind into proto value")
		eventKind = int32(asset.EventKind_EVENT_UNKNOWN)
	}

	assetKind, ok := asset.AssetKind_value[s.AssetKind]
	if !ok {
		logger.WithField("assetKind", s.AssetKind).Warn("unable to convert asset kind into proto value")
		assetKind = int32(asset.AssetKind_ASSET_UNKNOWN)
	}

	event := &asset.Event{
		Id:        s.ID,
		AssetKey:  s.AssetKey,
		EventKind: asset.EventKind(eventKind),
		AssetKind: asset.AssetKind(assetKind),
		Channel:   s.Channel,
		Timestamp: timestamppb.New(time.Unix(0, s.Timestamp)),
		Metadata:  s.Metadata,
	}

	err := asset.UnmarshalEventAsset(s.Asset, event, event.AssetKind)
	if err != nil {
		return nil, err
	}

	return event, nil
}

func (db *DB) addSingleEvent(event *asset.Event) error {
	exists, err := db.hasKey(eventResource, event.Id)
	if err != nil {
		return err
	}
	if exists {
		return errors.NewConflict("event", event.Id)
	}

	proxy, err := newStorableEvent(event)
	if err != nil {
		return err
	}

	bytes, err := json.Marshal(proxy)
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

	assetFilter := buildEventAssetFilter(filter)
	if assetFilter != nil {
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

		var stored storedAsset
		err = json.Unmarshal(queryResult.Value, &stored)
		if err != nil {
			return nil, "", err
		}

		proxy := new(storableEvent)
		err = json.Unmarshal(stored.Asset, proxy)
		if err != nil {
			return nil, "", err
		}

		event, err := db.newEventFromStorable(proxy)
		if err != nil {
			return nil, "", err
		}

		event.Channel = db.ccStub.GetChannelID()

		events = append(events, event)
	}

	return events, bookmark.Bookmark, nil
}

// buildEventAssetFilter creates a couchdb filter from EventQueryFilter.
// it will return nil if filter is nil or empty.
func buildEventAssetFilter(filter *asset.EventQueryFilter) map[string]interface{} {
	if filter == nil {
		return nil
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
		tsFilter := make(map[string]int64)
		if filter.Start != nil {
			tsFilter["$gte"] = filter.Start.AsTime().UnixNano()
		}
		if filter.End != nil {
			tsFilter["$lte"] = filter.End.AsTime().UnixNano()
		}
		assetFilter["timestamp"] = tsFilter
	}

	if len(assetFilter) > 0 {
		return assetFilter
	}
	return nil
}
