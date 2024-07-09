package main

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"github.com/arxon31/metrics-collector/internal/server/controller/rpc"
	"github.com/arxon31/metrics-proto/pkg/protobuf/metrics"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"os/signal"
	"syscall"

	"github.com/arxon31/metrics-collector/internal/encrypting"

	"github.com/arxon31/metrics-collector/pkg/logger"

	"github.com/arxon31/metrics-collector/internal/server/service/failover"

	"golang.org/x/sync/errgroup"

	"github.com/go-chi/chi/v5"

	"github.com/arxon31/metrics-collector/internal/repository"
	"github.com/arxon31/metrics-collector/internal/server/config"
	controllers "github.com/arxon31/metrics-collector/internal/server/controller/rest"
	"github.com/arxon31/metrics-collector/internal/server/service/pinger"
	"github.com/arxon31/metrics-collector/internal/server/service/provider"
	"github.com/arxon31/metrics-collector/internal/server/service/storage"
	"github.com/arxon31/metrics-collector/pkg/httpserver"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	exitCode := run()
	if exitCode != 0 {
		log.Fatal("exited with code", exitCode)
	}
}

func run() int {
	logger.Logger.Infoln("starting server")
	logger.Logger.Info(fmt.Sprintf("version: %s, build time: %s, build commit: %s", buildVersion, buildDate, buildCommit))

	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cfg, err := config.NewServerConfig()
	if err != nil {
		logger.Logger.Fatalf("failed to parse a config due to error: %v", err)
	}

	repo, err := repository.New(cfg.DBString)
	if err != nil {
		logger.Logger.Fatalf("failed to create repository due to error: %v", err)
	}

	pingerService := pinger.NewPingerService(repo)

	providerService := provider.NewProviderService(repo)

	storageService := storage.NewStorageService(repo)

	cryptoService := encrypting.NewService(cfg.CryptoKey)

	var privateKey *rsa.PrivateKey

	privateKey, err = cryptoService.GetPrivateKey()
	if err != nil {
		logger.Logger.Error(err)
		err = cryptoService.GenerateIfNotExist()
		if err != nil {
			logger.Logger.Fatal(err)
		}

		privateKey, err = cryptoService.GetPrivateKey()
		if err != nil {
			logger.Logger.Fatal(err)
		}
	}

	mux := chi.NewRouter()
	controller := controllers.NewController(mux, storageService, providerService, pingerService, cfg.HashKey, privateKey)

	server := httpserver.NewHTTPServer(controller, httpserver.WithAddr(cfg.Address))
	logger.Logger.Infof("server listening on: %s", cfg.Address)

	grpcController := rpc.NewGRPCController(storageService, providerService)
	listen, err := net.Listen("tcp", ":9090")
	if err != nil {
		logger.Logger.Fatal(err)
	}

	s := grpc.NewServer()

	metrics.RegisterMetricsCollectorServer(s, grpcController)
	reflection.Register(s)

	go func() {
		err := s.Serve(listen)
		if err != nil {
			logger.Logger.Error(err)
		}
	}()

	services := errgroup.Group{}

	if cfg.DBString == "" {
		failoverService := failover.NewService(repo, cfg.FileStoragePath, cfg.StoreInterval, cfg.Restore)
		services.Go(func() error {
			failoverService.Run(ctx)
			return nil
		})
	}

	select {
	case s := <-server.Notify():
		logger.Logger.Infof("server error: %v", s)
	case <-ctx.Done():
		err = services.Wait()
		if err != nil && !errors.Is(err, context.Canceled) {
			logger.Logger.Errorf("failed to gracefully shutdown services: %v", err)
		}
		logger.Logger.Infof("server terminated")
	}

	err = server.Shutdown()
	if err != nil {
		logger.Logger.Errorf("failed to gracefully shutdown server: %v", err)
		return 1
	}

	s.GracefulStop()

	err = services.Wait()
	if err != nil && !errors.Is(err, context.Canceled) {
		logger.Logger.Errorf("failed to gracefully shutdown services: %v", err)
		return 1
	}

	return 0
}
