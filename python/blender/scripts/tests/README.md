# Blender VSE Test Suite

This directory contains test files for the Blender Video Sequence Editor (VSE) Python API functionality.

## Test Files Overview

### Core Test Files

1. `test_transition_utils.py`
   - [x] Crossfade transitions
   - [x] Gamma crossfade transitions
   - [x] Wipe transitions
   - [ ] Audio transitions
   - [ ] Audio fades
   - [x] Color fade effects
   - [x] Test runner implementation

2. `test_flicker_effect.py`
   - [x] Video interleaving effects
   - [x] Multi-channel scene setup
   - [x] Flicker timing verification
   - [ ] Audio synchronization

### Demo and Example Files

1. `chapter_3_demos.py`
   - [ ] Trimming clips
   - [ ] Splitting clips
   - [ ] Removing segments
   - [ ] Slip editing
   - [ ] Moving strips
   - [ ] Scene cleanup verification

2. `auto_trailer_generator.py`
   - [ ] Trailer length calculation
   - [ ] Clip selection algorithm
   - [ ] Transition generation
   - [ ] Text overlay system
   - [ ] Audio crossfading
   - [ ] Final render verification

### Test Runners

1. `run_transition_tests.py`
   - [ ] CLI interface implementation
   - [ ] Individual test execution
   - [ ] Batch test execution
   - [ ] Test result reporting
   - Available tests:
     - [ ] crossfade
     - [ ] gamma_crossfade
     - [ ] wipe
     - [ ] audio_fade
     - [ ] audio_crossfade
     - [ ] fade_to_color

2. `chapter_3_test_runner.py`
   - [ ] Scene isolation system
   - [ ] Individual demo execution
   - [ ] Batch demo execution
   - [ ] Result verification
   - Available demos:
     - [ ] trim
     - [ ] split
     - [ ] remove
     - [ ] slip
     - [ ] move

## Cleanup Tasks

1. Media Directory Handling:
   - [ ] Consolidate directory finding code
   - [ ] Create unified utility function in `vse_utils.py`
   - [ ] Update all test files to use new function

2. Scene Setup:
   - [ ] Create standardized scene setup function
   - [ ] Add configurable channel handling
   - [ ] Implement cleanup routines
   - [ ] Update all tests to use new setup

3. Strip Selection:
   - [ ] Create unified strip selection utility
   - [ ] Add support for multiple selection modes
   - [ ] Implement error handling
   - [ ] Update affected files

## Running Tests

Most test files can be run directly from Blender's Text Editor. For specific test runners:

1. Transition Tests:
```python
import run_transition_tests
run_transition_tests.run_test('crossfade')  # or 'all' for all tests
```

2. Chapter 3 Demos:
```python
import chapter_3_test_runner
chapter_3_test_runner.run_test('trim')  # or 'all' for all demos
```

## Progress Tracking

- [ ] All core tests implemented
- [ ] All demos verified
- [ ] Test runners completed
- [ ] Code cleanup finished
- [ ] Documentation updated 