package dbal

import (
	"errors"
	"strconv"

	sq "github.com/Masterminds/squirrel"
	"github.com/go-playground/log/v7"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	orcerrors "github.com/owkin/orchestrator/lib/errors"
	"github.com/owkin/orchestrator/utils"
	"google.golang.org/protobuf/encoding/protojson"
)

// AddComputeTasks add one or multiple tasks to storage.
func (d *DBAL) AddComputeTasks(tasks ...*asset.ComputeTask) error {
	log.WithField("numTasks", len(tasks)).Debug("dbal: adding tasks in batch mode")
	err := d.insertTasks(tasks)
	if err != nil {
		return err
	}

	return d.insertParentTasks(tasks...)
}

// insertTasks insert tasks in database in batch mode.
func (d *DBAL) insertTasks(tasks []*asset.ComputeTask) error {
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

	return err
}

// insertParentTasks insert the parents of tasks in database in batch mode.
func (d *DBAL) insertParentTasks(tasks ...*asset.ComputeTask) error {
	parentRows := make([][]interface{}, 0)
	for _, t := range tasks {
		if t.ParentTaskKeys != nil {
			childTask, err := uuid.Parse(t.GetKey())
			if err != nil {
				return err
			}

			for idx, parentTaskKey := range t.ParentTaskKeys {
				parentTask, err := uuid.Parse(parentTaskKey)
				if err != nil {
					return err
				}
				parentRows = append(parentRows, []interface{}{parentTask, childTask, idx + 1})
			}
		}
	}

	_, err := d.tx.CopyFrom(
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

	stmt := getStatementBuilder().Select("id").
		From("compute_tasks").
		Where(sq.Eq{"channel": d.channel, "id": uniqueKeys})

	rows, err := d.query(stmt)
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
	stmt := getStatementBuilder().Select("asset").
		From("compute_tasks").
		Where(sq.Eq{"channel": d.channel, "id": keys})

	rows, err := d.query(stmt)
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
	filter := &asset.TaskQueryFilter{ComputePlanKey: key}
	tasks, _, err := d.QueryComputeTasks(nil, filter)
	return tasks, err
}

// QueryComputeTasks returns a paginated and filtered list of tasks.
func (d *DBAL) QueryComputeTasks(pagination *common.Pagination, filter *asset.TaskQueryFilter) ([]*asset.ComputeTask, common.PaginationToken, error) {
	stmt := getStatementBuilder().
		Select("asset").
		From("compute_tasks").
		Where(sq.Eq{"channel": d.channel}).
		OrderByClause("asset->>'creationDate' ASC, id")

	var offset int
	var err error

	if pagination != nil {
		offset, err = getOffset(pagination.Token)
		if err != nil {
			return nil, "", err
		}

		stmt = stmt.Offset(uint64(offset)).
			// Fetch page size + 1 elements to determine whether there is a next page
			Limit(uint64(pagination.Size + 1))
	}

	stmt = taskFilterToQuery(filter, stmt)

	rows, err := d.query(stmt)
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

		if pagination != nil && count == int(pagination.Size) {
			break
		}
	}
	if err = rows.Err(); err != nil {
		return nil, "", err
	}

	bookmark := ""
	if pagination != nil && count == int(pagination.Size) && rows.Next() {
		// there is more to fetch
		bookmark = strconv.Itoa(offset + count)
	}

	return tasks, bookmark, nil
}

// taskFilterToQuery convert as filter into query string and param list
func taskFilterToQuery(filter *asset.TaskQueryFilter, builder sq.SelectBuilder) sq.SelectBuilder {
	if filter == nil {
		return builder
	}

	if filter.Worker != "" {
		builder = builder.Where(sq.Eq{"worker": filter.Worker})
	}
	if filter.Status != 0 {
		builder = builder.Where(sq.Eq{"status": filter.Status.String()})
	}
	if filter.Category != 0 {
		builder = builder.Where(sq.Eq{"category": filter.Category.String()})
	}
	if filter.ComputePlanKey != "" {
		builder = builder.Where(sq.Eq{"compute_plan_id": filter.ComputePlanKey})
	}
	if filter.AlgoKey != "" {
		builder = builder.Where(sq.Eq{"asset->'algo'->>'key'": filter.AlgoKey})
	}

	return builder
}
