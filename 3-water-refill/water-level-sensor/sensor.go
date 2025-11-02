package water_level_sensor

import (
	"github.com/d2r2/go-i2c"
	"time"
)

// WaterLevelSensor encapsulates the two I2C devices of the Grove Water Level sensor.
// It reads the 8 low sections (0x77) and the 12 high sections (0x78), builds a 20-bit
// mask of touched sections and computes a water level percentage.
type WaterLevelSensor struct {
	lowDev    *i2c.I2C
	highDev   *i2c.I2C
	threshold byte
	lowBuf    []byte // len 8
	highBuf   []byte // len 12
}

// NewWaterLevelSensor opens I2C on the given bus for the two device addresses.
func NewWaterLevelSensor(bus int, lowAddr, highAddr byte, threshold int) (*WaterLevelSensor, error) {
	low, err := i2c.NewI2C(lowAddr, bus)
	if err != nil {
		return nil, err
	}
	high, err := i2c.NewI2C(highAddr, bus)
	if err != nil {
		_ = low.Close()
		return nil, err
	}
	return &WaterLevelSensor{
		lowDev:    low,
		highDev:   high,
		threshold: byte(threshold),
		lowBuf:    make([]byte, 8),
		highBuf:   make([]byte, 12),
	}, nil
}

// Close releases I2C devices.
func (s *WaterLevelSensor) Close() {
	if s.lowDev != nil {
		_ = s.lowDev.Close()
	}
	if s.highDev != nil {
		_ = s.highDev.Close()
	}
}

// readExactly reads len(buf) bytes from the given device.
func (s *WaterLevelSensor) readExactly(dev *i2c.I2C, buf []byte) error {
	n := len(buf)
	read := 0
	for read < n {
		m, err := dev.ReadBytes(buf[read:])
		if err != nil {
			return err
		}
		if m == 0 {
			return nil
		}
		read += m
	}
	return nil
}

// ReadSections reads low (8 bytes) and high (12 bytes) sections with small delays.
func (s *WaterLevelSensor) ReadSections() (low, high []byte, err error) {
	if err = s.readExactly(s.lowDev, s.lowBuf); err != nil {
		return nil, nil, err
	}
	time.Sleep(10 * time.Millisecond)
	if err = s.readExactly(s.highDev, s.highBuf); err != nil {
		return nil, nil, err
	}
	time.Sleep(10 * time.Millisecond)
	return append([]byte(nil), s.lowBuf...), append([]byte(nil), s.highBuf...), nil
}

// BuildTouchMask builds the 20-bit mask from low/high values using the threshold.
func (s *WaterLevelSensor) BuildTouchMask(low, high []byte) uint32 {
	var mask uint32
	for i := 0; i < len(low); i++ {
		if low[i] > s.threshold {
			mask |= 1 << uint(i)
		}
	}
	for i := 0; i < len(high); i++ {
		if high[i] > s.threshold {
			mask |= 1 << uint(8+i)
		}
	}
	return mask
}

// ComputeLevelPercent returns trigSection*5, where trigSection is the number of
// contiguous set bits from the LSB of mask.
func (s *WaterLevelSensor) ComputeLevelPercent(mask uint32) int {
	trigSection := 0
	for (mask & 0x01) != 0 {
		trigSection++
		mask >>= 1
	}
	return trigSection * 5
}

func (s *WaterLevelSensor) ReadWaterLevelPercent() (int, error) {
	low, high, err := s.ReadSections()
	if err != nil {
		return 0, err
	}
	mask := s.BuildTouchMask(low, high)
	return s.ComputeLevelPercent(mask), nil
}
