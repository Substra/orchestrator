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
func (s *Server) QueryObjective(ctx context.Context, q *ObjectiveQuery) (*Objective, error) {
	return s.objectiveService.GetObjective(q.GetKey())
}
