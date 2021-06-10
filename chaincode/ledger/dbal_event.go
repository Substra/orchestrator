// Copyright 2021 Owkin Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

	selector := couchAssetQuery{
		DocType: eventResource,
	}

	assetFilter := map[string]interface{}{}
	if filter.AssetKey != "" {
		assetFilter["assetKey"] = filter.AssetKey
	}
	if len(assetFilter) > 0 {
		selector.Asset = assetFilter
	}

	b, err := json.Marshal(selector)
	if err != nil {
		return nil, "", err
	}

	queryString := fmt.Sprintf(`{"selector":%s}`, string(b))
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
