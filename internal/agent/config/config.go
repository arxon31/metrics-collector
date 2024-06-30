package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/caarlos0/env/v10"
)

var (
	address        = flag.String("a", "localhost:8080", "server address")
	pollInterval   = flag.Int("p", 2, "agent poll interval")
	reportInterval = flag.Int("r", 10, "agent report interval")
	hashKey        = flag.String("k", "", "key to hash all sending data")
	rateLimit      = flag.Int("l", 100, "agent rate limit")
	cryptoKey      = flag.String("crypto-key", "", "key to encrypt all sending data")
)

const (
	PollIntervalEnv   = "POLL_INTERVAL"
	ReportIntervalEnv = "REPORT_INTERVAL"
)

type Config struct {
	Address        string `env:"ADDRESS"`
	PollInterval   time.Duration
	ReportInterval time.Duration
	HashKey        string `env:"KEY"`
	RateLimit      int    `env:"RATE_LIMIT"`
	CryptoKey      string `env:"CRYPTO_KEY"`
}

// NewAgentConfig creates new agent config
func NewAgentConfig() (*Config, error) {
	var config Config

	if err := env.Parse(&config); err != nil {
		return &config, err
	}
	flag.Parse()

	if config.Address == "" {
		config.Address = *address
	}
	if config.HashKey == "" {
		config.HashKey = *hashKey
	}
	if config.RateLimit == 0 {
		config.RateLimit = *rateLimit
	}

	if config.CryptoKey == "" {
		config.CryptoKey = *cryptoKey
	}

	config.PollInterval = time.Duration(*pollInterval) * time.Second
	pollIntervalString, pollExist := os.LookupEnv(PollIntervalEnv)
	if pollExist {
		pollIntervalInt, err := strconv.Atoi(pollIntervalString)
		if err != nil {
			return nil, fmt.Errorf("can not parse poll interval due to error: %v", err)
		}
		config.PollInterval = time.Duration(pollIntervalInt) * time.Second
	}

	config.ReportInterval = time.Duration(*reportInterval) * time.Second
	reportIntervalString, reportExist := os.LookupEnv(ReportIntervalEnv)
	if reportExist {
		reportIntervalInt, err := strconv.Atoi(reportIntervalString)
		if err != nil {
			return nil, fmt.Errorf("can not parse report interval due to error: %v", err)
		}
		config.ReportInterval = time.Duration(reportIntervalInt) * time.Second
	}

	return &config, nil
}
