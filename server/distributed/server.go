package distributed

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/owkin/orchestrator/lib/asset"
	"github.com/owkin/orchestrator/server/common"
	"github.com/owkin/orchestrator/server/common/logger"
	"github.com/owkin/orchestrator/server/distributed/wallet"
	"google.golang.org/grpc"
)

type AppServer struct {
	grpc *grpc.Server
}

func GetServer(networkConfig string, certificate string, key string, additionalOptions []grpc.ServerOption, orcConf *common.OrchestratorConfiguration) (*AppServer, error) {
	wallet := wallet.New(certificate, key)
	config := config.FromFile(networkConfig)

	chaincodeInterceptor, err := NewInterceptor(config, wallet)
	if err != nil {
		return nil, err

	}

	channelInterceptor := common.NewChannelInterceptor(orcConf)

	interceptor := grpc.ChainUnaryInterceptor(
		logger.AddLogger,
		common.LogRequest,
		common.InterceptDistributedErrors,
		common.InterceptMSPID,
		channelInterceptor.InterceptChannel,
		chaincodeInterceptor.Intercept,
	)

	serverOptions := append(additionalOptions, interceptor)

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
