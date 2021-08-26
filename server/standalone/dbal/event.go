package dbal

import (
	"strconv"

	"github.com/Masterminds/squirrel"
	"github.com/go-playground/log/v7"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
)

// AddEvents insert events in storage according to the most efficient way.
// Up to 5 events, they will be inserted one by one.
// For more than 5 events they will be processed in batch.
func (d *DBAL) AddEvents(events ...*asset.Event) error {
	if len(events) >= 5 {
		log.WithField("numEvents", len(events)).Debug("dbal: adding multiple events in batch mode")
		return d.addEvents(events)
	}

	for _, e := range events {
		err := d.addEvent(e)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *DBAL) addEvent(event *asset.Event) error {
	stmt := `insert into "events" ("id", "asset_key", "event", "channel") values ($1, $2, $3, $4)`
	_, err := d.tx.Exec(d.ctx, stmt, event.Id, event.AssetKey, event, d.channel)
	return err
}

// addEvents rely on COPY FROM directive a is faster for large number of items.
// According to the doc, it might even be faster for as few as 5 rows.
func (d *DBAL) addEvents(events []*asset.Event) error {
	_, err := d.tx.CopyFrom(
		d.ctx,
		pgx.Identifier{"events"},
		[]string{"id", "asset_key", "event", "channel"},
		pgx.CopyFromSlice(len(events), func(i int) ([]interface{}, error) {
			v, err := events[i].Value()
			if err != nil {
				return nil, err
			}
			// expect binary representation, not string
			id, err := uuid.Parse(events[i].Id)
			if err != nil {
				return nil, err
			}
			assetKey, err := uuid.Parse(events[i].AssetKey)
			if err != nil {
				return nil, err
			}
			return []interface{}{id, assetKey, v, d.channel}, nil
		}),
	)

	return err
}

func (d *DBAL) QueryEvents(p *common.Pagination, filter *asset.EventQueryFilter) ([]*asset.Event, common.PaginationToken, error) {
	var rows pgx.Rows
	var err error

	offset, err := getOffset(p.Token)
	if err != nil {
		return nil, "", err
	}

	pgDialect := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	builder := pgDialect.Select("event").
		From("events").
		Where(squirrel.Eq{"channel": d.channel}).
		OrderByClause("event->'timestamp' ASC").
		Offset(uint64(offset)).
		Limit(uint64(p.Size + 1))

	builder = eventFilterToQuery(filter, builder)

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, "", err
	}

	rows, err = d.tx.Query(d.ctx, query, args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var events []*asset.Event
	var count int

	for rows.Next() {
		event := new(asset.Event)

		err = rows.Scan(&event)
		if err != nil {
			return nil, "", err
		}

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
