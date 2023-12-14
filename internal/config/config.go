package config

import (
	"flag"
	"fmt"
	"github.com/caarlos0/env/v10"
	"os"
	"strconv"
	"time"
)

var (
	address        = flag.String("a", "localhost:8080", "server address")
	pollInterval   = flag.Int("p", 2, "agent poll interval")
	reportInterval = flag.Int("r", 10, "agent report interval")
)

const (
	PollIntervalEnv   = "POLL_INTERVAL"
	ReportIntervalEnv = "REPORT_INTERVAL"
)

type ServerConfig struct {
	Address string `env:"ADDRESS"`
}

type AgentConfig struct {
	Address        string `env:"ADDRESS"`
	PollInterval   time.Duration
	ReportInterval time.Duration
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
	pollIntervalString, pollExist := os.LookupEnv(PollIntervalEnv)

	if pollExist {
		pollIntervalInt, err := strconv.Atoi(pollIntervalString)
		if err != nil {
			return nil, fmt.Errorf("can not parse poll interval due to error: %v", err)
		}
		config.PollInterval = time.Duration(pollIntervalInt) * time.Second
	} else {
		flag.Parse()
		config.PollInterval = time.Duration(*pollInterval) * time.Second
	}

	reportIntervalString, reportExist := os.LookupEnv(ReportIntervalEnv)
	if reportExist {
		reportIntervalInt, err := strconv.Atoi(reportIntervalString)
		if err != nil {
			return nil, fmt.Errorf("can not parse report interval due to error: %v", err)
		}
		config.ReportInterval = time.Duration(reportIntervalInt) * time.Second
	} else {
		flag.Parse()
		config.ReportInterval = time.Duration(*reportInterval) * time.Second
	}

	return &config, nil
}
