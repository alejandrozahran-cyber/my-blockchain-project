package node

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Network struct {
		Name     string `yaml:"name"`
		ChainID  int    `yaml:"chain_id"`
		Port     int    `yaml:"port"`
		RPCURL   string `yaml:"rpc_url"`
		Bootnodes []string `yaml:"bootnodes"`
	} `yaml:"network"`

	Consensus struct {
		Type           string  `yaml:"type"`
		BlockTime      int     `yaml:"block_time"`
		ValidatorCount int     `yaml:"validator_count"`
		MinStake       float64 `yaml:"min_stake"`
	} `yaml:"consensus"`

	AIEngine struct {
		URL         string `yaml:"url"`
		Enabled     bool   `yaml:"enabled"`
		Timeout     int    `yaml:"timeout"`
		RetryCount  int    `yaml:"retry_count"`
	} `yaml:"ai_engine"`

	Database struct {
		Path string `yaml:"path"`
		Type string `yaml:"type"`
	} `yaml:"database"`

	API struct {
		Enabled      bool     `yaml:"enabled"`
		Host         string   `yaml:"host"`
		Port         int      `yaml:"port"`
		CorsDomains  []string `yaml:"cors_domains"`
		RateLimit    int      `yaml:"rate_limit"`
	} `yaml:"api"`

	P2P struct {
		MaxPeers   int    `yaml:"max_peers"`
		Discovery  bool   `yaml:"discovery"`
		ListenAddr string `yaml:"listen_addr"`
	} `yaml:"p2p"`
}

func DefaultConfig() *Config {
	cfg := &Config{}

	// Network
	cfg.Network.Name = "nusa-testnet"
	cfg.Network.ChainID = 2024
	cfg.Network.Port = 30303
	cfg.Network.RPCURL = "http://localhost:8545"
	cfg.Network.Bootnodes = []string{}

	// Consensus
	cfg.Consensus.Type = "PoVC"
	cfg.Consensus.BlockTime = 5
	cfg.Consensus.ValidatorCount = 3
	cfg.Consensus.MinStake = 1000

	// AI Engine
	cfg.AIEngine.URL = "http://localhost:8000"
	cfg.AIEngine.Enabled = true
	cfg.AIEngine.Timeout = 10
	cfg.AIEngine.RetryCount = 3

	// Database
	cfg.Database.Path = "./data/chaindata"
	cfg.Database.Type = "leveldb"

	// API
	cfg.API.Enabled = true
	cfg.API.Host = "0.0.0.0"
	cfg.API.Port = 8545
	cfg.API.CorsDomains = []string{"*"}
	cfg.API.RateLimit = 100

	// P2P
	cfg.P2P.MaxPeers = 50
	cfg.P2P.Discovery = true
	cfg.P2P.ListenAddr = "0.0.0.0:30303"

	return cfg
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func SaveConfig(cfg *Config, path string) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	// Create directory if not exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}