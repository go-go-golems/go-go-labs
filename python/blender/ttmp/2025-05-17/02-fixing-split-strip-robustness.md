# Debugging Blender VSE Python API - Handling Split Strip Edge Cases

## The Problem: Crashing on Split Operations

When trying to demonstrate VSE sequence clip operations in our `chapter-3-blender-book.py` script, we encountered a frustrating crash:

```
Splitting 'Clip2.008' at frame 404
Split strip 'Clip2.008' at frame 404
  Left part: 'Clip2.008', frames 338.0-404
  Right part: Not created or not found (e.g., split at end of content)
Original clip split into: 'Clip2.008' and 'None'
Split strip 'Audio2.008' at frame 404
  Left part: 'Audio2.008', frames 338.0-404
  Right part: Not created or not found (e.g., split at end of content)
Original audio clip split into: 'Audio2.008' and 'None'

--- Demonstration: Removing a Segment ---

Removing middle segment from 'Clip3.005'
  Original duration: 739 frames

Removing segment from 716 to 962 from 'Clip3.005'
Split strip 'Clip3.005' at frame 716
  Left part: 'Clip3.005', frames 470.0-716
  Right part: Not created or not found (e.g., split at end of content)
Traceback (most recent call last):
  [...]
AttributeError: 'NoneType' object has no attribute 'select'
```

The issue occurred during our `remove_segment()` function, which depends on `split_strip()`. We were making two sequential splits to remove a middle segment, but the script crashed with a `NoneType` error when trying to access the `.select` attribute.

## The Root Cause

Analyzing the error message and code, we discovered several potential issues:

1. The first `split_strip()` call worked - sort of - but only produced a left part and returned `None` for the right part
2. When `remove_segment()` tried to call `split_strip()` on `part_b` (which was `None`), it crashed
3. Our code wasn't handling this edge case gracefully, assuming both parts would always be created

The split operation was working differently than expected in these edge cases. While the split should theoretically create two parts - a left and right strip - in some cases it was only modifying the existing strip rather than creating a new one. This can happen when:

- Splitting exactly at a strip's end frame (nothing to split)
- The strip is too short to split meaningfully
- The split point is outside the strip's boundaries
- There's a rounding issue with floating-point frame numbers

## The Solution

We implemented several robustness improvements to handle these edge cases:

### 1. Enhanced `split_strip()` Functionality

The improved `split_strip()` function now:

```python
def split_strip(strip, frame, select_right=True):
    """
    Split a strip at the specified frame.
    
    This uses the sequencer.split operator to cut a strip into two parts.
    The split is made at the specified frame, and by default the right
    part remains selected after the cut.
    
    Args:
        strip (bpy.types.Strip): The strip to split.
        frame (int or float): Frame at which to make the cut (will be converted to int).
        select_right (bool): Whether to select the right part after cutting.
        
    Returns:
        tuple: (left_part, right_part) - the two resulting strips. Either can be None
               if the split operation failed or produced only one side (e.g., frame
               equals strip boundary).
    """
    # Ensure we have an integer frame number
    frame = int(frame)

    # If the frame is outside the strip bounds, do nothing and warn
    if frame <= strip.frame_start or frame >= strip.frame_final_end:
        print(f"  [WARN] Split frame {frame} is outside '{strip.name}' bounds (" \
              f"{strip.frame_start}-{strip.frame_final_end}). No split performed.")
        return (strip, None)

    seq = C.scene.sequence_editor

    # Record existing strips (by pointer) before the split so we can detect new ones
    pre_split_ids = {s.as_pointer() for s in seq.strips_all}

    # Select only the strip we want to split
    for s in seq.strips_all:
        s.select = (s == strip)

    # Set current frame to the split point
    C.scene.frame_current = frame

    # Perform the split – if it fails (e.g. strip locked) catch the exception
    try:
        bpy.ops.sequencer.split(frame=frame, channel=strip.channel, type='SOFT')
    except RuntimeError as e:
        print(f"  [ERROR] Split failed on '{strip.name}': {e}")
        return (strip, None)

    # Identify strips after split
    left_part = None
    right_part = None
    for s in seq.strips_all:
        if s.as_pointer() in pre_split_ids:
            # Existing strip – may have updated range; determine if it's left or right
            if s.channel == strip.channel:
                if s.frame_final_end == frame:
                    left_part = s
                elif s.frame_start == frame:
                    right_part = s
        else:
            # New strip created by split – decide if left or right
            if s.channel == strip.channel:
                if s.frame_final_end == frame:
                    left_part = s
                elif s.frame_start == frame:
                    right_part = s

    # Fallback heuristics if one side still missing
    if not left_part or not right_part:
        for s in seq.strips_all:
            if s.channel != strip.channel:
                continue
            if s.frame_final_end <= frame and (not left_part):
                left_part = s
            elif s.frame_start >= frame and (not right_part):
                right_part = s

    print(f"Split strip '{strip.name}' at frame {frame}")
    if left_part:
        print(f"  Left part: '{left_part.name}', frames {left_part.frame_start}-{left_part.frame_final_end}")
    else:
        print("  Left part: Not created or not found")
    if right_part:
        print(f"  Right part: '{right_part.name}', frames {right_part.frame_start}-{right_part.frame_final_end}")
    else:
        print("  Right part: Not created or not found (e.g., split at end of content)")

    # Manage selection state
    if left_part and right_part:
        if select_right:
            left_part.select = False
            right_part.select = True
        else:
            left_part.select = True
            right_part.select = False

    return (left_part, right_part)
```

Key improvements:
- **Boundary Checking**: We now check if the frame is outside the strip's range before attempting to split
- **Exception Handling**: We catch any runtime errors from the operator
- **Enhanced Strip Detection**: We use multiple methods to find the resulting strips:
  - Tracking existing strips via memory pointer before/after split
  - Checking strip ranges relative to the split frame
  - Implementing fallback heuristics if detection methods fail
- **Improved Reporting**: Better console output when splits don't work as expected

### 2. Hardened `remove_segment()` Function

The `remove_segment()` function now includes proper null checking:

```python
def remove_segment(strip, start_frame, end_frame):
    """
    Remove a segment from a strip and close the gap.
    
    If the splits fail (e.g., frame outside strip bounds), the function will
    abort gracefully and return (strip, None).
    """
    print(f"\nRemoving segment from {start_frame} to {end_frame} from '{strip.name}'")

    # Split at the start of segment
    part_a, part_b = split_strip(strip, start_frame, select_right=True)
    if not part_b:
        print("  [WARN] Unable to create middle/right part at first split – aborting segment removal.")
        return (part_a, None)

    # Split at the end of segment (working on part_b)
    part_b, part_c = split_strip(part_b, end_frame, select_right=True)
    if not part_b:
        print("  [WARN] Second split produced no middle segment – aborting segment removal.")
        return (part_a, part_c)

    # Remove the middle segment (part_b)
    seq = C.scene.sequence_editor
    if part_b and part_b in seq.strips_all:
        seq.strips.remove(part_b)
        print(f"  Removed segment '{part_b.name}'")

    # Move part_c to start at the end of part_a
    if part_c and part_a:
        part_c.frame_start = part_a.frame_final_end
        print(f"  Moved '{part_c.name}' to frame {part_c.frame_start} (closing the gap)")

    return (part_a, part_c)
```

Key improvements:
- **Null Checking**: We check if `part_b` is None after the first split and abort
- **Graceful Failure**: If either split fails, we return what we have and warn the user
- **Existence Verification**: We verify strips exist before removing them
- **Better Error Reporting**: More descriptive warning messages about failed steps

## Lessons Learned

Working with Blender's VSE operators through Python requires careful handling due to several factors:

1. **Operator-Based API Quirks**: Unlike pure data operations, Blender's operator system (bpy.ops) can have side effects, success/failure contexts, and dependencies on state that aren't immediately obvious.

2. **Timeline Boundary Issues**: Split operations near strip boundaries may silently fail or behave differently. Always verify frame values are valid relative to strips.

3. **Soft vs. Hard Errors**: In many cases, Blender will simply not perform an operation rather than throwing an error. Our `split_strip` function didn't initially detect these "soft failures."

4. **Defensive Programming Is Essential**: When working with Blender's operator system, always:
   - Check input values for validity
   - Verify operation results rather than assuming success
   - Implement graceful fallbacks for when things don't work as expected
   - Add extra debugging output to identify edge cases

The improved code is much more robust, providing both better error handling and clearer feedback when operations can't be performed as requested. While the VSE Python API is powerful, its operator-based nature requires careful handling to create production-ready scripts.

## Practical Implications

This bugfix makes our educational script more reliable, but the lessons apply to any VSE scripting work:

- Always check boundary conditions (strip edges, empty sequences)
- Validate operation results by inspecting the sequence editor state
- Use memory pointers (as_pointer()) to track objects through operations
- Implement graceful degradation when operations partially succeed
- Add detailed console output to help diagnose issues

These principles lead to more robust automation scripts that can handle real-world editing tasks without crashing or producing unexpected results. 