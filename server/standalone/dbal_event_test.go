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
	"testing"

	"github.com/Masterminds/squirrel"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/stretchr/testify/assert"
)

func TestEventFilterToQuery(t *testing.T) {
	cases := map[string]struct {
		filter        *asset.EventQueryFilter
		queryContains string
		params        []interface{}
	}{
		"empty":         {&asset.EventQueryFilter{}, "", nil},
		"single filter": {&asset.EventQueryFilter{AssetKey: "uuid"}, "event->>'assetKey' = $1", []interface{}{"uuid"}},
		"two filter": {
			&asset.EventQueryFilter{AssetKind: asset.AssetKind_ASSET_COMPUTE_TASK, EventKind: asset.EventKind_EVENT_ASSET_CREATED},
			"event->>'assetKind' = $1 AND event->>'eventKind' = $2",
			[]interface{}{asset.AssetKind_ASSET_COMPUTE_TASK.String(), asset.EventKind_EVENT_ASSET_CREATED.String()}},
		"three filter": {
			&asset.EventQueryFilter{AssetKey: "uuid", AssetKind: asset.AssetKind_ASSET_COMPUTE_PLAN, EventKind: asset.EventKind_EVENT_ASSET_UPDATED},
			"event->>'assetKey' = $1 AND event->>'assetKind' = $2 AND event->>'eventKind' = $3",
			[]interface{}{"uuid", asset.AssetKind_ASSET_COMPUTE_PLAN.String(), asset.EventKind_EVENT_ASSET_UPDATED.String()},
		},
	}

	pgDialect := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			builder := pgDialect.Select("event").From("events")
			builder = eventFilterToQuery(c.filter, builder)
			query, params, err := builder.ToSql()
			assert.NoError(t, err)
			assert.Contains(t, query, c.queryContains)
			assert.Equal(t, c.params, params)
		})
	}
}
