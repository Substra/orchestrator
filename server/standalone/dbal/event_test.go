package dbal

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/jackc/pgconn"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	"github.com/owkin/orchestrator/utils"
	"github.com/pashagolub/pgxmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestEventFilterToQuery(t *testing.T) {
	cases := map[string]struct {
		filter        *asset.EventQueryFilter
		queryContains string
		params        []interface{}
	}{
		"empty":         {&asset.EventQueryFilter{}, "", nil},
		"single filter": {&asset.EventQueryFilter{AssetKey: "uuid"}, "asset_key = $1", []interface{}{"uuid"}},
		"two filter": {
			&asset.EventQueryFilter{AssetKind: asset.AssetKind_ASSET_COMPUTE_TASK, EventKind: asset.EventKind_EVENT_ASSET_CREATED},
			"asset_kind = $1 AND event_kind = $2",
			[]interface{}{asset.AssetKind_ASSET_COMPUTE_TASK.String(), asset.EventKind_EVENT_ASSET_CREATED.String()}},
		"three filter": {
			&asset.EventQueryFilter{AssetKey: "uuid", AssetKind: asset.AssetKind_ASSET_COMPUTE_PLAN, EventKind: asset.EventKind_EVENT_ASSET_UPDATED},
			"asset_key = $1 AND asset_kind = $2 AND event_kind = $3",
			[]interface{}{"uuid", asset.AssetKind_ASSET_COMPUTE_PLAN.String(), asset.EventKind_EVENT_ASSET_UPDATED.String()},
		},
		"time filter": {
			&asset.EventQueryFilter{Start: timestamppb.New(time.Unix(1337, 0)), End: timestamppb.New(time.Unix(7331, 0))},
			"timestamp >= $1 AND timestamp <= $2",
			[]interface{}{time.Unix(1337, 0).UTC(), time.Unix(7331, 0).UTC()},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			builder := getStatementBuilder().Select("id").From("events")
			builder = eventFilterToQuery(c.filter, builder)
			query, params, err := builder.ToSql()
			assert.NoError(t, err)
			assert.Contains(t, query, c.queryContains)
			assert.Equal(t, c.params, params)
		})
	}
}

func makeEventRows() *pgxmock.Rows {
	return pgxmock.NewRows([]string{"id", "asset_key", "asset_kind", "event_kind", "timestamp", "asset", "metadata"}).
		AddRow("id1", "13e88e4f-a287-4e8f-a96e-ea0c03f91e86", "ASSET_ALGO", "EVENT_ASSET_CREATED", time.Unix(1, 0).UTC(), []byte(`{}`), map[string]string{}).
		AddRow("id2", "7623fc2d-33fd-4b00-a6a0-65f5ec2eee20", "ASSET_MODEL", "EVENT_ASSET_UPDATED", time.Unix(2, 0).UTC(), []byte(`{}`), map[string]string{})
}

func TestEventQuery(t *testing.T) {
	mock, err := pgxmock.NewConn()
	assert.NoError(t, err)
	defer mock.Close(context.Background())

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT id, asset_key, asset_kind, event_kind, timestamp, asset, metadata FROM events .* ORDER BY timestamp ASC, id ASC`).
		WithArgs(testChannel).
		WillReturnRows(makeEventRows())

	tx, err := mock.Begin(context.Background())
	require.NoError(t, err)

	dbal := &DBAL{ctx: context.TODO(), tx: tx, channel: testChannel}

	res, _, err := dbal.QueryEvents(common.NewPagination("", 10), &asset.EventQueryFilter{}, asset.SortOrder_ASCENDING)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())

	for _, event := range res {
		assert.Equal(t, testChannel, event.Channel)
	}
}

func TestQueryEventsNilFilter(t *testing.T) {
	mock, err := pgxmock.NewConn()
	assert.NoError(t, err)
	defer mock.Close(context.Background())

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT id, asset_key, asset_kind, event_kind, timestamp, asset, metadata FROM events`).
		WithArgs(testChannel).
		WillReturnRows(makeEventRows())

	tx, err := mock.Begin(context.Background())
	require.NoError(t, err)

	dbal := &DBAL{ctx: context.TODO(), tx: tx, channel: testChannel}

	_, _, err = dbal.QueryEvents(common.NewPagination("", 10), nil, asset.SortOrder_ASCENDING)
	assert.NoError(t, err)
}

func TestGetEventPosition(t *testing.T) {
	eventID := "912aa0f8-ad56-4446-ac25-fe9b924561aa"
	eventPosition := int64(1234)

	conn, err := utils.NewMockConn()
	require.NoError(t, err)

	query := `SELECT position FROM events WHERE channel = $1 AND id = $2`
	conn.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(testChannel, eventID).
		WillReturnRows(pgxmock.NewRows([]string{"position"}).AddRow(eventPosition))

	dbal := &DBAL{ctx: context.TODO(), channel: testChannel, conn: conn}
	result, err := dbal.getEventPosition(eventID)
	assert.NoError(t, err)
	assert.Equal(t, eventPosition, result)

	conn.AssertExpectations(t)
}

func TestReplayBatchOfEvents(t *testing.T) {
	conn, err := utils.NewMockConn()
	require.NoError(t, err)

	startAfterPosition := int64(80)

	eventPosition := startAfterPosition + 1
	event := &asset.Event{
		Id:        "b2b30b36-b7f3-4839-9c6f-36ddafcf19fc",
		AssetKey:  "56a3dc56-f493-47e5-8a61-46a120e5403c",
		AssetKind: asset.AssetKind_ASSET_ALGO,
		EventKind: asset.EventKind_EVENT_ASSET_CREATED,
		Channel:   testChannel,
		Timestamp: timestamppb.New(time.Unix(15000, 0)),
		Asset:     &asset.Event_Algo{Algo: &asset.Algo{}},
		Metadata:  map[string]string{},
	}

	marshalledAsset, err := asset.MarshalEventAsset(event)
	require.NoError(t, err)

	rows := pgxmock.NewRows([]string{"position", "id", "asset_key", "asset_kind", "event_kind", "timestamp", "asset", "metadata"}).
		AddRow(eventPosition, event.Id, event.AssetKey, event.AssetKind, event.EventKind, event.Timestamp.AsTime(), marshalledAsset, event.Metadata)

	query := "SELECT position, id, asset_key, asset_kind, event_kind, timestamp, asset, metadata " +
		"FROM events " +
		"WHERE channel = $1 AND position > $2 " +
		"ORDER BY position " +
		fmt.Sprintf("LIMIT %d", replayEventsBatchSize+1)
	conn.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(testChannel, startAfterPosition).
		WillReturnRows(rows)

	stream := new(asset.MockEventService_SubscribeToEventsServer)
	matchEventID := mock.MatchedBy(func(e *asset.Event) bool {
		return proto.Equal(e, event)
	})
	stream.On("Send", matchEventID).Return(nil)

	dbal := &DBAL{ctx: context.TODO(), conn: conn, channel: testChannel}
	lastProcessedPos, hasNextBatch, err := dbal.replayBatchOfEvents(startAfterPosition, stream)
	assert.NoError(t, err)
	assert.False(t, hasNextBatch)
	assert.Equal(t, eventPosition, lastProcessedPos)

	stream.AssertExpectations(t)
	conn.AssertExpectations(t)
}

func getPgNotificationFrom(n *eventNotification) (*pgconn.Notification, error) {
	marshalledPayload, err := json.Marshal(n)
	if err != nil {
		return nil, err
	}

	return &pgconn.Notification{
		PID:     123,
		Channel: "events",
		Payload: string(marshalledPayload),
	}, nil
}

func makeEventRowsFrom(e *asset.Event) (*pgxmock.Rows, error) {
	marshalledAsset, err := asset.MarshalEventAsset(e)
	if err != nil {
		return nil, err
	}

	rows := pgxmock.NewRows([]string{"id", "asset_key", "asset_kind", "event_kind", "timestamp", "asset", "metadata"}).
		AddRow(e.Id, e.AssetKey, e.AssetKind, e.EventKind, e.Timestamp.AsTime(), marshalledAsset, e.Metadata)
	return rows, nil
}

func TestForwardEventNotification(t *testing.T) {
	ctx := context.TODO()

	conn, err := utils.NewMockConn()
	require.NoError(t, err)

	notif := &eventNotification{
		EventPosition: 50,
		Channel:       testChannel,
	}
	event := &asset.Event{
		Id:        "b2b30b36-b7f3-4839-9c6f-36ddafcf19fc",
		AssetKey:  "56a3dc56-f493-47e5-8a61-46a120e5403c",
		AssetKind: asset.AssetKind_ASSET_ALGO,
		EventKind: asset.EventKind_EVENT_ASSET_CREATED,
		Channel:   testChannel,
		Timestamp: timestamppb.New(time.Unix(15000, 0)),
		Asset:     &asset.Event_Algo{Algo: &asset.Algo{}},
		Metadata:  map[string]string{},
	}

	pgNotif, err := getPgNotificationFrom(notif)
	require.NoError(t, err)

	conn.On("WaitForNotification", ctx).Return(pgNotif, nil)

	rows, err := makeEventRowsFrom(event)
	require.NoError(t, err)

	conn.ExpectQuery("SELECT .* FROM events").
		WithArgs(testChannel, notif.EventPosition).
		WillReturnRows(rows)

	stream := new(asset.MockEventService_SubscribeToEventsServer)
	matchEvent := mock.MatchedBy(func(e *asset.Event) bool {
		return proto.Equal(e, event)
	})
	stream.On("Send", matchEvent).Return(nil)

	dbal := &DBAL{ctx: ctx, conn: conn, channel: testChannel}
	lastProcessedPos, err := dbal.forwardEventNotification(1, stream)
	assert.NoError(t, err)
	assert.Equal(t, notif.EventPosition, lastProcessedPos)

	conn.AssertExpectations(t)
	stream.AssertExpectations(t)
}

func TestForwardEventNotificationIgnoreEvent(t *testing.T) {
	cases := map[string]struct {
		initialLastProcessedPos int64
		notif                   *eventNotification
	}{
		"event not in channel": {
			notif: &eventNotification{Channel: "not test channel"},
		},
		"event already processed": {
			initialLastProcessedPos: 6,
			notif:                   &eventNotification{EventPosition: 5, Channel: testChannel},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			ctx := context.TODO()

			conn, err := utils.NewMockConn()
			require.NoError(t, err)

			pgNotif, err := getPgNotificationFrom(c.notif)
			require.NoError(t, err)
			conn.On("WaitForNotification", ctx).Return(pgNotif, nil)

			dbal := &DBAL{ctx: ctx, conn: conn, channel: testChannel}
			lastProcessedPos, err := dbal.forwardEventNotification(c.initialLastProcessedPos, nil)
			assert.NoError(t, err)
			assert.Equal(t, c.initialLastProcessedPos, lastProcessedPos)

			conn.AssertExpectations(t)
		})
	}
}

func TestWaitForEventNotification(t *testing.T) {
	ctx := context.TODO()

	conn, err := utils.NewMockConn()
	require.NoError(t, err)

	notif := &eventNotification{
		EventPosition: 5,
		Channel:       testChannel,
	}

	pgNotif, err := getPgNotificationFrom(notif)
	require.NoError(t, err)
	conn.On("WaitForNotification", ctx).Return(pgNotif, nil)

	dbal := &DBAL{ctx: ctx, conn: conn, channel: testChannel}
	received, err := dbal.waitForEventNotification()
	assert.NoError(t, err)
	assert.Equal(t, received, notif)

	conn.AssertExpectations(t)
}

func TestGetEventByPosition(t *testing.T) {
	conn, err := utils.NewMockConn()
	require.NoError(t, err)

	position := int64(9)
	event := &asset.Event{
		Id:        "b2b30b36-b7f3-4839-9c6f-36ddafcf19fc",
		AssetKey:  "56a3dc56-f493-47e5-8a61-46a120e5403c",
		AssetKind: asset.AssetKind_ASSET_ALGO,
		EventKind: asset.EventKind_EVENT_ASSET_CREATED,
		Channel:   testChannel,
		Timestamp: timestamppb.New(time.Unix(15000, 0)),
		Asset:     &asset.Event_Algo{Algo: &asset.Algo{}},
		Metadata:  map[string]string{},
	}

	query := `SELECT id, asset_key, asset_kind, event_kind, timestamp, asset, metadata FROM events WHERE channel = $1 AND position = $2`
	rows, err := makeEventRowsFrom(event)
	require.NoError(t, err)

	conn.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(testChannel, position).WillReturnRows(rows)

	dbal := &DBAL{ctx: context.TODO(), channel: testChannel, conn: conn}
	retrieved, err := dbal.getEventByPosition(position)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(event, retrieved))

	conn.AssertExpectations(t)
}
