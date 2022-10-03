package dbal

import (
	"errors"
	"strconv"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/rs/zerolog/log"
	"github.com/substra/orchestrator/lib/asset"
	"github.com/substra/orchestrator/lib/common"
	orcerrors "github.com/substra/orchestrator/lib/errors"
	"github.com/substra/orchestrator/lib/persistence"
	"github.com/substra/orchestrator/utils"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const computeTaskOutputAssetsTable = "compute_task_output_assets"

type sqlComputeTask struct {
	Key            string
	AlgoKey        string
	Owner          string
	ComputePlanKey string
	ParentTaskKeys []string
	Rank           int32
	Status         asset.ComputeTaskStatus
	Worker         string
	CreationDate   time.Time
	LogsPermission asset.Permission
	Data           []byte
	Metadata       map[string]string
}

func (t *sqlComputeTask) toComputeTask() (*asset.ComputeTask, error) {
	task := new(asset.ComputeTask)

	// task data is stored as a marshalled task object
	err := protojson.Unmarshal(t.Data, task)
	if err != nil {
		return nil, err
	}

	task.Key = t.Key
	task.AlgoKey = t.AlgoKey
	task.Owner = t.Owner
	task.ComputePlanKey = t.ComputePlanKey
	task.ParentTaskKeys = t.ParentTaskKeys
	task.Rank = t.Rank
	task.Status = t.Status
	task.Worker = t.Worker
	task.CreationDate = timestamppb.New(t.CreationDate)
	task.LogsPermission = &t.LogsPermission
	task.Metadata = t.Metadata

	return task, nil
}

// AddComputeTasks add one or multiple tasks to storage.
func (d *DBAL) AddComputeTasks(tasks ...*asset.ComputeTask) error {
	log.Ctx(d.ctx).Debug().Int("numTasks", len(tasks)).Msg("dbal: adding tasks in batch mode")
	err := d.insertTasks(tasks)
	if err != nil {
		return err
	}

	err = d.insertParentTasks(tasks...)
	if err != nil {
		return err
	}

	err = d.insertTaskInputs(tasks)
	if err != nil {
		return err
	}

	return d.insertTaskOutputs(tasks)
}

// insertTasks insert tasks in database in batch mode.
func (d *DBAL) insertTasks(tasks []*asset.ComputeTask) error {
	// insert tasks
	_, err := d.tx.CopyFrom(
		d.ctx,
		pgx.Identifier{"compute_tasks"},
		[]string{"key", "channel", "algo_key", "owner", "compute_plan_key", "rank", "status", "worker", "creation_date", "logs_permission", "task_data", "metadata"},
		pgx.CopyFromSlice(len(tasks), func(i int) ([]interface{}, error) {
			return getCopyableComputeTaskValues(d.channel, tasks[i])
		}),
	)
	return err
}

func getCopyableComputeTaskValues(channel string, task *asset.ComputeTask) ([]interface{}, error) {
	// expect binary representation, not string
	key, err := uuid.Parse(task.Key)
	if err != nil {
		return nil, err
	}

	algoKey, err := uuid.Parse(task.AlgoKey)
	if err != nil {
		return nil, err
	}

	computePlanKey, err := uuid.Parse(task.ComputePlanKey)
	if err != nil {
		return nil, err
	}

	logsPermission, err := protojson.Marshal(task.LogsPermission)
	if err != nil {
		return nil, err
	}

	// store task data in a marshalled task object, empty fields will be omitted
	taskData, err := protojson.Marshal(&asset.ComputeTask{Data: task.Data})
	if err != nil {
		return nil, err
	}

	return []interface{}{
		key,
		channel,
		algoKey,
		task.Owner,
		computePlanKey,
		task.Rank,
		task.Status.String(),
		task.Worker,
		task.CreationDate.AsTime(),
		logsPermission,
		taskData,
		task.Metadata,
	}, nil
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
		[]string{"parent_task_key", "child_task_key", "position"},
		pgx.CopyFromRows(parentRows),
	)

	return err
}

// UpdateComputeTaskStatus updates the status of an existing task.
func (d *DBAL) UpdateComputeTaskStatus(taskKey string, taskStatus asset.ComputeTaskStatus) error {
	stmt := getStatementBuilder().
		Update("compute_tasks").
		Set("status", taskStatus).
		Where(sq.Eq{"channel": d.channel, "key": taskKey})

	return d.exec(stmt)
}

// GetExistingComputeTaskKeys returns the keys of tasks already in storage among those given as input.
func (d *DBAL) GetExistingComputeTaskKeys(keys []string) ([]string, error) {
	if len(keys) == 0 {
		return []string{}, nil
	}

	uniqueKeys := utils.Unique(keys)

	stmt := getStatementBuilder().
		Select("key").
		From("compute_tasks").
		Where(sq.Eq{"channel": d.channel, "key": uniqueKeys})

	rows, err := d.query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	existingKeys := []string{}
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
	stmt := getStatementBuilder().
		Select("key", "compute_plan_key", "status", "worker", "owner", "rank", "creation_date",
			"logs_permission", "task_data", "metadata", "algo_key", "parent_task_keys").
		From("expanded_compute_tasks").
		Where(sq.Eq{"channel": d.channel, "key": key})

	row, err := d.queryRow(stmt)
	if err != nil {
		return nil, err
	}

	ct := new(sqlComputeTask)
	err = row.Scan(&ct.Key, &ct.ComputePlanKey, &ct.Status, &ct.Worker, &ct.Owner, &ct.Rank, &ct.CreationDate,
		&ct.LogsPermission, &ct.Data, &ct.Metadata, &ct.AlgoKey, &ct.ParentTaskKeys)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, orcerrors.NewNotFound("computetask", key)
		}
		return nil, err
	}

	res, err := ct.toComputeTask()
	if err != nil {
		return nil, err
	}

	err = d.populateTasksIO(res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// GetComputeTaskChildren returns the children of the task identified by the given key.
// Warning: this function doesn't populate the task input/output fields, not the algo input/output fields.
func (d *DBAL) GetComputeTaskChildren(key string) ([]*asset.ComputeTask, error) {
	stmt := getStatementBuilder().
		Select("key", "compute_plan_key", "status", "worker", "owner", "rank", "creation_date",
			"logs_permission", "task_data", "metadata", "algo_key", "parent_task_keys").
		From("expanded_compute_tasks t").
		Join("compute_task_parents p ON t.key = p.child_task_key").
		Where(sq.Eq{"t.channel": d.channel, "p.parent_task_key": key}).
		OrderByClause("p.position ASC")

	rows, err := d.query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := []*asset.ComputeTask{}
	for rows.Next() {
		ct := new(sqlComputeTask)

		err = rows.Scan(
			&ct.Key, &ct.ComputePlanKey, &ct.Status, &ct.Worker, &ct.Owner, &ct.Rank, &ct.CreationDate,
			&ct.LogsPermission, &ct.Data, &ct.Metadata, &ct.AlgoKey, &ct.ParentTaskKeys)
		if err != nil {
			return nil, err
		}

		task, err := ct.toComputeTask()
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
	stmt := getStatementBuilder().
		Select("key").
		From("compute_tasks").
		Where(sq.Eq{"channel": d.channel, "compute_plan_key": key})

	rows, err := d.query(stmt)
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

func (d *DBAL) CountComputeTaskRegisteredOutputs(key string) (persistence.ComputeTaskOutputCounter, error) {
	counter := make(persistence.ComputeTaskOutputCounter)

	stmt := getStatementBuilder().
		Select("compute_task_output_identifier, count(1)").
		From("compute_task_output_assets").
		GroupBy("compute_task_output_identifier").
		Where(sq.Eq{"compute_task_key": key})

	rows, err := d.query(stmt)
	if err != nil {
		return counter, err
	}
	defer rows.Close()

	for rows.Next() {
		var identifier string
		var count int
		err := rows.Scan(&identifier, &count)
		if err != nil {
			return counter, err
		}
		counter[identifier] = count
	}
	if err := rows.Err(); err != nil {
		return counter, err
	}

	return counter, nil
}

// queryBaseComputeTasks will return tasks without inputs/outputs, their keys and pagination token
func (d *DBAL) queryBaseComputeTasks(pagination *common.Pagination, filterer func(sq.SelectBuilder) sq.SelectBuilder) ([]*asset.ComputeTask, common.PaginationToken, error) {
	stmt := getStatementBuilder().
		Select("key", "compute_plan_key", "status", "worker", "owner", "rank", "creation_date",
			"logs_permission", "task_data", "metadata", "algo_key", "parent_task_keys").
		From("expanded_compute_tasks").
		Where(sq.Eq{"channel": d.channel}).
		OrderByClause("creation_date ASC, key")

	var (
		offset int
		err    error
		tasks  []*asset.ComputeTask
	)

	if pagination != nil {
		offset, err = getOffset(pagination.Token)
		if err != nil {
			return nil, "", err
		}

		stmt = stmt.Offset(uint64(offset)).
			// Fetch page size + 1 elements to determine whether there is a next page
			Limit(uint64(pagination.Size + 1))

		tasks = make([]*asset.ComputeTask, 0, pagination.Size)
	} else {
		tasks = make([]*asset.ComputeTask, 0)
	}

	stmt = filterer(stmt)

	rows, err := d.query(stmt)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var count int

	for rows.Next() {
		ct := new(sqlComputeTask)

		err = rows.Scan(
			&ct.Key, &ct.ComputePlanKey, &ct.Status, &ct.Worker, &ct.Owner, &ct.Rank, &ct.CreationDate,
			&ct.LogsPermission, &ct.Data, &ct.Metadata, &ct.AlgoKey, &ct.ParentTaskKeys)
		if err != nil {
			return nil, "", err
		}

		task, err := ct.toComputeTask()
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

func (d *DBAL) queryComputeTasks(pagination *common.Pagination, filterer func(sq.SelectBuilder) sq.SelectBuilder) ([]*asset.ComputeTask, common.PaginationToken, error) {
	tasks, bookmark, err := d.queryBaseComputeTasks(pagination, filterer)
	if err != nil {
		return nil, "", err
	}

	err = d.populateTasksIO(tasks...)
	if err != nil {
		return nil, "", err
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
	if filter.ComputePlanKey != "" {
		builder = builder.Where(sq.Eq{"compute_plan_key": filter.ComputePlanKey})
	}
	if filter.AlgoKey != "" {
		builder = builder.Where(sq.Eq{"algo_key": filter.AlgoKey})
	}

	return builder
}

// QueryComputeTasks returns a paginated and filtered list of tasks.
func (d *DBAL) QueryComputeTasks(pagination *common.Pagination, filter *asset.TaskQueryFilter) ([]*asset.ComputeTask, common.PaginationToken, error) {
	return d.queryComputeTasks(pagination, func(builder sq.SelectBuilder) sq.SelectBuilder {
		return taskFilterToQuery(filter, builder)
	})
}

// GetComputePlanTasks returns the tasks of the compute plan identified by the given key
func (d *DBAL) GetComputePlanTasks(key string) ([]*asset.ComputeTask, error) {
	tasks, _, err := d.queryComputeTasks(nil, func(builder sq.SelectBuilder) sq.SelectBuilder {
		return builder.Where(sq.Eq{"compute_plan_key": key})
	})
	return tasks, err
}

// GetComputeTasks returns the list of unique compute tasks identified by the provided keys.
// It should not be used where pagination is expected!
func (d *DBAL) GetComputeTasks(keys []string) ([]*asset.ComputeTask, error) {
	tasks, _, err := d.queryComputeTasks(nil, func(builder sq.SelectBuilder) sq.SelectBuilder {
		return builder.Where(sq.Eq{"key": keys})
	})
	return tasks, err
}

func (d *DBAL) AddComputeTaskOutputAsset(output *asset.ComputeTaskOutputAsset) error {
	stmt := getStatementBuilder().
		Insert(computeTaskOutputAssetsTable).
		Columns("compute_task_key", "compute_task_output_identifier", "position", "asset_kind", "asset_key").
		Select(
			getStatementBuilder().
				Select().
				Column("?", output.ComputeTaskKey).
				Column("?", output.ComputeTaskOutputIdentifier).
				Column("coalesce(max(position)+1, 0)").
				Column("?", output.AssetKind).
				Column("?", output.AssetKey).
				From(computeTaskOutputAssetsTable).
				Where(sq.Eq{"compute_task_key": output.ComputeTaskKey, "compute_task_output_identifier": output.ComputeTaskOutputIdentifier}),
		)

	return d.exec(stmt)
}

func (d *DBAL) GetComputeTaskOutputAssets(taskKey, identifier string) ([]*asset.ComputeTaskOutputAsset, error) {
	stmt := getStatementBuilder().
		Select("asset_kind", "asset_key").
		From(computeTaskOutputAssetsTable).
		Where(sq.Eq{"compute_task_key": taskKey, "compute_task_output_identifier": identifier}).
		OrderBy("position ASC")

	rows, err := d.query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	outputAssets := []*asset.ComputeTaskOutputAsset{}

	for rows.Next() {
		out := &asset.ComputeTaskOutputAsset{
			ComputeTaskKey:              taskKey,
			ComputeTaskOutputIdentifier: identifier,
		}

		err = rows.Scan(&out.AssetKind, &out.AssetKey)
		if err != nil {
			return nil, err
		}

		outputAssets = append(outputAssets, out)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return outputAssets, nil
}
