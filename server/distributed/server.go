package distributed

import (
	"strings"

	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/owkin/orchestrator/lib/asset"
	orcerrors "github.com/owkin/orchestrator/lib/errors"
	"github.com/owkin/orchestrator/server/common"
	"github.com/owkin/orchestrator/server/common/logger"
	"github.com/owkin/orchestrator/server/common/trace"
	"github.com/owkin/orchestrator/server/distributed/wallet"
	"google.golang.org/grpc"
)

type AppServer struct {
	grpc *grpc.Server
}

func GetServer(networkConfig string, certificate string, key string, params common.AppParameters) (*AppServer, error) {
	wallet := wallet.New(certificate, key)
	config := config.FromFile(networkConfig)

	chaincodeInterceptor, err := NewInterceptor(config, wallet)
	if err != nil {
		return nil, err

	}

	channelInterceptor := common.NewChannelInterceptor(params.Config)

	retryInterceptor := common.NewRetryInterceptor(params.RetryBudget, shouldRetry)

	interceptor := grpc.ChainUnaryInterceptor(
		trace.InterceptRequestID,
		logger.AddLogger,
		common.LogRequest,
		common.InterceptDistributedErrors,
		common.InterceptMSPID,
		channelInterceptor.InterceptChannel,
		retryInterceptor.Intercept,
		chaincodeInterceptor.Intercept,
	)

	serverOptions := append(params.GrpcOptions, interceptor)

	server := grpc.NewServer(serverOptions...)

	// Register application services
	asset.RegisterNodeServiceServer(server, NewNodeAdapter())
	asset.RegisterObjectiveServiceServer(server, NewObjectiveAdapter())
	asset.RegisterDataSampleServiceServer(server, NewDataSampleAdapter())
	asset.RegisterAlgoServiceServer(server, NewAlgoAdapter())
	asset.RegisterDataManagerServiceServer(server, NewDataManagerAdapter())
	asset.RegisterDatasetServiceServer(server, NewDatasetAdapter())
	asset.RegisterComputeTaskServiceServer(server, NewComputeTaskAdapter())
	asset.RegisterModelServiceServer(server, NewModelAdapter())
	asset.RegisterComputePlanServiceServer(server, NewComputePlanAdapter())
	asset.RegisterPerformanceServiceServer(server, NewPerformanceAdapter())
	asset.RegisterEventServiceServer(server, NewEventAdapter())

	return &AppServer{server}, nil
}

func (a *AppServer) GetGrpcServer() *grpc.Server {
	return a.grpc
}

func (a *AppServer) Stop() {
	a.grpc.Stop()
}

// shouldRetry will trigger a retry on specific orchestration error.
func shouldRetry(err error) bool {
	msg := err.Error()
	switch {
	case strings.Contains(msg, orcerrors.ErrReferenceNotFound.Error()):
		// Reference not found might be due to an out of sync ledger
		return true
	case strings.Contains(msg, orcerrors.ErrNotFound.Error()):
		// Asset not found might be due to an out of sync ledger
		return true
	case strings.Contains(msg, orcerrors.ErrIncompatibleTaskStatus.Error()):
		// Task status mismatch might be due to an out of sync ledger
		return true
	default:
		return false
	}
}
