# Raspberry Pi Pico W with Pico Display (240x135) Reference Guide

This comprehensive guide covers working with the Raspberry Pi Pico W (RP2040) in combination with the Pimoroni Pico Display Pack (240x135 LCD). It provides detailed examples and best practices for MicroPython development.

## Table of Contents
- [Hardware Overview](#hardware-overview)
- [Initial Setup](#initial-setup)
- [Display Basics](#display-basics)
- [Button Handling](#button-handling)
- [Graphics Programming](#graphics-programming)
- [WiFi Connectivity](#wifi-connectivity)
- [Advanced Topics](#advanced-topics)

## Hardware Overview

### Pico W Specifications
- RP2040 microcontroller
- Dual-core ARM Cortex M0+ processor
- 264KB RAM
- 2MB Flash storage
- 2.4GHz WiFi
- Operating voltage: 3.3V
- GPIO: 26 multi-function GPIO pins

### Pico Display Pack Features
- 240x135 pixel IPS LCD screen
- 4 tactile buttons (A, B, X, Y)
- RGB LED
- Display connector using GPIO pins
- Dimensions: 66.1mm x 18.5mm

## Initial Setup

### Required Libraries
```python
from picographics import PicoGraphics, DISPLAY_PICO_DISPLAY
from pimoroni import Button
import network
import time
```

### Basic Display Initialization
```python
display = PicoGraphics(display=DISPLAY_PICO_DISPLAY)
display.set_backlight(1.0)  # Full brightness (0.0 to 1.0)

# Create button objects
button_a = Button(12)
button_b = Button(13)
button_x = Button(14)
button_y = Button(15)

# Create color constants
BLACK = display.create_pen(0, 0, 0)
WHITE = display.create_pen(255, 255, 255)
RED = display.create_pen(255, 0, 0)
GREEN = display.create_pen(0, 255, 0)
BLUE = display.create_pen(0, 0, 255)
```

## Display Basics

### Screen Coordinates
The display uses a coordinate system where (0,0) is at the top-left corner:
- X-axis: 0 to 239 (left to right)
- Y-axis: 0 to 134 (top to bottom)

### Basic Drawing Operations
```python
# Clear the display
display.set_pen(BLACK)
display.clear()

# Draw a pixel
display.set_pen(WHITE)
display.pixel(120, 67)  # Center point

# Draw a line
display.line(0, 0, 239, 134)  # Diagonal line

# Draw a rectangle
display.rectangle(10, 10, 50, 30)  # x, y, width, height

# Draw a circle
display.circle(120, 67, 30)  # x, y, radius

# Add text
display.set_font("bitmap8")
display.text("Hello, Pico!", 10, 10, scale=2)

# Update display
display.update()
```

### Text Rendering
```python
# Available fonts
FONTS = ["bitmap6", "bitmap8", "bitmap14_outline", "sans"]

def draw_text(text, x, y, font="bitmap8", scale=1):
    display.set_font(font)
    display.text(text, x, y, scale=scale)
    width = display.measure_text(text, scale=scale)
    return width
```

## Button Handling

### Basic Button Reading
```python
def check_buttons():
    if button_a.read():
        return "A"
    elif button_b.read():
        return "B"
    elif button_x.read():
        return "X"
    elif button_y.read():
        return "Y"
    return None

# Example usage with debouncing
def button_loop():
    last_press = time.ticks_ms()
    debounce_time = 200  # milliseconds
    
    while True:
        current_time = time.ticks_ms()
        if time.ticks_diff(current_time, last_press) > debounce_time:
            button = check_buttons()
            if button:
                print(f"Button {button} pressed")
                last_press = current_time
        time.sleep(0.01)
```

### Event-Based Button Handling
```python
class ButtonHandler:
    def __init__(self):
        self.callbacks = {
            'A': None,
            'B': None,
            'X': None,
            'Y': None
        }
        self.last_press = time.ticks_ms()
        self.debounce_time = 200
    
    def register_callback(self, button, callback):
        self.callbacks[button] = callback
    
    def update(self):
        current_time = time.ticks_ms()
        if time.ticks_diff(current_time, self.last_press) > self.debounce_time:
            button = check_buttons()
            if button and self.callbacks[button]:
                self.callbacks[button]()
                self.last_press = current_time
```

## Graphics Programming

### Drawing Shapes
```python
def draw_rounded_rect(x, y, width, height, radius):
    display.set_pen(WHITE)
    # Main rectangle
    display.rectangle(x + radius, y, width - 2*radius, height)
    display.rectangle(x, y + radius, width, height - 2*radius)
    # Corners
    display.circle(x + radius, y + radius, radius)
    display.circle(x + width - radius, y + radius, radius)
    display.circle(x + radius, y + height - radius, radius)
    display.circle(x + width - radius, y + height - radius, radius)

def draw_progress_bar(x, y, width, height, progress):
    """Draw a progress bar (0.0 to 1.0)"""
    display.set_pen(WHITE)
    display.rectangle(x, y, width, height)
    display.set_pen(BLUE)
    bar_width = int(width * max(0, min(1, progress)))
    if bar_width > 0:
        display.rectangle(x + 2, y + 2, bar_width - 4, height - 4)
```

### Animation
```python
def bounce_ball():
    x, y = 120, 67
    dx, dy = 2, 2
    radius = 10
    
    while True:
        display.set_pen(BLACK)
        display.clear()
        
        # Update position
        x += dx
        y += dy
        
        # Bounce off edges
        if x - radius <= 0 or x + radius >= 239:
            dx = -dx
        if y - radius <= 0 or y + radius >= 134:
            dy = -dy
        
        # Draw ball
        display.set_pen(WHITE)
        display.circle(int(x), int(y), radius)
        display.update()
        time.sleep(0.01)
```

## WiFi Connectivity

### Basic WiFi Setup
```python
def connect_wifi(ssid, password):
    wlan = network.WLAN(network.STA_IF)
    wlan.active(True)
    wlan.connect(ssid, password)
    
    # Wait for connection with timeout
    max_wait = 10
    while max_wait > 0:
        if wlan.status() < 0 or wlan.status() >= 3:
            break
        max_wait -= 1
        print('Waiting for connection...')
        time.sleep(1)
    
    if wlan.status() != 3:
        raise RuntimeError('WiFi connection failed')
    
    status = wlan.ifconfig()
    print(f'Connected to {ssid}')
    print(f'IP: {status[0]}')
    return wlan
```

## Advanced Topics

### Power Management
```python
def set_display_power(on):
    if on:
        display.set_backlight(1.0)
    else:
        display.set_backlight(0.0)

def power_saving_mode():
    """Enable power saving features"""
    display.set_backlight(0.5)  # Reduce brightness
    # Add other power-saving measures as needed
```

### Complete Example: Menu System
```python
class MenuItem:
    def __init__(self, text, callback):
        self.text = text
        self.callback = callback

class Menu:
    def __init__(self):
        self.items = []
        self.selected = 0
    
    def add_item(self, text, callback):
        self.items.append(MenuItem(text, callback))
    
    def draw(self):
        display.set_pen(BLACK)
        display.clear()
        
        y = 10
        for i, item in enumerate(self.items):
            display.set_pen(WHITE)
            if i == self.selected:
                display.rectangle(0, y-2, 240, 20)
                display.set_pen(BLACK)
            display.text(item.text, 10, y, scale=2)
            y += 25
        
        display.update()
    
    def navigate(self, direction):
        self.selected = (self.selected + direction) % len(self.items)
        self.draw()
    
    def select(self):
        if 0 <= self.selected < len(self.items):
            self.items[self.selected].callback()

# Example usage
def create_menu():
    menu = Menu()
    menu.add_item("Option 1", lambda: print("Selected 1"))
    menu.add_item("Option 2", lambda: print("Selected 2"))
    menu.add_item("Option 3", lambda: print("Selected 3"))
    
    # Button handlers
    button_handler = ButtonHandler()
    button_handler.register_callback('A', lambda: menu.navigate(-1))
    button_handler.register_callback('B', lambda: menu.navigate(1))
    button_handler.register_callback('X', menu.select)
    
    return menu, button_handler
```

## Best Practices

1. **Memory Management**
   - Clear unused variables
   - Use garbage collection when needed
   - Avoid creating large objects in loops

2. **Display Updates**
   - Batch drawing operations
   - Only update the display when needed
   - Use double-buffering for smooth animations

3. **Button Handling**
   - Always implement debouncing
   - Use event-based approaches for complex applications
   - Consider long-press and double-press scenarios

4. **Error Handling**
   - Implement proper try-except blocks
   - Handle WiFi connection failures gracefully
   - Monitor memory usage

## Troubleshooting

Common issues and solutions:

1. **Display Not Updating**
   - Check if `display.update()` is called
   - Verify display initialization
   - Check power connections

2. **Button Issues**
   - Verify button pin configurations
   - Check debouncing implementation
   - Test button hardware

3. **WiFi Connection Problems**
   - Verify credentials
   - Check signal strength
   - Monitor connection status

4. **Performance Issues**
   - Reduce update frequency
   - Optimize drawing operations
   - Use appropriate sleep intervals
