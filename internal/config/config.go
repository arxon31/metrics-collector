package config

import (
	"flag"
	"log"
	"strings"
	"time"
)

var (
	address        = flag.String("a", "localhost:8080", "server address")
	pollInterval   = flag.Duration("p", 2*time.Second, "agent poll interval")
	reportInterval = flag.Duration("r", 10*time.Second, "agent report interval")
)

type Config struct {
	Server ServerConfig
	Agent  AgentConfig
}

type ServerConfig struct {
	Address string
	Port    string
}

type AgentConfig struct {
	Address        string
	Port           string
	PollInterval   time.Duration
	ReportInterval time.Duration
}

func New() *Config {
	flag.Parse()

	addr := strings.Split(*address, ":")
	if len(addr) != 2 {
		log.Fatal("invalid address")
	}

	return &Config{
		Server: ServerConfig{
			Address: addr[0],
			Port:    addr[1],
		},
		Agent: AgentConfig{
			Address:        addr[0],
			Port:           addr[1],
			PollInterval:   *pollInterval,
			ReportInterval: *reportInterval,
		},
	}
}
