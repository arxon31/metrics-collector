package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/arxon31/metrics-collector/internal/agent/service/generator"

	"go.uber.org/zap"

	"github.com/arxon31/metrics-collector/internal/agent/config"
	"github.com/arxon31/metrics-collector/internal/agent/service/compressor"
	"github.com/arxon31/metrics-collector/internal/agent/service/hasher"
	"github.com/arxon31/metrics-collector/internal/agent/service/poller"
	"github.com/arxon31/metrics-collector/internal/agent/service/reporter"
	"github.com/arxon31/metrics-collector/internal/repository/memory"
	"github.com/arxon31/metrics-collector/pkg/httpclient"
)

const retryCount = 3

func main() {
	exitCode := run()
	if exitCode != 0 {
		log.Fatal("exited with code", exitCode)
	}
}

func run() int {
	logger := initLogger()
	logger.Info("starting agent")

	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cfg, err := config.NewAgentConfig()
	if err != nil {
		logger.Error("failed to parse a config due to error: %v", err)
		return 1
	}

	repo := memory.NewMapStorage()

	reportClient := httpclient.NewClient()

	hashService := hasher.NewHasherService(cfg.HashKey)
	compressService := compressor.NewCompressorService()

	pollService := poller.New(logger, repo)

	generateService := generator.New(cfg.Address, repo, hashService, compressService, logger)

	reportService := reporter.NewReporter(logger, cfg.RateLimit, reportClient)

	pollTicker := time.NewTicker(cfg.PollInterval)
	defer pollTicker.Stop()

	reportTicker := time.NewTicker(cfg.ReportInterval)
	defer reportTicker.Stop()

WORKLOOP:
	for {
		select {
		case <-ctx.Done():
			break WORKLOOP
		case <-pollTicker.C:
			pollService.Poll(ctx)
		case <-reportTicker.C:
			reportService.Report(generateService.Generate(ctx))
		}
	}

	logger.Infof("agent stopped")

	return 0
}

func initLogger() *zap.SugaredLogger {
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(err)
	}
	return logger.Sugar()
}
