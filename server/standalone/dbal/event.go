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
	"google.golang.org/protobuf/types/known/timestamppb"
)

type sqlEvent struct {
	ID        string
	AssetKey  string
	AssetKind asset.AssetKind
	EventKind asset.EventKind
	Channel   string
	Timestamp time.Time
	Metadata  map[string]string
	Asset     []byte
}

func (e *sqlEvent) toEvent() (*asset.Event, error) {
	event := &asset.Event{
		Id:        e.ID,
		AssetKey:  e.AssetKey,
		AssetKind: e.AssetKind,
		EventKind: e.EventKind,
		Channel:   e.Channel,
		Timestamp: timestamppb.New(e.Timestamp),
		Metadata:  e.Metadata,
	}

	err := asset.UnmarshalEventAsset(e.Asset, event, event.AssetKind)
	if err != nil {
		return nil, err
	}

	return event, nil
}

// AddEvents insert events in storage in batch mode.
func (d *DBAL) AddEvents(events ...*asset.Event) error {
	log.WithField("numEvents", len(events)).Debug("dbal: adding multiple events in batch mode")

	// Relying on COPY FROM directive is faster for a large number of items.
	_, err := d.tx.CopyFrom(
		d.ctx,
		pgx.Identifier{"events"},
		[]string{"id", "asset_key", "asset_kind", "event_kind", "channel", "timestamp", "asset", "metadata"},
		pgx.CopyFromSlice(len(events), func(i int) ([]interface{}, error) {
			event := events[i]

			// expect binary representation, not string
			id, err := uuid.Parse(event.Id)
			if err != nil {
				return nil, err
			}

			eventAsset, err := asset.MarshalEventAsset(event)
			if err != nil {
				return nil, err
			}

			return []interface{}{
				id,
				event.AssetKey,
				event.AssetKind.String(),
				event.EventKind.String(),
				d.channel,
				event.Timestamp.AsTime(),
				eventAsset,
				event.Metadata,
			}, nil
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
	orderBy := fmt.Sprintf("timestamp %s, id %s", order, order)

	stmt := getStatementBuilder().
		Select("id", "asset_key", "asset_kind", "event_kind", "timestamp", "asset", "metadata").
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
		ev := sqlEvent{Channel: d.channel}

		err = rows.Scan(&ev.ID, &ev.AssetKey, &ev.AssetKind, &ev.EventKind, &ev.Timestamp, &ev.Asset, &ev.Metadata)
		if err != nil {
			return nil, "", err
		}

		event, err := ev.toEvent()
		if err != nil {
			return nil, "", err
		}

		events = append(events, event)
		count++

		if count == int(p.Size) {
			break
		}
	}
	if err = rows.Err(); err != nil {
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
		builder = builder.Where(sq.Eq{"asset_kind": filter.AssetKind.String()})
	}
	if filter.EventKind != asset.EventKind_EVENT_UNKNOWN {
		builder = builder.Where(sq.Eq{"event_kind": filter.EventKind.String()})
	}
	if filter.Metadata != nil {
		builder = builder.Where(sq.Expr("metadata @> ?", filter.Metadata))
	}
	if filter.Start != nil {
		builder = builder.Where(sq.GtOrEq{"timestamp": filter.Start.AsTime()})
	}
	if filter.End != nil {
		builder = builder.Where(sq.LtOrEq{"timestamp": filter.End.AsTime()})
	}

	return builder
}
