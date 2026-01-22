package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	APIKey      string              `json:"api_key"`
	APISecret   string              `json:"api_secret"`
	AccessToken string              `json:"access_token"`
	Instruments []InstrumentConfig  `json:"instruments"`
}

type InstrumentConfig struct {
	Symbol   string `json:"symbol"`
	Exchange string `json:"exchange"`
	LotSize  int    `json:"lot_size"`
	Product  string `json:"product"` // MIS, NRML, CNC, etc.
}

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

	// Validate config
	if config.APIKey == "" {
		return nil, fmt.Errorf("api_key is required in config.json")
	}
	if config.AccessToken == "" {
		return nil, fmt.Errorf("access_token is required in config.json")
	}
	if len(config.Instruments) == 0 {
		return nil, fmt.Errorf("at least one instrument is required in config.json")
	}

	// Validate each instrument
	for i, inst := range config.Instruments {
		if inst.Symbol == "" {
			return nil, fmt.Errorf("instrument %d: symbol is required", i+1)
		}
		if inst.Exchange == "" {
			return nil, fmt.Errorf("instrument %d: exchange is required", i+1)
		}
		if inst.LotSize <= 0 {
			return nil, fmt.Errorf("instrument %d: lot_size must be greater than 0", i+1)
		}
		// Product is optional - defaults to MIS if not specified
	}

	return &config, nil
}