package config

import (
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	addr = "localhost:9090"
	key  = "key"
	rl   = 200
	poll = 15
	rep  = 20
)

func TestNewAgentConfig(t *testing.T) {

	t.Run("must_return_config_from_flags", func(t *testing.T) {
		config, err := NewAgentConfig()
		require.Nil(t, err)
		require.IsType(t, &Config{}, config)
		require.Equal(t, "localhost:8080", config.Address)
		require.Equal(t, "", config.HashKey)
		require.Equal(t, 100, config.RateLimit)
		require.Equal(t, float64(2), config.PollInterval.Seconds())
		require.Equal(t, float64(10), config.ReportInterval.Seconds())
	})

	t.Run("must_return_config_from_env", func(t *testing.T) {
		setup()
		config, err := NewAgentConfig()
		require.Nil(t, err)
		require.IsType(t, &Config{}, config)
		require.Equal(t, addr, config.Address)
		require.Equal(t, key, config.HashKey)
		require.Equal(t, rl, config.RateLimit)
		require.Equal(t, float64(poll), config.PollInterval.Seconds())
		require.Equal(t, float64(rep), config.ReportInterval.Seconds())
	})

}

func setup() {
	os.Setenv("ADDRESS", addr)
	os.Setenv("KEY", key)
	os.Setenv("RATE_LIMIT", strconv.Itoa(rl))
	os.Setenv("POLL_INTERVAL", strconv.Itoa(poll))
	os.Setenv("REPORT_INTERVAL", strconv.Itoa(rep))
}
