package config

import (
	"encoding/json"
	"os"
	"path"
)

type WaterConfig struct {
	MaxAllowedLevelDrop    int   `json:"max_allowed_level_drop"`
	RefillIfBelowLevel     int   `json:"refill_if_below_level"`
	RefillTimeMilliseconds int   `json:"refill_time_milliseconds"`
	EmptyIfAboveLevel      int   `json:"empty_if_above_level"`
	EmptyTimeMilliseconds  int   `json:"empty_time_milliseconds"`
	FillForSeconds         int64 `json:"fill_for_seconds"`
	EmptyForSeconds        int64 `json:"empty_for_seconds"`
}

func createDefaultConfig(configPath string) (*WaterConfig, error) {
	defaultConfig := WaterConfig{
		MaxAllowedLevelDrop:    100,
		RefillIfBelowLevel:     75,
		RefillTimeMilliseconds: 4000,
		EmptyIfAboveLevel:      55,
		EmptyTimeMilliseconds:  5000,
		FillForSeconds:         30,
		EmptyForSeconds:        30,
	}

	file, err := os.Create(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	return &defaultConfig, encoder.Encode(defaultConfig)
}

func GetConfig() (*WaterConfig, error) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configPath := path.Join(homedir, "water_refill.config")
	file, err := os.Open(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			if _, err := createDefaultConfig(configPath); err != nil {
				return nil, err
			}
			file, err = os.Open(configPath)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	defer file.Close()

	config := &WaterConfig{}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(config); err != nil {
		if config, err = createDefaultConfig(configPath); err != nil {
			return nil, err
		}
		file, err = os.Open(configPath)
		if err != nil {
			return nil, err
		}
	}

	if config.RefillTimeMilliseconds > 20000 {
		config.RefillTimeMilliseconds = 20000
	}
	if config.RefillTimeMilliseconds < 500 {
		config.RefillTimeMilliseconds = 500
	}
	return config, nil
}
