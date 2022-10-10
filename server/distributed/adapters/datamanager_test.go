package adapters

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/substra/orchestrator/lib/asset"
	"github.com/substra/orchestrator/lib/errors"
	commonInterceptors "github.com/substra/orchestrator/server/common/interceptors"
	"github.com/substra/orchestrator/server/distributed/chaincode"
	"github.com/substra/orchestrator/server/distributed/interceptors"
	"github.com/substra/orchestrator/utils"
)

func TestRegisterDatamanager(t *testing.T) {
	adapter := NewDataManagerAdapter()

	newObj := &asset.NewDataManager{
		Key: "uuid",
	}

	newCtx := context.TODO()
	invocator := &chaincode.MockInvocator{}

	invocator.On("Call", utils.AnyContext, "orchestrator.datamanager:RegisterDataManager", newObj, &asset.DataManager{}).
		Once().
		Run(func(args mock.Arguments) {
			dm := args.Get(3).(*asset.DataManager)
			dm.Key = "uuid"
			dm.Owner = "test"
		}).
		Return(nil)

	ctx := interceptors.WithInvocator(newCtx, invocator)

	dm, err := adapter.RegisterDataManager(ctx, newObj)
	assert.NoError(t, err, "Registration should pass")

	assert.Equal(t, "uuid", dm.Key)
	assert.Equal(t, "test", dm.Owner)
}

func TestHandleDatamanagerConflictAfterTimeout(t *testing.T) {
	adapter := NewDataManagerAdapter()

	newObj := &asset.NewDataManager{
		Key: "uuid",
	}

	newCtx := commonInterceptors.WithLastError(context.Background(), FabricTimeout)
	invocator := &chaincode.MockInvocator{}

	invocator.On("Call", utils.AnyContext, "orchestrator.datamanager:RegisterDataManager", newObj, &asset.DataManager{}).
		Once().
		Return(errors.NewError(errors.ErrConflict, "test"))
	invocator.On("Call", utils.AnyContext, "orchestrator.datamanager:GetDataManager", &asset.GetDataManagerParam{Key: newObj.Key}, &asset.DataManager{}).
		Once().
		Return(nil)

	ctx := interceptors.WithInvocator(newCtx, invocator)

	_, err := adapter.RegisterDataManager(ctx, newObj)
	assert.NoError(t, err, "Registration should pass")
}

func TestUpdateDataManager(t *testing.T) {
	adapter := NewDataManagerAdapter()

	updatedA := &asset.UpdateDataManagerParam{
		Key:  "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83",
		Name: "Updated datamanager name",
	}

	newCtx := context.TODO()
	invocator := &chaincode.MockInvocator{}

	invocator.On("Call", utils.AnyContext, "orchestrator.datamanager:UpdateDataManager", updatedA, nil).Return(nil)

	ctx := interceptors.WithInvocator(newCtx, invocator)

	_, err := adapter.UpdateDataManager(ctx, updatedA)
	assert.NoError(t, err, "Update should pass")
}

func TestArchiveDataManager(t *testing.T) {
	adapter := NewDataManagerAdapter()

	archiveDM := &asset.ArchiveDataManagerParam{
		Key:      "4c67ad88-309a-48b4-8bc4-c2e2c1a87a83",
		Archived: true,
	}

	newCtx := context.TODO()
	invocator := &chaincode.MockInvocator{}

	invocator.On("Call", utils.AnyContext, "orchestrator.datamanager:ArchiveDataManager", archiveDM, nil).Return(nil)

	ctx := interceptors.WithInvocator(newCtx, invocator)

	_, err := adapter.ArchiveDataManager(ctx, archiveDM)
	assert.NoError(t, err, "Archive should pass")
}
