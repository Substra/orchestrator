package distributed

import (
	"context"
	"strings"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/errors"
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
	invocator, err := ExtractInvocator(ctx)
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

// UpdateDataManager will update the objective of an existing DataManager
func (s *DataManagerAdapter) UpdateDataManager(ctx context.Context, d *asset.DataManagerUpdateParam) (*asset.DataManagerUpdateResponse, error) {
	invocator, err := ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.datamanager:UpdateDataManager"

	err = invocator.Call(ctx, method, d, nil)

	return &asset.DataManagerUpdateResponse{}, err
}

// GetDataManager fetches a datamanager by its key
func (s *DataManagerAdapter) GetDataManager(ctx context.Context, params *asset.GetDataManagerParam) (*asset.DataManager, error) {
	invocator, err := ExtractInvocator(ctx)
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
	invocator, err := ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.datamanager:QueryDataManagers"
	response := &asset.QueryDataManagersResponse{}

	err = invocator.Call(ctx, method, params, response)

	return response, err
}
