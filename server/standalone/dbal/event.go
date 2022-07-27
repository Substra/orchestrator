package dbal

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/go-playground/log/v7"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	commonserv "github.com/owkin/orchestrator/server/common"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var replayEventsBatchSize = commonserv.MustParseInt(
	commonserv.GetEnvOrFallback("REPLAY_EVENTS_BATCH_SIZE", "100"),
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

	_, err := d.tx.Exec(d.ctx, "LOCK TABLE events IN SHARE ROW EXCLUSIVE MODE")
	if err != nil {
		return err
	}

	stmt := getStatementBuilder().Select("COALESCE(MAX(position), 0) + 1").From("events")
	row, err := d.queryRow(stmt)
	if err != nil {
		return err
	}

	var position uint64
	err = row.Scan(&position)
	if err != nil {
		return err
	}

	// Relying on COPY FROM directive is faster for a large number of items.
	_, err = d.tx.CopyFrom(
		d.ctx,
		pgx.Identifier{"events"},
		[]string{"id", "asset_key", "asset_kind", "event_kind", "channel", "timestamp", "asset", "metadata", "position"},
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

			value := []interface{}{
				id,
				event.AssetKey,
				event.AssetKind.String(),
				event.EventKind.String(),
				d.channel,
				event.Timestamp.AsTime(),
				eventAsset,
				event.Metadata,
				position,
			}

			position++
			return value, nil
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

// SubscribeToEvents replays already existing events starting from startEventID (excluded),
// then it waits and forward newly created events.
func (d *DBAL) SubscribeToEvents(startEventID string, stream asset.EventService_SubscribeToEventsServer) error {
	// Start listening to event notifications before fetching already existing events
	// from the database to prevent missing any event.
	err := d.startListeningToEventNotifications()
	if err != nil {
		return err
	}

	// lastProcessedPos stores the position of the last processed event
	lastProcessedPos := int64(0)

	if startEventID != "" {
		lastProcessedPos, err = d.getEventPosition(startEventID)
		if err != nil {
			return err
		}
	}

	hasNextBatch := true
	for hasNextBatch {
		lastProcessedPos, hasNextBatch, err = d.replayBatchOfEvents(lastProcessedPos, stream)
		if err != nil {
			return err
		}
	}

	for {
		if err = d.ctx.Err(); err != nil {
			return err
		}

		lastProcessedPos, err = d.forwardEventNotification(lastProcessedPos, stream)
		if err != nil {
			return err
		}
	}
}

func (d *DBAL) getEventPosition(eventID string) (int64, error) {
	stmt := getStatementBuilder().
		Select("position").
		From("events").
		Where(sq.Eq{"id": eventID, "channel": d.channel})

	query, args, err := stmt.ToSql()
	if err != nil {
		return 0, err
	}

	row := d.conn.QueryRow(context.Background(), query, args...)

	var position int64
	err = row.Scan(&position)

	return position, err
}

// replayBatchOfEvents fetches a batch of already existing events from the database and send them in the provided stream.
// Events are replayed based on position order, starting right after startAfterPosition.
func (d *DBAL) replayBatchOfEvents(startAfterPosition int64, stream asset.EventService_SubscribeToEventsServer) (lastProcessedPos int64, hasNextBatch bool, err error) {
	stmt := getStatementBuilder().
		Select("position", "id", "asset_key", "asset_kind", "event_kind", "timestamp", "asset", "metadata").
		From("events").
		Where(sq.Eq{"channel": d.channel}).
		Where(sq.Gt{"position": startAfterPosition}).
		OrderBy("position").
		// Fetch replayEventsBatchSize size + 1 elements to determine whether there is a next batch to fetch
		Limit(uint64(replayEventsBatchSize + 1))

	query, args, err := stmt.ToSql()
	if err != nil {
		return 0, false, err
	}

	rows, err := d.conn.Query(context.Background(), query, args...)
	if err != nil {
		return 0, false, err
	}
	defer rows.Close()

	count := 0

	for rows.Next() {
		var position int64
		ev := sqlEvent{Channel: d.channel}

		err = rows.Scan(&position, &ev.ID, &ev.AssetKey, &ev.AssetKind, &ev.EventKind, &ev.Timestamp, &ev.Asset, &ev.Metadata)
		if err != nil {
			return 0, false, err
		}

		event, err := ev.toEvent()
		if err != nil {
			return 0, false, err
		}

		err = stream.Send(event)
		if err != nil {
			return 0, false, err
		}

		lastProcessedPos = position
		count++

		if count == replayEventsBatchSize {
			break
		}
	}

	if err := rows.Err(); err != nil {
		return 0, false, err
	}

	hasNextBatch = count == replayEventsBatchSize && rows.Next()
	return lastProcessedPos, hasNextBatch, err
}

// startListeningToEventNotifications starts listening to PostgreSQL notifications
func (d *DBAL) startListeningToEventNotifications() error {
	_, err := d.conn.Exec(context.Background(), "LISTEN events")
	return err
}

// forwardEventNotification waits for the reception of a notification indicating a new event,
// and then sends the corresponding event into the provided stream.
func (d *DBAL) forwardEventNotification(lastProcessedPos int64, stream asset.EventService_SubscribeToEventsServer) (int64, error) {
	notif, err := d.waitForEventNotification()
	if err != nil {
		return lastProcessedPos, err
	}

	if notif.Channel != d.channel {
		return lastProcessedPos, nil
	}

	// since events are inserted with a strictly increasing position value,
	// this ensures that an already forwarded event cannot be sent again
	if notif.EventPosition <= lastProcessedPos {
		return lastProcessedPos, nil
	}

	event, err := d.getEventByPosition(notif.EventPosition)
	if err != nil {
		return lastProcessedPos, err
	}

	err = stream.Send(event)
	if err != nil {
		return lastProcessedPos, err
	}

	lastProcessedPos = notif.EventPosition

	return lastProcessedPos, nil
}

// eventNotification is sent with PostgreSQL NOTIFY when an event is inserted in the events table
type eventNotification struct {
	EventPosition int64  `json:"event_position"`
	Channel       string `json:"channel"`
}

// waitForEventNotification returns an *eventNotification upon reception
// of a PostgreSQL notification.
func (d *DBAL) waitForEventNotification() (*eventNotification, error) {
	notif, err := d.conn.WaitForNotification(d.ctx)
	if err != nil {
		return nil, err
	}

	eventNotif := new(eventNotification)
	err = json.Unmarshal([]byte(notif.Payload), eventNotif)
	return eventNotif, err
}

func (d *DBAL) getEventByPosition(position int64) (*asset.Event, error) {
	stmt := getStatementBuilder().
		Select("id", "asset_key", "asset_kind", "event_kind", "timestamp", "asset", "metadata").
		From("events").
		Where(sq.Eq{"position": position, "channel": d.channel})

	query, args, err := stmt.ToSql()
	if err != nil {
		return nil, err
	}

	row := d.conn.QueryRow(context.Background(), query, args...)

	ev := sqlEvent{Channel: d.channel}
	err = row.Scan(&ev.ID, &ev.AssetKey, &ev.AssetKind, &ev.EventKind, &ev.Timestamp, &ev.Asset, &ev.Metadata)
	if err != nil {
		return nil, err
	}

	return ev.toEvent()
}
