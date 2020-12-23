package objective

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	persistenceHelper "github.com/substrafoundation/substra-orchestrator/lib/persistence/testing"
)

func TestRegistration(t *testing.T) {
	mockDB := new(persistenceHelper.MockDatabase)
	service := NewService(mockDB)

	objective := Objective{
		Key: "objKey",
	}

	mockDB.On("PutState", "objKey", mock.Anything).Return(nil).Once()

	service.RegisterObjective(&objective)
}

func TestQuery(t *testing.T) {
	mockDB := new(persistenceHelper.MockDatabase)
	service := NewService(mockDB)

	objective := Objective{
		Key:  "objKey",
		Name: "Test",
	}

	objBytes, err := json.Marshal(&objective)
	require.Nil(t, err)

	mockDB.On("GetState", "objKey").Return(objBytes, nil).Once()

	o, err := service.GetObjective("objKey")
	require.Nil(t, err)
	assert.Equal(t, o.Name, objective.Name)
}
