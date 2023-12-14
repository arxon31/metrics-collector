package config

import (
	"flag"
	"github.com/caarlos0/env/v10"
	"time"
)

var (
	address        = flag.String("a", "localhost:8080", "server address")
	pollInterval   = flag.Int("p", 2, "agent poll interval")
	reportInterval = flag.Int("r", 10, "agent report interval")
)

type ServerConfig struct {
	Address string `env:"ADDRESS"`
}

type AgentConfig struct {
	Address        string        `env:"ADDRESS"`
	PollInterval   time.Duration `env:"POLL_INTERVAL"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL"`
}

func NewServerConfig() (*ServerConfig, error) {

	var config ServerConfig

	if err := env.Parse(&config); err != nil {
		return &config, err
	}

	if config.Address == "" {
		flag.Parse()
		config.Address = *address
	}

	return &config, nil
}

func NewAgentConfig() (*AgentConfig, error) {
	var config AgentConfig

	if err := env.Parse(&config); err != nil {
		return &config, err
	}

	if config.Address == "" {
		flag.Parse()
		config.Address = *address
	}
	if config.PollInterval == 0 {
		flag.Parse()
		config.PollInterval = time.Duration(*pollInterval) * time.Second
	}
	if config.ReportInterval == 0 {
		flag.Parse()
		config.ReportInterval = time.Duration(*reportInterval) * time.Second
	}

	return &config, nil
}
