package dbal

import (
	"errors"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4"
	"github.com/owkin/orchestrator/lib/asset"
	orcerrors "github.com/owkin/orchestrator/lib/errors"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type sqlOrganization struct {
	ID           string
	Address      string
	CreationDate time.Time
}

func (s *sqlOrganization) toOrganization() *asset.Organization {
	return &asset.Organization{
		Id:           s.ID,
		Address:      s.Address,
		CreationDate: timestamppb.New(s.CreationDate),
	}
}

// AddOrganization implements persistence.OrganizationDBAL
func (d *DBAL) AddOrganization(organization *asset.Organization) error {
	stmt := getStatementBuilder().
		Insert("organizations").
		Columns("id", "address", "channel", "creation_date").
		Values(organization.GetId(), organization.GetAddress(), d.channel, organization.GetCreationDate().AsTime())

	return d.exec(stmt)
}

// OrganizationExists implements persistence.OrganizationDBAL
func (d *DBAL) OrganizationExists(id string) (bool, error) {
	stmt := getStatementBuilder().
		Select("COUNT(id)").
		From("organizations").
		Where(sq.Eq{"id": id, "channel": d.channel})

	row, err := d.queryRow(stmt)
	if err != nil {
		return false, err
	}

	var count int
	err = row.Scan(&count)

	return count == 1, err
}

// GetAllOrganizations implements persistence.OrganizationDBAL
func (d *DBAL) GetAllOrganizations() ([]*asset.Organization, error) {
	stmt := getStatementBuilder().
		Select("id", "address", "creation_date").
		From("organizations").
		Where(sq.Eq{"channel": d.channel})

	rows, err := d.query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var organizations []*asset.Organization

	for rows.Next() {
		scanned := sqlOrganization{}

		err = rows.Scan(&scanned.ID, &scanned.Address, &scanned.CreationDate)
		if err != nil {
			return nil, err
		}

		organizations = append(organizations, scanned.toOrganization())
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return organizations, nil
}

// GetOrganization implements persistence.OrganizationDBAL
func (d *DBAL) GetOrganization(id string) (*asset.Organization, error) {
	stmt := getStatementBuilder().
		Select("id", "address", "creation_date").
		From("organizations").
		Where(sq.Eq{"id": id, "channel": d.channel})

	row, err := d.queryRow(stmt)
	if err != nil {
		return nil, err
	}

	scanned := sqlOrganization{}
	err = row.Scan(&scanned.ID, &scanned.Address, &scanned.CreationDate)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, orcerrors.NewNotFound("organization", id)
		}
		return nil, err
	}

	return scanned.toOrganization(), nil
}
