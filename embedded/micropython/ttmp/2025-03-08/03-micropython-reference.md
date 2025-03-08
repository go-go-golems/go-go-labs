# MicroPython on RP2040 Reference Guide

This comprehensive guide covers working with MicroPython on the RP2040 microcontroller, with a focus on the Raspberry Pi Pico and Pico W. It provides detailed examples and best practices for MicroPython development.

## Table of Contents
- [Hardware Overview](#hardware-overview)
- [Core Modules](#core-modules)
- [Hardware Interfaces](#hardware-interfaces)
- [Networking](#networking)
- [File System and Storage](#file-system-and-storage)
- [Advanced Features](#advanced-features)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

## Hardware Overview

### RP2040 Specifications
- Dual-core ARM Cortex M0+ processor running at up to 133MHz
- 264KB of SRAM
- 2MB of onboard Flash storage
- Hardware support for:
  - USB 1.1 Host/Device
  - UART, SPI, and I2C
  - Programmable I/O (PIO)
  - 30 GPIO pins
  - 3 ADC inputs (12-bit)
  - 16 PWM channels
  - Temperature sensor

### Pico W Additional Features
- 2.4GHz WiFi (802.11n)
- WPA3 security support
- Infineon CYW43439 wireless chip
- Onboard antenna

## Core Modules

### machine Module
The `machine` module is the primary interface for hardware control. It provides access to:

#### Pin Control
```python
from machine import Pin

# Create an output pin
led = Pin(25, Pin.OUT)  # Onboard LED on Pico
led.on()   # Turn LED on
led.off()  # Turn LED off
led.toggle()  # Toggle LED state

# Create an input pin
button = Pin(15, Pin.IN, Pin.PULL_UP)
value = button.value()  # Read pin state (0 or 1)

# Pin with interrupt
def callback(pin):
    print("Pin change detected!")

button.irq(trigger=Pin.IRQ_FALLING, handler=callback)
```

#### ADC (Analog-to-Digital Converter)
```python
from machine import ADC

# Create an ADC object
adc = ADC(26)  # ADC0 on GPIO26
value = adc.read_u16()  # Read 16-bit value (0-65535)

# Temperature sensor
temp_sensor = ADC(4)  # Internal temperature sensor
temp = 27 - (temp_sensor.read_u16() * 3.3 / 65535 - 0.706)/0.001721
```

#### PWM (Pulse Width Modulation)
```python
from machine import PWM

# Create a PWM object
pwm = PWM(Pin(15))
pwm.freq(1000)     # Set frequency to 1kHz
pwm.duty_u16(32768)  # Set duty cycle to 50% (range 0-65535)

# LED brightness control
led_pwm = PWM(Pin(25))
led_pwm.freq(1000)
for i in range(65535):
    led_pwm.duty_u16(i)
    time.sleep_ms(1)
```

#### I2C Interface
```python
from machine import I2C

# Software I2C
i2c = I2C(0, scl=Pin(17), sda=Pin(16), freq=400000)

# Scan for devices
devices = i2c.scan()

# Read from device
data = i2c.readfrom(device_addr, 4)  # Read 4 bytes

# Write to device
i2c.writeto(device_addr, bytes([0x3C, 0x42]))
```

#### SPI Interface
```python
from machine import SPI

# Initialize SPI
spi = SPI(0,
          baudrate=1000000,
          polarity=0,
          phase=0,
          bits=8,
          firstbit=SPI.MSB,
          sck=Pin(18),
          mosi=Pin(19),
          miso=Pin(16))

# Write and read data
spi.write(b'data')
result = spi.read(4)  # Read 4 bytes
```

#### UART (Serial) Interface
```python
from machine import UART

# Initialize UART
uart = UART(0, baudrate=115200)
uart.init(baudrate=115200, bits=8, parity=None, stop=1)

# Send and receive data
uart.write('Hello')
if uart.any():
    data = uart.read()
```

#### Real-Time Clock (RTC)
```python
from machine import RTC

rtc = RTC()
rtc.datetime((2024, 3, 8, 5, 15, 45, 0, 0))  # Set date/time
current = rtc.datetime()  # Get current date/time
```

### time Module
Essential for timing and delays:

```python
import time

# Delays
time.sleep(1)       # Sleep for 1 second
time.sleep_ms(500)  # Sleep for 500 milliseconds
time.sleep_us(10)   # Sleep for 10 microseconds

# Timing
start = time.ticks_ms()
# ... do something ...
duration = time.ticks_diff(time.ticks_ms(), start)
```

### micropython Module
Provides MicroPython-specific functions:

```python
import micropython

# Enable emergency exception buffer
micropython.alloc_emergency_exception_buf(100)

# Memory information
micropython.mem_info()

# Optimize code
@micropython.native
def optimized_function():
    pass

@micropython.viper
def highly_optimized_function():
    pass
```

## Hardware Interfaces

### Neopixel Support
Built-in support for WS2812 LEDs:

```python
from machine import Pin
import neopixel

# Initialize strip of 8 pixels
np = neopixel.NeoPixel(Pin(28), 8)

# Set pixel colors (R, G, B)
np[0] = (255, 0, 0)  # Red
np[1] = (0, 255, 0)  # Green
np[2] = (0, 0, 255)  # Blue

# Update strip
np.write()
```

### PIO (Programmable I/O)
Advanced hardware interface programming:

```python
from machine import Pin
import rp2

@rp2.asm_pio(set_init=rp2.PIO.OUT_LOW)
def blink_1hz():
    # Assembly program for 1Hz blink
    wrap_target()
    set(pins, 1)   [31]
    nop()          [31]
    nop()          [31]
    nop()          [31]
    set(pins, 0)   [31]
    nop()          [31]
    nop()          [31]
    nop()          [31]
    wrap()

# Create StateMachine with the program
sm = rp2.StateMachine(0, blink_1hz, freq=2000, set_base=Pin(25))
sm.active(1)  # Start the state machine
```

## Networking

### WiFi (Pico W)
```python
import network
import time

# Initialize WiFi interface
wlan = network.WLAN(network.STA_IF)
wlan.active(True)

# Connect to network
wlan.connect('SSID', 'PASSWORD')

# Wait for connection
max_wait = 10
while max_wait > 0:
    if wlan.status() < 0 or wlan.status() >= 3:
        break
    max_wait -= 1
    print('Waiting for connection...')
    time.sleep(1)

# Check connection status
if wlan.status() != 3:
    raise RuntimeError('Network connection failed')
else:
    print('Connected')
    status = wlan.ifconfig()
    print('IP:', status[0])
```

### Socket Programming
```python
import socket

# Create a TCP/IP socket
s = socket.socket()

# Connect to a server
s.connect(('example.com', 80))

# Send HTTP request
s.send(b'GET / HTTP/1.0\r\n\r\n')

# Receive data
data = s.recv(1024)
print(data)

# Close connection
s.close()
```

### Web Server Example
```python
import socket

def web_server():
    s = socket.socket()
    s.bind(('', 80))
    s.listen(5)
    
    while True:
        conn, addr = s.accept()
        request = conn.recv(1024)
        response = """HTTP/1.0 200 OK
Content-Type: text/html

<html><body><h1>Hello from Pico W!</h1></body></html>
"""
        conn.send(response)
        conn.close()
```

## File System and Storage

### Basic File Operations
```python
# Write to file
with open('data.txt', 'w') as f:
    f.write('Hello, Pico!')

# Read from file
with open('data.txt', 'r') as f:
    content = f.read()

# List directory contents
import os
files = os.listdir()

# Create directory
os.mkdir('data')

# Remove file
os.remove('old.txt')
```

## Best Practices

1. **Memory Management**
   - Use `gc.collect()` after large operations
   - Avoid creating large strings or lists
   - Use `bytearray` for large data buffers
   - Monitor memory usage with `micropython.mem_info()`

2. **Power Management**
   - Use sleep modes when possible
   - Minimize wireless usage on Pico W
   - Implement watchdog for reliability
   ```python
   from machine import WDT
   wdt = WDT(timeout=2000)  # 2 second timeout
   while True:
       wdt.feed()
       # Your code here
   ```

3. **Error Handling**
   ```python
   try:
       # Potentially failing operation
       result = risky_function()
   except Exception as e:
       print('Error:', e)
       # Handle error or reset device
       machine.reset()
   ```

4. **Code Organization**
   - Use modules for better organization
   - Implement boot.py for initialization
   - Use main.py for primary application
   - Keep critical functions in separate files

## Troubleshooting

1. **Common Issues**
   - Reset device: `machine.reset()`
   - Soft reset: Ctrl+D in REPL
   - Hard reset: BOOTSEL button
   - Check memory: `micropython.mem_info()`

2. **Debug Tools**
   ```python
   # Print debugging
   def debug_print(*args):
       print("DEBUG:", *args)
   
   # Memory debugging
   micropython.mem_info(1)  # Verbose output
   ```

3. **WiFi Diagnostics**
   ```python
   def check_wifi():
       if wlan.status() == network.STAT_GOT_IP:
           return 'Connected'
       elif wlan.status() == network.STAT_CONNECTING:
           return 'Connecting'
       elif wlan.status() == network.STAT_WRONG_PASSWORD:
           return 'Wrong Password'
       elif wlan.status() == network.STAT_NO_AP_FOUND:
           return 'No AP Found'
       else:
           return f'Failed, status: {wlan.status()}'
   ``` 