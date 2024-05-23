package main

import (
	"context"
	"github.com/arxon31/metrics-collector/internal/agent/config"
	"github.com/arxon31/metrics-collector/internal/agent/service/compressor"
	"github.com/arxon31/metrics-collector/internal/agent/service/hasher"
	"github.com/arxon31/metrics-collector/internal/agent/service/poller"
	"github.com/arxon31/metrics-collector/internal/agent/service/reporter"
	"github.com/arxon31/metrics-collector/internal/repository/memory"
	"github.com/arxon31/metrics-collector/pkg/httpclient"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const retryCount = 3

func main() {
	os.Exit(run())
}

func run() int {
	logger := initLogger()
	logger.Info("starting agent")

	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	services := errgroup.Group{}

	cfg, err := config.NewAgentConfig()
	if err != nil {
		logger.Error("failed to parse a config due to error: %v", err)
		return 1
	}

	repo := memory.NewMapStorage()

	reportClient := httpclient.NewClient(httpclient.WithRetries(retryCount))

	hashService := hasher.NewHasherService(cfg.HashKey)
	compressService := compressor.NewCompressorService()

	pollService := poller.New(logger, repo, repo, cfg, hashService, compressService)
	services.Go(func() error {
		pollService.Run(ctx)
		return nil
	})

	reportService := reporter.NewReporter(logger, cfg.RateLimit, cfg.ReportInterval, reportClient, pollService.GetReqChan())
	services.Go(func() error {
		reportService.Run(ctx)
		return nil
	})

	select {
	case <-ctx.Done():
		err = services.Wait()
		if err != nil {
			logger.Info("services stopped due to: %v", err)
		}
		logger.Infof("agent stopped due to signal: %v", ctx.Err())
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
