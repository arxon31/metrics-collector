package main

import (
	"github.com/arxon31/metrics-collector/internal/httpserver"
	"log"
)

func main() {
	log.Println("starting server...")

	params := &httpserver.Params{
		Address: "localhost",
		Port:    "8080",
	}

	server := httpserver.New(params)
	log.Printf("server is listening on %s:%s", params.Address, params.Port)

	server.Run()
}
