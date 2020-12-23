package objective

import (
	"encoding/json"

	"github.com/substrafoundation/substra-orchestrator/lib/persistence"
)

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

func (s *Service) RegisterObjective(o *Objective) error {
	b, err := json.Marshal(o)
	if err != nil {
		return err
	}

	s.db.PutState(o.GetKey(), b)
	return nil
}

func (s *Service) GetObjective(id string) (*Objective, error) {
	o := Objective{}

	b, err := s.db.GetState(id)
	if err != nil {
		return &o, err
	}

	err = json.Unmarshal(b, &o)
	return &o, err
}
