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

func TestAddOrganization(t *testing.T) {
	mock, err := pgxmock.NewConn()
	require.NoError(t, err)

	organization := &asset.Organization{
		Id:           "1e8c1074-7fc4-4350-afcb-dc2d4849694c",
		CreationDate: timestamppb.New(time.Unix(900, 0)),
	}

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO organizations`).WithArgs(organization.Id, organization.Address, testChannel, organization.CreationDate.AsTime()).WillReturnResult(pgxmock.NewResult("INSERT", 1))

	tx, err := mock.Begin(context.Background())
	require.NoError(t, err)

	dbal := &DBAL{ctx: context.TODO(), tx: tx, channel: testChannel}

	err = dbal.AddOrganization(organization)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrganizationExists(t *testing.T) {
	mock, err := pgxmock.NewConn()
	require.NoError(t, err)

	organizationID := "45e80360-a9e5-11ec-b909-0242ac120002"

	mock.ExpectBegin()

	rows := pgxmock.NewRows([]string{"count"}).
		AddRow(1)
	mock.ExpectQuery(`SELECT COUNT`).WillReturnRows(rows)

	tx, err := mock.Begin(context.Background())
	require.NoError(t, err)

	dbal := &DBAL{ctx: context.TODO(), tx: tx, channel: testChannel}
	exists, err := dbal.OrganizationExists(organizationID)

	assert.NoError(t, err)
	assert.True(t, exists)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAllOrganizations(t *testing.T) {
	mock, err := pgxmock.NewConn()
	require.NoError(t, err)

	organization1 := &asset.Organization{
		Id:           "45e80360-a9e5-11ec-b909-0242ac120002",
		Address:      "substra-backend.org-1.com",
		CreationDate: timestamppb.New(time.Unix(800, 0)),
	}
	organization2 := &asset.Organization{
		Id:           "cb5ca026-a9ca-4bcf-9bdb-01711d5c6862",
		Address:      "substra-backend.org-2.com",
		CreationDate: timestamppb.New(time.Unix(900, 0)),
	}

	mock.ExpectBegin()

	rows := pgxmock.NewRows([]string{"id", "address", "creation_date"}).
		AddRow(organization1.Id, organization1.Address, organization1.CreationDate.AsTime()).
		AddRow(organization2.Id, organization2.Address, organization2.CreationDate.AsTime())
	mock.ExpectQuery(`SELECT id, address, creation_date FROM organizations`).WillReturnRows(rows)

	tx, err := mock.Begin(context.Background())
	require.NoError(t, err)

	dbal := &DBAL{ctx: context.TODO(), tx: tx, channel: testChannel}

	res, err := dbal.GetAllOrganizations()
	assert.NoError(t, err)
	assert.Len(t, res, 2)
	assert.Equal(t, organization1, res[0])
	assert.Equal(t, organization2, res[1])

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetOrganization(t *testing.T) {
	mock, err := pgxmock.NewConn()
	require.NoError(t, err)

	organization := &asset.Organization{
		Id:           "45e80360-a9e5-11ec-b909-0242ac120002",
		Address:      "substra-backend.org-1.com",
		CreationDate: timestamppb.New(time.Unix(800, 0)),
	}

	mock.ExpectBegin()

	rows := pgxmock.NewRows([]string{"id", "address", "creation_date"}).
		AddRow(organization.Id, organization.Address, organization.CreationDate.AsTime())
	mock.ExpectQuery(`SELECT id, address, creation_date FROM organizations`).WillReturnRows(rows)

	tx, err := mock.Begin(context.Background())
	require.NoError(t, err)

	dbal := &DBAL{ctx: context.TODO(), tx: tx, channel: testChannel}
	res, err := dbal.GetOrganization(organization.Id)

	assert.NoError(t, err)
	assert.Equal(t, res, organization)

	assert.NoError(t, mock.ExpectationsWereMet())
}
