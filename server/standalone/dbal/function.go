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

type sqlFunction struct {
	Key          string
	Name         string
	Description  asset.Addressable
	Function     asset.Addressable
	Permissions  asset.Permissions
	Owner        string
	CreationDate time.Time
	Metadata     map[string]string
	Status       asset.FunctionStatus
}

func (a *sqlFunction) toFunction() *asset.Function {
	return &asset.Function{
		Key:          a.Key,
		Name:         a.Name,
		Description:  &a.Description,
		Function:     &a.Function,
		Permissions:  &a.Permissions,
		Owner:        a.Owner,
		CreationDate: timestamppb.New(a.CreationDate),
		Metadata:     a.Metadata,
		Status:       a.Status,
	}
}

// AddFunction implements persistence.FunctionDBAL
func (d *DBAL) AddFunction(function *asset.Function) error {
	err := d.addAddressable(function.Description)
	if err != nil {
		return err
	}

	err = d.addAddressable(function.Function)
	if err != nil {
		return err
	}

	stmt := getStatementBuilder().
		Insert("functions").
		Columns("key", "channel", "name", "description", "functionAddress", "permissions", "owner", "creation_date", "metadata", "status").
		Values(function.Key, d.channel, function.Name, function.Description.StorageAddress, function.Function.StorageAddress, function.Permissions, function.Owner, function.CreationDate.AsTime(), function.Metadata, function.Status.String())

	err = d.exec(stmt)
	if err != nil {
		return err
	}

	err = d.insertFunctionInputs(function.Key, function.Inputs)
	if err != nil {
		return err
	}

	err = d.insertFunctionOutputs(function.Key, function.Outputs)
	if err != nil {
		return err
	}

	return nil
}

// GetFunction implements persistence.FunctionDBAL
func (d *DBAL) GetFunction(key string) (*asset.Function, error) {
	stmt := getStatementBuilder().
		Select("key", "name", "description_address", "description_checksum", "function_address", "function_checksum", "permissions", "owner", "creation_date", "metadata", "status").
		From("expanded_functions").
		Where(sq.Eq{"key": key, "channel": d.channel})

	row, err := d.queryRow(stmt)
	if err != nil {
		return nil, err
	}

	al := sqlFunction{}
	err = row.Scan(&al.Key, &al.Name, &al.Description.StorageAddress, &al.Description.Checksum, &al.Function.StorageAddress, &al.Function.Checksum, &al.Permissions, &al.Owner, &al.CreationDate, &al.Metadata, &al.Status)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, orcerrors.NewNotFound(asset.FunctionKind, key)
		}
		return nil, err
	}

	res := al.toFunction()

	err = d.populateFunctionsIO(res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// QueryFunctions implements persistence.FunctionDBAL
func (d *DBAL) QueryFunctions(p *common.Pagination, filter *asset.FunctionQueryFilter) ([]*asset.Function, common.PaginationToken, error) {
	functions, bookmark, err := d.queryFunctions(p, filter)
	if err != nil {
		return nil, "", err
	}

	err = d.populateFunctionsIO(functions...)
	if err != nil {
		return nil, "", err
	}

	return functions, bookmark, nil
}

// FunctionExists implements persistence.FunctionDBAL
func (d *DBAL) FunctionExists(key string) (bool, error) {
	stmt := getStatementBuilder().
		Select("COUNT(key)").
		From("functions").
		Where(sq.Eq{"key": key, "channel": d.channel})

	row, err := d.queryRow(stmt)
	if err != nil {
		return false, err
	}

	var count int
	err = row.Scan(&count)

	return count == 1, err
}

func (d *DBAL) queryFunctions(p *common.Pagination, filter *asset.FunctionQueryFilter) ([]*asset.Function, common.PaginationToken, error) {
	offset, err := getOffset(p.Token)
	if err != nil {
		return nil, "", err
	}

	stmt := getStatementBuilder().
		Select("key", "name", "description_address", "description_checksum", "function_address", "function_checksum", "permissions", "owner", "creation_date", "metadata", "status").
		From("expanded_functions").
		Where(sq.Eq{"channel": d.channel}).
		OrderByClause("creation_date ASC, key").
		Offset(uint64(offset)).
		// Fetch page size + 1 elements to determine whether there is a next page
		Limit(uint64(p.Size + 1))

	if filter != nil {
		if filter.ComputePlanKey != "" {
			stmt = stmt.Where(sq.Expr(
				"key IN (SELECT DISTINCT(function_key) FROM compute_tasks WHERE compute_plan_key = ?)",
				filter.ComputePlanKey,
			))
		}
	}

	rows, err := d.query(stmt)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	functions := make([]*asset.Function, 0, p.Size)
	var count int

	for rows.Next() {
		al := sqlFunction{}

		err = rows.Scan(&al.Key, &al.Name, &al.Description.StorageAddress, &al.Description.Checksum, &al.Function.StorageAddress, &al.Function.Checksum, &al.Permissions, &al.Owner, &al.CreationDate, &al.Metadata, &al.Status)
		if err != nil {
			return nil, "", err
		}

		functions = append(functions, al.toFunction())
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

	return functions, bookmark, nil
}

// UpdateFunction updates the mutable fields of an function in the DB. List of mutable fields: name, status.
func (d *DBAL) UpdateFunction(function *asset.Function) error {
	stmt := getStatementBuilder().
		Update("functions").
		Set("name", function.Name).
		Where(sq.Eq{"channel": d.channel, "key": function.Key})

	return d.exec(stmt)
}

func (d *DBAL) UpdateFunctionStatus(functionKey string, status asset.FunctionStatus) error {
	stmt := getStatementBuilder().
		Update("functions").
		Set("status", status.String()).
		Where(sq.Eq{"channel": d.channel, "key": functionKey})

	return d.exec(stmt)
}
