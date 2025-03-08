"""
Duck Hunt Game for Raspberry Pi Pico W with Pico Display
A simple duck hunting game where targets fly across the screen and you must shoot them with a crosshair.
"""
from picographics import PicoGraphics, DISPLAY_PICO_DISPLAY
from pimoroni import Button
import time
import random

# Initialize display
display = PicoGraphics(display=DISPLAY_PICO_DISPLAY)
display.set_backlight(1.0)  # Full brightness

# Initialize buttons
button_a = Button(12)  # Move left
button_b = Button(13)  # Move right
button_x = Button(14)  # Move up
button_y = Button(15)  # Shoot

# Colors
BLACK = display.create_pen(0, 0, 0)
WHITE = display.create_pen(255, 255, 255)
RED = display.create_pen(255, 0, 0)
GREEN = display.create_pen(0, 255, 0)
BLUE = display.create_pen(0, 0, 255)
YELLOW = display.create_pen(255, 255, 0)
BROWN = display.create_pen(139, 69, 19)

# Game constants
DISPLAY_WIDTH = 240
DISPLAY_HEIGHT = 135
MAX_DUCKS = 3
DUCK_SPEED_MIN = 1
DUCK_SPEED_MAX = 3
CROSSHAIR_SPEED = 3
DUCK_SIZE = 10

# Game state
score = 0
game_over = False
ducks = []
crosshair_x = DISPLAY_WIDTH // 2
crosshair_y = DISPLAY_HEIGHT // 2

class Duck:
    def __init__(self):
        # Start from either left or right side
        self.direction = int(random.choice([-1, 1])) # type: ignore
        self.x = DISPLAY_WIDTH if self.direction < 0 else 0
        self.y = random.randint(20, DISPLAY_HEIGHT - 20)
        self.speed = random.uniform(DUCK_SPEED_MIN, DUCK_SPEED_MAX)
        self.size = DUCK_SIZE
        self.alive = True
    
    def update(self):
        # Move duck
        self.x += self.speed * self.direction
        
        # Check if duck has left the screen
        if (self.direction > 0 and self.x > DISPLAY_WIDTH) or \
           (self.direction < 0 and self.x < 0):
            return False  # Duck has escaped
        return True  # Duck is still on screen
    
    def draw(self):
        # Simple duck shape (triangle for body, circle for head)
        if self.alive:
            display.set_pen(YELLOW)
            
            # Draw body (triangle)
            if self.direction > 0:  # Flying right
                display.triangle(
                    int(self.x - self.size), int(self.y),
                    int(self.x), int(self.y - self.size // 2),
                    int(self.x), int(self.y + self.size // 2)
                )
            else:  # Flying left
                display.triangle(
                    int(self.x + self.size), int(self.y),
                    int(self.x), int(self.y - self.size // 2),
                    int(self.x), int(self.y + self.size // 2)
                )
            
            # Draw head (circle)
            head_x = self.x + (self.size // 2 * self.direction)
            display.circle(int(head_x), int(self.y), self.size // 2)
    
    def check_hit(self, x, y):
        # Check if crosshair coordinates are within duck's hitbox
        distance = ((self.x - x) ** 2 + (self.y - y) ** 2) ** 0.5
        return distance < self.size * 1.5  # Slightly larger hitbox

def draw_crosshair(x, y):
    display.set_pen(RED)
    size = 8
    # Draw crosshair
    display.line(x - size, y, x + size, y)  # Horizontal line
    display.line(x, y - size, x, y + size)  # Vertical line
    # Draw circle around center
    display.circle(x, y, 3)

def draw_score():
    display.set_pen(WHITE)
    display.set_font("bitmap8")
    display.text(f"Score: {score}", 5, 5, scale=1)

def draw_game_over():
    display.set_pen(BLACK)
    display.clear()
    display.set_pen(WHITE)
    display.set_font("bitmap8")
    display.text("GAME OVER", DISPLAY_WIDTH // 2 - 40, DISPLAY_HEIGHT // 2 - 20, scale=2)
    display.text(f"Final Score: {score}", DISPLAY_WIDTH // 2 - 50, DISPLAY_HEIGHT // 2 + 10, scale=1)
    display.update()

def spawn_duck():
    if len(ducks) < MAX_DUCKS and random.random() < 0.03:  # 3% chance each frame
        ducks.append(Duck())

def check_buttons():
    global crosshair_x, crosshair_y, score, game_over
    
    # Move crosshair
    if button_a.read():
        crosshair_x = max(0, crosshair_x - CROSSHAIR_SPEED)
    if button_b.read():
        crosshair_x = min(DISPLAY_WIDTH, crosshair_x + CROSSHAIR_SPEED)
    if button_x.read():
        crosshair_y = max(0, crosshair_y - CROSSHAIR_SPEED)
    if button_y.read():
        # Shoot
        for duck in ducks:
            if duck.alive and duck.check_hit(crosshair_x, crosshair_y):
                duck.alive = False
                score += 1
                time.sleep(0.1)  # Small delay for feedback

def main_game_loop():
    global game_over, ducks, score
    
    # Reset game state
    game_over = False
    ducks = []
    score = 0
    
    # Game loop
    while not game_over:
        # Clear screen
        display.set_pen(BLUE)  # Sky background
        display.clear()
        
        # Draw ground
        display.set_pen(GREEN)
        display.rectangle(0, DISPLAY_HEIGHT - 20, DISPLAY_WIDTH, 20)
        
        # Check button inputs
        check_buttons()
        
        # Spawn ducks
        spawn_duck()
        
        # Update and draw ducks
        for i in range(len(ducks) - 1, -1, -1):
            duck = ducks[i]
            if not duck.update() or not duck.alive:
                ducks.pop(i)
            else:
                duck.draw()
        
        # Draw crosshair
        draw_crosshair(crosshair_x, crosshair_y)
        
        # Draw score
        draw_score()
        
        # Update display
        display.update()
        
        # Add a small delay to control frame rate
        time.sleep(0.01)

# Start the game
while True:
    main_game_loop()
    draw_game_over()
    time.sleep(3)  # Wait before starting a new game 