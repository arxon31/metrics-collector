package config

import (
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	addr  = "localhost:9090"
	store = 200
	fPath = "/tmp/metrics.json"
	rest  = false
	dbstr = "PostgresString"
	key   = "key"
)

func TestNewServerConfig(t *testing.T) {

	t.Run("must_return_config_from_flags", func(t *testing.T) {
		config, err := NewServerConfig()
		require.Nil(t, err)
		require.IsType(t, &Config{}, config)
		require.Equal(t, "localhost:8080", config.Address)
		require.Equal(t, float64(300), config.StoreInterval.Seconds())
		require.Equal(t, "/tmp/metrics-db.json", config.FileStoragePath)
		require.Equal(t, true, config.Restore)
		require.Equal(t, "", config.DBString)
		require.Equal(t, "", config.HashKey)
	})

	t.Run("must_return_config_from_env", func(t *testing.T) {
		setup()
		config, err := NewServerConfig()
		require.Nil(t, err)
		require.IsType(t, &Config{}, config)
		require.Equal(t, addr, config.Address)
		require.Equal(t, float64(store), config.StoreInterval.Seconds())
		require.Equal(t, fPath, config.FileStoragePath)
		require.Equal(t, rest, config.Restore)
		require.Equal(t, dbstr, config.DBString)
		require.Equal(t, key, config.HashKey)
	})

}

func setup() {
	os.Setenv("ADDRESS", addr)
	os.Setenv("STORE_INTERVAL", strconv.Itoa(store))
	os.Setenv("FILE_STORAGE_PATH", fPath)
	os.Setenv("RESTORE", strconv.FormatBool(rest))
	os.Setenv("DATABASE_DSN", dbstr)
	os.Setenv("KEY", key)
}
