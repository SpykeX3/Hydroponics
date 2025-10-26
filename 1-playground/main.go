package main

import (
	"fmt"
	i2c "github.com/d2r2/go-i2c"
	"log"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/host/v3"
	"periph.io/x/host/v3/rpi"
	"time"
)

type Led1602 struct {
	e, rs, d0, d1, d2, d3, d4, d5, d6, d7 gpio.PinIO
}

func (l Led1602) Write(b byte, rs gpio.Level) error {
	fmt.Printf("Writing byte: 0x%02x\n", b)
	err := l.rs.Out(rs)
	if err != nil {
		return err
	}

	pins := []gpio.PinIO{l.d0, l.d1, l.d2, l.d3, l.d4, l.d5, l.d6, l.d7}
	for i, p := range pins {
		level := gpio.Low
		if b&(1<<i) != 0 {
			level = gpio.High
		}
		fmt.Println("Writing pin", p, p.Number(), "to", level)
		err := p.Out(level)
		if err != nil {
			return err
		}
	}
	return l.Tick()
}

func (l Led1602) Clear() error {
	return l.Write(0x01, gpio.Low)
}

func (l Led1602) Tick() error {
	time.Sleep(10 * time.Millisecond)
	err := l.e.Out(gpio.High)
	if err != nil {
		return err
	}
	time.Sleep(10 * time.Millisecond)
	err = l.e.Out(gpio.Low)
	if err != nil {
		return err
	}
	time.Sleep(10 * time.Millisecond)
	return nil
}

func (l Led1602) Blink() error {
	return l.Write(0x0f, gpio.Low)
}

func main() {
	_, err := host.Init()
	if err != nil {
		panic(err)
	}
	t := time.NewTicker(2000 * time.Millisecond)

	for l := gpio.Low; ; l = !l {
		err := rpi.P1_7.Out(l)
		if err != nil {
			fmt.Println("Error setting pin state:", err)
			return
		}
		fmt.Println("Set to", l)
		<-t.C
	}
	rpi.P1_7.Out(gpio.Low)

}

func i2cTest() {
	i2c, err := i2c.NewI2C(0x48, 1)
	if err != nil {
		log.Fatal(err)
	}
	defer i2c.Close()
	_, err = i2c.WriteBytes([]byte{0x01, 0x84, 0x83})
	checkRegister(i2c, 0x01)
	for {
		checkRegister(i2c, 0x00)
		time.Sleep(1 * time.Second)
	}
}

func checkRegister(i2c *i2c.I2C, register byte) error {
	_, err := i2c.WriteBytes([]byte{register})
	if err != nil {
		return err
	}
	bytes := make([]byte, 2)
	_, err = i2c.ReadBytes(bytes)
	if err != nil {
		return err
	}
	log.Printf("Register %1x: 0b%08b %08b", register, bytes[0], bytes[1])
	var vRaw int16
	vRaw = int16(bytes[0])<<8 | int16(bytes[1])
	log.Println(float32(vRaw) / (1 << 16))
	return nil
}
