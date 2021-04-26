// Copyright 2021 Owkin Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package standalone

import (
	"database/sql"
	"strconv"

	"github.com/Masterminds/squirrel"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
)

// AddComputeTask stores a new ComputeTask in DB
func (d *DBAL) AddComputeTask(t *asset.ComputeTask) error {
	stmt := `insert into "compute_tasks" ("id", "asset", "channel") values ($1, $2, $3)`
	_, err := d.tx.Exec(stmt, t.GetKey(), t, d.channel)
	return err
}

// UpdateComputeTask updates an existing task
func (d *DBAL) UpdateComputeTask(t *asset.ComputeTask) error {
	stmt := `update "compute_tasks" set asset=$3 where id=$1 and channel=$2`
	_, err := d.tx.Exec(stmt, t.GetKey(), d.channel, t)
	return err
}

// ComputeTaskExists returns true if a task with the given ID exists
func (d *DBAL) ComputeTaskExists(key string) (bool, error) {
	row := d.tx.QueryRow(`select count(id) from "compute_tasks" where id=$1 and channel=$2`, key, d.channel)

	var count int
	err := row.Scan(&count)

	return count == 1, err
}

// GetComputeTask returns a single task by its key
func (d *DBAL) GetComputeTask(key string) (*asset.ComputeTask, error) {
	row := d.tx.QueryRow(`select asset from "compute_tasks" where id=$1 and channel=$2`, key, d.channel)

	task := new(asset.ComputeTask)
	err := row.Scan(task)
	if err != nil {
		return nil, err
	}

	return task, nil
}

// GetComputeTasks returns all compute tasks identified by the provided IDs.
// It should not be used where pagination is expected!
func (d *DBAL) GetComputeTasks(keys []string) ([]*asset.ComputeTask, error) {
	pgDialect := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	query, args, err := pgDialect.Select("asset").
		From("compute_tasks").
		Where(squirrel.Eq{"channel": d.channel, "id": keys}).ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := d.tx.Query(query, args...)
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

	return tasks, nil
}

// GetComputeTaskChildren returns the children of the task identified by the given key
func (d *DBAL) GetComputeTaskChildren(key string) ([]*asset.ComputeTask, error) {
	rows, err := d.tx.Query(`select asset from "compute_tasks" where asset->'parentTaskKeys' ? $1 and channel=$2`, key, d.channel)
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

	return tasks, nil
}

// QueryComputeTasks returns a paginated and filtered list of tasks.
func (d *DBAL) QueryComputeTasks(p *common.Pagination, filter *asset.TaskQueryFilter) ([]*asset.ComputeTask, common.PaginationToken, error) {
	var rows *sql.Rows
	var err error

	offset, err := getOffset(p.Token)
	if err != nil {
		return nil, "", err
	}

	pgDialect := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	builder := pgDialect.Select("asset").
		From("compute_tasks").
		Where(squirrel.Eq{"channel": d.channel}).
		OrderByClause("created_at ASC").
		Offset(uint64(offset)).
		Limit(uint64(p.Size + 1))

	builder = filterToQuery(filter, builder)

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, "", err
	}

	rows, err = d.tx.Query(query, args...)
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

	bookmark := ""
	if count == int(p.Size) && rows.Next() {
		// there is more to fetch
		bookmark = strconv.Itoa(offset + count)
	}

	return tasks, bookmark, nil
}

// filterToQuery convert as filter into query string and param list
func filterToQuery(filter *asset.TaskQueryFilter, builder squirrel.SelectBuilder) squirrel.SelectBuilder {
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

	return builder
}
