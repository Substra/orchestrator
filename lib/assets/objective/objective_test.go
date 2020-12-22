package objective

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/substrafoundation/substra-orchestrator/lib/persistence"
	persistenceHelper "github.com/substrafoundation/substra-orchestrator/lib/persistence/testing"
	"golang.org/x/net/context"
)

func getServer(db persistence.Database) *Server {
	factory := func(_ interface{}) (persistence.Database, error) {
		return db, nil
	}

	return &Server{dbFactory: factory}
}

func TestRegistration(t *testing.T) {
	mockDB := new(persistenceHelper.MockDatabase)

	objective := Objective{
		Key: "objKey",
	}

	mockDB.On("PutState", "objKey", mock.Anything).Return(nil).Once()

	server := getServer(mockDB)
	ctx := new(context.Context)
	server.RegisterObjective(*ctx, &objective)
}

func TestQuery(t *testing.T) {
	mockDB := new(persistenceHelper.MockDatabase)

	objective := Objective{
		Key:  "objKey",
		Name: "Test",
	}

	objBytes, err := json.Marshal(objective)
	require.Nil(t, err)

	mockDB.On("GetState", "objKey").Return(objBytes, nil).Once()

	server := getServer(mockDB)

	ctx := new(context.Context)
	o, err := server.QueryObjective(*ctx, &ObjectiveQuery{Key: "objKey"})
	require.Nil(t, err)
	assert.Equal(t, o.Name, objective.Name)
}
