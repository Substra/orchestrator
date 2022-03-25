package dbal

import (
	"fmt"
	"strconv"
	"time"

	sq "github.com/Masterminds/squirrel"
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

func (d *DBAL) QueryEvents(p *common.Pagination, filter *asset.EventQueryFilter, sortOrder asset.SortOrder) ([]*asset.Event, common.PaginationToken, error) {
	offset, err := getOffset(p.Token)
	if err != nil {
		return nil, "", err
	}

	order := PgSortAsc
	if sortOrder == asset.SortOrder_DESCENDING {
		order = PgSortDesc
	}
	orderBy := fmt.Sprintf("cast(event->>'timestamp' as timestamptz) %s, id %s", order, order)

	stmt := getStatementBuilder().Select("event").
		From("events").
		Where(sq.Eq{"channel": d.channel}).
		OrderByClause(orderBy).
		Offset(uint64(offset)).
		// Fetch page size + 1 elements to determine whether there is a next page
		Limit(uint64(p.Size + 1))

	stmt = eventFilterToQuery(filter, stmt)

	rows, err := d.query(stmt)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var events []*asset.Event
	var count int

	for rows.Next() {
		event := new(asset.Event)

		err = rows.Scan(event)
		if err != nil {
			return nil, "", err
		}
		event.Channel = d.channel

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
func eventFilterToQuery(filter *asset.EventQueryFilter, builder sq.SelectBuilder) sq.SelectBuilder {
	if filter == nil {
		return builder
	}

	if filter.AssetKey != "" {
		builder = builder.Where(sq.Eq{"asset_key": filter.AssetKey})
	}
	if filter.AssetKind != asset.AssetKind_ASSET_UNKNOWN {
		builder = builder.Where(sq.Eq{"event->>'assetKind'": filter.AssetKind.String()})
	}
	if filter.EventKind != asset.EventKind_EVENT_UNKNOWN {
		builder = builder.Where(sq.Eq{"event->>'eventKind'": filter.EventKind.String()})
	}
	if filter.Metadata != nil {
		builder = builder.Where(sq.Expr("event->'metadata' @> ?", filter.Metadata))
	}
	if filter.Start != nil {
		builder = builder.Where(sq.Expr("cast(event->>'timestamp' as timestamptz) >= cast(? as timestamptz)", filter.Start.AsTime().Format(time.RFC3339Nano)))
	}
	if filter.End != nil {
		builder = builder.Where(sq.Expr("cast(event->>'timestamp' as timestamptz) <= cast(? as timestamptz)", filter.End.AsTime().Format(time.RFC3339Nano)))
	}

	return builder
}
