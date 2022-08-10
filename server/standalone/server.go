package standalone

import (
	"errors"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/substra/orchestrator/lib/asset"
	"github.com/substra/orchestrator/server/common"
	commonInterceptors "github.com/substra/orchestrator/server/common/interceptors"
	"github.com/substra/orchestrator/server/standalone/dbal"
	"github.com/substra/orchestrator/server/standalone/handlers"
	"github.com/substra/orchestrator/server/standalone/interceptors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
)

type AppServer struct {
	grpc *grpc.Server
	amqp *common.Session
	db   *dbal.Database
}

func GetServer(dbURL string, rabbitDSN string, params common.AppParameters, healthcheck *health.Server) (*AppServer, error) {
	pgDB, err := dbal.InitDatabase(dbURL)
	if err != nil {
		return nil, err
	}

	session := common.NewSession(rabbitDSN)

	channelInterceptor := commonInterceptors.NewChannelInterceptor(params.Config)

	MSPIDInterceptor, err := commonInterceptors.NewMSPIDInterceptor()
	if err != nil {
		return nil, err
	}

	// providerInterceptor will wrap gRPC requests and inject a ServiceProvider in request's context
	providerInterceptor := interceptors.NewProviderInterceptor(pgDB, session, healthcheck)

	retryInterceptor := commonInterceptors.NewRetryInterceptor(params.RetryBudget, shouldRetry)

	unaryInterceptor := grpc.ChainUnaryInterceptor(
		commonInterceptors.InterceptRequestID,
		grpc_prometheus.UnaryServerInterceptor,
		commonInterceptors.UnaryServerLoggerInterceptor,
		commonInterceptors.UnaryServerRequestLogger,
		commonInterceptors.InterceptStandaloneErrors,
		MSPIDInterceptor.UnaryServerInterceptor,
		channelInterceptor.UnaryServerInterceptor,
		retryInterceptor.UnaryServerInterceptor,
		providerInterceptor.UnaryServerInterceptor,
	)

	dbConnInterceptor := interceptors.NewDatabaseConnInterceptor(dbURL)
	streamInterceptor := grpc.ChainStreamInterceptor(
		grpc_prometheus.StreamServerInterceptor,
		commonInterceptors.StreamServerLoggerInterceptor,
		commonInterceptors.StreamServerRequestLogger,
		MSPIDInterceptor.StreamServerInterceptor,
		channelInterceptor.StreamServerInterceptor,
		dbConnInterceptor.StreamServerInterceptor,
	)
	serverOptions := append(params.GrpcOptions, unaryInterceptor, streamInterceptor)

	server := grpc.NewServer(serverOptions...)

	// Register application services
	asset.RegisterOrganizationServiceServer(server, handlers.NewOrganizationServer())
	asset.RegisterDataSampleServiceServer(server, handlers.NewDataSampleServer())
	asset.RegisterAlgoServiceServer(server, handlers.NewAlgoServer())
	asset.RegisterDataManagerServiceServer(server, handlers.NewDataManagerServer())
	asset.RegisterDatasetServiceServer(server, handlers.NewDatasetServer())
	asset.RegisterComputeTaskServiceServer(server, handlers.NewComputeTaskServer())
	asset.RegisterModelServiceServer(server, handlers.NewModelServer())
	asset.RegisterComputePlanServiceServer(server, handlers.NewComputePlanServer())
	asset.RegisterPerformanceServiceServer(server, handlers.NewPerformanceServer())
	asset.RegisterEventServiceServer(server, handlers.NewEventServer())
	asset.RegisterInfoServiceServer(server, handlers.NewInfoServer())
	asset.RegisterFailureReportServiceServer(server, handlers.NewFailureReportServer())

	return &AppServer{
		grpc: server,
		amqp: session,
		db:   pgDB,
	}, nil
}

func (a *AppServer) GetGrpcServer() *grpc.Server {
	return a.grpc
}

func (a *AppServer) Stop() {
	a.grpc.Stop()
	a.db.Close()
	a.amqp.Close()
}

// shouldRetry is used as RetryInterceptor's checker function
// and allow a retry on transaction serialization failure.
func shouldRetry(err error) bool {
	var pgErr *pgconn.PgError

	return errors.As(err, &pgErr) && pgErr.Code == pgerrcode.SerializationFailure
}
