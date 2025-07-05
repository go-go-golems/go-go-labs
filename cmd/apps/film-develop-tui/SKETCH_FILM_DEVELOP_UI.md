# Film Development TUI - Complete Specification

## Overview
A minimal terminal user interface for calculating B&W film development parameters using ILFOSOL 3, ILFOSTOP, and Sprint Fixer.

## Core Features
- Film type and EI rating selection
- Tank size calculation based on roll count and format
- Chemical dilution calculations
- Fixer capacity tracking
- Development time display

## Data Model

### Film Database
```yaml
films:
  hp5_plus:
    name: "HP5 PLUS"
    ei_ratings: [200, 400, 800]
    times_20c:
      1_plus_9: { 200: "5:00", 400: "6:30", 800: "13:30" }
      1_plus_14: { 200: "7:00", 400: "11:00", 800: "19:30" }
  
  fp4_plus:
    name: "FP4 PLUS"
    ei_ratings: [125]
    times_20c:
      1_plus_9: { 125: "4:15" }
      1_plus_14: { 125: "7:30" }
  
  delta_100:
    name: "DELTA 100"
    ei_ratings: [100]
    times_20c:
      1_plus_9: { 100: "5:00" }
      1_plus_14: { 100: "7:30" }

  delta_400:
    name: "DELTA 400"
    ei_ratings: [200, 400, 800]
    times_20c:
      1_plus_9: { 200: "5:30", 400: "7:00", 800: "14:00" }
      1_plus_14: { 200: "8:00", 400: "12:00", 800: "20:30" }

  delta_3200:
    name: "DELTA 3200"
    ei_ratings: [400, 800, 1600, 3200, 6400]
    times_20c:
      1_plus_9: { 400: "6:00", 800: "7:30", 1600: "10:00", 3200: "11:00", 6400: "18:00" }
      1_plus_14: { 400: "11:00", 800: "13:00", 1600: "15:30", 3200: "17:00", 6400: "23:00" }

  pan_f_plus:
    name: "PAN F PLUS"
    ei_ratings: [50]
    times_20c:
      1_plus_14: { 50: "4:30" }

  sfx_200:
    name: "SFX 200"
    ei_ratings: [200, 400]
    times_20c:
      1_plus_9: { 200: "6:00", 400: "8:30" }
      1_plus_14: { 200: "9:00", 400: "13:30" }
```

### Tank Size Calculation
```yaml
tank_sizes:
  35mm:
    1_roll: 300ml
    2_rolls: 500ml
    3_rolls: 600ml
    4_rolls: 700ml
    5_rolls: 800ml
    6_rolls: 900ml
  
  120mm:
    1_roll: 500ml
    2_rolls: 700ml
    3_rolls: 900ml
    4_rolls: 1000ml
    5_rolls: 1200ml
    6_rolls: 1400ml
```

### Chemical Dilutions
```yaml
chemicals:
  ilfosol_3:
    dilutions: ["1+9", "1+14"]
    default: "1+9"
    type: "one_shot"
  
  ilfostop:
    dilution: "1+19"
    time: "0:10"
    type: "reusable"
    capacity: "15 rolls per liter"
  
  sprint_fixer:
    dilution: "1+4"
    time: "2:30"
    type: "reusable"
    capacity: "24 rolls per liter"
```

## State Machine

### Application States
```
MAIN_SCREEN
├── FILM_SELECTION
│   ├── EI_SELECTION
│   │   ├── ROLL_SELECTION
│   │   │   ├── MIXED_ROLL_INPUT
│   │   │   └── CALCULATED_SCREEN
│   │   └── [back to FILM_SELECTION]
│   └── [back to MAIN_SCREEN]
├── FIXER_TRACKING
└── SETTINGS
```

### State Transitions
```yaml
states:
  main_screen:
    actions:
      'f': film_selection
      'u': fixer_tracking
      's': settings
      'q': quit
  
  film_selection:
    actions:
      '1-7': set_film_type -> ei_selection
      'esc': main_screen
      'q': quit
  
  ei_selection:
    actions:
      '1-n': set_ei_rating -> roll_selection
      'esc': film_selection
      'q': quit
  
  roll_selection:
    actions:
      '1-6': set_35mm_rolls -> calculated_screen
      'a-f': set_120mm_rolls -> calculated_screen
      'm': mixed_roll_input
      'esc': ei_selection
      'q': quit
  
  calculated_screen:
    actions:
      'u': use_fixer
      'r': roll_selection
      'f': film_selection
      'q': quit
```

## Screen Layouts

### 1. Main Screen (MAIN_SCREEN)
```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                            🎞️  Film Development Calculator                        │
└─────────────────────────────────────────────────────────────────────────────────┘

┌─── Film Setup ──────────────────────────────────────────────────────────────────┐
│                                                                                 │
│  Film Type:    [ Not Selected ]                    EI:  [ -- ]                 │
│  Rolls:        [ -- ]                              Tank: [ --ml ]              │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘

┌─── Chemicals (20°C) ────────────────────────────────────────────────────────────┐
│                                                                                 │
│  ILFOSOL 3     │  ILFOSTOP      │  SPRINT FIXER                                │
│  1+9 dilution  │  1+19 dilution │  1+4 dilution                                │
│  --ml conc     │  --ml conc     │  --ml conc                                   │
│  --ml water    │  --ml water    │  --ml water                                  │
│  Time: --:--   │  Time: 0:10    │  Time: 2:30                                  │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘

┌─── Fixer Usage ─────────────────────────────────────────────────────────────────┐
│                                                                                 │
│  Capacity: 24 rolls per liter    Used: 0 rolls    Remaining: -- rolls          │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘

┌─── Actions ─────────────────────────────────────────────────────────────────────┐
│                                                                                 │
│  [F] Film Type    [U] Fixer Usage    [S] Settings    [Q] Quit                  │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘
```

### 2. Film Selection Screen (FILM_SELECTION)
```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                            🎞️  Film Development Calculator                        │
└─────────────────────────────────────────────────────────────────────────────────┘

┌─── Select Film Type ────────────────────────────────────────────────────────────┐
│                                                                                 │
│  [1] HP5 PLUS        (EI 200/400/800)    📈 Most popular                       │
│  [2] FP4 PLUS        (EI 125)            🎯 Fine grain                         │
│  [3] DELTA 100       (EI 100)            🔍 Ultra fine                         │
│  [4] DELTA 400       (EI 200/400/800)    ⚖️  Versatile                         │
│  [5] DELTA 3200      (EI 400-6400)       🌙 High speed                         │
│  [6] PAN F PLUS      (EI 50)             💎 Finest grain                       │
│  [7] SFX 200         (EI 200/400)        🔴 Infrared                           │
│                                                                                 │
│  [ESC] Back                                                                     │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘

┌─── Actions ─────────────────────────────────────────────────────────────────────┐
│                                                                                 │
│  [1-7] Select Film    [ESC] Back    [Q] Quit                                   │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘
```

### 3. EI Selection Screen (EI_SELECTION)
```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                            🎞️  Film Development Calculator                        │
└─────────────────────────────────────────────────────────────────────────────────┘

┌─── Film Setup ──────────────────────────────────────────────────────────────────┐
│                                                                                 │
│  Film Type:    [ HP5 PLUS ]                        EI:  [ Not Set ]           │
│  Rolls:        [ -- ]                              Tank: [ --ml ]              │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘

┌─── Select EI Rating ────────────────────────────────────────────────────────────┐
│                                                                                 │
│  [1] EI 200  (5:00 @ 1+9)     🌞 Bright light, fine grain                     │
│  [2] EI 400  (6:30 @ 1+9)     📷 Standard, most common                        │
│  [3] EI 800  (13:30 @ 1+9)    🌆 Low light, pushed grain                      │
│                                                                                 │
│  [ESC] Back to film selection                                                   │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘

┌─── Actions ─────────────────────────────────────────────────────────────────────┐
│                                                                                 │
│  [1-3] Select EI    [ESC] Back    [Q] Quit                                     │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘
```

### 4. Roll Selection Screen (ROLL_SELECTION)
```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                            🎞️  Film Development Calculator                        │
└─────────────────────────────────────────────────────────────────────────────────┘

┌─── Film Setup ──────────────────────────────────────────────────────────────────┐
│                                                                                 │
│  Film Type:    [ HP5 PLUS ]                        EI:  [ 400 ]                │
│  Rolls:        [ -- ]                              Tank: [ --ml ]              │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘

┌─── Number of Rolls ─────────────────────────────────────────────────────────────┐
│                                                                                 │
│  35mm Rolls:                           120mm Rolls:                            │
│  [1] 1 Roll (300ml)  [4] 4 Rolls       [A] 1 Roll (500ml)  [D] 4 Rolls        │
│  [2] 2 Rolls (500ml) [5] 5 Rolls       [B] 2 Rolls (700ml) [E] 5 Rolls        │
│  [3] 3 Rolls (600ml) [6] 6 Rolls       [C] 3 Rolls (900ml) [F] 6 Rolls        │
│                                                                                 │
│  Mixed batches: [M] Custom mix                                                  │
│                                                                                 │
│  [ESC] Back to EI selection                                                     │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘

┌─── Actions ─────────────────────────────────────────────────────────────────────┐
│                                                                                 │
│  [1-6] 35mm    [A-F] 120mm    [M] Mixed    [ESC] Back    [Q] Quit              │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘
```

### 5. Mixed Roll Input Screen (MIXED_ROLL_INPUT)
```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                            🎞️  Film Development Calculator                        │
└─────────────────────────────────────────────────────────────────────────────────┘

┌─── Custom Mix Setup ────────────────────────────────────────────────────────────┐
│                                                                                 │
│  35mm Rolls: [ 0 ]    (↑/↓ or +/- to adjust)                                  │
│  120mm Rolls: [ 0 ]   (↑/↓ or +/- to adjust)                                  │
│                                                                                 │
│  Total Tank Size: [ 0ml ]                                                      │
│                                                                                 │
│  [ENTER] Confirm    [ESC] Back    [R] Reset                                    │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘

┌─── Actions ─────────────────────────────────────────────────────────────────────┐
│                                                                                 │
│  [↑↓] Adjust 35mm    [+/-] Adjust 120mm    [ENTER] Confirm    [ESC] Back       │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘
```

### 6. Calculated Results Screen (CALCULATED_SCREEN)
```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                            🎞️  Film Development Calculator                        │
└─────────────────────────────────────────────────────────────────────────────────┘

┌─── Film Setup ──────────────────────────────────────────────────────────────────┐
│                                                                                 │
│  Film Type:    [ HP5 PLUS ]                        EI:  [ 400 ]                │
│  Rolls:        [ 1x 120mm ]                        Tank: [ 500ml ]             │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘

┌─── Chemicals (20°C) ────────────────────────────────────────────────────────────┐
│                                                                                 │
│  ILFOSOL 3     │  ILFOSTOP      │  SPRINT FIXER                                │
│  1+9 dilution  │  1+19 dilution │  1+4 dilution                                │
│  50ml conc     │  25ml conc     │  100ml conc                                  │
│  450ml water   │  475ml water   │  400ml water                                 │
│  Time: 6:30    │  Time: 0:10    │  Time: 2:30                                  │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘

┌─── Fixer Usage ─────────────────────────────────────────────────────────────────┐
│                                                                                 │
│  Capacity: 24 rolls per liter    Used: 0 rolls    Remaining: 24 rolls          │
│  This batch uses: 1 roll         After use: 23 rolls remaining                 │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘

┌─── Actions ─────────────────────────────────────────────────────────────────────┐
│                                                                                 │
│  [U] Use Fixer    [R] Change Rolls    [F] Change Film    [Q] Quit              │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘
```

