package dbal

import (
	"context"
	"testing"
	"time"

	"github.com/Masterminds/squirrel"
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
			"event->>'assetKind' = $1 AND event->>'eventKind' = $2",
			[]interface{}{asset.AssetKind_ASSET_COMPUTE_TASK.String(), asset.EventKind_EVENT_ASSET_CREATED.String()}},
		"three filter": {
			&asset.EventQueryFilter{AssetKey: "uuid", AssetKind: asset.AssetKind_ASSET_COMPUTE_PLAN, EventKind: asset.EventKind_EVENT_ASSET_UPDATED},
			"asset_key = $1 AND event->>'assetKind' = $2 AND event->>'eventKind' = $3",
			[]interface{}{"uuid", asset.AssetKind_ASSET_COMPUTE_PLAN.String(), asset.EventKind_EVENT_ASSET_UPDATED.String()},
		},
		"time filter": {
			&asset.EventQueryFilter{Start: timestamppb.New(time.Unix(1337, 0)), End: timestamppb.New(time.Unix(7331, 0))},
			"cast(event->>'timestamp' as timestamptz) >= cast($1 as timestamptz) AND cast(event->>'timestamp' as timestamptz) <= cast($2 as timestamptz)",
			[]interface{}{time.Unix(1337, 0).UTC().Format(time.RFC3339Nano), time.Unix(7331, 0).UTC().Format(time.RFC3339Nano)},
		},
	}

	pgDialect := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			builder := pgDialect.Select("event").From("events")
			builder = eventFilterToQuery(c.filter, builder)
			query, params, err := builder.ToSql()
			assert.NoError(t, err)
			assert.Contains(t, query, c.queryContains)
			assert.Equal(t, c.params, params)
		})
	}
}

func TestEventQuery(t *testing.T) {
	mock, err := pgxmock.NewConn()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mock.Close(context.Background())

	mock.ExpectBegin()

	rows := pgxmock.NewRows([]string{"asset"}).
		AddRow([]byte("{}")).
		AddRow([]byte("{}"))

	mock.ExpectQuery(`SELECT event FROM events .* ORDER BY cast\(event->>'timestamp' as timestamptz\) ASC, id ASC`).WithArgs(testChannel).WillReturnRows(rows)

	tx, err := mock.Begin(context.Background())
	require.NoError(t, err)

	dbal := &DBAL{ctx: context.TODO(), tx: tx, channel: testChannel}

	res, _, err := dbal.QueryEvents(common.NewPagination("", 10), &asset.EventQueryFilter{}, asset.SortOrder_ASCENDING)
	assert.NoError(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
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

	rows := pgxmock.NewRows([]string{"asset"})

	mock.ExpectQuery(`SELECT event FROM events`).WithArgs(testChannel).WillReturnRows(rows)

	tx, err := mock.Begin(context.Background())
	require.NoError(t, err)

	dbal := &DBAL{ctx: context.TODO(), tx: tx, channel: testChannel}

	_, _, err = dbal.QueryEvents(common.NewPagination("", 10), nil, asset.SortOrder_ASCENDING)
	assert.NoError(t, err)
}
