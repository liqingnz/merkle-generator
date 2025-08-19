// Package util provides configuration utilities
package util

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	RPC RPCConfig `yaml:"rpc"`
	CSV CSVConfig `yaml:"csv"`
}

// RPCConfig contains RPC connection settings
type RPCConfig struct {
	Endpoint        string `yaml:"endpoint"`
	ContractAddress string `yaml:"contract_address"`
}

// CSVConfig contains CSV file settings for merkle tree data
type CSVConfig struct {
	FilePath string `yaml:"file_path"`
}

// LoadConfig loads configuration from a YAML file
func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// ValidateConfig validates the configuration
func (c *Config) ValidateConfig() error {
	if c.RPC.Endpoint == "" {
		return fmt.Errorf("rpc.endpoint is required")
	}

	if c.RPC.ContractAddress == "" {
		return fmt.Errorf("rpc.contract_address is required")
	}

	if c.CSV.FilePath == "" {
		return fmt.Errorf("csv.file_path is required")
	}

	return nil
}
