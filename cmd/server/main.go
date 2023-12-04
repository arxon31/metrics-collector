package main

import (
	"github.com/arxon31/metrics-collector/internal/metric"
	"github.com/arxon31/metrics-collector/internal/service"
	"log"
	"net/http"
)

func main() {
	log.Println("Starting server...")
	mux := http.NewServeMux()
	storage := metric.NewMetricStorage()
	service := service.NewService(storage)
	handler := metric.NewHandler(service)
	handler.Register(mux)
	log.Println("Listening...")
	http.ListenAndServe("localhost:8080", mux)
}
