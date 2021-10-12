package handlers

import (
	"context"

	"github.com/owkin/orchestrator/lib/asset"
	libCommon "github.com/owkin/orchestrator/lib/common"
	"github.com/owkin/orchestrator/server/common"
	"github.com/owkin/orchestrator/server/common/logger"

	"github.com/owkin/orchestrator/server/standalone/interceptors"
)

// DataManagerServer is the gRPC facade to DataManager manipulation
type DataManagerServer struct {
	asset.UnimplementedDataManagerServiceServer
}

// NewDataManagerServer creates a gRPC server
func NewDataManagerServer() *DataManagerServer {
	return &DataManagerServer{}
}

// RegisterDataManager will persist new DataManagers
func (s *DataManagerServer) RegisterDataManager(ctx context.Context, d *asset.NewDataManager) (*asset.DataManager, error) {
	logger.Get(ctx).WithField("datamanager", d).Debug("Register DataManager")

	mspid, err := common.ExtractMSPID(ctx)
	if err != nil {
		return nil, err
	}
	services, err := interceptors.ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}

	dm, err := services.GetDataManagerService().RegisterDataManager(d, mspid)
	if err != nil {
		return nil, err
	}

	return dm, nil
}

// GetDataManager fetches a datamanager by its key
func (s *DataManagerServer) GetDataManager(ctx context.Context, params *asset.GetDataManagerParam) (*asset.DataManager, error) {
	services, err := interceptors.ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}
	return services.GetDataManagerService().GetDataManager(params.GetKey())
}

// QueryDataManagers returns a paginated list of all known datamanagers
func (s *DataManagerServer) QueryDataManagers(ctx context.Context, params *asset.QueryDataManagersParam) (*asset.QueryDataManagersResponse, error) {
	services, err := interceptors.ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}

	datamanagers, paginationToken, err := services.GetDataManagerService().QueryDataManagers(libCommon.NewPagination(params.PageToken, params.PageSize))
	if err != nil {
		return nil, err
	}

	return &asset.QueryDataManagersResponse{
		DataManagers:  datamanagers,
		NextPageToken: paginationToken,
	}, nil
}
