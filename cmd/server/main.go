package main

import (
	"context"
	"github.com/arxon31/metrics-collector/internal/repository"
	"github.com/arxon31/metrics-collector/internal/server"
	"github.com/arxon31/metrics-collector/internal/server/config"
	"log"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
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

	store, err := repository.New(cfg.DBString, sugared)
	if err != nil {
		sugared.Fatalln("can not create repository due to error", err)
	}

	server := server.New(cfg, sugared, store)
	sugared.Infof("server is listening on %s, with store interval %.1fs, file storage path: %s, restore %t, database_dsn: %s",
		cfg.Address, cfg.StoreInterval.Seconds(), cfg.FileStoragePath, cfg.Restore, cfg.DBString)

	server.Run(ctx)

}
