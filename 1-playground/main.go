package main

import (
	i2c "github.com/d2r2/go-i2c"
	"log"
	"periph.io/x/host/v3"
	"time"
)

func main() {
	_, err := host.Init()
	if err != nil {
		panic(err)
	}
	/*
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
	*/
	i2c2ATest()

}

func i2c2ATest() {
	i2c, err := i2c.NewI2C(0x48, 1)
	if err != nil {
		log.Fatal(err)
	}
	defer i2c.Close()
	_, err = i2c.WriteBytes([]byte{0x01, 0x84, 0x83})
	checkConverterRegister(i2c, 0x01)
	for {
		checkConverterRegister(i2c, 0x00)
		time.Sleep(1 * time.Second)
	}
}

func checkConverterRegister(i2c *i2c.I2C, register byte) error {
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
