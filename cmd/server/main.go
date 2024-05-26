// TODO: отдельный пакет для логгера

package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/arxon31/metrics-collector/internal/repository"
	"github.com/arxon31/metrics-collector/internal/server/config"
	controllers "github.com/arxon31/metrics-collector/internal/server/controller/http"
	"github.com/arxon31/metrics-collector/internal/server/service/failover"
	"github.com/arxon31/metrics-collector/internal/server/service/pinger"
	"github.com/arxon31/metrics-collector/internal/server/service/provider"
	"github.com/arxon31/metrics-collector/internal/server/service/storage"
	"github.com/arxon31/metrics-collector/pkg/httpserver"
)

var Logger *zap.SugaredLogger

func main() {
	os.Exit(run())
}

func run() int {
	logger := initLogger()
	logger.Infoln("starting server")

	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	services := errgroup.Group{}

	cfg, err := config.NewServerConfig()
	if err != nil {
		logger.Fatalf("failed to parse a config due to error: %v", err)
	}

	repo, err := repository.New(cfg.DBString, logger)
	if err != nil {
		logger.Fatalf("failed to create repository due to error: %v", err)
	}

	if cfg.DBString == "" {
		failoverService := failover.NewService(repo, cfg.FileStoragePath, cfg.StoreInterval, cfg.Restore, logger)
		services.Go(func() error {
			failoverService.Run(ctx)
			return nil
		})
	}

	pingerService := pinger.NewPingerService(repo)

	providerService := provider.NewProviderService(repo, logger)

	storageService := storage.NewStorageService(repo, logger)

	mux := chi.NewRouter()
	controller := controllers.NewController(mux, storageService, providerService, pingerService, logger, cfg.HashKey)

	server := httpserver.NewHTTPServer(controller, httpserver.WithAddr(cfg.Address))
	logger.Infof("server listening on: %s", cfg.Address)

	select {
	case s := <-server.Notify():
		logger.Infof("server error: %v", s)
		return 1
	case <-ctx.Done():
		err := services.Wait()
		if err != nil && !errors.Is(err, context.Canceled) {
			logger.Errorf("failed to gracefully shutdown server: %v", err)
			return 1
		}
		logger.Infof("server terminated")
	}

	err = server.Shutdown()
	if err != nil {
		logger.Errorf("failed to gracefully shutdown server: %v", err)
		return 1
	}

	return 0
}

func initLogger() *zap.SugaredLogger {
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(err)
	}
	return logger.Sugar()
}
