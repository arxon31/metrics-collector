package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/arxon31/metrics-collector/internal/config"
	"github.com/arxon31/metrics-collector/internal/httpserver"
)

func main() {
	log.Println("starting server...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
		<-stop
		cancel()
	}()

	cfg, err := config.NewServerConfig()
	if err != nil {
		log.Fatalf("failed to parse config due to error: %v", err)
	}

	params := httpserver.Params(*cfg)

	server := httpserver.New(&params)
	log.Printf("server is listening on %s", params.Address)

	server.Run(ctx)
}
