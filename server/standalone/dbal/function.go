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
	Archive      asset.Addressable
	Permissions  asset.Permissions
	Owner        string
	CreationDate time.Time
	Metadata     map[string]string
	Status       asset.FunctionStatus
	Image        asset.Addressable
}

func (a *sqlFunction) toFunction() *asset.Function {
	return &asset.Function{
		Key:          a.Key,
		Name:         a.Name,
		Description:  &a.Description,
		Archive:      &a.Archive,
		Permissions:  &a.Permissions,
		Owner:        a.Owner,
		CreationDate: timestamppb.New(a.CreationDate),
		Metadata:     a.Metadata,
		Status:       a.Status,
		Image:        &a.Image,
	}
}

// AddFunction implements persistence.FunctionDBAL
func (d *DBAL) AddFunction(function *asset.Function) error {
	err := d.addAddressable(function.Description)
	if err != nil {
		return err
	}

	err = d.addAddressable(function.Archive)
	if err != nil {
		return err
	}

	err = d.getOrAddAddressable(function.Image)
	if err != nil {
		return err
	}

	stmt := getStatementBuilder().
		Insert("functions").
		Columns("key", "channel", "name", "description", "archive_address", "permissions", "owner", "creation_date", "metadata", "status", "image_address").
		Values(function.Key, d.channel, function.Name, function.Description.StorageAddress, function.Archive.StorageAddress, function.Permissions, function.Owner, function.CreationDate.AsTime(), function.Metadata, function.Status.String(), function.Image.StorageAddress)

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
		Select("key", "name", "description_address", "description_checksum", "archive_address", "archive_checksum", "permissions", "owner", "creation_date", "metadata", "status", "image_address", "image_checksum").
		From("expanded_functions").
		Where(sq.Eq{"key": key, "channel": d.channel})

	row, err := d.queryRow(stmt)
	if err != nil {
		return nil, err
	}

	al := sqlFunction{}

	err = row.Scan(&al.Key, &al.Name, &al.Description.StorageAddress, &al.Description.Checksum, &al.Archive.StorageAddress, &al.Archive.Checksum, &al.Permissions, &al.Owner, &al.CreationDate, &al.Metadata, &al.Status, &al.Image.StorageAddress, &al.Image.Checksum)

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
		Select("key", "name", "description_address", "description_checksum", "archive_address", "archive_checksum", "permissions", "owner", "creation_date", "metadata", "status", "image_address", "image_checksum").
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

		err = rows.Scan(&al.Key, &al.Name, &al.Description.StorageAddress, &al.Description.Checksum, &al.Archive.StorageAddress, &al.Archive.Checksum, &al.Permissions, &al.Owner, &al.CreationDate, &al.Metadata, &al.Status, &al.Image.StorageAddress, &al.Image.Checksum)
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

// UpdateFunction updates the mutable fields of a function in the DB. List of mutable fields: name, status, image.
func (d *DBAL) UpdateFunction(function *asset.Function) error {
	var err error
	if function.Image.StorageAddress != "" {
		err = d.updateAddressable(function.Image)
		if err != nil {
			return err
		}
		stmt := getStatementBuilder().
			Update("functions").
			Set("name", function.Name).
			Set("status", function.Status.String()).
			Set("image_address", function.Image.StorageAddress).
			Where(sq.Eq{"channel": d.channel, "key": function.Key})
		return d.exec(stmt)
	}
	stmt := getStatementBuilder().
		Update("functions").
		Set("name", function.Name).
		Set("status", function.Status.String()).
		Where(sq.Eq{"channel": d.channel, "key": function.Key})
	return d.exec(stmt)

}
