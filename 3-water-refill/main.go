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

var pumpInPin = rpi.P1_11
var pumpOutPin = rpi.P1_13

func main() {

	db, err := db.OpenStateDB()
	if err != nil {
		log.Panicln(err)
	}
	defer db.Close()

	if _, err = host.Init(); err != nil {
		panic(err)
	}

	defer func() {
		if r := recover(); r != nil {
			log.Printf("Panic recovered: %v", r)
			// Ensure both pump pins are set to Low for safety
			pumpInPin.Out(gpio.Low)
			pumpOutPin.Out(gpio.Low)
			log.Println("Emergency: Both pump pins set to Low")
		}
	}()

	sensor, err := wlevel.NewWaterLevelSensor(1, addrLow, addrHigh, threshold)
	if err != nil {
		log.Panicln(err)
	}
	defer sensor.Close()

	// Handle Ctrl+C gracefully
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	signal.Notify(stop, os.Interrupt, syscall.SIGINT)

	levelPercent, err := sensor.ReadWaterLevelPercent()
	if err != nil {
		log.Panicln(err)
	}
	log.Printf("Read water level = %d%%", levelPercent)

	cfg, err := config.GetConfig()
	if err != nil {
		log.Panicln(err)
	}

	log.Println("Config:", cfg)

	err = waterLevelSanityCheck(levelPercent, db, cfg)
	if err != nil {
		log.Panicln(err)
	}
	err = db.InsertReading(levelPercent)
	if err != nil {
		log.Panicln(err)
	}

	state, previousState := calculateDesiredState(db, cfg)
	if state != previousState {
		log.Printf("Changing state from %s to %s", previousState, state)
		err = db.InsertState(state)
		if err != nil {
			log.Panicln(err)
		}
	} else {
		log.Println("State remains ", state)
	}

	switch state {
	case "FILL":
		if refillRequired(levelPercent, cfg) {
			err := refill(cfg)
			if err != nil {
				log.Panicln(err)
			}
		}
	case "EMPTY":
		if emptyingRequired(levelPercent, cfg) {
			err := empty(cfg)
			if err != nil {
				log.Panicln(err)
			}
		}
	}

	log.Println("Done")
}

func emptyingRequired(percent int, cfg *config.WaterConfig) bool {
	return percent > cfg.EmptyIfAboveLevel
}

func empty(cfg *config.WaterConfig) error {
	log.Println("Turning on the out-pump.")
	err := pumpOutPin.Out(gpio.High)
	if err != nil {
		pumpOutPin.Out(gpio.Low)
		return err
	}
	time.Sleep(time.Duration(cfg.EmptyTimeMilliseconds) * time.Millisecond)
	log.Println("Turning off the out-pump.")
	err = pumpOutPin.Out(gpio.Low)
	if err != nil {
		return err
	}
	return nil
}

func calculateDesiredState(stateDB *db.StateDB, cfg *config.WaterConfig) (db.State, db.State) {
	if stateDB.IsStateEmpty() {
		log.Println("State database is empty")
		return db.StateFill, db.StateEmpty
	}
	lastState, timestamp, err := stateDB.GetLastState()
	if err != nil {
		log.Fatal(err)
	}
	if lastState == db.StateFill {
		if time.Now().Unix()-timestamp > cfg.FillForSeconds {
			return db.StateEmpty, lastState
		}
		return lastState, lastState
	} else {
		if time.Now().Unix()-timestamp > cfg.EmptyForSeconds {
			return db.StateFill, lastState
		}
		return lastState, lastState
	}

}

func refill(cfg *config.WaterConfig) error {
	log.Println("Turning on the refill pump.")
	err := pumpInPin.Out(gpio.High)
	if err != nil {
		pumpInPin.Out(gpio.Low)
		return err
	}
	time.Sleep(time.Duration(cfg.RefillTimeMilliseconds) * time.Millisecond)
	log.Println("Turning off the refill pump.")
	err = pumpInPin.Out(gpio.Low)
	if err != nil {
		return err
	}
	return nil
}

func refillRequired(percent int, config *config.WaterConfig) bool {
	return percent < config.RefillIfBelowLevel
}

func waterLevelSanityCheck(percent int, readingsDB *db.StateDB, config *config.WaterConfig) error {
	if percent == 0 {
		return errors.New("water level is 0%")
	}
	if readingsDB.IsReadingsEmpty() {
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
