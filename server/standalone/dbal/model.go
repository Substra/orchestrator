package dbal

import (
	"errors"
	"strconv"
	"time"

	"github.com/jackc/pgtype"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	orcerrors "github.com/owkin/orchestrator/lib/errors"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type sqlModel struct {
	Key            string
	Category       asset.ModelCategory
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
		Category:       m.Category,
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
		Select("key", "compute_task_key", "category", "address", "checksum", "permissions", "owner", "creation_date").
		From("expanded_models").
		Where(sq.Eq{"channel": d.channel, "key": key})

	row, err := d.queryRow(stmt)
	if err != nil {
		return nil, err
	}

	m := new(sqlModel)
	err = row.Scan(&m.Key, &m.ComputeTaskKey, &m.Category, &m.Address, &m.Checksum, &m.Permissions, &m.Owner, &m.CreationDate)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, orcerrors.NewNotFound("model", key)
		}
		return nil, err
	}

	return m.toModel(), nil
}

func (d *DBAL) QueryModels(c asset.ModelCategory, p *common.Pagination) ([]*asset.Model, common.PaginationToken, error) {
	offset, err := getOffset(p.Token)
	if err != nil {
		return nil, "", err
	}

	stmt := getStatementBuilder().
		Select("key", "compute_task_key", "category", "address", "checksum", "permissions", "owner", "creation_date").
		From("expanded_models").
		Where(sq.Eq{"channel": d.channel}).
		OrderByClause("creation_date ASC, key").
		Offset(uint64(offset)).
		// Fetch page size + 1 elements to determine whether there is a next page
		Limit(uint64(p.Size + 1))

	if c != asset.ModelCategory_MODEL_UNKNOWN {
		stmt = stmt.Where(sq.Eq{"category": c.String()})
	}

	rows, err := d.query(stmt)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var models []*asset.Model
	var count int

	for rows.Next() {
		m := new(sqlModel)

		err = rows.Scan(&m.Key, &m.ComputeTaskKey, &m.Category, &m.Address, &m.Checksum, &m.Permissions, &m.Owner, &m.CreationDate)
		if err != nil {
			return nil, "", err
		}

		models = append(models, m.toModel())
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

	return models, bookmark, nil
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
		Select("key", "compute_task_key", "category", "address", "checksum", "permissions", "owner", "creation_date").
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

		err = rows.Scan(&m.Key, &m.ComputeTaskKey, &m.Category, &m.Address, &m.Checksum, &m.Permissions, &m.Owner, &m.CreationDate)
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
	err := d.addAddressable(model.Address)
	if err != nil {
		return err
	}

	stmt := getStatementBuilder().
		Insert("models").
		Columns("key", "channel", "compute_task_key", "category", "address", "permissions", "owner", "creation_date").
		Values(model.Key, d.channel, model.ComputeTaskKey, model.Category, model.Address.StorageAddress, model.Permissions, model.Owner, model.CreationDate.AsTime())

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
		Set("category", model.Category).
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
		err = d.addAddressable(model.Address)
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
