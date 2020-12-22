package node

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/substrafoundation/substra-orchestrator/lib/persistence"
	persistenceHelper "github.com/substrafoundation/substra-orchestrator/lib/persistence/testing"
	"golang.org/x/net/context"
)

func TestRegistration(t *testing.T) {
	mockDB := new(persistenceHelper.MockDatabase)
	factory := func(_ interface{}) (persistence.Database, error) {
		return mockDB, nil
	}

	node := Node{
		Id:       "uuid1",
		ModelKey: "test",
		Foo:      "bar",
	}

	mockDB.On("PutState", "uuid1", mock.Anything).Return(nil).Once()

	server := &Server{dbFactory: factory}

	ctx := new(context.Context)
	server.RegisterNode(*ctx, &node)
}
