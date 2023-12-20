package main

import (
	"context"
	"github.com/arxon31/metrics-collector/internal/storage/mem"
	"go.uber.org/zap"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/arxon31/metrics-collector/internal/config"
	"github.com/arxon31/metrics-collector/internal/httpserver"
)

func main() {
	const op = "main()"
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Sync()

	sugared := logger.Sugar()

	sugared.Infoln("starting server...")

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
		sugared.Fatalln("failed to parse config due to error: %v", err)
	}

	storage := mem.NewMapStorage()

	params := httpserver.Params(*cfg)

	server := httpserver.New(&params, sugared, storage, storage)
	sugared.Infof("server is listening on %s", params.Address)

	server.Run(ctx)
}
