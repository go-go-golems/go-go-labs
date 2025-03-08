import time
from picographics import PicoGraphics, DISPLAY_PICO_DISPLAY
from pimoroni import Button
from machine import Timer

# Initialize display
display = PicoGraphics(display=DISPLAY_PICO_DISPLAY)
WIDTH, HEIGHT = display.get_bounds()  # 240x135 in landscape mode

# Setup buttons - Standard Pimoroni button configuration
button_a = Button(12)
button_b = Button(13)
button_c = Button(14)
button_d = Button(15)

# Define colors
BLACK = display.create_pen(0, 0, 0)
WHITE = display.create_pen(255, 255, 255)
RED = display.create_pen(255, 0, 0)
GREEN = display.create_pen(0, 255, 0)
BLUE = display.create_pen(0, 0, 255)
YELLOW = display.create_pen(255, 255, 0)
LIGHT_BLUE = display.create_pen(100, 150, 255)

# Chemical presets
DEVELOPERS = ["D-76", "Rodinal", "HC-110", "XTOL", "Pyro"]
FIXERS = ["Rapid Fix", "TF-4", "Ilford", "Kodak"]

# Font sizes
FONT_SMALL = "bitmap6"
FONT_MEDIUM = "bitmap8"
FONT_LARGE = "bitmap14_outline"

# Screen IDs
SCREEN_HOME = 0
SCREEN_DEVELOP = 1
SCREEN_FIX = 2
SCREEN_RESULTS = 3 
SCREEN_LOGS = 4

# Global variables
current_screen = SCREEN_HOME
current_experiment = 1
current_developer = 0
current_fixer = 0
develop_time = 0
fix_time = 0
timer_running = False
timer_started = 0
results = []  # Store results for later retrieval
last_button_time = 0  # For debouncing

def debounce():
    """Implement button debouncing"""
    global last_button_time
    current_time = time.ticks_ms()
    if time.ticks_diff(current_time, last_button_time) > 200:  # 200ms debounce
        last_button_time = current_time
        return True
    return False

# Timer interrupt for counting seconds
timer = Timer(-1)  # Use any available timer ID

def update_timer(timer):
    """Timer callback function"""
    global develop_time, fix_time, timer_running, timer_started, current_screen
    
    if timer_running:
        current_time = time.ticks_ms()
        elapsed = (current_time - timer_started) // 1000  # Convert to seconds
        
        if current_screen == SCREEN_DEVELOP:
            develop_time = elapsed
        elif current_screen == SCREEN_FIX:
            fix_time = elapsed
            
        draw_screen()

def setup():
    """Initialize the application"""
    # Setup timer interrupt to trigger every 100ms
    timer.init(freq=10, mode=Timer.PERIODIC, callback=update_timer)
    draw_screen()

def center_text(text, y, font=FONT_MEDIUM, scale=1):
    """Helper to center text horizontally"""
    display.set_font(font)
    text_width = display.measure_text(text, scale=scale)
    x = (WIDTH - text_width) // 2
    display.text(text, x, y, scale=scale)
    return text_width

def draw_button_labels(labels):
    """Draw the button labels at the bottom of the screen"""
    # Button labels at the bottom
    y = HEIGHT - 15
    
    # Calculate spacing for 4 buttons
    button_width = WIDTH // 4
    
    for i, label in enumerate(labels):
        if label:
            display.set_pen(WHITE)
            display.set_font(FONT_SMALL)
            
            # Center within button space
            text_width = display.measure_text(label, scale=1)
            x = (i * button_width) + ((button_width - text_width) // 2)
            
            display.text(label, x, y, scale=1)

def draw_home_screen():
    """Draw the home setup screen"""
    # Clear display with black background
    display.set_pen(BLACK)
    display.clear()
    
    # Title - Experiment number
    display.set_pen(LIGHT_BLUE)
    center_text(f"Experiment #{current_experiment}", 10, FONT_MEDIUM, 2)
    
    # Developer section
    y = 40
    display.set_pen(WHITE)
    display.set_font(FONT_MEDIUM)
    display.text("Developer:", 20, y, scale=1)
    y += 15
    display.set_pen(WHITE)
    display.set_font(FONT_MEDIUM)
    display.text(DEVELOPERS[current_developer], 110, y, scale=1)
    
    # Fixer section
    y += 25
    display.set_pen(WHITE)
    display.set_font(FONT_MEDIUM)
    display.text("Fixer:", 20, y, scale=1)
    y += 15
    display.set_pen(WHITE)
    display.set_font(FONT_MEDIUM)
    display.text(FIXERS[current_fixer], 110, y, scale=1)
    
    # Draw button labels
    draw_button_labels(["A: Dev+", "B: Fix+", "C: Start", "D: Exp+"])
    
    # Update display
    display.update()

def draw_develop_screen():
    """Draw the development timer screen"""
    # Clear display with black background
    display.set_pen(BLACK)
    display.clear()
    
    # Title
    display.set_pen(GREEN)
    center_text("DEVELOPING", 10, FONT_MEDIUM, 2)
    
    # Developer info
    y = 35
    display.set_pen(WHITE)
    display.set_font(FONT_SMALL)
    dev_text = f"Developer: {DEVELOPERS[current_developer]}"
    center_text(dev_text, y, FONT_SMALL, 1)
    
    # Timer display - large centered
    y = 60
    display.set_pen(WHITE)
    display.set_font(FONT_LARGE)
    time_text = format_time(develop_time)
    center_text(time_text, y, FONT_LARGE, 2)
    
    # Draw button labels
    draw_button_labels(["", "", "C: To Fix", "D: Cancel"])
    
    # Update display
    display.update()

def draw_fix_screen():
    """Draw the fix timer screen"""
    # Clear display with black background
    display.set_pen(BLACK)
    display.clear()
    
    # Title
    display.set_pen(BLUE)
    center_text("FIXING", 10, FONT_MEDIUM, 2)
    
    # Fixer info
    y = 35
    display.set_pen(WHITE)
    display.set_font(FONT_SMALL)
    fix_text = f"Fixer: {FIXERS[current_fixer]}"
    center_text(fix_text, y, FONT_SMALL, 1)
    
    # Timer display - large centered
    y = 60
    display.set_pen(WHITE)
    display.set_font(FONT_LARGE)
    time_text = format_time(fix_time)
    center_text(time_text, y, FONT_LARGE, 2)
    
    # Draw button labels
    draw_button_labels(["", "", "C: Done", "D: Cancel"])
    
    # Update display
    display.update()

def draw_results_screen():
    """Draw the results summary screen"""
    # Clear display with black background
    display.set_pen(BLACK)
    display.clear()
    
    # Title
    display.set_pen(GREEN)
    center_text("COMPLETE", 10, FONT_MEDIUM, 2)
    
    # Result summary
    y = 35
    display.set_pen(WHITE)
    display.set_font(FONT_SMALL)
    
    # Place information in a grid layout
    left_margin = 20
    right_margin = WIDTH // 2 + 10
    
    # Experiment number
    display.text(f"Experiment #{current_experiment}", left_margin, y, scale=1)
    
    # Developer info
    y += 15
    display.text(f"Developer:", left_margin, y, scale=1)
    display.text(f"{DEVELOPERS[current_developer]}", right_margin, y, scale=1)
    
    # Development time
    y += 15
    display.text(f"Dev Time:", left_margin, y, scale=1)
    display.text(f"{format_time(develop_time)}", right_margin, y, scale=1)
    
    # Fix time
    y += 15
    display.text(f"Fix Time:", left_margin, y, scale=1)
    display.text(f"{format_time(fix_time)}", right_margin, y, scale=1)
    
    # Draw button labels
    draw_button_labels(["A: Home", "", "C: Save", "D: Discard"])
    
    # Update display
    display.update()

def draw_logs_screen():
    """Draw the saved logs screen"""
    # Clear display with black background
    display.set_pen(BLACK)
    display.clear()
    
    # Title
    display.set_pen(LIGHT_BLUE)
    center_text("Saved Logs", 10, FONT_MEDIUM, 2)
    
    # Display recent results (limited to last 3)
    y = 35
    display.set_pen(WHITE)
    display.set_font(FONT_SMALL)
    
    if len(results) == 0:
        center_text("No saved results", y + 20, FONT_SMALL, 1)
    else:
        # Show last 3 results (or fewer if less are available)
        for i in range(min(3, len(results))):
            result = results[-(i+1)]  # Get results from newest to oldest
            
            # Format: #{id}: {developer}
            log_title = f"#{result['experiment']}: {result['developer']}"
            display.text(log_title, 20, y, scale=1)
            y += 12
            
            # Format: Dev: {time} • Fix: {time}
            log_detail = f"Dev: {format_time(result['develop_time'])} • Fix: {format_time(result['fix_time'])}"
            display.text(log_detail, 30, y, scale=1)
            y += 20  # Extra space between entries
    
    # Draw button labels
    draw_button_labels(["", "B: Back", "", "D: Clear"])
    
    # Update display
    display.update()

def draw_screen():
    """Draw the appropriate screen based on current state"""
    if current_screen == SCREEN_HOME:
        draw_home_screen()
    elif current_screen == SCREEN_DEVELOP:
        draw_develop_screen()
    elif current_screen == SCREEN_FIX:
        draw_fix_screen()
    elif current_screen == SCREEN_RESULTS:
        draw_results_screen()
    elif current_screen == SCREEN_LOGS:
        draw_logs_screen()

def format_time(seconds):
    """Format seconds into MM:SS format"""
    minutes = seconds // 60
    seconds = seconds % 60
    return f"{minutes:02d}:{seconds:02d}"

def start_timer():
    """Start the development timer"""
    global timer_running, timer_started, current_screen
    
    timer_running = True
    timer_started = time.ticks_ms()
    current_screen = SCREEN_DEVELOP
    draw_screen()

def to_fix_stage():
    """Move from development to fixing stage"""
    global timer_running, timer_started, current_screen
    
    # Stop development timer
    timer_running = False
    
    # Start fix timer
    timer_started = time.ticks_ms()
    timer_running = True
    current_screen = SCREEN_FIX
    draw_screen()

def complete_process():
    """Complete the fixing process and show results"""
    global timer_running, current_screen
    
    timer_running = False
    current_screen = SCREEN_RESULTS
    draw_screen()

def save_result():
    """Save the current result to memory and file"""
    global results, current_experiment
    
    # Save current result
    result = {
        "experiment": current_experiment,
        "developer": DEVELOPERS[current_developer],
        "fixer": FIXERS[current_fixer],
        "develop_time": develop_time,
        "fix_time": fix_time,
        "timestamp": time.time()
    }
    
    results.append(result)
    
    # Try to save to file
    try:
        with open("film_results.txt", "a") as f:
            f.write(f"{result['experiment']},{result['developer']},{result['fixer']},{result['develop_time']},{result['fix_time']},{result['timestamp']}\n")
    except:
        # If file write fails, we'll still have results in memory
        pass
    
    # Increment experiment number for next test
    current_experiment += 1
    reset_timer()

def reset_timer():
    """Reset the timer and go to home screen"""
    global timer_running, current_screen, develop_time, fix_time
    
    timer_running = False
    current_screen = SCREEN_HOME
    develop_time = 0
    fix_time = 0
    draw_screen()

def clear_logs():
    """Clear all saved results"""
    global results, current_screen
    
    results = []
    
    # Try to delete or clear the file
    try:
        with open("film_results.txt", "w") as f:
            pass  # Just open and close to clear
    except:
        pass  # Ignore errors
    
    draw_screen()

def handle_button_a():
    """Handle Button A press based on current screen"""
    global current_developer
    
    if current_screen == SCREEN_HOME:
        # Cycle through developers
        current_developer = (current_developer + 1) % len(DEVELOPERS)
        draw_screen()
    elif current_screen == SCREEN_RESULTS:
        # Return to home without saving
        reset_timer()

def handle_button_b():
    """Handle Button B press based on current screen"""
    global current_fixer, current_screen
    
    if current_screen == SCREEN_HOME:
        # Cycle through fixers
        current_fixer = (current_fixer + 1) % len(FIXERS)
        draw_screen()
    elif current_screen == SCREEN_LOGS:
        # Return to home screen
        current_screen = SCREEN_HOME
        draw_screen()

def handle_button_c():
    """Handle Button C press based on current screen"""
    if current_screen == SCREEN_HOME:
        # Start timer
        start_timer()
    elif current_screen == SCREEN_DEVELOP:
        # Go to fix stage
        to_fix_stage()
    elif current_screen == SCREEN_FIX:
        # Complete process
        complete_process()
    elif current_screen == SCREEN_RESULTS:
        # Save result and return to home
        save_result()

def handle_button_d():
    """Handle Button D press based on current screen"""
    global current_experiment
    
    if current_screen == SCREEN_HOME:
        # Increment experiment number
        current_experiment += 1
        draw_screen()
    elif current_screen == SCREEN_DEVELOP or current_screen == SCREEN_FIX:
        # Cancel current test
        reset_timer()
    elif current_screen == SCREEN_RESULTS:
        # Discard result and return to home
        reset_timer()
    elif current_screen == SCREEN_LOGS:
        # Clear all logs
        clear_logs()

def check_buttons():
    """Check for button presses and handle accordingly"""
    if button_a.read():
        if debounce():
            handle_button_a()
    
    if button_b.read():
        if debounce():
            handle_button_b()
    
    if button_c.read():
        if debounce():
            handle_button_c()
    
    if button_d.read():
        if debounce():
            handle_button_d()

def load_saved_results():
    """Try to load saved results from file on startup"""
    global results
    
    try:
        with open("film_results.txt", "r") as f:
            for line in f:
                parts = line.strip().split(",")
                if len(parts) >= 6:
                    result = {
                        "experiment": int(parts[0]),
                        "developer": parts[1],
                        "fixer": parts[2],
                        "develop_time": int(parts[3]),
                        "fix_time": int(parts[4]),
                        "timestamp": float(parts[5])
                    }
                    results.append(result)
    except:
        # If file doesn't exist or read fails, just continue
        pass

def main():
    """Main program loop"""
    # Initial setup
    load_saved_results()
    setup()
    
    while True:
        check_buttons()
        time.sleep(0.01)  # Small delay to prevent high CPU usage

if __name__ == "__main__":
    main() 