from picographics import PicoGraphics, DISPLAY_PICO_DISPLAY
from pimoroni import Button

# Initialize the display
display = PicoGraphics(display=DISPLAY_PICO_DISPLAY)

# Create color constants
WHITE = display.create_pen(255, 255, 255)
BLACK = display.create_pen(0, 0, 0)

# Clear the display first
display.set_pen(BLACK)
display.clear()

# Set the pen color and draw text
display.set_pen(WHITE)
display.set_font("bitmap8")  # Use a smaller font
display.text("Hello,", 10, 20, scale=1)
display.text("Pimoroni!", 10, 40, scale=1)

# Draw a rectangle around the text to show display boundaries
display.set_pen(WHITE)
display.rectangle(5, 15, 100, 35)

# Update the display to show changes
display.update()