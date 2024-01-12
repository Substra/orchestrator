// server binary exposing a gRPC interface to manage distributed learning asset.
package main

import (
	"context"
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

	utils.InitLogging()

	serverOptions := []grpc.ServerOption{}
	if tlsOptions := common.GetTLSOptions(); tlsOptions != nil {
		serverOptions = append(serverOptions, tlsOptions)
	}
	serverOptions = append(serverOptions, common.GetKeepAliveOptions())

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

	app = getStandaloneServer(params, healthcheck)

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
		Addr:              fmt.Sprintf(":%s", httpPort),
		ReadHeaderTimeout: 2 * time.Second,
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
		log.Error().Err(err).Msg("Server returned an error")
	}
}
