package main

import (
	"context"
	"fmt"
	"github.com/arxon31/metrics-collector/internal/agent/service/encryptor"
	"github.com/arxon31/metrics-collector/internal/encrypting"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/arxon31/metrics-collector/pkg/logger"

	"github.com/arxon31/metrics-collector/internal/agent/service/generator"

	"github.com/arxon31/metrics-collector/internal/agent/config"
	"github.com/arxon31/metrics-collector/internal/agent/service/compressor"
	"github.com/arxon31/metrics-collector/internal/agent/service/hasher"
	"github.com/arxon31/metrics-collector/internal/agent/service/poller"
	"github.com/arxon31/metrics-collector/internal/agent/service/reporter"
	"github.com/arxon31/metrics-collector/internal/repository/memory"
	"github.com/arxon31/metrics-collector/pkg/httpclient"
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
	logger.Logger.Info("starting agent")
	logger.Logger.Info(fmt.Sprintf("version: %s, build time: %s, build commit: %s", buildVersion, buildDate, buildCommit))

	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cfg, err := config.NewAgentConfig()
	if err != nil {
		logger.Logger.Error("failed to parse a config due to error: %v", err)
		return 1
	}

	repo := memory.NewMapStorage()

	reportClient := httpclient.NewClient()

	hashService := hasher.NewHasherService(cfg.HashKey)

	compressService := compressor.NewCompressorService()

	cryptoService := encrypting.NewService(cfg.CryptoKey)

	publicKey, err := cryptoService.GetPublicKey()
	if err != nil {
		logger.Logger.Error("failed to get public key due to error: %v", err)
		return 1
	}

	encryptorService := encryptor.NewEncryptorService(publicKey)

	pollService := poller.New(repo)

	generateService := generator.New(cfg.Address, repo, hashService, compressService, encryptorService)

	reportService := reporter.NewReporter(cfg.RateLimit, reportClient)

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

	logger.Logger.Info("agent stopped")

	return 0
}
