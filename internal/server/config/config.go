package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/caarlos0/env/v10"

	"github.com/arxon31/metrics-collector/pkg/logger"
)

var (
	address         = flag.String("a", "localhost:8080", "server address")
	storeInterval   = flag.Int("i", 300, "store interval")
	fileStoragePath = flag.String("f", "/tmp/metrics-db.json", "file storage path")
	restore         = flag.Bool("r", true, "restore from file-db")
	dbstring        = flag.String("d", "", "database connection string")
	hashKey         = flag.String("k", "", "key for hash counting")
	cryptoKeyPath   = flag.String("crypto-key", "", "key to decrypt all sending data")
	configFilePath  = flag.String("c", "./config/server_cfg.json", "config file path")
)

const (
	storeIntervalEnv = "STORE_INTERVAL"
	restoreEnv       = "RESTORE"
)

type Config struct {
	Address         string `env:"ADDRESS" ,json:"address"`
	StoreInterval   time.Duration
	FileStoragePath string `env:"FILE_STORAGE_PATH" ,json:"store_file"`
	Restore         bool   `env:"RESTORE"`
	DBString        string `env:"DATABASE_DSN" ,json:"database_dsn"`
	HashKey         string `env:"KEY" ,json:"hash_key"`
	CryptoKey       string `env:"CRYPTO_KEY" ,json:"crypto_key"`
}

// NewServerConfig creates new server config
func NewServerConfig() (*Config, error) {
	var config Config

	flag.Parse()

	if *configFilePath != "" {
		logger.Logger.Info("config file path: ", *configFilePath)

		file, err := os.Open(*configFilePath)
		if err != nil {
			logger.Logger.Error(err)
		}
		defer file.Close()

		configBytes, err := io.ReadAll(file)
		if err != nil {
			logger.Logger.Error(err)
		}

		err = json.Unmarshal(configBytes, &config)
		if err != nil {
			logger.Logger.Error(err)
		}
	}

	if err := env.Parse(&config); err != nil {
		return &config, err
	}

	if config.Address == "" {
		config.Address = *address
	}

	if config.FileStoragePath == "" {
		config.FileStoragePath = *fileStoragePath
	}

	if config.DBString == "" {
		config.DBString = *dbstring
	}

	if config.HashKey == "" {
		config.HashKey = *hashKey
	}

	if config.CryptoKey == "" {
		config.CryptoKey = *cryptoKeyPath
	}

	config.Restore = *restore
	restoreString, isRestoreExist := os.LookupEnv(restoreEnv)
	if isRestoreExist {
		restoreBool, err := strconv.ParseBool(restoreString)
		if err != nil {
			return nil, fmt.Errorf("can not parse poll interval due to error: %v", err)
		}
		config.Restore = restoreBool
	}

	config.StoreInterval = time.Duration(*storeInterval) * time.Second
	storeIntervalString, isStoreIntervalExist := os.LookupEnv(storeIntervalEnv)
	if isStoreIntervalExist {
		storeIntervalInt, err := strconv.Atoi(storeIntervalString)
		if err != nil {
			return nil, fmt.Errorf("can not parse poll interval due to error: %v", err)
		}
		config.StoreInterval = time.Duration(storeIntervalInt) * time.Second
	}

	return &config, nil
}
