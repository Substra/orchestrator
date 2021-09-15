package dbal

import (
	"errors"
	"strconv"

	"github.com/Masterminds/squirrel"
	"github.com/go-playground/log/v7"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	orcerrors "github.com/owkin/orchestrator/lib/errors"
	"github.com/owkin/orchestrator/utils"
	"google.golang.org/protobuf/encoding/protojson"
)

// AddComputeTasks insert tasks in storage according to the most efficient way.
// Up to 5 tasks, they will be inserted one by one.
// For more than 5 tasks they will be processed in batch.
func (d *DBAL) AddComputeTasks(tasks ...*asset.ComputeTask) error {
	if len(tasks) >= 5 {
		log.WithField("numTasks", len(tasks)).Debug("dbal: adding multiple tasks in batch mode")
		return d.addTasks(tasks)
	}

	for _, t := range tasks {
		err := d.addTask(t)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *DBAL) addTask(t *asset.ComputeTask) error {
	stmt := `insert into "compute_tasks" ("id", "asset", "channel") values ($1, $2, $3)`
	_, err := d.tx.Exec(d.ctx, stmt, t.GetKey(), t, d.channel)
	return err
}

func (d *DBAL) addTasks(tasks []*asset.ComputeTask) error {
	_, err := d.tx.CopyFrom(
		d.ctx,
		pgx.Identifier{"compute_tasks"},
		[]string{"id", "asset", "channel"},
		pgx.CopyFromSlice(len(tasks), func(i int) ([]interface{}, error) {
			v, err := protojson.Marshal(tasks[i])
			if err != nil {
				return nil, err
			}
			// expect binary representation, not string
			id, err := uuid.Parse(tasks[i].Key)
			if err != nil {
				return nil, err
			}
			return []interface{}{id, v, d.channel}, nil
		}),
	)

	return err
}

// UpdateComputeTask updates an existing task
func (d *DBAL) UpdateComputeTask(t *asset.ComputeTask) error {
	stmt := `update "compute_tasks" set asset=$3 where id=$1 and channel=$2`
	_, err := d.tx.Exec(d.ctx, stmt, t.GetKey(), d.channel, t)
	return err
}

// ComputeTaskExists returns true if a task with the given ID exists
func (d *DBAL) ComputeTaskExists(key string) (bool, error) {
	row := d.tx.QueryRow(d.ctx, `select count(id) from "compute_tasks" where id=$1 and channel=$2`, key, d.channel)

	var count int
	err := row.Scan(&count)

	return count == 1, err
}

// GetExistingComputeTaskKeys returns the keys of tasks already in storage among those given as input.
func (d *DBAL) GetExistingComputeTaskKeys(keys []string) ([]string, error) {
	existingKeys := []string{}

	uniqueKeys := utils.UniqueString(keys)
	pgDialect := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	query, args, err := pgDialect.Select("id").
		From("compute_tasks").
		Where(squirrel.Eq{"channel": d.channel, "id": uniqueKeys}).ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := d.tx.Query(d.ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var key string
		err = rows.Scan(&key)
		if err != nil {
			return nil, err
		}

		existingKeys = append(existingKeys, key)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return existingKeys, nil
}

// GetComputeTask returns a single task by its key
func (d *DBAL) GetComputeTask(key string) (*asset.ComputeTask, error) {
	row := d.tx.QueryRow(d.ctx, `select asset from "compute_tasks" where id=$1 and channel=$2`, key, d.channel)

	task := new(asset.ComputeTask)
	err := row.Scan(task)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, orcerrors.NewNotFound("computetask", key)
		}
		return nil, err
	}

	return task, nil
}

// GetComputeTasks returns the list of unique compute tasks identified by the provided keys.
// It should not be used where pagination is expected!
func (d *DBAL) GetComputeTasks(keys []string) ([]*asset.ComputeTask, error) {
	pgDialect := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	query, args, err := pgDialect.Select("asset").
		From("compute_tasks").
		Where(squirrel.Eq{"channel": d.channel, "id": keys}).ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := d.tx.Query(d.ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*asset.ComputeTask

	for rows.Next() {
		task := new(asset.ComputeTask)

		err = rows.Scan(&task)
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

// GetComputeTaskChildren returns the children of the task identified by the given key
func (d *DBAL) GetComputeTaskChildren(key string) ([]*asset.ComputeTask, error) {
	rows, err := d.tx.Query(d.ctx, `select asset from "compute_tasks" where asset->'parentTaskKeys' ? $1 and channel=$2`, key, d.channel)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := []*asset.ComputeTask{}
	for rows.Next() {
		task := new(asset.ComputeTask)
		err := rows.Scan(task)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

// GetComputePlanTasksKeys returns the list of task keys from the provided compute plan
func (d *DBAL) GetComputePlanTasksKeys(key string) ([]string, error) {
	rows, err := d.tx.Query(d.ctx, `select id from "compute_tasks" where asset->>'computePlanKey' = $1 and channel=$2`, key, d.channel)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	keys := []string{}
	for rows.Next() {
		var key string
		err := rows.Scan(&key)
		if err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return keys, nil
}

// GetComputePlanTasks returns the tasks of the compute plan identified by the given key
func (d *DBAL) GetComputePlanTasks(key string) ([]*asset.ComputeTask, error) {
	rows, err := d.tx.Query(d.ctx, `select asset from "compute_tasks" where asset->>'computePlanKey' = $1 and channel=$2`, key, d.channel)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := []*asset.ComputeTask{}
	for rows.Next() {
		task := new(asset.ComputeTask)
		err := rows.Scan(task)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

// QueryComputeTasks returns a paginated and filtered list of tasks.
func (d *DBAL) QueryComputeTasks(p *common.Pagination, filter *asset.TaskQueryFilter) ([]*asset.ComputeTask, common.PaginationToken, error) {
	var rows pgx.Rows
	var err error

	offset, err := getOffset(p.Token)
	if err != nil {
		return nil, "", err
	}

	pgDialect := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	builder := pgDialect.Select("asset").
		From("compute_tasks").
		Where(squirrel.Eq{"channel": d.channel}).
		OrderByClause("asset->>'creationDate' ASC").
		Offset(uint64(offset)).
		Limit(uint64(p.Size + 1))

	builder = taskFilterToQuery(filter, builder)

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, "", err
	}

	rows, err = d.tx.Query(d.ctx, query, args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var tasks []*asset.ComputeTask
	var count int

	for rows.Next() {
		task := new(asset.ComputeTask)

		err = rows.Scan(&task)
		if err != nil {
			return nil, "", err
		}

		tasks = append(tasks, task)
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

	return tasks, bookmark, nil
}

// taskFilterToQuery convert as filter into query string and param list
func taskFilterToQuery(filter *asset.TaskQueryFilter, builder squirrel.SelectBuilder) squirrel.SelectBuilder {
	if filter == nil {
		return builder
	}

	if filter.Worker != "" {
		builder = builder.Where(squirrel.Eq{"asset->>'worker'": filter.Worker})
	}
	if filter.Status != 0 {
		builder = builder.Where(squirrel.Eq{"asset->>'status'": filter.Status.String()})
	}
	if filter.Category != 0 {
		builder = builder.Where(squirrel.Eq{"asset->>'category'": filter.Category.String()})
	}
	if filter.ComputePlanKey != "" {
		builder = builder.Where(squirrel.Eq{"asset->>'computePlanKey'": filter.ComputePlanKey})
	}
	if filter.AlgoKey != "" {
		builder = builder.Where(squirrel.Eq{"asset->'algo'->>'key'": filter.AlgoKey})
	}

	return builder
}
