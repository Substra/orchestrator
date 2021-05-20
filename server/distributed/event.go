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

package distributed

import (
	"context"

	"github.com/owkin/orchestrator/lib/asset"
)

// EventAdapter is a grpc server exposing the same Event interface as in standalone,
// but relies on a remote chaincode to actually manage the asset.
type EventAdapter struct {
	asset.UnimplementedEventServiceServer
}

// NewEventAdapter creates a Server
func NewEventAdapter() *EventAdapter {
	return &EventAdapter{}
}

func (a *EventAdapter) QueryEvents(ctx context.Context, query *asset.QueryEventsParam) (*asset.QueryEventsResponse, error) {
	invocator, err := ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.event:QueryEvents"

	response := &asset.QueryEventsResponse{}

	err = invocator.Call(method, query, response)

	return response, err
}
