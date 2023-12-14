package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/arxon31/metrics-collector/internal/agent"
	"github.com/arxon31/metrics-collector/internal/config"
)

func main() {
	log.Println("starting agent...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
		<-stop

		cancel()
	}()

	cfg, err := config.NewAgentConfig()
	if err != nil {
		log.Fatalf("failed to parse a config due to error: %v", err)
	}

	params := agent.Params(*cfg)
	a := agent.New(&params)
	log.Printf("a is posting to %s with poll interval %.1fs and report interval %.1fs", params.Address, params.PollInterval.Seconds(), params.ReportInterval.Seconds())
	a.Run(ctx)

}
