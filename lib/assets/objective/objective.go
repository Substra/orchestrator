package objective

import (
	"encoding/json"
	"log"

	"github.com/substrafoundation/substra-orchestrator/lib/persistence"
	"golang.org/x/net/context"
)

var objs []Objective

// Server is exported
type Server struct {
	UnimplementedObjectiveServiceServer
	dbFactory persistence.Factory
}

func NewServer(dbFactory persistence.Factory) *Server {
	return &Server{dbFactory: dbFactory}
}

// RegisterObjective is exported
func (s *Server) RegisterObjective(ctx context.Context, o *Objective) (*Objective, error) {
	db, err := s.dbFactory(ctx)
	if err != nil {
		log.Printf("Cannot derive DB from context: %v\n", err)
		return nil, err
	}

	log.Println(o)
	log.Printf("objective: %s, %s, %s", o.GetKey(), o.GetName(), o.GetTestDataset())

	objBytes, err := json.Marshal(o)
	if err != nil {
		return nil, err
	}

	err = db.PutState(o.Key, objBytes)
	return o, err
}

// QueryObjective is exported
func (s *Server) QueryObjective(ctx context.Context, q *ObjectiveQuery) (*Objective, error) {
	db, err := s.dbFactory(ctx)
	if err != nil {
		log.Printf("Cannot derive DB from context: %v\n", err)
		return nil, err
	}

	r, err := db.GetState(q.Key)
	if err != nil {
		log.Printf("Failed to fetch from db: %v\n", err)
		return nil, err
	}

	obj := new(Objective)
	err = json.Unmarshal(r, obj)
	return obj, err
}
