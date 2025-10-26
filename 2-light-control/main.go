package main

import (
	"encoding/json"
	"log"
	"os"
	"path"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/host/v3"
	"periph.io/x/host/v3/rpi"
	"time"
)

type LightConfig struct {
	StartTime  time.Duration `json:"start_time"`
	FinishTime time.Duration `json:"finish_time"`
}

type lightConfigStr struct {
	StartTime  string `json:"start_time"`
	FinishTime string `json:"finish_time"`
}

func getConfig() (*LightConfig, error) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	file, err := os.Open(path.Join(homedir, "light.config"))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	config := &lightConfigStr{}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(config); err != nil {
		return nil, err
	}
	lightStartTime, err := time.ParseDuration(config.StartTime)
	if err != nil {
		return nil, err
	}
	lightFinishTime, err := time.ParseDuration(config.FinishTime)
	if err != nil {
		return nil, err
	}
	return &LightConfig{
		StartTime:  lightStartTime,
		FinishTime: lightFinishTime,
	}, nil
}

func main() {
	_, err := host.Init()
	if err != nil {
		panic(err)
	}

	config, err := getConfig()
	if err != nil {
		log.Printf("Failed to load config: %v", err)
	}

	now := time.Now().Local().Truncate(time.Second)
	today0 := now.Truncate(24 * time.Hour)

	start := today0.Add(config.StartTime)
	finish := today0.Add(config.FinishTime)

	log.Println("Start:", start)
	log.Println("Finish:", finish)
	log.Println("Now:", now)

	if now.Before(start) || now.After(finish) {
		err := rpi.P1_7.Out(gpio.Low)
		if err != nil {
			log.Panicln("Error setting pin state:", err)
		}
		return
	}
	log.Println("Turning on the light.")
	err = rpi.P1_7.Out(gpio.High)
}
