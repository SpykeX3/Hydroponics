package config

import (
	"encoding/json"
	"os"
	"path"
)

type WaterConfig struct {
	MaxAllowedLevelDrop    int `json:"max_allowed_level_drop"`
	RefillIfBelowLevel     int `json:"refill_if_below_level"`
	RefillTimeMilliseconds int `json:"refill_time_milliseconds"`
}

func createDefaultConfig(configPath string) error {
	defaultConfig := WaterConfig{
		MaxAllowedLevelDrop:    25,
		RefillIfBelowLevel:     55,
		RefillTimeMilliseconds: 2000,
	}

	file, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	return encoder.Encode(defaultConfig)
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
			if err := createDefaultConfig(configPath); err != nil {
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
		return nil, err
	}

	if config.RefillTimeMilliseconds > 5000 {
		config.RefillTimeMilliseconds = 5000
	}
	if config.RefillTimeMilliseconds < 500 {
		config.RefillTimeMilliseconds = 500
	}
	return config, nil
}
