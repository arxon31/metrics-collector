package main

import (
	"github.com/arxon31/metrics-collector/internal/config"
	"github.com/arxon31/metrics-collector/internal/httpserver"
	"log"
)

func main() {
	log.Println("starting server...")

	cfg := config.New()

	params := httpserver.Params(cfg.Server)

	server := httpserver.New(&params)
	log.Printf("server is listening on %s", params.Address)

	server.Run()
}
