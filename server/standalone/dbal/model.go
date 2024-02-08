package dbal

import (
	"errors"
	"time"

	"github.com/jackc/pgtype"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4"
	"github.com/substra/orchestrator/lib/asset"
	orcerrors "github.com/substra/orchestrator/lib/errors"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type sqlModel struct {
	Key            string
	ComputeTaskKey string
	Address        pgtype.Text
	Checksum       pgtype.Text
	Permissions    asset.Permissions
	Owner          string
	CreationDate   time.Time
}

func (m *sqlModel) toModel() *asset.Model {
	model := &asset.Model{
		Key:            m.Key,
		ComputeTaskKey: m.ComputeTaskKey,
		Permissions:    &m.Permissions,
		Owner:          m.Owner,
		CreationDate:   timestamppb.New(m.CreationDate),
	}

	if m.Address.Status == pgtype.Present {
		model.Address = &asset.Addressable{
			Checksum:       m.Checksum.String,
			StorageAddress: m.Address.String,
		}
	}

	return model
}

func (d *DBAL) GetModel(key string) (*asset.Model, error) {
	stmt := getStatementBuilder().
		Select("key", "compute_task_key", "address", "checksum", "permissions", "owner", "creation_date").
		From("expanded_models").
		Where(sq.Eq{"channel": d.channel, "key": key})

	row, err := d.queryRow(stmt)
	if err != nil {
		return nil, err
	}

	m := new(sqlModel)
	err = row.Scan(&m.Key, &m.ComputeTaskKey, &m.Address, &m.Checksum, &m.Permissions, &m.Owner, &m.CreationDate)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, orcerrors.NewNotFound("model", key)
		}
		return nil, err
	}

	return m.toModel(), nil
}

func (d *DBAL) ModelExists(key string) (bool, error) {
	stmt := getStatementBuilder().
		Select("COUNT(key)").
		From("models").
		Where(sq.Eq{"channel": d.channel, "key": key})

	row, err := d.queryRow(stmt)
	if err != nil {
		return false, err
	}

	var count int
	err = row.Scan(&count)

	return count == 1, err
}

func (d *DBAL) GetComputeTaskOutputModels(key string) ([]*asset.Model, error) {
	stmt := getStatementBuilder().
		Select("key", "compute_task_key", "address", "checksum", "permissions", "owner", "creation_date").
		From("expanded_models").
		Where(sq.Eq{"channel": d.channel, "compute_task_key": key})

	rows, err := d.query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	models := []*asset.Model{}
	for rows.Next() {
		m := new(sqlModel)

		err = rows.Scan(&m.Key, &m.ComputeTaskKey, &m.Address, &m.Checksum, &m.Permissions, &m.Owner, &m.CreationDate)
		if err != nil {
			return nil, err
		}
		models = append(models, m.toModel())
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return models, nil
}

func (d *DBAL) AddModel(model *asset.Model, identifier string) error {
	err := d.addAddressable(model.Address, false)
	if err != nil {
		return err
	}

	stmt := getStatementBuilder().
		Insert("models").
		Columns("key", "channel", "compute_task_key", "address", "permissions", "owner", "creation_date").
		Values(model.Key, d.channel, model.ComputeTaskKey, model.Address.StorageAddress, model.Permissions, model.Owner, model.CreationDate.AsTime())

	return d.exec(stmt)
}

func (d *DBAL) UpdateModel(model *asset.Model) error {
	selectAddressStmt := getStatementBuilder().
		Select("address").
		From("models").
		Where(sq.Eq{"channel": d.channel, "key": model.Key})

	row, err := d.queryRow(selectAddressStmt)
	if err != nil {
		return err
	}

	var previousAddress string
	err = row.Scan(&previousAddress)
	if err != nil {
		return err
	}

	updateStmt := getStatementBuilder().
		Update("models").
		Set("compute_task_key", model.ComputeTaskKey).
		Set("address", nil).
		Set("permissions", model.Permissions).
		Set("owner", model.Owner).
		Where(sq.Eq{"channel": d.channel, "key": model.Key})

	err = d.exec(updateStmt)
	if err != nil {
		return err
	}

	err = d.deleteAddressable(previousAddress)
	if err != nil {
		return err
	}

	if model.Address != nil {
		err = d.addAddressable(model.Address, false)
		if err != nil {
			return err
		}

		updateAddressStmt := getStatementBuilder().
			Update("models").
			Set("address", model.Address.StorageAddress).
			Where(sq.Eq{"channel": d.channel, "key": model.Key})

		return d.exec(updateAddressStmt)
	}

	return nil
}
