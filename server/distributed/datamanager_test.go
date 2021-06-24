package distributed

import (
	"context"
	"testing"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRegisterDatamanager(t *testing.T) {
	adapter := NewDataManagerAdapter()

	newObj := &asset.NewDataManager{
		Key: "uuid",
	}

	newCtx := context.TODO()
	invocator := &mockedInvocator{}

	invocator.On("Call", "orchestrator.datamanager:RegisterDataManager", newObj, &asset.DataManager{}).
		Once().
		Run(func(args mock.Arguments) {
			dm := args.Get(2).(*asset.DataManager)
			dm.Key = "uuid"
			dm.Owner = "test"
		}).
		Return(nil)

	ctx := context.WithValue(newCtx, ctxInvocatorKey, invocator)

	dm, err := adapter.RegisterDataManager(ctx, newObj)
	assert.NoError(t, err, "Registration should pass")

	assert.Equal(t, "uuid", dm.Key)
	assert.Equal(t, "test", dm.Owner)
}
