## Cardputer Bike Trainer — NimBLE BLE Central for M5Stack Cardputer

This project is a small ESP-IDF application that runs on the **M5Stack Cardputer** (ESP32-S3, 8MB flash) and acts as a **BLE Central** using the NimBLE stack.
It is based on Espressif's `blecent` example and is currently a **headless** app: it does not use the Cardputer display or keyboard yet, but it can scan for and connect to BLE peripherals and log its behaviour over the serial console.

You can use it either with a real BLE peripheral (for example, a bike trainer that exposes the right GATT services) or with the included Python test script (`blecent_test.py`), which emulates a BLE GATT server on your PC.

---

## How it Works (High-Level)

- On boot, the app initializes the NimBLE host and configures the ESP32-S3 Bluetooth controller in **BLE-only** mode.
- It starts a **passive scan** for nearby BLE devices.
- When it finds a device that:
  - Is connectable, and
  - Advertises the **Alert Notification Service** (UUID `0x1811`) as a primary service,
  it initiates a connection to that device.
- After connecting, it can:
  - Read characteristics.
  - Write to characteristics.
  - Subscribe to notifications.
  - Perform a small sequence of GATT operations against the peer.
- All logs and debug output are printed to the serial console using ESP-IDF logging macros, so you can observe the BLE procedures in real time.

The current code does **not yet integrate** with the Cardputer UI framework (`HalCardputer`, `mooncake`, etc.). That means there is no on-screen feedback. All interaction is via BLE, visible in the logs.

---

## Project Layout

- `CMakeLists.txt` — standard ESP-IDF project definition (project name `blecent`).
- `sdkconfig` — configuration for ESP-IDF (target set to `esp32s3`, 8MB flash, custom partition table).
- `partitions.csv` — partition table aligned with the official Cardputer demo:
  - 4MB factory app partition.
  - 1MB FAT `storage` partition.
- `main/`:
  - `main.c` — main BLE central application logic.
  - `misc.c`, `peer.c` — helper code for NimBLE peer and GATT handling.
- `blecent_test.py` — Python utility that implements a BLE GATT server for testing the central.

---

## Prerequisites

- **Hardware**
  - M5Stack Cardputer (ESP32-S3 with 8MB flash).
  - USB-C cable for power and flashing.

- **Software**
  - ESP-IDF **v4.4.x** (recommended: the same version used by `M5Cardputer-UserDemo`, e.g. 4.4.6).
  - Python 3.x with `pip`.
  - Build tools installed by ESP-IDF installer (CMake, Ninja, etc.).

Make sure your ESP-IDF environment is activated (e.g. `source export.sh` from your ESP-IDF install) before building this project.

---

## Configuring for Cardputer

This project is already configured for the Cardputer:

- **Target chip**: `esp32s3` (set in `sdkconfig`).
- **Flash size**: 8MB (matches the Cardputer module).
- **Partition table**: `partitions.csv`, the same layout used by `M5Cardputer-UserDemo`:
  - `nvs` and `phy_init` data partitions.
  - `factory` app partition at `0x10000` with size `4M`.
  - `storage` FAT partition of `1M`.
- **BLE stack**: NimBLE enabled, Bluedroid disabled.

If you ever need to reconfigure:

```bash
idf.py set-target esp32s3
idf.py menuconfig
```

Then verify under:

- **Serial flasher config → Flash size**: `8 MB`.
- **Partition Table → Partition Table**: `Custom partition table CSV`.
- **Partition Table → Custom partition CSV file**: `partitions.csv`.

---

## Building the Firmware

From the project root (`go-go-labs/embedded/cardputer-bike-trainer`):

```bash
idf.py set-target esp32s3        # safe to run even if already set
idf.py build
```

This will:

- Configure the project for ESP32-S3 (if not already done).
- Compile the NimBLE BLE central application.
- Produce firmware images under `build/` (e.g. `blecent.bin`).

---

## Flashing to the M5Stack Cardputer

Connect the Cardputer via USB-C, then determine the serial port (e.g. `/dev/ttyACM0`, `/dev/ttyUSB0`, etc.).

Flash and monitor in one step:

```bash
idf.py -p /dev/ttyACM0 flash monitor
```

If you prefer separate steps:

```bash
idf.py -p /dev/ttyACM0 flash
idf.py -p /dev/ttyACM0 monitor
```

Replace `/dev/ttyACM0` with the correct port for your system.  
To exit the monitor, press `Ctrl+]`.

---

## Using the Python Test Utility (Optional)

If you don't have a suitable BLE peripheral handy, you can use the included Python script `blecent_test.py` to emulate one on your PC.

1. Install required Python dependencies:

```bash
python -m pip install --user -r "$IDF_PATH"/requirements.txt -r "$IDF_PATH"/tools/ble/requirements.txt
```

2. Run the Python GATT server (Linux with BlueZ + D-Bus is required):

```bash
python blecent_test.py
```

3. With the Cardputer firmware running and the script active:
   - The Cardputer (central) should discover and connect to the Python GATT server.
   - You will see log messages on both:
     - The Cardputer serial monitor (NimBLE central logs).
     - The Python console (GATT server logs).

This is a convenient way to verify that the BLE central logic and GATT operations are working end to end.

---

## Next Steps / Extending for Full Cardputer UI

Right now, this project proves that the **BLE central logic** works on the Cardputer hardware using NimBLE and the correct flash/partition layout.

Future improvements could include:

- Integrating with the Cardputer HAL (`HalCardputer`) and Mooncake framework from `M5Cardputer-UserDemo`.
- Adding a simple on-screen status app that shows:
  - Scan status.
  - Connected device name/address.
  - Basic GATT state or notifications.
- Mapping keyboard shortcuts to actions (e.g. start/stop scanning).

Those enhancements are intentionally left for later so that this example stays small and focused on BLE central behaviour first.


