// Copyright 2020 Owkin Inc.
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

package event

import (
	"github.com/google/uuid"
	"github.com/owkin/orchestrator/lib/asset"
)

// Kind represent the kind of orchestration events
type Kind string

const (
	// AssetCreated is emitted when a new asset is created
	AssetCreated Kind = "asset_created"
	// AssetUpdated is emitted when an existing asset is updated
	AssetUpdated = "asset_updated"
	// AssetDisabled is emitted when an asset is disabled (ie: not accessible anymore)
	AssetDisabled = "asset_disabled"
)

// Event is an occurence of an orchestration event.
// It is triggered during orchestration and allows a consumer to react to the orchestration process.
type Event struct {
	ID        string     `json:"id"`
	AssetKind asset.Kind `json:"asset_kind"`
	AssetKey  string     `json:"asset_key"`
	EventKind Kind       `json:"event_kind"`
	Channel   string     `json:"channel"`
	// Metadata can hold asset specific data
	Metadata map[string]string `json:"metadata"`
}

// NewEvent creates a new orchestration event related to given asset.
func NewEvent(eventKind Kind, assetKey string, assetKind asset.Kind) *Event {
	return &Event{
		ID:        uuid.NewString(),
		AssetKind: assetKind,
		AssetKey:  assetKey,
		EventKind: eventKind,
	}
}

// WithMetadata embeds given metadata in the event.
func (e *Event) WithMetadata(meta map[string]string) *Event {
	e.Metadata = meta
	return e
}
