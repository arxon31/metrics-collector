// TODO: отдельный пакет для логгера

package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os/signal"
	"syscall"

	"github.com/arxon31/metrics-collector/internal/server/service/failover"

	"golang.org/x/sync/errgroup"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/arxon31/metrics-collector/internal/repository"
	"github.com/arxon31/metrics-collector/internal/server/config"
	controllers "github.com/arxon31/metrics-collector/internal/server/controller/http"
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
	logger := initLogger()
	logger.Infoln("starting server")
	logger.Info(fmt.Sprintf("version: %s, build time: %s, build commit: %s", buildVersion, buildDate, buildCommit))

	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cfg, err := config.NewServerConfig()
	if err != nil {
		logger.Fatalf("failed to parse a config due to error: %v", err)
	}

	repo, err := repository.New(cfg.DBString, logger)
	if err != nil {
		logger.Fatalf("failed to create repository due to error: %v", err)
	}

	pingerService := pinger.NewPingerService(repo)

	providerService := provider.NewProviderService(repo, logger)

	storageService := storage.NewStorageService(repo, logger)

	mux := chi.NewRouter()
	controller := controllers.NewController(mux, storageService, providerService, pingerService, logger, cfg.HashKey)

	server := httpserver.NewHTTPServer(controller, httpserver.WithAddr(cfg.Address))
	logger.Infof("server listening on: %s", cfg.Address)

	services := errgroup.Group{}

	if cfg.DBString == "" {
		failoverService := failover.NewService(repo, cfg.FileStoragePath, cfg.StoreInterval, cfg.Restore, logger)
		services.Go(func() error {
			failoverService.Run(ctx)
			return nil
		})
	}

	select {
	case s := <-server.Notify():
		logger.Infof("server error: %v", s)
		return 1
	case <-ctx.Done():
		err = services.Wait()
		if err != nil && !errors.Is(err, context.Canceled) {
			logger.Errorf("failed to gracefully shutdown services: %v", err)
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
