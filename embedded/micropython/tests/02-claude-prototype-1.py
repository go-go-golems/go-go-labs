import time
from picographics import PicoGraphics, DISPLAY_PICO_DISPLAY
from pimoroni import Button
import machine
from machine import Pin, Timer

# Initialize display
display = PicoGraphics(display=DISPLAY_PICO_DISPLAY)
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
current_stage = 0  # 0=Develop, 1=Fix, 2=Complete
timer_started = 0
results = []  # Store results for later retrieval

# Timer interrupt for counting seconds
timer = Timer(-1)  # Use any available timer ID

def update_timer(timer):
    global develop_time, fix_time, timer_running, timer_started, current_stage
    
    if timer_running:
        current_time = time.ticks_ms()
        elapsed = (current_time - timer_started) // 1000  # Convert to seconds
        
        if current_stage == 0:
            develop_time = elapsed
        elif current_stage == 1:
            fix_time = elapsed
            
        draw_screen()

def setup():
    # Setup timer interrupt to trigger every 100ms
    timer.init(freq=10, mode=Timer.PERIODIC, callback=update_timer)
    draw_screen()

def draw_screen():
    display.set_pen(BLACK)
    display.clear()
    
    # Set up vertical layout with smaller text
    margin = 5
    line_height = 18  # Smaller line height for vertical layout
    y_pos = margin
    
    if timer_running:
        # Header - Current stage
        display.set_pen(WHITE)
        current_stage_text = "DEVELOP" if current_stage == 0 else "FIX"
        display.text(current_stage_text, (WIDTH // 2) - 30, y_pos, scale=1)
        y_pos += line_height
        
        # Chemical being used
        display.set_pen(YELLOW)
        chemical = DEVELOPERS[current_developer] if current_stage == 0 else FIXERS[current_fixer]
        display.text(chemical, (WIDTH // 2) - (len(chemical) * 4), y_pos, scale=1)
        y_pos += line_height + 5
        
        # Timer display (larger)
        display.set_pen(WHITE)
        time_text = format_time(develop_time if current_stage == 0 else fix_time)
        display.text(time_text, (WIDTH // 2) - 40, y_pos, scale=3)  # Keep timer large
        y_pos += 30  # Larger offset for the larger text
        
        # Secondary timer if needed
        if current_stage == 1:
            display.set_pen(GREEN)
            display.text("Dev: " + format_time(develop_time), margin, y_pos, scale=1)
            y_pos += line_height
        
        # Instructions at bottom
        y_pos = HEIGHT - (line_height * 2)
        display.set_pen(BLUE)
        if current_stage == 0:
            display.text("C: Next Stage", margin, y_pos, scale=1)
        else:
            display.text("C: Complete", margin, y_pos, scale=1)
        y_pos += line_height
        display.text("D: Cancel", margin, y_pos, scale=1)
    elif current_stage == 2:  # Complete screen
        display.set_pen(WHITE)
        display.text("COMPLETE", (WIDTH // 2) - 35, y_pos, scale=1)
        y_pos += line_height
        
        display.set_pen(GREEN)
        display.text("Dev: " + DEVELOPERS[current_developer], margin, y_pos, scale=1)
        y_pos += line_height
        display.text("Time: " + format_time(develop_time), margin, y_pos, scale=1)
        y_pos += line_height + 5
        
        display.set_pen(BLUE)
        display.text("Fix: " + FIXERS[current_fixer], margin, y_pos, scale=1)
        y_pos += line_height
        display.text("Time: " + format_time(fix_time), margin, y_pos, scale=1)
        y_pos += line_height + 10
        
        # Instructions at bottom
        y_pos = HEIGHT - (line_height * 2)
        display.set_pen(YELLOW)
        display.text("C: Save", margin, y_pos, scale=1)
        y_pos += line_height
        display.text("D: Discard", margin, y_pos, scale=1)
    else:  # Idle screen
        # Show experiment number at top
        display.set_pen(WHITE)
        exp_text = "Experiment #" + str(current_experiment)
        display.text(exp_text, (WIDTH // 2) - (len(exp_text) * 4), y_pos, scale=1)
        y_pos += line_height + 5
        
        # Developer selection
        display.set_pen(GREEN)
        display.text("Developer:", margin, y_pos, scale=1)
        y_pos += line_height
        dev_text = DEVELOPERS[current_developer]
        display.text(dev_text, margin + 10, y_pos, scale=1)
        y_pos += line_height + 5
        
        # Fixer selection
        display.set_pen(BLUE)
        display.text("Fixer:", margin, y_pos, scale=1)
        y_pos += line_height
        fix_text = FIXERS[current_fixer]
        display.text(fix_text, margin + 10, y_pos, scale=1)
        y_pos += line_height + 5
        
        # Instructions at bottom
        y_pos = HEIGHT - (line_height * 4)
        display.set_pen(YELLOW)
        display.text("A: Change Developer", margin, y_pos, scale=1)
        y_pos += line_height
        display.text("B: Change Fixer", margin, y_pos, scale=1)
        y_pos += line_height
        display.text("C: Start Timer", margin, y_pos, scale=1)
        y_pos += line_height
        display.text("D: Next Experiment", margin, y_pos, scale=1)
    
    display.update()

def format_time(seconds):
    minutes = seconds // 60
    seconds = seconds % 60
    return f"{minutes:02d}:{seconds:02d}"

def start_timer():
    global timer_running, timer_started, current_stage
    
    timer_running = True
    timer_started = time.ticks_ms()
    current_stage = 0
    draw_screen()

def stop_timer():
    global timer_running, current_stage
    
    timer_running = False
    if current_stage == 0:
        current_stage = 1
        timer_started = time.ticks_ms()
        timer_running = True
    elif current_stage == 1:
        current_stage = 2
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
    current_stage = 0
    develop_time = 0
    fix_time = 0
    draw_screen()

def check_buttons():
    if button_a.read():
        if current_stage == 0:
            # Cycle through developers
            global current_developer
            current_developer = (current_developer + 1) % len(DEVELOPERS)
        time.sleep(0.2)  # Debounce
        draw_screen()
    
    if button_b.read():
        if current_stage == 0:
            # Cycle through fixers
            global current_fixer
            current_fixer = (current_fixer + 1) % len(FIXERS)
        time.sleep(0.2)  # Debounce
        draw_screen()
    
    if button_c.read():
        if current_stage == 0:
            start_timer()
        elif current_stage == 1:
            stop_timer()  # Move to FIX stage
        elif current_stage == 2:
            save_result()  # Save and reset
        time.sleep(0.2)  # Debounce
    
    if button_d.read():
        if current_stage == 0:
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