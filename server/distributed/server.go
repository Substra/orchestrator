package distributed

import (
	"context"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/server/common"
	"github.com/owkin/orchestrator/server/common/logger"
	"github.com/owkin/orchestrator/server/common/trace"
	"github.com/owkin/orchestrator/server/distributed/wallet"
	"google.golang.org/grpc"
)

type AppServer struct {
	grpc          *grpc.Server
	ccInterceptor *Interceptor
}

func GetServer(networkConfig string, certificate string, key string, gatewayTimeout time.Duration, params common.AppParameters) (*AppServer, error) {
	wallet := wallet.New(certificate, key)
	config := config.FromFile(networkConfig)

	chaincodeInterceptor := NewInterceptor(config, wallet, gatewayTimeout)

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

	return &AppServer{server, chaincodeInterceptor}, nil
}

func (a *AppServer) GetGrpcServer() *grpc.Server {
	return a.grpc
}

func (a *AppServer) Stop() {
	a.grpc.Stop()
	a.ccInterceptor.Close()
}

// shouldRetry will trigger a retry on specific orchestration error.
func shouldRetry(err error) bool {
	st, ok := status.FromError(err)
	switch {
	case ok && st.Code == int32(status.Timeout):
		return true
	default:
		return false
	}
}

// isTimeoutRetry will return true if we are in a retry and the last error was a fabric timeout
func isFabricTimeoutRetry(ctx context.Context) bool {
	prevErr := common.GetLastError(ctx)
	if prevErr == nil {
		return false
	}

	st, ok := status.FromError(prevErr)
	return ok && st.Code == int32(status.Timeout)
}
