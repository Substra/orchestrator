package objective

import (
	"encoding/json"

	"github.com/substrafoundation/substra-orchestrator/lib/persistence"
)

const resource = "objectives"

// API defines the methods to act on Objectives
type API interface {
	RegisterObjective(*Objective) error
	GetObjective(string) (*Objective, error)
}

// Service is the objective manipulation entry point
// it implements the API interface
type Service struct {
	db persistence.Database
}

// NewService will create a new service with given persistence layer
func NewService(db persistence.Database) *Service {
	return &Service{db: db}
}

// RegisterObjective persist an objective
func (s *Service) RegisterObjective(o *Objective) error {
	b, err := json.Marshal(o)
	if err != nil {
		return err
	}

	s.db.PutState(resource, o.GetKey(), b)
	return nil
}

// GetObjective retrieves an objective by its ID
func (s *Service) GetObjective(id string) (*Objective, error) {
	o := Objective{}

	b, err := s.db.GetState(resource, id)
	if err != nil {
		return &o, err
	}

	err = json.Unmarshal(b, &o)
	return &o, err
}
