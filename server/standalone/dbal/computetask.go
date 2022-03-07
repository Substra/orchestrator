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
	stmt := `insert into "compute_tasks" ("id", "channel", "category", "compute_plan_id", "status", "worker", "asset") values ($1, $2, $3, $4, $5, $6, $7)`
	_, err := d.tx.Exec(d.ctx, stmt, t.GetKey(), d.channel, t.Category, t.ComputePlanKey, t.Status, t.Worker, t)
	if err != nil {
		return err
	}

	err = d.insertParentTasks(t)
	return err
}

func (d *DBAL) addTasks(tasks []*asset.ComputeTask) error {
	_, err := d.tx.CopyFrom(
		d.ctx,
		pgx.Identifier{"compute_tasks"},
		[]string{"id", "channel", "category", "compute_plan_id", "status", "worker", "asset"},
		pgx.CopyFromSlice(len(tasks), func(i int) ([]interface{}, error) {
			task := tasks[i]
			v, err := protojson.Marshal(task)
			if err != nil {
				return nil, err
			}

			// expect binary representation, not string
			id, err := uuid.Parse(task.Key)
			if err != nil {
				return nil, err
			}
			computePlanKey, err := uuid.Parse(task.ComputePlanKey)
			if err != nil {
				return nil, err
			}

			return []interface{}{id, d.channel, task.Category, computePlanKey, task.Status, task.Worker, v}, nil
		}),
	)

	if err != nil {
		return err
	}

	// Add compute_task_parents rows
	parentRows := make([][]interface{}, 0)
	for _, t := range tasks {
		if t.ParentTaskKeys != nil {
			childTask, err := uuid.Parse(t.GetKey())
			if err != nil {
				return err
			}
			position := 1
			for _, parentTaskKey := range t.ParentTaskKeys {
				parentTask, err := uuid.Parse(parentTaskKey)
				if err != nil {
					return err
				}
				parentRows = append(parentRows, []interface{}{parentTask, childTask, position})
				position++
			}
		}
	}

	_, err = d.tx.CopyFrom(
		d.ctx,
		pgx.Identifier{"compute_task_parents"},
		[]string{"parent_task_id", "child_task_id", "position"},
		pgx.CopyFromRows(parentRows),
	)

	return err
}

// deleteParentTasks deletes all parent tasks for a given task
func (d *DBAL) deleteParentTasks(taskKey string) error {
	stmt := `delete from compute_task_parents where child_task_id = $1`
	_, err := d.tx.Exec(d.ctx, stmt, taskKey)
	return err
}

// insertParentTasks inserts all the parents tasks of a compute task, one by one
func (d *DBAL) insertParentTasks(t *asset.ComputeTask) error {
	position := 1
	for _, parentTaskKey := range t.ParentTaskKeys {
		stmt := `insert into compute_task_parents(parent_task_id, child_task_id, position) values ($1, $2, $3)`
		_, err := d.tx.Exec(d.ctx, stmt, parentTaskKey, t.GetKey(), position)
		if err != nil {
			return err
		}
		position++
	}
	return nil
}

// UpdateComputeTask updates an existing task
func (d *DBAL) UpdateComputeTask(t *asset.ComputeTask) error {
	err := d.deleteParentTasks(t.GetKey())
	if err != nil {
		return err
	}

	err = d.insertParentTasks(t)
	if err != nil {
		return err
	}

	stmt := `update "compute_tasks" set category=$3, compute_plan_id=$4, status=$5, worker=$6, asset=$7 where id=$1 and channel=$2`
	_, err = d.tx.Exec(d.ctx, stmt, t.GetKey(), d.channel, t.Category, t.ComputePlanKey, t.Status, t.Worker, t)
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

		err = rows.Scan(task)
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
	rows, err := d.tx.Query(d.ctx, `
	select ct.asset
	from compute_tasks ct
	join compute_task_parents ctp on ct.id = ctp.child_task_id
	 where ctp.parent_task_id = $1
	 and ct.channel = $2
	order by ctp.position asc;`, key, d.channel)
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
	rows, err := d.tx.Query(d.ctx, `select id from "compute_tasks" where compute_plan_id = $1 and channel=$2`, key, d.channel)
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
	rows, err := d.tx.Query(d.ctx, `select asset from "compute_tasks" where compute_plan_id = $1 and channel=$2`, key, d.channel)
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
		OrderByClause("asset->>'creationDate' ASC, id").
		Offset(uint64(offset)).
		// Fetch page size + 1 elements to determine whether there is a next page
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

		err = rows.Scan(task)
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
		builder = builder.Where(squirrel.Eq{"worker": filter.Worker})
	}
	if filter.Status != 0 {
		builder = builder.Where(squirrel.Eq{"status": filter.Status.String()})
	}
	if filter.Category != 0 {
		builder = builder.Where(squirrel.Eq{"category": filter.Category.String()})
	}
	if filter.ComputePlanKey != "" {
		builder = builder.Where(squirrel.Eq{"compute_plan_id": filter.ComputePlanKey})
	}
	if filter.AlgoKey != "" {
		builder = builder.Where(squirrel.Eq{"asset->'algo'->>'key'": filter.AlgoKey})
	}

	return builder
}
