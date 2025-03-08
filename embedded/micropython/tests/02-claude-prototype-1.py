import time
from picographics import PicoGraphics, DISPLAY_PIMIRONI_PICO_DISPLAY
from pimoroni import Button
import machine
from machine import Pin, Timer

# Initialize display
display = PicoGraphics(display=DISPLAY_PIMIRONI_PICO_DISPLAY)
WIDTH, HEIGHT = display.get_bounds()

# Setup buttons - Assuming standard Pimoroni button configuration
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

# Chemical presets (you can customize these)
DEVELOPERS = ["D-76", "Rodinal", "HC-110", "XTOL", "Pyro"]
FIXERS = ["Rapid Fix", "TF-4", "Ilford", "Kodak"]

# Global variables
current_experiment = 1
current_developer = 0
current_fixer = 0
develop_time = 0
fix_time = 0
timer_running = False
timer_started = 0
current_stage = "IDLE"  # IDLE, DEVELOP, FIX
results = []  # Store results for later retrieval

# Timer interrupt for counting seconds
timer = Timer()

def update_timer(timer):
    global develop_time, fix_time, timer_running, timer_started, current_stage
    
    if timer_running:
        current_time = time.ticks_ms()
        elapsed = (current_time - timer_started) // 1000  # Convert to seconds
        
        if current_stage == "DEVELOP":
            develop_time = elapsed
        elif current_stage == "FIX":
            fix_time = elapsed
            
        draw_screen()

def setup():
    # Setup timer interrupt to trigger every 100ms
    timer.init(freq=10, mode=Timer.PERIODIC, callback=update_timer)
    draw_screen()

def draw_screen():
    display.set_pen(BLACK)
    display.clear()
    display.set_pen(WHITE)
    
    # Draw header
    display.set_pen(BLUE)
    display.text(f"Experiment #{current_experiment}", 10, 10, 240, 2)
    display.set_pen(WHITE)
    
    # Draw chemical selection or current timers
    if current_stage == "IDLE":
        display.text("Developer:", 10, 40, 240, 2)
        display.text(DEVELOPERS[current_developer], 120, 40, 240, 2)
        
        display.text("Fixer:", 10, 70, 240, 2)
        display.text(FIXERS[current_fixer], 120, 70, 240, 2)
        
        # Button labels
        display.set_pen(YELLOW)
        display.text("A: Dev+", 10, HEIGHT - 60, 240, 1)
        display.text("B: Fix+", 120, HEIGHT - 60, 240, 1)
        display.text("C: Start", 10, HEIGHT - 40, 240, 1)
        display.text("D: Exp+", 120, HEIGHT - 40, 240, 1)
    else:
        # Show stage and timer
        if current_stage == "DEVELOP":
            display.set_pen(GREEN)
            display.text("DEVELOPING", 10, 40, 240, 2)
            display.text(f"Developer: {DEVELOPERS[current_developer]}", 10, 70, 240, 1)
            display.set_pen(WHITE)
            display.text(f"Time: {format_time(develop_time)}", 10, 100, 240, 3)
        elif current_stage == "FIX":
            display.set_pen(BLUE)
            display.text("FIXING", 10, 40, 240, 2)
            display.text(f"Fixer: {FIXERS[current_fixer]}", 10, 70, 240, 1)
            display.set_pen(WHITE)
            display.text(f"Time: {format_time(fix_time)}", 10, 100, 240, 3)
        elif current_stage == "COMPLETE":
            display.set_pen(GREEN)
            display.text("COMPLETE", 10, 40, 240, 2)
            display.text(f"Developer: {format_time(develop_time)}", 10, 70, 240, 1)
            display.text(f"Fixer: {format_time(fix_time)}", 10, 100, 240, 1)
        
        # Button labels for active timer
        display.set_pen(YELLOW)
        if current_stage == "DEVELOP":
            display.text("A: -", 10, HEIGHT - 60, 240, 1)
            display.text("B: -", 120, HEIGHT - 60, 240, 1)
            display.text("C: To Fix", 10, HEIGHT - 40, 240, 1)
            display.text("D: Cancel", 120, HEIGHT - 40, 240, 1)
        elif current_stage == "FIX":
            display.text("A: -", 10, HEIGHT - 60, 240, 1)
            display.text("B: -", 120, HEIGHT - 60, 240, 1)
            display.text("C: Done", 10, HEIGHT - 40, 240, 1)
            display.text("D: Cancel", 120, HEIGHT - 40, 240, 1)
        elif current_stage == "COMPLETE":
            display.text("A: -", 10, HEIGHT - 60, 240, 1)
            display.text("B: -", 120, HEIGHT - 60, 240, 1)
            display.text("C: Save", 10, HEIGHT - 40, 240, 1)
            display.text("D: Discard", 120, HEIGHT - 40, 240, 1)
    
    display.update()

def format_time(seconds):
    minutes = seconds // 60
    seconds = seconds % 60
    return f"{minutes:02d}:{seconds:02d}"

def start_timer():
    global timer_running, timer_started, current_stage
    
    timer_running = True
    timer_started = time.ticks_ms()
    current_stage = "DEVELOP"
    draw_screen()

def stop_timer():
    global timer_running, current_stage
    
    timer_running = False
    if current_stage == "DEVELOP":
        current_stage = "FIX"
        timer_started = time.ticks_ms()
        timer_running = True
    elif current_stage == "FIX":
        current_stage = "COMPLETE"
        timer_running = False
    draw_screen()

def save_result():
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
    global timer_running, current_stage, develop_time, fix_time
    
    timer_running = False
    current_stage = "IDLE"
    develop_time = 0
    fix_time = 0
    draw_screen()

def check_buttons():
    if button_a.read():
        if current_stage == "IDLE":
            # Cycle through developers
            global current_developer
            current_developer = (current_developer + 1) % len(DEVELOPERS)
        time.sleep(0.2)  # Debounce
        draw_screen()
    
    if button_b.read():
        if current_stage == "IDLE":
            # Cycle through fixers
            global current_fixer
            current_fixer = (current_fixer + 1) % len(FIXERS)
        time.sleep(0.2)  # Debounce
        draw_screen()
    
    if button_c.read():
        if current_stage == "IDLE":
            start_timer()
        elif current_stage == "DEVELOP":
            stop_timer()  # Move to FIX stage
        elif current_stage == "FIX":
            stop_timer()  # Move to COMPLETE stage
        elif current_stage == "COMPLETE":
            save_result()  # Save and reset
        time.sleep(0.2)  # Debounce
    
    if button_d.read():
        if current_stage == "IDLE":
            # Increment experiment number
            global current_experiment
            current_experiment += 1
        else:
            # Cancel current test
            reset_timer()
        time.sleep(0.2)  # Debounce
        draw_screen()

def main():
    setup()
    
    while True:
        check_buttons()
        time.sleep(0.01)  # Small delay to prevent high CPU usage

if __name__ == "__main__":
    main()