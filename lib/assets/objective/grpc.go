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

package objective

import (
	"log"

	"golang.org/x/net/context"
)

// Server is the gRPC facade to Objective manipulation
type Server struct {
	UnimplementedObjectiveServiceServer
	objectiveService *Service
}

// NewServer creates a grpc server
func NewServer(service *Service) *Server {
	return &Server{objectiveService: service}
}

// RegisterObjective will persiste a new objective
func (s *Server) RegisterObjective(ctx context.Context, o *Objective) (*Objective, error) {
	log.Println(o)
	log.Printf("objective: %s, %s, %s", o.GetKey(), o.GetName(), o.GetTestDataset())

	err := s.objectiveService.RegisterObjective(o)
	return o, err
}

// QueryObjective fetches an objective by its key
func (s *Server) QueryObjective(ctx context.Context, key string) (*Objective, error) {
	return s.objectiveService.GetObjective(key)
}
