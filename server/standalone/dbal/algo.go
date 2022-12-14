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

type sqlAlgo struct {
	Key          string
	Name         string
	Description  asset.Addressable
	Algorithm    asset.Addressable
	Permissions  asset.Permissions
	Owner        string
	CreationDate time.Time
	Metadata     map[string]string
}

func (a *sqlAlgo) toAlgo() *asset.Algo {
	return &asset.Algo{
		Key:          a.Key,
		Name:         a.Name,
		Description:  &a.Description,
		Algorithm:    &a.Algorithm,
		Permissions:  &a.Permissions,
		Owner:        a.Owner,
		CreationDate: timestamppb.New(a.CreationDate),
		Metadata:     a.Metadata,
	}
}

// AddAlgo implements persistence.AlgoDBAL
func (d *DBAL) AddAlgo(algo *asset.Algo) error {
	err := d.addAddressable(algo.Description)
	if err != nil {
		return err
	}

	err = d.addAddressable(algo.Algorithm)
	if err != nil {
		return err
	}

	stmt := getStatementBuilder().
		Insert("algos").
		Columns("key", "channel", "name", "description", "algorithm", "permissions", "owner", "creation_date", "metadata").
		Values(algo.Key, d.channel, algo.Name, algo.Description.StorageAddress, algo.Algorithm.StorageAddress, algo.Permissions, algo.Owner, algo.CreationDate.AsTime(), algo.Metadata)

	err = d.exec(stmt)
	if err != nil {
		return err
	}

	err = d.insertAlgoInputs(algo.Key, algo.Inputs)
	if err != nil {
		return err
	}

	err = d.insertAlgoOutputs(algo.Key, algo.Outputs)
	if err != nil {
		return err
	}

	return nil
}

// GetAlgo implements persistence.AlgoDBAL
func (d *DBAL) GetAlgo(key string) (*asset.Algo, error) {
	stmt := getStatementBuilder().
		Select("key", "name", "description_address", "description_checksum", "algorithm_address", "algorithm_checksum", "permissions", "owner", "creation_date", "metadata").
		From("expanded_algos").
		Where(sq.Eq{"key": key, "channel": d.channel})

	row, err := d.queryRow(stmt)
	if err != nil {
		return nil, err
	}

	al := sqlAlgo{}
	err = row.Scan(&al.Key, &al.Name, &al.Description.StorageAddress, &al.Description.Checksum, &al.Algorithm.StorageAddress, &al.Algorithm.Checksum, &al.Permissions, &al.Owner, &al.CreationDate, &al.Metadata)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, orcerrors.NewNotFound(asset.AlgoKind, key)
		}
		return nil, err
	}

	res := al.toAlgo()

	err = d.populateAlgosIO(res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// QueryAlgos implements persistence.AlgoDBAL
func (d *DBAL) QueryAlgos(p *common.Pagination, filter *asset.AlgoQueryFilter) ([]*asset.Algo, common.PaginationToken, error) {
	algos, bookmark, err := d.queryAlgos(p, filter)
	if err != nil {
		return nil, "", err
	}

	err = d.populateAlgosIO(algos...)
	if err != nil {
		return nil, "", err
	}

	return algos, bookmark, nil
}

// AlgoExists implements persistence.AlgoDBAL
func (d *DBAL) AlgoExists(key string) (bool, error) {
	stmt := getStatementBuilder().
		Select("COUNT(key)").
		From("algos").
		Where(sq.Eq{"key": key, "channel": d.channel})

	row, err := d.queryRow(stmt)
	if err != nil {
		return false, err
	}

	var count int
	err = row.Scan(&count)

	return count == 1, err
}

func (d *DBAL) queryAlgos(p *common.Pagination, filter *asset.AlgoQueryFilter) ([]*asset.Algo, common.PaginationToken, error) {
	offset, err := getOffset(p.Token)
	if err != nil {
		return nil, "", err
	}

	stmt := getStatementBuilder().
		Select("key", "name", "description_address", "description_checksum", "algorithm_address", "algorithm_checksum", "permissions", "owner", "creation_date", "metadata").
		From("expanded_algos").
		Where(sq.Eq{"channel": d.channel}).
		OrderByClause("creation_date ASC, key").
		Offset(uint64(offset)).
		// Fetch page size + 1 elements to determine whether there is a next page
		Limit(uint64(p.Size + 1))

	if filter != nil {
		if filter.ComputePlanKey != "" {
			stmt = stmt.Where(sq.Expr(
				"key IN (SELECT DISTINCT(algo_key) FROM compute_tasks WHERE compute_plan_key = ?)",
				filter.ComputePlanKey,
			))
		}
	}

	rows, err := d.query(stmt)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	algos := make([]*asset.Algo, 0, p.Size)
	var count int

	for rows.Next() {
		al := sqlAlgo{}

		err = rows.Scan(&al.Key, &al.Name, &al.Description.StorageAddress, &al.Description.Checksum, &al.Algorithm.StorageAddress, &al.Algorithm.Checksum, &al.Permissions, &al.Owner, &al.CreationDate, &al.Metadata)
		if err != nil {
			return nil, "", err
		}

		algos = append(algos, al.toAlgo())
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

// UpdateAlgo updates the mutable fields of an algo in the DB. List of mutable fields: name.
func (d *DBAL) UpdateAlgo(algo *asset.Algo) error {
	stmt := getStatementBuilder().
		Update("algos").
		Set("name", algo.Name).
		Where(sq.Eq{"channel": d.channel, "key": algo.Key})

	return d.exec(stmt)
}
