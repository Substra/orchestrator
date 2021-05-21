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

package dbal

import (
	"github.com/owkin/orchestrator/lib/asset"
)

func (d *DBAL) AddPerformance(perf *asset.Performance) error {
	stmt := `insert into "performances" ("id", "asset", "channel") values ($1, $2, $3)`
	_, err := d.tx.Exec(stmt, perf.ComputeTaskKey, perf, d.channel)
	return err
}

func (d *DBAL) GetComputeTaskPerformance(key string) (*asset.Performance, error) {
	row := d.tx.QueryRow(`select asset from "performances" where id=$1 and channel=$2`, key, d.channel)

	perf := new(asset.Performance)
	err := row.Scan(perf)
	if err != nil {
		return nil, err
	}

	return perf, nil
}
