# Changelog

## Pico W and Pico Display Documentation
Added comprehensive documentation for Raspberry Pi Pico W and Pico Display (240x135) development. The documentation includes hardware specifications, setup instructions, display programming, button handling, graphics programming, WiFi connectivity, and advanced topics with practical examples.

- Added detailed hardware specifications for both Pico W and Pico Display
- Included code examples for display initialization and basic operations
- Added button handling with debouncing and event system
- Provided graphics programming examples including animations
- Added WiFi connectivity setup guide
- Included power management and menu system examples
- Added best practices and troubleshooting sections 

## Fixed Pimoroni Display Tutorial Example
Fixed issues in the Pimoroni Display tutorial example:
- Corrected display constant from DISPLAY_PIMORONI_PICO_DISPLAY to DISPLAY_PICO_DISPLAY
- Fixed pen creation method to use create_pen instead of direct RGB values
- Added clear comments explaining each step 

## Improved Text Display for Small Screen
Optimized text display for the 240x135 Pico Display:
- Split text into multiple lines for better readability
- Used smaller font and scale settings
- Added black background for contrast
- Added visual boundary rectangle
- Adjusted text positioning for better centering 