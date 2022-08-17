// server binary exposing a gRPC interface to manage distributed learning asset.
// It can run in either standalone or distributed mode.
// In standalone mode it handles all the logic while in distributed mode everything is delegated to a chaincode.
package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	// #nosec: profiling tool is exposed on a separate port
	_ "net/http/pprof"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
	"github.com/substra/orchestrator/server/common"
	"github.com/substra/orchestrator/server/distributed"
	"github.com/substra/orchestrator/server/standalone"
	"github.com/substra/orchestrator/utils"
	"golang.org/x/sync/errgroup"
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
		log.Fatal().Err(err).Msg("failed to create standalone server")
	}

	return server
}

func getStandaloneServer(params common.AppParameters, healthcheck *health.Server) common.Runnable {
	dbURL := common.MustGetEnv("DATABASE_URL")

	server, err := standalone.GetServer(dbURL, params, healthcheck)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create standalone server")
	}

	return server
}

func main() {
	var app common.Runnable
	var httpServer *http.Server
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
	if tlsOptions := common.GetTLSOptions(); tlsOptions != nil {
		serverOptions = append(serverOptions, tlsOptions)
	}

	orchestrationConfig := common.NewConfig(common.MustGetEnv("CHANNEL_CONFIG"))

	retryBudget := common.MustParseDuration(common.MustGetEnv("TX_RETRY_BUDGET"))

	params := common.AppParameters{
		GrpcOptions: serverOptions,
		Config:      orchestrationConfig,
		RetryBudget: retryBudget,
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGTERM)
	defer signal.Stop(interrupt)

	healthcheck := health.NewServer()

	if standaloneMode {
		app = getStandaloneServer(params, healthcheck)
	} else {
		app = getDistributedServer(params)
	}

	// Register reflection service
	reflection.Register(app.GetGrpcServer())

	// Register metrics
	grpc_prometheus.Register(app.GetGrpcServer())
	grpc_prometheus.EnableHandlingTimeHistogram()

	// Register healthcheck service
	healthpb.RegisterHealthServer(app.GetGrpcServer(), healthcheck)

	if metricsEnabled, _ := utils.GetenvBool("METRICS_ENABLED"); metricsEnabled {
		http.Handle("/metrics", promhttp.Handler())
	}

	httpServer = &http.Server{
		Addr: fmt.Sprintf(":%s", httpPort),
	}

	g, ctx := errgroup.WithContext(ctx)

	// Expose HTTP endpoints
	g.Go(func() error {
		err := httpServer.ListenAndServe()
		log.Info().Str("port", httpPort).Msg("HTTP server serving")
		if err != http.ErrServerClosed {
			log.Error().Err(err).Str("port", httpPort).Msg("failed to serve HTTP endpoints")
			return err
		}
		return nil
	})

	// Expose GRPC endpoints
	g.Go(func() error {
		listen, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
		if err != nil {
			log.Fatal().Err(err).Str("port", grpcPort).Msg("failed to listen")
		}
		log.Info().Str("address", listen.Addr().String()).Msg("gRPC server listening")
		return app.GetGrpcServer().Serve(listen)
	})

	select {
	case <-interrupt:
		log.Info().Msg("Received interruption signal")
		break
	case <-ctx.Done():
		break
	}

	log.Warn().Msg("Server shutting down")

	healthcheck.Shutdown()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer shutdownCancel()

	err := httpServer.Shutdown(shutdownCtx)
	if err != nil {
		log.Warn().Err(err).Msg("error during graceful shutdown of the HTTP server")
	}

	app.GetGrpcServer().GracefulStop()
	app.Stop()

	err = g.Wait()
	if err != nil {
		log.Fatal().Err(err).Msg("Server returned an error")
	}
}
