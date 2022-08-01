package distributed

import (
	"time"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/server/common"
	"github.com/owkin/orchestrator/server/common/trace"
	"github.com/owkin/orchestrator/server/distributed/adapters"
	"github.com/owkin/orchestrator/server/distributed/chaincode"
	"github.com/owkin/orchestrator/server/distributed/interceptors"
	"google.golang.org/grpc"
)

type AppServer struct {
	grpc                 *grpc.Server
	invocatorInterceptor *interceptors.InvocatorInterceptor
}

func GetServer(networkConfig string, certificate string, key string, gatewayTimeout time.Duration, params common.AppParameters) (*AppServer, error) {
	wallet := chaincode.NewWallet(certificate, key)
	config := config.FromFile(networkConfig)

	invocatorInterceptor := interceptors.NewInvocatorInterceptor(config, wallet, gatewayTimeout)

	channelInterceptor := common.NewChannelInterceptor(params.Config)
	MSPIDInterceptor, err := common.NewMSPIDInterceptor()
	if err != nil {
		return nil, err
	}

	retryInterceptor := common.NewRetryInterceptor(params.RetryBudget, shouldRetry)

	unaryInterceptor := grpc.ChainUnaryInterceptor(
		trace.InterceptRequestID,
		grpc_prometheus.UnaryServerInterceptor,
		common.UnaryServerLoggerInterceptor,
		common.UnaryServerRequestLogger,
		common.InterceptDistributedErrors,
		MSPIDInterceptor.UnaryServerInterceptor,
		channelInterceptor.UnaryServerInterceptor,
		retryInterceptor.Intercept,
		invocatorInterceptor.UnaryServerInterceptor,
	)

	chaincodeDataInterceptor := interceptors.NewChaincodeDataInterceptor(wallet, config)

	streamInterceptor := grpc.ChainStreamInterceptor(
		grpc_prometheus.StreamServerInterceptor,
		common.StreamServerLoggerInterceptor,
		common.StreamServerRequestLogger,
		MSPIDInterceptor.StreamServerInterceptor,
		channelInterceptor.StreamServerInterceptor,
		chaincodeDataInterceptor.StreamServerInterceptor,
	)

	serverOptions := append(params.GrpcOptions, unaryInterceptor, streamInterceptor)

	server := grpc.NewServer(serverOptions...)

	// Register application services
	asset.RegisterOrganizationServiceServer(server, adapters.NewOrganizationAdapter())
	asset.RegisterDataSampleServiceServer(server, adapters.NewDataSampleAdapter())
	asset.RegisterAlgoServiceServer(server, adapters.NewAlgoAdapter())
	asset.RegisterDataManagerServiceServer(server, adapters.NewDataManagerAdapter())
	asset.RegisterDatasetServiceServer(server, adapters.NewDatasetAdapter())
	asset.RegisterComputeTaskServiceServer(server, adapters.NewComputeTaskAdapter())
	asset.RegisterModelServiceServer(server, adapters.NewModelAdapter())
	asset.RegisterComputePlanServiceServer(server, adapters.NewComputePlanAdapter())
	asset.RegisterPerformanceServiceServer(server, adapters.NewPerformanceAdapter())
	asset.RegisterEventServiceServer(server, adapters.NewEventAdapter())
	asset.RegisterInfoServiceServer(server, adapters.NewInfoAdapter())
	asset.RegisterFailureReportServiceServer(server, adapters.NewFailureReportAdapter())

	return &AppServer{server, invocatorInterceptor}, nil
}

func (a *AppServer) GetGrpcServer() *grpc.Server {
	return a.grpc
}

func (a *AppServer) Stop() {
	a.grpc.Stop()
	a.invocatorInterceptor.Close()
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
