package config

import (
	"flag"
	"log"
	"strings"
	"time"
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

	poll := time.Duration(*pollInterval) * time.Second
	report := time.Duration(*reportInterval) * time.Second

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
			PollInterval:   poll,
			ReportInterval: report,
		},
	}
}
