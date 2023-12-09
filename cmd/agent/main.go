package main

import (
	"context"
	agent2 "github.com/arxon31/metrics-collector/internal/agent"
	"github.com/arxon31/metrics-collector/internal/config"
	"log"
)

func main() {
	Cfg := config.New()

	params := agent2.Params(Cfg.Agent)
	agent := agent2.New(context.Background(), &params)
	log.Printf("agent is posting to %s:%s with poll interval %s and report interval %s", params.Address, params.Port, params.PollInterval, params.ReportInterval)
	agent.Run()

}
