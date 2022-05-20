package dbal

import (
	"context"
	"testing"
	"time"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	"github.com/pashagolub/pgxmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
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

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

	for _, event := range res {
		assert.Equal(t, testChannel, event.Channel)
	}
}

func TestQueryEventsNilFilter(t *testing.T) {
	mock, err := pgxmock.NewConn()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
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
