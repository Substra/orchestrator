package distributed

import (
	"context"

	"github.com/owkin/orchestrator/lib/asset"
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

	err = invocator.Call(method, d, response)

	return response, err
}

// UpdateDataManager will update the objective of an existing DataManager
func (s *DataManagerAdapter) UpdateDataManager(ctx context.Context, d *asset.DataManagerUpdateParam) (*asset.DataManagerUpdateResponse, error) {
	invocator, err := ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.datamanager:UpdateDataManager"

	err = invocator.Call(method, d, nil)

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

	err = invocator.Call(method, params, response)

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

	err = invocator.Call(method, params, response)

	return response, err
}
