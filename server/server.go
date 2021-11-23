// server binary exposing a gRPC interface to manage distributed learning asset.
// It can run in either standalone or distributed mode.
// In standalone mode it handle all the logic while in distributed mode everything is delegated to a chaincode.
package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"

	// #nosec: profiling tool is exposed on a separate port
	_ "net/http/pprof"

	"github.com/go-playground/log/v7"
	"github.com/owkin/orchestrator/server/common"
	"github.com/owkin/orchestrator/server/distributed"
	"github.com/owkin/orchestrator/server/standalone"
	"github.com/owkin/orchestrator/utils"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

const httpPort = "8484"
const grpcPort = "9000"

func getDistributedServer(params common.AppParameters) common.Runnable {
	networkConfig := common.MustGetEnv("NETWORK_CONFIG")
	certificate := common.MustGetEnv("FABRIC_CERT")
	key := common.MustGetEnv("FABRIC_KEY")
	gatewayTimeout := common.MustParseDuration(common.MustGetEnv("FABRIC_GATEWAY_TIMEOUT"))

	server, err := distributed.GetServer(networkConfig, certificate, key, gatewayTimeout, params)
	if err != nil {
		log.WithError(err).Fatal("failed to create standalone server")
	}

	return server
}

func getStandaloneServer(params common.AppParameters, healthcheck *health.Server) common.Runnable {
	dbURL := common.MustGetEnv("DATABASE_URL")
	rabbitDSN := common.MustGetEnv("AMQP_DSN")

	server, err := standalone.GetServer(dbURL, rabbitDSN, params, healthcheck)
	if err != nil {
		log.WithError(err).Fatal("failed to create standalone server")
	}

	return server
}

func main() {
	var standaloneMode = false

	utils.InitLogging()

	flag.BoolVar(&standaloneMode, "standalone", true, "Run the server in standalone mode")
	flag.BoolVar(&standaloneMode, "s", true, "Run the server in standalone mode (shorthand)")

	flag.Parse()

	mode, ok := common.GetEnv("MODE")
	if ok {
		switch mode {
		case "distributed":
			standaloneMode = false
		case "standalone":
			standaloneMode = true
		}
	}

	serverOptions := []grpc.ServerOption{}
	if tlsOptions := getTLSOptions(); tlsOptions != nil {
		serverOptions = append(serverOptions, tlsOptions)
	}

	orchestrationConfig := common.NewConfig(common.MustGetEnv("CHANNEL_CONFIG"))

	retryBudget := common.MustParseDuration(common.MustGetEnv("TX_RETRY_BUDGET"))

	params := common.AppParameters{
		GrpcOptions: serverOptions,
		Config:      orchestrationConfig,
		RetryBudget: retryBudget,
	}

	healthcheck := health.NewServer()

	var app common.Runnable
	if standaloneMode {
		app = getStandaloneServer(params, healthcheck)
	} else {
		app = getDistributedServer(params)
	}
	defer app.Stop()

	// Register reflection service
	reflection.Register(app.GetGrpcServer())

	// Register healthcheck service
	healthpb.RegisterHealthServer(app.GetGrpcServer(), healthcheck)

	if metricsEnabled, _ := utils.GetenvBool("METRICS_ENABLED"); metricsEnabled {
		http.Handle("/metrics", promhttp.Handler())
	}

	// Expose HTTP endpoints
	go func() {
		err := http.ListenAndServe(fmt.Sprintf(":%s", httpPort), nil)
		if err != nil {
			log.WithError(err).WithField("port", httpPort).Error("failed to serve HTTP endpoints")
		}
	}()

	listen, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		log.WithError(err).WithField("port", grpcPort).Fatal("failed to listen")
	}

	log.WithField("address", listen.Addr().String()).Info("gRPC server listening")
	if err := app.GetGrpcServer().Serve(listen); err != nil {
		log.WithError(err).WithField("port", grpcPort).Fatal("failed to serve gRPC endpoints")
	}
}
