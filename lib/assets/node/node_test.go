package node

import (
	"testing"

	"github.com/stretchr/testify/mock"
	persistenceHelper "github.com/substrafoundation/substra-orchestrator/lib/persistence/testing"
)

func TestRegistration(t *testing.T) {
	mockDB := new(persistenceHelper.MockDatabase)
	node := Node{
		Id:       "uuid1",
		ModelKey: "test",
		Foo:      "bar",
	}

	mockDB.On("PutState", "uuid1", mock.Anything).Return(nil).Once()

	service := NewService(mockDB)

	service.RegisterNode(&node)
}
