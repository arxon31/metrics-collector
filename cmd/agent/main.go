package main

import (
	"context"
	agent2 "github.com/arxon31/metrics-collector/internal/agent"
	"time"
)

func main() {
	params := agent2.Params{
		Address:        "localhost",
		Port:           "8080",
		PollInterval:   2 * time.Second,
		ReportInterval: 10 * time.Second,
	}
	agent := agent2.New(context.Background(), &params)
	agent.Run()

}
