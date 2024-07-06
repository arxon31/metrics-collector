package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/arxon31/metrics-collector/pkg/logger"

	"github.com/caarlos0/env/v10"
)

var (
	address        = flag.String("a", "localhost:8080", "server address")
	pollInterval   = flag.Int("p", 2, "agent poll interval")
	reportInterval = flag.Int("r", 10, "agent report interval")
	hashKey        = flag.String("k", "", "key to hash all sending data")
	rateLimit      = flag.Int("l", 100, "agent rate limit")
	cryptoKeyPath  = flag.String("crypto-key", "", "key to encrypt all sending data")
	configFilePath = flag.String("c", "", "config file path")
)

const (
	PollIntervalEnv   = "POLL_INTERVAL"
	ReportIntervalEnv = "REPORT_INTERVAL"
)

type Config struct {
	Address        string `env:"ADDRESS" ,json:"address"`
	PollInterval   time.Duration
	ReportInterval time.Duration
	HashKey        string `env:"KEY" ,json:"hash_key"`
	RateLimit      int    `env:"RATE_LIMIT" ,json:"rate_limit"`
	CryptoKey      string `env:"CRYPTO_KEY" ,json:"crypto_key"`
}

// NewAgentConfig creates new agent config
func NewAgentConfig() (*Config, error) {
	var config Config

	flag.Parse()

	if *configFilePath != "" {
		err := configFromFile(&config)
		if err != nil {
			logger.Logger.Info(err)
		}
	}

	if err := env.Parse(&config); err != nil {
		return &config, err
	}

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
		config.CryptoKey = *cryptoKeyPath
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

func configFromFile(cfg *Config) error {
	logger.Logger.Info("config file path: ", *configFilePath)

	file, err := os.Open(*configFilePath)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	configBytes, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	err = json.Unmarshal(configBytes, &cfg)
	if err != nil {
		return fmt.Errorf("unmarshal file: %w", err)
	}

	return nil
}
