package dbal

import (
	"errors"
	"strconv"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	orcerrors "github.com/owkin/orchestrator/lib/errors"
)

func (d *DBAL) GetModel(key string) (*asset.Model, error) {
	row := d.tx.QueryRow(d.ctx, `select asset from "models" where id=$1 and channel=$2`, key, d.channel)

	model := new(asset.Model)
	err := row.Scan(model)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, orcerrors.NewNotFound("model", key)
		}
		return nil, err
	}

	return model, nil
}

func (d *DBAL) QueryModels(c asset.ModelCategory, p *common.Pagination) ([]*asset.Model, common.PaginationToken, error) {
	offset, err := getOffset(p.Token)
	if err != nil {
		return nil, "", err
	}

	pgDialect := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	stmt := pgDialect.Select("asset").
		From("models").
		Where(sq.Eq{"channel": d.channel}).
		OrderByClause("asset->>'creationDate' ASC, id").
		Offset(uint64(offset)).
		// Fetch page size + 1 elements to determine whether there is a next page
		Limit(uint64(p.Size + 1))

	if c != asset.ModelCategory_MODEL_UNKNOWN {
		stmt = stmt.Where(sq.Eq{"asset->>'category'": c.String()})
	}

	rows, err := d.query(stmt)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var models []*asset.Model
	var count int

	for rows.Next() {
		model := new(asset.Model)

		err = rows.Scan(model)
		if err != nil {
			return nil, "", err
		}

		models = append(models, model)
		count++

		if count == int(p.Size) {
			break
		}
	}
	if err := rows.Err(); err != nil {
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
	row := d.tx.QueryRow(d.ctx, `select count(id) from "models" where id=$1 and channel=$2`, key, d.channel)

	var count int
	err := row.Scan(&count)

	return count == 1, err
}

func (d *DBAL) GetComputeTaskOutputModels(key string) ([]*asset.Model, error) {
	rows, err := d.tx.Query(d.ctx, `select asset from "models" where compute_task_id = $1 and channel=$2`, key, d.channel)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	models := []*asset.Model{}
	for rows.Next() {
		model := new(asset.Model)
		err := rows.Scan(model)
		if err != nil {
			return nil, err
		}
		models = append(models, model)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return models, nil
}

func (d *DBAL) AddModel(model *asset.Model) error {
	stmt := `insert into "models" ("id", "asset", "channel", "compute_task_id") values ($1, $2, $3, $4)`
	_, err := d.tx.Exec(d.ctx, stmt, model.GetKey(), model, d.channel, model.ComputeTaskKey)
	return err
}

func (d *DBAL) UpdateModel(model *asset.Model) error {
	stmt := `update "models" set asset = $2, compute_task_id = $3 where id = $1 and channel = $4`
	_, err := d.tx.Exec(d.ctx, stmt, model.GetKey(), model, model.ComputeTaskKey, d.channel)
	return err
}
