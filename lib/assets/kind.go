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

package assets

// Kind represent the type of assets handled by the orchestrator
type Kind = string

var (
	// NodeKind is the type of Node assets
	NodeKind Kind = "node"
	// ObjectiveKind is the type of Objective assets
	ObjectiveKind = "objective"
)
