# Film Development Timer - Complete User Manual

## Table of Contents
1. [Overview](#overview)
2. [Getting Started](#getting-started)
3. [The Chemistry System](#the-chemistry-system)
4. [Interface Guide](#interface-guide)
5. [Step-by-Step Usage](#step-by-step-usage)
6. [Advanced Features](#advanced-features)
7. [Troubleshooting](#troubleshooting)
8. [Tips & Best Practices](#tips--best-practices)
9. [Technical Reference](#technical-reference)

---

## Overview

The Film Development Timer is a specialized mobile-friendly web application designed for processing 17 rolls of film using a single liter of Cinestill Cs41 color negative developer. The app implements a precise "bleed-and-feed" replenishment system that maximizes chemistry usage while maintaining consistent development quality.

### What This App Does
- **Tracks 17 rolls** with progressively adjusted development times
- **Times each process step** with precision timing and audio alerts
- **Guides replenishment** with built-in chemistry management reminders
- **Prevents mistakes** by showing exactly what to do at each stage
- **Works offline** once loaded, perfect for darkroom use

### Key Benefits
- **Cost effective**: Get maximum rolls from expensive chemistry
- **Consistent results**: Scientifically calculated time adjustments
- **Mobile optimized**: Large buttons and dark theme for darkroom use
- **Foolproof routine**: Eliminates guesswork and timing errors

---

## Getting Started

### System Requirements
- Any modern web browser (Chrome, Safari, Firefox, Edge)
- No installation required - runs directly in browser
- Works on phones, tablets, and computers
- Audio alerts require device sound enabled

### First Launch
1. **Open the application** in your web browser
2. **Test audio** by starting/stopping a timer to ensure notifications work
3. **Familiarize yourself** with the interface before mixing chemistry
4. **Bookmark the page** for easy darkroom access

### Initial Setup Checklist
Before using the timer, ensure you have:
- [ ] Cinestill Cs41 kit ready to mix
- [ ] 600mL development tank
- [ ] Accurate thermometer (102°F/39°C target)
- [ ] 1L amber storage bottle labeled "WORKING DEVELOPER"
- [ ] Small bottle for 400mL reserve labeled "RESERVE - FRESH"
- [ ] 25mL measuring cylinder or graduated beaker
- [ ] Waste container for discarded chemistry

---

## The Chemistry System

### The Science Behind the Timer

This application implements a **bleed-and-feed replenishment system** that balances chemistry exhaustion with fresh developer injection. Here's how it works:

#### Base Formula
- **1 liter total** Cinestill Cs41 developer mixed at once
- **600mL working volume** in your development tank
- **400mL reserve** stored separately for replenishment
- **25mL replacement** after each roll (≈4% of tank volume)

#### Why This Works
- **Cinestill's guidelines**: 4%/roll time increase for 500mL, 2%/roll for 1L
- **Your system**: 4% volume replacement splits the difference
- **Result**: ~3%/roll time increase with consistent quality
- **Capacity**: Exactly 17 rolls using all chemistry

### Chemistry Preparation

#### Mixing Day
1. **Mix entire 1L** of Cs41 developer to specifications
2. **Heat to 102°F/39°C** and maintain temperature
3. **Pour 600mL** into your development tank for Roll #1
4. **Store 400mL** in amber bottle as reserve
5. **Label everything** clearly with mixing date

#### Shelf Life Management
- **Use within 10-14 days** of mixing for best results
- **Store in cool, dark place** when not in use
- **Keep bottles full** to minimize air exposure
- **Track mixing date** - after 2 weeks, expect degraded performance

---

## Interface Guide

### Main Screen Layout

#### Header Section
- **App title** and temperature reminder (102°F/39°C)
- **Roll counter** showing current progress (e.g., "Roll 5/17")
- **Film format** and push/pull status for current roll

#### Roll Navigation
- **Previous/Next buttons** to move between rolls
- **Roll information** showing film type (120/35mm) and exposure adjustment
- **Special notes** highlighting important information for specific rolls

#### Process Steps Panel
Five clickable cards showing each development stage:
1. **Pre-soak** (1:00) - Plain water temperature stabilization
2. **Developer** (varies) - Main development with calculated time
3. **Bleach-fix** (6:30) - Fixed time, no chemistry fatigue
4. **Wash** (3:00) - Running water rinse
5. **Stabilizer** (1:00) - Final treatment before drying

#### Timer Display
- **Large digital readout** showing time remaining
- **Negative time support** for tracking overtime
- **Color coding**: White for normal time, red for overtime
- **Status indicators**: Current timer state and warnings

#### Control Buttons
- **Start/Resume**: Begin timing or continue after pause
- **Pause**: Temporarily stop timer while preserving time
- **Stop**: Reset timer to zero and stop counting
- **Complete**: Mark current step as finished and advance

#### Quick Actions
- **Reset Roll**: Start over with current roll from step 1
- **Replenishment reminder**: Shows after step 5 completion

### Visual Indicators

#### Step Status
- **⏸️ Waiting**: Step not yet started (gray border)
- **⏱️ Active**: Currently selected step (blue border)
- **✅ Complete**: Finished step (green border)

#### Timer States
- **Normal time**: White text counting down
- **Overtime**: Red text with negative values and warning
- **Paused**: Timer stopped but preserving remaining time

---

## Step-by-Step Usage

### Starting a New Roll

1. **Navigate to correct roll** using Previous/Next buttons
2. **Verify film information** matches your actual film
3. **Check chemistry temperature** (102°F/39°C)
4. **Click "Pre-soak"** step to select it
5. **Load film** into development tank
6. **Start timer** and begin pre-soak

### During Each Step

#### Pre-soak (1:00)
- **Purpose**: Stabilize film and tank temperature
- **Process**: Fill tank with plain water at 102°F
- **Timing**: Exactly 1 minute
- **Action**: Gentle agitation for first 10 seconds

#### Developer (varies by roll)
- **Purpose**: Convert exposed silver halides to metallic silver
- **Process**: Pour developer quickly, start timer immediately
- **Timing**: Shown on timer (e.g., 3:30 for Roll 2)
- **Agitation**: First 10 seconds of every 30-second interval
- **Critical**: Most important step for image quality

#### Bleach-fix (6:30)
- **Purpose**: Remove unexposed silver and fix the image
- **Process**: Pour blix immediately after developer
- **Timing**: Always 6:30 (no adjustment needed)
- **Agitation**: Continuous for first minute, then 10s every 30s

#### Wash (3:00)
- **Purpose**: Remove processing chemicals
- **Process**: Running water at temperature
- **Timing**: 3-4 minutes minimum
- **Flow**: Ensure complete water exchange

#### Stabilizer (1:00)
- **Purpose**: Final conditioning and protection
- **Process**: Final chemical bath
- **Timing**: 1 minute with gentle agitation
- **After**: Hang film to dry in dust-free environment

### Between Rolls: Replenishment Routine

After completing Step 5 (Stabilizer), the app shows replenishment instructions:

1. **Pour used 600mL** back into "WORKING DEVELOPER" bottle
2. **Measure 25mL** from working developer and discard
3. **Add 25mL fresh** developer from "RESERVE" bottle
4. **Cap both bottles** and invert once to mix
5. **Update your notes** if tracking roll details

### Advancing to Next Roll

1. **Complete replenishment** (except after Roll 17)
2. **Click "Next"** button to advance
3. **Verify new roll information** matches your next film
4. **Reset timer** automatically clears to start fresh
5. **Begin with Pre-soak** for the new roll

---

## Advanced Features

### Pause and Resume
**When to use**: Emergency interruptions, phone calls, or processing issues

**How it works**:
- **Pause**: Stops timer while preserving exact remaining time
- **Resume**: Continues from where you left off
- **Status**: Shows "PAUSED" state clearly

**Best practice**: Try to minimize pauses during Developer step

### Step Selection
**Purpose**: Jump to any step without following sequence

**Usage scenarios**:
- **Skip ahead**: Click "Wash" if you finish developer early
- **Go back**: Return to previous step if needed
- **Emergency**: Jump to critical step immediately

**How to use**:
1. **Click any step card** to select it immediately
2. **Timer resets** to that step's duration
3. **Previous timers stop** automatically

### Negative Time (Overtime)
**What it shows**: Exact amount of over-development

**Display**:
- **Red numbers**: Negative time like "-0:15"
- **Warning message**: "⚠️ OVERTIME" indicator
- **Continues counting**: No automatic stop

**Uses**:
- **Track actual times**: Know how long steps really took
- **Adjust technique**: Identify steps that consistently run over
- **Emergency reference**: Know exactly how much extra time occurred

### Audio Notifications
**When alerts play**: Timer crosses from positive to negative time

**Sound**: Brief notification tone (if device audio enabled)

**Troubleshooting**:
- **No sound**: Check device volume and browser permissions
- **Multiple alerts**: Only plays once per timer crossing zero
- **Disable**: Mute device if alerts not wanted

---

## Troubleshooting

### Timer Issues

#### Timer Won't Start
**Possible causes**:
- Step shows "varies" (Developer step with invalid time)
- Browser frozen or slow
- Audio permission blocking

**Solutions**:
1. **Refresh browser** and try again
2. **Check step selection** - click step card to select properly
3. **Verify time display** shows valid time (not "varies")

#### Timer Stops Unexpectedly
**Possible causes**:
- Browser background mode
- Device sleep mode
- Network connectivity (shouldn't affect timer)

**Solutions**:
1. **Keep browser active** - don't switch apps during timing
2. **Adjust device settings** to prevent sleep
3. **Use pause/resume** if interruption needed

#### Wrong Time Showing
**Check these items**:
- **Correct roll selected**: Verify roll number matches your film
- **Right step active**: Ensure proper step is highlighted
- **App not cached**: Refresh if updates aren't showing

### Chemistry Problems

#### Inconsistent Results
**Timer-related causes**:
- **Wrong development time**: Double-check roll number
- **Temperature drift**: Verify 102°F maintenance
- **Timing errors**: Use pause rather than stopping mid-step

**Solutions**:
1. **Follow timer exactly** - resist urges to "adjust"
2. **Pre-warm tanks** before each roll
3. **Track actual times** using overtime feature

#### Running Out of Reserve Early
**Calculation check**:
- 25mL × 16 rolls = 400mL total reserve needed
- If running short, measure more carefully

**Prevention**:
1. **Measure exactly**: Use proper graduated cylinder
2. **Track usage**: Note actual amounts used
3. **Start with full liter**: Ensure complete mixing

### Technical Issues

#### App Won't Load
**Browser compatibility**:
- **Use modern browser**: Chrome, Safari, Firefox, Edge
- **Clear cache**: Force refresh with Ctrl+F5 (PC) or Cmd+Shift+R (Mac)
- **Try different device**: Phone, tablet, or computer

#### Buttons Not Responsive
**Touch issues**:
- **Screen sensitivity**: Clean screen if touch problems
- **Button size**: Designed for fingers, not stylus
- **Network delay**: App works offline once loaded

#### Lost Progress
**Data persistence**:
- **Refresh resets**: App doesn't save between sessions
- **Manual tracking**: Keep paper backup of completed rolls
- **Browser bookmarks**: Save app URL for quick access

---

## Tips & Best Practices

### Darkroom Setup

#### Before Starting
- **Organize workspace**: Chemistry, timer, materials in easy reach
- **Test equipment**: Verify thermometer, measure cylinders
- **Set up lighting**: Red safelight only, minimize screen brightness
- **Prepare materials**: Pre-cut film, load reels if needed

#### During Processing
- **Follow sequence**: Don't skip or rearrange steps
- **Watch temperature**: Check periodically, adjust if needed
- **Use timer religiously**: Don't estimate times
- **Stay organized**: Keep used/fresh chemistry separated

### Chemistry Management

#### Maximizing Results
- **Mix fresh**: Don't use old chemistry for Roll 1
- **Store properly**: Amber bottles, minimal air exposure
- **Track dates**: Note mixing date on bottles
- **Monitor quality**: Watch for color shifts in results

#### Waste Reduction
- **Measure precisely**: 25mL exactly, not "approximately"
- **Use all 17 rolls**: Don't stop early if chemistry is fresh
- **Plan batches**: Process multiple rolls when possible
- **Save time**: Batch process same film types together

### Timing Strategies

#### Efficient Processing
- **Batch preparation**: Load multiple reels between rolls
- **Temperature control**: Use water bath to maintain heat
- **Rhythm development**: Get into consistent timing routine
- **Emergency planning**: Know how to handle interruptions

#### Quality Control
- **Consistency**: Use same agitation pattern every time
- **Documentation**: Note any deviations from timer
- **Results tracking**: Keep log of actual processing times
- **Problem solving**: Adjust technique based on overtime patterns

### Mobile Usage

#### Darkroom Considerations
- **Screen brightness**: Minimize to avoid fogging film
- **Battery life**: Charge device before long sessions
- **Notifications**: Disable other apps to prevent interference
- **Backup timing**: Have analog timer as backup

#### Practical Tips
- **Large buttons**: Designed for wet/gloved hands
- **Audio alerts**: Use sound to know when timer finishes
- **Simple interface**: Easy to navigate in red light
- **Offline ready**: Works without internet once loaded

---

## Technical Reference

### Development Time Schedule

| Roll | Film | Push/Pull | Dev Time | Notes |
|------|------|-----------|----------|-------|
| 1 | 120 | +1 stop | 4:55 | Fresh chemistry—no fatigue compensation |
| 2 | 120 | normal | 3:30 | First "tired" run; replenished |
| 3 | 120 | normal | 3:36 | |
| 4 | 120 | normal | 3:43 | |
| 5 | 120 | normal | 3:49 | |
| 6 | 120 | normal | 3:55 | |
| 7 | 120 | normal | 4:01 | |
| 8 | 35mm | normal | 4:08 | |
| 9 | 35mm | normal | 4:14 | |
| 10 | 35mm | normal | 4:20 | |
| 11 | 35mm | normal | 4:27 | |
| 12 | 35mm | normal | 4:33 | |
| 13 | 35mm | normal | 4:39 | |
| 14 | 35mm | normal | 4:46 | |
| 15 | 35mm | normal | 4:52 | |
| 16 | 35mm | normal | 4:58 | |
| 17 | 35mm | normal | 5:04 | Final roll uses last 25mL from reserve |

### Chemistry Calculations

#### Volume Management
- **Total chemistry**: 1000mL
- **Working volume**: 600mL (tank capacity)
- **Reserve volume**: 400mL (storage)
- **Replenishment**: 25mL per roll
- **Total replenishment**: 25mL × 16 = 400mL (perfect match)

#### Time Increase Formula
- **Base time**: 3:30 for fresh chemistry
- **Increase rate**: ~3% per roll after replenishment
- **Fatigue compensation**: Built into schedule
- **Safety margin**: Stays within Kodak C-41 tolerance (±0.15 log D)

### Process Parameters

#### Temperature Requirements
- **Developer**: 102°F ± 0.5°F (39°C ± 0.3°C)
- **All other steps**: 102°F ± 2°F (39°C ± 1°C)
- **Critical**: Developer temperature most important

#### Agitation Pattern
- **Developer**: First 10 seconds, then 10 seconds every 30 seconds
- **Bleach-fix**: Continuous first minute, then 10s/30s
- **Wash**: Continuous flow or frequent changes
- **Stabilizer**: Gentle continuous or 10s/30s

#### Timing Tolerances
- **Pre-soak**: ±15 seconds acceptable
- **Developer**: ±5 seconds maximum deviation
- **Bleach-fix**: ±30 seconds acceptable
- **Wash**: Minimum time critical, longer okay
- **Stabilizer**: ±15 seconds acceptable

---

*This manual covers the complete operation of the Film Development Timer application. For additional support or questions about C-41 processing chemistry, consult Cinestill's official documentation or experienced color film processors.*

**Version**: 1.0  
**Last Updated**: May 2025  
**Compatibility**: All modern web browsers