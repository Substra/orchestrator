package handlers

import (
	"context"

	"github.com/substra/orchestrator/lib/asset"
	commonInterceptors "github.com/substra/orchestrator/server/common/interceptors"
	"github.com/substra/orchestrator/server/standalone/interceptors"
)

// FailureReportServer is the gRPC facade to FailureReport manipulation
type FailureReportServer struct {
	asset.UnimplementedFailureReportServiceServer
}

// NewFailureReportServer creates a gRPC server
func NewFailureReportServer() *FailureReportServer {
	return &FailureReportServer{}
}

func (s *FailureReportServer) RegisterFailureReport(ctx context.Context, newFailureReport *asset.NewFailureReport) (*asset.FailureReport, error) {
	mspid, err := commonInterceptors.ExtractMSPID(ctx)
	if err != nil {
		return nil, err
	}
	services, err := interceptors.ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}

	return services.GetFailureReportService().RegisterFailureReport(newFailureReport, mspid)
}

func (s *FailureReportServer) GetFailureReport(ctx context.Context, in *asset.GetFailureReportParam) (*asset.FailureReport, error) {
	services, err := interceptors.ExtractProvider(ctx)
	if err != nil {
		return nil, err
	}
	return services.GetFailureReportService().GetFailureReport(in.AssetKey)
}
