package dbal

import (
	"context"
	"testing"
	"time"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/pashagolub/pgxmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestAddNode(t *testing.T) {
	mock, err := pgxmock.NewConn()
	require.NoError(t, err)
	defer mock.Close(context.Background())

	node := &asset.Node{
		Id:           "1e8c1074-7fc4-4350-afcb-dc2d4849694c",
		CreationDate: timestamppb.New(time.Unix(900, 0)),
	}

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO nodes`).WithArgs(node.Id, testChannel, node.CreationDate.AsTime()).WillReturnResult(pgxmock.NewResult("INSERT", 1))

	tx, err := mock.Begin(context.Background())
	require.NoError(t, err)

	dbal := &DBAL{ctx: context.TODO(), tx: tx, channel: testChannel}

	err = dbal.AddNode(node)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNodeExists(t *testing.T) {
	mock, err := pgxmock.NewConn()
	require.NoError(t, err)
	defer mock.Close(context.Background())

	nodeID := "45e80360-a9e5-11ec-b909-0242ac120002"

	mock.ExpectBegin()

	rows := pgxmock.NewRows([]string{"count"}).
		AddRow(1)
	mock.ExpectQuery(`SELECT COUNT`).WillReturnRows(rows)

	tx, err := mock.Begin(context.Background())
	require.NoError(t, err)

	dbal := &DBAL{ctx: context.TODO(), tx: tx, channel: testChannel}
	exists, err := dbal.NodeExists(nodeID)

	assert.NoError(t, err)
	assert.True(t, exists)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAllNodes(t *testing.T) {
	mock, err := pgxmock.NewConn()
	require.NoError(t, err)
	defer mock.Close(context.Background())

	node1 := &asset.Node{
		Id:           "45e80360-a9e5-11ec-b909-0242ac120002",
		CreationDate: timestamppb.New(time.Unix(800, 0)),
	}
	node2 := &asset.Node{
		Id:           "cb5ca026-a9ca-4bcf-9bdb-01711d5c6862",
		CreationDate: timestamppb.New(time.Unix(900, 0)),
	}

	mock.ExpectBegin()

	rows := pgxmock.NewRows([]string{"id", "creation_date"}).
		AddRow(node1.Id, node1.CreationDate.AsTime()).
		AddRow(node2.Id, node2.CreationDate.AsTime())
	mock.ExpectQuery(`SELECT id, creation_date FROM nodes`).WillReturnRows(rows)

	tx, err := mock.Begin(context.Background())
	require.NoError(t, err)

	dbal := &DBAL{ctx: context.TODO(), tx: tx, channel: testChannel}

	res, err := dbal.GetAllNodes()
	assert.NoError(t, err)
	assert.Len(t, res, 2)
	assert.Equal(t, node1, res[0])
	assert.Equal(t, node2, res[1])

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetNode(t *testing.T) {
	mock, err := pgxmock.NewConn()
	require.NoError(t, err)
	defer mock.Close(context.Background())

	node := &asset.Node{
		Id:           "45e80360-a9e5-11ec-b909-0242ac120002",
		CreationDate: timestamppb.New(time.Unix(800, 0)),
	}

	mock.ExpectBegin()

	rows := pgxmock.NewRows([]string{"id", "creation_date"}).
		AddRow(node.Id, node.CreationDate.AsTime())
	mock.ExpectQuery(`SELECT id, creation_date FROM nodes`).WillReturnRows(rows)

	tx, err := mock.Begin(context.Background())
	require.NoError(t, err)

	dbal := &DBAL{ctx: context.TODO(), tx: tx, channel: testChannel}
	res, err := dbal.GetNode(node.Id)

	assert.NoError(t, err)
	assert.Equal(t, res, node)

	assert.NoError(t, mock.ExpectationsWereMet())
}
