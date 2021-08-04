package dbal

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v4"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	orcerrors "github.com/owkin/orchestrator/lib/errors"
)

// AddObjective implements persistence.ObjectiveDBAL
func (d *DBAL) AddObjective(obj *asset.Objective) error {
	stmt := `insert into "objectives" ("id", "asset", "channel") values ($1, $2, $3)`
	_, err := d.tx.Exec(d.ctx, stmt, obj.GetKey(), obj, d.channel)

	return err
}

// GetObjective implements persistence.ObjectiveDBAL
func (d *DBAL) GetObjective(key string) (*asset.Objective, error) {
	row := d.tx.QueryRow(d.ctx, `select "asset" from "objectives" where id=$1 and channel=$2`, key, d.channel)

	objective := new(asset.Objective)
	err := row.Scan(&objective)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("objective not found: %w", orcerrors.ErrNotFound)
		}
		return nil, err
	}

	return objective, nil
}

// ObjectiveExists implements persistence.ObjectiveDBAL
func (d *DBAL) ObjectiveExists(key string) (bool, error) {
	row := d.tx.QueryRow(d.ctx, `select count(id) from "objectives" where id=$1 and channel=$2`, key, d.channel)

	var count int
	err := row.Scan(&count)

	return count == 1, err
}

// QueryObjectives implements persistence.ObjectiveDBAL
func (d *DBAL) QueryObjectives(p *common.Pagination) ([]*asset.Objective, common.PaginationToken, error) {
	offset, err := getOffset(p.Token)
	if err != nil {
		return nil, "", err
	}

	query := `select "asset" from "objectives" where channel=$3 order by created_at asc limit $1 offset $2`
	rows, err := d.tx.Query(d.ctx, query, p.Size+1, offset, d.channel)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var objectives []*asset.Objective
	var count int

	for rows.Next() {
		objective := new(asset.Objective)

		err = rows.Scan(&objective)
		if err != nil {
			return nil, "", err
		}

		objectives = append(objectives, objective)
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

	return objectives, bookmark, nil
}

// GetLeaderboard returns for an objective all its certified ComputeTask with ComputeTaskCategory: TEST_TASK with a done status
func (d *DBAL) GetLeaderboard(key string) (*asset.Leaderboard, error) {
	objective, err := d.GetObjective(key)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("objective not found: %w", orcerrors.ErrNotFound)
		}
		return nil, err
	}

	lb := &asset.Leaderboard{}

	lb.Objective = objective

	var boardItems []*asset.BoardItem

	query := `select c.asset->'algo' as algo,
	c.asset->'test'->>'objectiveKey'  as objective_key,
	c.asset->>'key' as compute_task_key,
	cast(p.asset->>'performanceValue' as double precision) as perf from "compute_tasks" c
	inner join performances p on p.asset->>'computeTaskKey' = c.asset->>'key'
	where c.asset->>'category' = 'TASK_TEST'
	and c.asset->>'status' = 'STATUS_DONE'
	and c.asset->'test'->>'certified' = 'true'
	and c.asset->'test'->>'objectiveKey' = $1`

	rows, err := d.tx.Query(d.ctx, query, key)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		boardItem := new(asset.BoardItem)

		err = rows.Scan(&boardItem.Algo, &boardItem.ObjectiveKey, &boardItem.ComputeTaskKey, &boardItem.Perf)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, fmt.Errorf("%w: No board item found", orcerrors.ErrNotFound)
			}
			return nil, fmt.Errorf("failed to scan BoardItem: %w", err)
		}
		boardItems = append(boardItems, boardItem)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	lb.BoardItems = boardItems

	return lb, nil
}
