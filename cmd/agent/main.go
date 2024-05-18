package main

import (
	"context"
	"github.com/arxon31/metrics-collector/internal/agent/config"
	"log"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/arxon31/metrics-collector/internal/agent"
)

func main() {

	ctx, cancel := context.WithCancel(context.Background())

	go listenStopSignals(cancel)

	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Sync()

	sugared := logger.Sugar()

	sugared.Infoln("starting agent...")

	cfg, err := config.NewAgentConfig()
	if err != nil {
		log.Fatalf("failed to parse a config due to error: %v", err)
	}

	a := agent.New(cfg, sugared)
	log.Printf("a is posting to %s with poll interval %.1fs, report interval %.1fs and %d workers",
		cfg.Address,
		cfg.PollInterval.Seconds(),
		cfg.ReportInterval.Seconds(),
		cfg.RateLimit)

	a.Run(ctx)

}

func listenStopSignals(cancelFunc context.CancelFunc) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	cancelFunc()
}
