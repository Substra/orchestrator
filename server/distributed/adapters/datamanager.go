package adapters

import (
	"context"
	"strings"

	"github.com/substra/orchestrator/lib/asset"
	"github.com/substra/orchestrator/lib/errors"
	"github.com/substra/orchestrator/server/distributed/interceptors"
)

// DataManagerAdapter is a grpc server exposing the same dataManager interface,
// but relies on a remote chaincode to actually manage the asset.
type DataManagerAdapter struct {
	asset.UnimplementedDataManagerServiceServer
}

// NewDataManagerAdapter creates a Server
func NewDataManagerAdapter() *DataManagerAdapter {
	return &DataManagerAdapter{}
}

// RegisterDataManager will persist new DataManagers
func (s *DataManagerAdapter) RegisterDataManager(ctx context.Context, d *asset.NewDataManager) (*asset.DataManager, error) {
	invocator, err := interceptors.ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.datamanager:RegisterDataManager"

	response := &asset.DataManager{}

	err = invocator.Call(ctx, method, d, response)

	if err != nil && isFabricTimeoutRetry(ctx) && strings.Contains(err.Error(), errors.ErrConflict) {
		// In this very specific case we are in a retry context after a timeout.
		// We can assume that the previous request succeeded and created the asset.
		// So we convert the error in a success response.
		err = invocator.Call(ctx, "orchestrator.datamanager:GetDataManager", &asset.GetDataManagerParam{Key: d.Key}, response)
		return response, err
	}

	return response, err
}

// GetDataManager fetches a datamanager by its key
func (s *DataManagerAdapter) GetDataManager(ctx context.Context, params *asset.GetDataManagerParam) (*asset.DataManager, error) {
	invocator, err := interceptors.ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.datamanager:GetDataManager"
	response := &asset.DataManager{}

	err = invocator.Call(ctx, method, params, response)

	return response, err
}

// QueryDataManagers returns a paginated list of all known datamanagers
func (s *DataManagerAdapter) QueryDataManagers(ctx context.Context, params *asset.QueryDataManagersParam) (*asset.QueryDataManagersResponse, error) {
	invocator, err := interceptors.ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.datamanager:QueryDataManagers"
	response := &asset.QueryDataManagersResponse{}

	err = invocator.Call(ctx, method, params, response)

	return response, err
}

// UpdateDataManager will update a DataManager from the state
func (s *DataManagerAdapter) UpdateDataManager(ctx context.Context, params *asset.UpdateDataManagerParam) (*asset.UpdateDataManagerResponse, error) {
	invocator, err := interceptors.ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.datamanager:UpdateDataManager"

	response := &asset.UpdateDataManagerResponse{}

	err = invocator.Call(ctx, method, params, nil)

	return response, err
}

// ArchiveDataManager will archive a DataManager from the state
func (s *DataManagerAdapter) ArchiveDataManager(ctx context.Context, param *asset.ArchiveDataManagerParam) (*asset.ArchiveDataManagerResponse, error) {
	invocator, err := interceptors.ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.datamanager:ArchiveDataManager"

	response := &asset.ArchiveDataManagerResponse{}

	err = invocator.Call(ctx, method, param, nil)

	return response, err
}
