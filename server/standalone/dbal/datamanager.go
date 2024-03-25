package dbal

import (
	"errors"
	"strconv"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4"
	"github.com/substra/orchestrator/lib/asset"
	"github.com/substra/orchestrator/lib/common"
	orcerrors "github.com/substra/orchestrator/lib/errors"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type sqlDataManager struct {
	Key            string
	Name           string
	Owner          string
	Permissions    asset.Permissions
	Description    asset.Addressable
	Opener         asset.Addressable
	CreationDate   time.Time
	LogsPermission asset.Permission
	Metadata       map[string]string
}

func (dm *sqlDataManager) toDataManager() *asset.DataManager {
	return &asset.DataManager{
		Key:            dm.Key,
		Name:           dm.Name,
		Owner:          dm.Owner,
		Permissions:    &dm.Permissions,
		Description:    &dm.Description,
		Opener:         &dm.Opener,
		CreationDate:   timestamppb.New(dm.CreationDate),
		LogsPermission: &dm.LogsPermission,
		Metadata:       dm.Metadata,
	}
}

// AddDataManager implements persistence.DataManagerDBAL
func (d *DBAL) AddDataManager(datamanager *asset.DataManager) error {
	err := d.addAddressable(datamanager.Description, false)
	if err != nil {
		return err
	}

	err = d.addAddressable(datamanager.Opener, false)
	if err != nil {
		return err
	}

	stmt := getStatementBuilder().
		Insert("datamanagers").
		Columns("key", "channel", "name", "owner", "permissions", "description", "opener", "creation_date", "logs_permission", "metadata").
		Values(datamanager.Key, d.channel, datamanager.Name, datamanager.Owner, datamanager.Permissions, datamanager.Description.StorageAddress, datamanager.Opener.StorageAddress, datamanager.CreationDate.AsTime(), datamanager.LogsPermission, datamanager.Metadata)

	return d.exec(stmt)
}

// DataManagerExists implements persistence.DataManagerDBAL
func (d *DBAL) DataManagerExists(key string) (bool, error) {
	stmt := getStatementBuilder().
		Select("COUNT(key)").
		From("datamanagers").
		Where(sq.Eq{"channel": d.channel, "key": key})

	row, err := d.queryRow(stmt)
	if err != nil {
		return false, err
	}

	var count int
	err = row.Scan(&count)

	return count == 1, err
}

// GetDataManager implements persistence.DataManagerDBAL
func (d *DBAL) GetDataManager(key string) (*asset.DataManager, error) {
	stmt := getStatementBuilder().
		Select("key", "name", "owner", "permissions", "description_address", "description_checksum", "opener_address", "opener_checksum", "creation_date", "logs_permission", "metadata").
		From("expanded_datamanagers").
		Where(sq.Eq{"channel": d.channel, "key": key})

	row, err := d.queryRow(stmt)
	if err != nil {
		return nil, err
	}

	dm := new(sqlDataManager)
	err = row.Scan(&dm.Key, &dm.Name, &dm.Owner, &dm.Permissions, &dm.Description.StorageAddress, &dm.Description.Checksum, &dm.Opener.StorageAddress, &dm.Opener.Checksum, &dm.CreationDate, &dm.LogsPermission, &dm.Metadata)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, orcerrors.NewNotFound("datamanager", key)
		}
		return nil, err
	}

	return dm.toDataManager(), nil
}

// QueryDataManagers implements persistence.DataManagerDBAL
func (d *DBAL) QueryDataManagers(p *common.Pagination) ([]*asset.DataManager, common.PaginationToken, error) {
	offset, err := getOffset(p.Token)
	if err != nil {
		return nil, "", err
	}

	stmt := getStatementBuilder().
		Select("key", "name", "owner", "permissions", "description_address", "description_checksum", "opener_address", "opener_checksum", "creation_date", "logs_permission", "metadata").
		From("expanded_datamanagers").
		Where(sq.Eq{"channel": d.channel}).
		OrderByClause("creation_date ASC, key").
		Offset(uint64(offset)).
		// Fetch page size + 1 elements to determine whether there is a next page
		Limit(uint64(p.Size + 1))

	rows, err := d.query(stmt)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var datamanagers []*asset.DataManager
	var count int

	for rows.Next() {
		dm := new(sqlDataManager)

		err = rows.Scan(&dm.Key, &dm.Name, &dm.Owner, &dm.Permissions, &dm.Description.StorageAddress, &dm.Description.Checksum, &dm.Opener.StorageAddress, &dm.Opener.Checksum, &dm.CreationDate, &dm.LogsPermission, &dm.Metadata)
		if err != nil {
			return nil, "", err
		}

		datamanagers = append(datamanagers, dm.toDataManager())
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
		bookmark = strconv.Itoa(offset + count)
	}

	return datamanagers, bookmark, nil
}

// UpdateDataManager updates the mutable fields of a data manager in the DB. List of mutable fields: name.
func (d *DBAL) UpdateDataManager(datamanager *asset.DataManager) error {
	stmt := getStatementBuilder().
		Update("datamanagers").
		Set("name", datamanager.Name).
		Where(sq.Eq{"channel": d.channel, "key": datamanager.Key})

	return d.exec(stmt)
}
