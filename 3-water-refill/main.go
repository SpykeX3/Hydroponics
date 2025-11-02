package main

import (
	"3-water-refill/config"
	"3-water-refill/db"
	wlevel "3-water-refill/water-level-sensor"
	"errors"
	"log"
	"os"
	"os/signal"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/host/v3"
	"periph.io/x/host/v3/rpi"
	"syscall"
	"time"
)

const (
	threshold       = 100
	addrHigh  uint8 = 0x78 // 12 sections
	addrLow   uint8 = 0x77 // 8 sections
)

var pumpPin = rpi.P1_11

func main() {

	db, err := db.OpenReadingsDB()
	if err != nil {
		log.Fatal(err)
	}

	if _, err = host.Init(); err != nil {
		panic(err)
	}
	sensor, err := wlevel.NewWaterLevelSensor(1, addrLow, addrHigh, threshold)
	if err != nil {
		log.Fatal(err)
	}
	defer sensor.Close()

	// Handle Ctrl+C gracefully
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	signal.Notify(stop, os.Interrupt, syscall.SIGINT)

	levelPercent, err := sensor.ReadWaterLevelPercent()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Read water level = %d%%", levelPercent)

	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Config:", cfg)

	err = waterLevelSanityCheck(levelPercent, db, cfg)
	if err != nil {
		log.Fatal(err)
	}
	err = db.InsertReading(levelPercent)
	if err != nil {
		log.Fatal(err)
	}

	if refillRequired(levelPercent, cfg) {
		refill(cfg)
	}

	log.Println("Done")
}

func refill(cfg *config.WaterConfig) error {
	log.Println("Turning on the pump.")
	err := pumpPin.Out(gpio.High)
	if err != nil {
		pumpPin.Out(gpio.Low)
		return err
	}
	time.Sleep(time.Duration(cfg.RefillTimeMilliseconds) * time.Millisecond)
	log.Println("Turning off the pump.")
	err = pumpPin.Out(gpio.Low)
	if err != nil {
		return err
	}
	return nil
}

func refillRequired(percent int, config *config.WaterConfig) bool {
	return percent < config.RefillIfBelowLevel
}

func waterLevelSanityCheck(percent int, readingsDB *db.ReadingsDB, config *config.WaterConfig) error {
	if percent == 0 {
		return errors.New("water level is 0%")
	}
	if readingsDB.IsEmpty() {
		log.Println("Readings database is empty")
		return nil
	}
	lastReading, err := readingsDB.GetLastReading()
	if err != nil {
		return err
	}
	log.Printf("Last reading: %d%%", lastReading)
	if lastReading-percent > config.MaxAllowedLevelDrop {
		return errors.New("suspicious water level drop")
	}
	return nil
}
