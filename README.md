# Hydroponics

A compilation of small Go programs to automate parts of a hydroponics setup on a Raspberry Pi. The repo currently contains:

- 1-playground — quick experiments with GPIO/I²C
- 2-light-control — turns a grow light on/off based on a daily schedule
- 3-water-refill — reads a Grove Water Level sensor and runs a pump to refill to a safe level

> Note: This README documents only what is evident from the code. Any missing operational or hardware details are marked as TODO for future clarification.

## Requirements

Software
- Go toolchain (module files declare `go 1.24`). TODO: Confirm the exact Go version used in production.
- Raspberry Pi OS or Linux on hardware that exposes Raspberry Pi-compatible GPIO and I²C.
- I²C enabled (e.g., via `raspi-config`) and user permitted to access `/dev/i2c-1`.
- Optional: permissions for GPIO without root, or run with sufficient privileges.

Hardware (as inferred from code)
- Raspberry Pi with 40-pin header.
- Grow light switched via a GPIO-controlled relay on pin `rpi.P1_7`. TODO: Confirm the relay board and BCM pin mapping.
- Water pump switched via a GPIO-controlled relay on pin `rpi.P1_11`. TODO: Confirm wiring and BCM pin mapping.
- Grove Water Level sensor (20 segments; two I²C devices at 0x77 and 0x78) connected to I²C bus 1.

## Project Structure

```
.
├── 1-playground/                 # Misc experiments (GPIO/I²C)
│   ├── main.go                   # I²C ADC/register read demo
│   └── lcd.go                    # (playground code)
├── 2-light-control/
│   └── main.go                   # Light on/off based on schedule
├── 3-water-refill/
│   ├── main.go                   # Reads water level, logs, and refills via pump
│   ├── config/config.go          # Loads/creates JSON config in $HOME/water_refill.config
│   ├── db/database.go            # SQLite (modernc.org/sqlite) readings DB in $HOME/readings.db
│   └── water-level-sensor/       # Grove Water Level sensor helper
│       └── sensor.go
└── README.md
```

## Build

Each subfolder is its own Go module. Build from within each directory.

- Light control
  ```bash
  cd 2-light-control
  go build -o light-control
  ```

- Water refill
  ```bash
  cd 3-water-refill
  go build -o water-refill
  ```

## Run

- Light control (scheduled on/off based on config)
  ```bash
  cd 2-light-control
  ./light-control
  ```
  Behavior:
  - Uses Raspberry Pi pin `rpi.P1_7` to control the light relay.
  - Turns the light ON only if current time is between configured `start_time` and `finish_time`. Otherwise, sets the pin LOW (light OFF).

- Water refill (measure and top-up)
  ```bash
  cd 3-water-refill
  ./water-refill
  ```
  Behavior:
  - Initializes host peripherals and Grove Water Level sensor at I²C addresses 0x77 (low 8 sections) and 0x78 (high 12 sections) on bus 1.
  - Computes a water level percentage (5% per contiguous touched section from the bottom).
  - Performs a sanity check against the last reading to avoid sudden suspicious drops.
  - Logs the reading to a local SQLite DB file at `$HOME/readings.db`.
  - If current level is below threshold, activates the pump on pin `rpi.P1_11` for a configured duration, then turns it off.

### Running as a service

- TODO: Provide a `cron` unit example for each binary.
- TODO: Document required udev rules or group memberships to access GPIO/I²C without root.

## Configuration

Configuration files are created with default values on first run (if not present) and stored in the user home directory.

- 2-light-control
  - Config path: `$HOME/light.config`
  - JSON schema (string durations, `time.ParseDuration` syntax):
    ```json
    {
      "start_time": "7h",
      "finish_time": "22h"
    }
    ```

- 3-water-refill
  - Config path: `$HOME/water_refill.config`
  - JSON schema:
    ```json
    {
      "max_allowed_level_drop": 25,
      "refill_if_below_level": 55,
      "refill_time_milliseconds": 2000
    }
    ```
  - Notes:
    - `refill_time_milliseconds` is clamped in code to the [500, 5000] ms range.
    - The water level sensor threshold is currently hard-coded to `100` in the program.

## Data Storage

- SQLite database at `$HOME/readings.db` with table `readings(id INTEGER PRIMARY KEY AUTOINCREMENT, reading INTEGER, timestamp INTEGER)`.
- The application appends a new row on each run of `3-water-refill`.
