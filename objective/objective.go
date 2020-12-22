package objective

import (
	"log"

	"golang.org/x/net/context"
)

var objs []Objective

// Server is exported
type Server struct{}

// RegisterObjective is exported
func (s *Server) RegisterObjective(ctx context.Context, o *Objective) (*Objective, error) {
	log.Println(o)
	log.Printf("objective: %s, %s, %s", o.GetKey(), o.GetName(), o.GetTestDataset())
	objs = append(objs, *o)
	return o, nil
}

// QueryObjective is exported
func (s *Server) QueryObjective(ctx context.Context, o *Objective) (*Objective, error) {
	return &objs[0], nil
}
