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

// AddAlgo implements persistence.AlgoDBAL
func (d *DBAL) AddAlgo(algo *asset.Algo) error {
	stmt := `insert into "algos" ("id", "asset", "channel") values ($1, $2, $3)`
	_, err := d.tx.Exec(d.ctx, stmt, algo.GetKey(), algo, d.channel)
	return err
}

// GetAlgo implements persistence.AlgoDBAL
func (d *DBAL) GetAlgo(key string) (*asset.Algo, error) {
	row := d.tx.QueryRow(d.ctx, `select "asset" from "algos" where id=$1 and channel=$2`, key, d.channel)

	algo := new(asset.Algo)
	err := row.Scan(algo)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, orcerrors.NewNotFound("algo", key)
		}
		return nil, err
	}

	return algo, nil
}

// QueryAlgos implements persistence.AlgoDBAL
func (d *DBAL) QueryAlgos(p *common.Pagination, filter *asset.AlgoQueryFilter) ([]*asset.Algo, common.PaginationToken, error) {
	offset, err := getOffset(p.Token)
	if err != nil {
		return nil, "", err
	}

	stmt := getStatementBuilder().Select("asset").
		From("algos").
		Where(sq.Eq{"channel": d.channel}).
		OrderByClause("asset->>'creationDate' ASC, id").
		Offset(uint64(offset)).
		// Fetch page size + 1 elements to determine whether there is a next page
		Limit(uint64(p.Size + 1))

	if filter.Category != asset.AlgoCategory_ALGO_UNKNOWN {
		stmt = stmt.Where(sq.Eq{"asset->>'category'": filter.Category.String()})
	}

	if filter.ComputePlanKey != "" {
		stmt = stmt.Where(sq.Expr(
			"id IN (SELECT DISTINCT(asset->'algo'->>'key')::uuid FROM compute_tasks WHERE compute_plan_id = ?)",
			filter.ComputePlanKey,
		))
	}

	rows, err := d.query(stmt)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var algos []*asset.Algo
	var count int

	for rows.Next() {
		algo := new(asset.Algo)

		err = rows.Scan(algo)
		if err != nil {
			return nil, "", err
		}

		algos = append(algos, algo)
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

	return algos, bookmark, nil
}

// AlgoExists implements persistence.AlgoDBAL
func (d *DBAL) AlgoExists(key string) (bool, error) {
	row := d.tx.QueryRow(d.ctx, `select count(id) from "algos" where id=$1 and channel=$2`, key, d.channel)

	var count int
	err := row.Scan(&count)

	return count == 1, err
}
