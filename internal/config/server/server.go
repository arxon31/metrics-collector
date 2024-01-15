package server

import (
	"flag"
	"fmt"
	"github.com/caarlos0/env/v10"
	"os"
	"strconv"
	"time"
)

var (
	address         = flag.String("a", "localhost:8080", "server address")
	storeInterval   = flag.Int("i", 300, "store interval")
	fileStoragePath = flag.String("f", "/tmp/metrics-db.json", "file storage path")
	restore         = flag.Bool("r", true, "restore from file-db")
)

const (
	storeIntervalEnv = "STORE_INTERVAL"
	restoreEnv       = "RESTORE"
)

type Config struct {
	Address         string `env:"ADDRESS"`
	StoreInterval   time.Duration
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
}

func NewServerConfig() (*Config, error) {
	var config Config

	if err := env.Parse(&config); err != nil {
		return &config, err
	}

	flag.Parse()

	if config.Address == "" {
		config.Address = *address
	}

	if config.FileStoragePath == "" {
		config.FileStoragePath = *fileStoragePath
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
