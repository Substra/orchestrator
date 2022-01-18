package distributed

import (
	"context"
	"strings"

	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/lib/errors"
)

// FailureReportAdapter is a gRPC server exposing the same FailureReport interface,
// but relies on a remote chaincode to actually manage the asset.
type FailureReportAdapter struct {
	asset.UnimplementedFailureReportServiceServer
}

// NewFailureReportAdapter creates a Server
func NewFailureReportAdapter() *FailureReportAdapter {
	return &FailureReportAdapter{}
}

func (a *FailureReportAdapter) RegisterFailureReport(ctx context.Context, newFailureReport *asset.NewFailureReport) (*asset.FailureReport, error) {
	invocator, err := ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.failurereport:RegisterFailureReport"

	failureReport := &asset.FailureReport{}

	err = invocator.Call(ctx, method, newFailureReport, failureReport)

	if err != nil && isFabricTimeoutRetry(ctx) && strings.Contains(err.Error(), errors.ErrConflict) {
		// In this very specific case we are in a retry context after a timeout.
		// We can assume that the previous request succeeded and created the asset.
		// So we convert the error in a success response.
		err = invocator.Call(ctx, "orchestrator.failurereport:GetFailureReport", &asset.GetFailureReportParam{ComputeTaskKey: newFailureReport.ComputeTaskKey}, failureReport)
		return failureReport, err
	}

	return failureReport, err
}

func (a *FailureReportAdapter) GetFailureReport(ctx context.Context, param *asset.GetFailureReportParam) (*asset.FailureReport, error) {
	invocator, err := ExtractInvocator(ctx)
	if err != nil {
		return nil, err
	}
	method := "orchestrator.failurereport:GetFailureReport"

	failureReport := &asset.FailureReport{}

	err = invocator.Call(ctx, method, param, failureReport)

	return failureReport, err
}
