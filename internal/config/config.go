// Package config handles application configuration loading and validation
package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config represents the main application configuration
type Config struct {
	APIKey      string              `json:"api_key"`
	APISecret   string              `json:"api_secret"`
	AccessToken string              `json:"access_token"`
	Instruments []InstrumentConfig  `json:"instruments"`
}

// InstrumentConfig represents configuration for a single trading instrument
type InstrumentConfig struct {
	Symbol   string `json:"symbol"`
	Exchange string `json:"exchange"`
	LotSize  int    `json:"lot_size"`
	Product  string `json:"product"` // MIS, NRML, CNC, etc.
}

// Load reads and validates configuration from a JSON file
func Load(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening config file: %w", err)
	}
	defer file.Close()

	var config Config
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return nil, fmt.Errorf("error decoding config: %w", err)
	}

	// Validate configuration
	if err := validateConfig(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// validateConfig performs comprehensive validation of the configuration
func validateConfig(config *Config) error {
	if config.APIKey == "" {
		return fmt.Errorf("api_key is required in config.json")
	}
	if config.AccessToken == "" {
		return fmt.Errorf("access_token is required in config.json")
	}
	if len(config.Instruments) == 0 {
		return fmt.Errorf("at least one instrument is required in config.json")
	}

	// Validate each instrument
	for i, inst := range config.Instruments {
		if inst.Symbol == "" {
			return fmt.Errorf("instrument %d: symbol is required", i+1)
		}
		if inst.Exchange == "" {
			return fmt.Errorf("instrument %d: exchange is required", i+1)
		}
		if inst.LotSize <= 0 {
			return fmt.Errorf("instrument %d: lot_size must be greater than 0", i+1)
		}
		// Product is optional - defaults to MIS if not specified
	}

	return nil
}