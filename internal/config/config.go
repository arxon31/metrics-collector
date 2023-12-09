package config

import (
	"flag"
	"github.com/caarlos0/env/v10"
	"log"
)

var (
	address        = flag.String("a", "localhost:8080", "server address")
	pollInterval   = flag.Int("p", 2, "agent poll interval")
	reportInterval = flag.Int("r", 10, "agent report interval")
)

type Config struct {
	Server ServerConfig
	Agent  AgentConfig
}

type ServerConfig struct {
	Address string `env:"ADDRESS"`
}

type AgentConfig struct {
	Address        string `env:"ADDRESS"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
}

func New() *Config {
	var config Config

	if err := env.Parse(&config.Server); err != nil {
		log.Fatalf("failed to parse server config: %v", err)
	}
	if err := env.Parse(&config.Agent); err != nil {
		log.Fatalf("failed to parse agent config: %v", err)
	}

	if config.Server.Address == "" {
		flag.Parse()
		config.Server.Address = *address
	}
	if config.Agent.Address == "" {
		flag.Parse()
		config.Agent.Address = *address
	}
	if config.Agent.PollInterval == 0 {
		flag.Parse()
		config.Agent.PollInterval = *pollInterval
	}
	if config.Agent.ReportInterval == 0 {
		flag.Parse()
		config.Agent.ReportInterval = *reportInterval
	}

	return &config
}
