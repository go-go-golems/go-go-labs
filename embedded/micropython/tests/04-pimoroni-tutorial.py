from picographics import PicoGraphics, DISPLAY_PICO_DISPLAY
from pimoroni import Button
import time

# Initialize the display
display = PicoGraphics(display=DISPLAY_PICO_DISPLAY)

# Initialize buttons
button_a = Button(12)
button_b = Button(13)
button_x = Button(14)
button_y = Button(15)

# Create color constants
WHITE = display.create_pen(255, 255, 255)
BLACK = display.create_pen(0, 0, 0)
RED = display.create_pen(255, 0, 0)
GREEN = display.create_pen(0, 255, 0)
BLUE = display.create_pen(0, 0, 255)
YELLOW = display.create_pen(255, 255, 0)

# Current rectangle color
current_color = WHITE

while True:
    # Clear the display
    display.set_pen(BLACK)
    display.clear()

    # Draw the text
    display.set_pen(WHITE)
    display.set_font("bitmap8")
    display.text("Hello,", 10, 20, scale=1)
    display.text("Pimoroni!", 10, 40, scale=1)

    # Check buttons and update color
    if button_a.read():
        current_color = RED
    elif button_b.read():
        current_color = GREEN
    elif button_x.read():
        current_color = BLUE
    elif button_y.read():
        current_color = YELLOW

    # Draw rectangle with current color
    display.set_pen(current_color)
    display.rectangle(5, 80, 100, 120)

    # Update the display
    display.update()
    
    # Small delay to prevent too rapid updates
    time.sleep(0.1)