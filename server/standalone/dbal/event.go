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

package dbal

import (
	"database/sql"
	"strconv"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
)

func (d *DBAL) AddEvent(event *asset.Event) error {
	stmt := `insert into "events" ("id", "asset_key", "event", "channel") values ($1, $2, $3, $4)`
	_, err := d.tx.Exec(stmt, event.Id, event.AssetKey, event, d.channel)
	return err
}

func (d *DBAL) QueryEvents(p *common.Pagination, filter *asset.EventQueryFilter) ([]*asset.Event, common.PaginationToken, error) {
	var rows *sql.Rows
	var err error

	offset, err := getOffset(p.Token)
	if err != nil {
		return nil, "", err
	}

	pgDialect := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	builder := pgDialect.Select("event", "created_at").
		From("events").
		Where(squirrel.Eq{"channel": d.channel}).
		OrderByClause("created_at ASC").
		Offset(uint64(offset)).
		Limit(uint64(p.Size + 1))

	builder = eventFilterToQuery(filter, builder)

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, "", err
	}

	rows, err = d.tx.Query(query, args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var events []*asset.Event
	var count int

	for rows.Next() {
		event := new(asset.Event)
		creation := new(time.Time)

		err = rows.Scan(&event, &creation)
		if err != nil {
			return nil, "", err
		}

		event.Timestamp = uint64(creation.Unix())
		events = append(events, event)
		count++

		if count == int(p.Size) {
			break
		}
	}
	if err := rows.Err(); err != nil {
		return nil, "", err
	}

	bookmark := ""
	if count == int(p.Size) && rows.Next() {
		// there is more to fetch
		bookmark = strconv.Itoa(offset + count)
	}

	return events, bookmark, nil
}

// eventFilterToQuery convert as filter into query string and param list
func eventFilterToQuery(filter *asset.EventQueryFilter, builder squirrel.SelectBuilder) squirrel.SelectBuilder {
	if filter == nil {
		return builder
	}

	if filter.AssetKey != "" {
		builder = builder.Where(squirrel.Eq{"event->>'assetKey'": filter.AssetKey})
	}
	if filter.AssetKind != asset.AssetKind_ASSET_UNKNOWN {
		builder = builder.Where(squirrel.Eq{"event->>'assetKind'": filter.AssetKind.String()})
	}
	if filter.EventKind != asset.EventKind_EVENT_UNKNOWN {
		builder = builder.Where(squirrel.Eq{"event->>'eventKind'": filter.EventKind.String()})
	}

	return builder
}
