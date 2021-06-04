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

package handlers

import (
	"context"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/common"
	"github.com/owkin/orchestrator/server/standalone/interceptors"
)

// EventServer is the gRPC facade to Model manipulation
type EventServer struct {
	asset.UnimplementedEventServiceServer
}

// NewEventServer creates a grpc server
func NewEventServer() *EventServer {
	return &EventServer{}
}

func (s *EventServer) QueryEvents(ctx context.Context, params *asset.QueryEventsParam) (*asset.QueryEventsResponse, error) {
	services, err := interceptors.ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}

	events, paginationToken, err := services.GetEventService().QueryEvents(common.NewPagination(params.PageToken, params.PageSize), params.Filter)
	if err != nil {
		return nil, err
	}

	return &asset.QueryEventsResponse{
		Events:        events,
		NextPageToken: paginationToken,
	}, nil
}
