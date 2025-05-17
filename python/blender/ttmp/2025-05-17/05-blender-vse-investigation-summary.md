# Blender VSE API Investigations

## Module Import Issues

We investigated why our attempt to import the module `appendix_explore_dna_video_sequence_editor` failed, while directly testing the script seemed to work.

### Findings:

1. **Python Module Naming Rules vs. File Naming:**
   - The file on disk uses hyphens: `appendix-explore-dna-video-sequence-editor.py`
   - We attempted to import it as: `appendix_explore_dna_video_sequence_editor` (with underscores)
   - Python's import system doesn't automatically convert hyphens to underscores

2. **Unexpected Import Behavior:**
   - Surprisingly, `import appendix-explore-dna-video-sequence-editor` works in Blender's Python!
   - However, this is unusual and not recommended as it violates Python's standard module naming conventions
   - Standard Python would reject hyphenated module names entirely

3. **File Existence Verification:**
   - The underscored filename doesn't exist on disk: `appendix_explore_dna_video_sequence_editor.py`
   - Only the hyphenated version exists: `appendix-explore-dna-video-sequence-editor.py`

### Best Practices for Module Naming:

1. Always use underscores instead of hyphens in Python module filenames
2. Module names should follow the same rules as Python variable names
3. If you must work with files containing hyphens, use one of these approaches:
   - Rename the file to use underscores
   - Use file-based execution methods like `exec(open(filename).read())` instead of imports

## Segment Removal Issues

We investigated why the segment removal operation (cutting out the middle part of a strip) was failing in our original implementation.

### Findings about Blender's Split Operation:

1. **Non-Intuitive Strip Creation:**
   - When splitting a strip at frame N, Blender doesn't create two clearly separated strips at different timeline positions
   - Instead, it often keeps both parts starting at the same frame_start position
   - It uses frame_offset_start/end values to determine which portion of the source content each strip displays

2. **Demonstrating the Problem with Real Data:**
   - First Split at frame 185 of a strip that was at frames 1-740:
     - Left part: frame_start=1, frame_final_end=185, frame_offset_end=555
     - Right part: frame_start=1, frame_final_start=185, frame_offset_start=184
     
   - Second Split at frame 370 of the right part:
     - Middle part: frame_start=1, frame_final_start=185, frame_final_end=370, frame_offset_start=184, frame_offset_end=370
     - End part: frame_start=1, frame_final_start=370, frame_offset_start=369

3. **Identification Challenges:**
   - Our original code assumed the right part would always be at frame_start=N after splitting
   - It also assumed frame_final_end=N would reliably identify the left part
   - These assumptions failed because Blender's behavior is more complex

### Our Solutions:

1. **Multiple Identification Criteria:**
   - Added several fallback mechanisms to identify left and right parts after splitting
   - Used combinations of frame_final_start, frame_final_end, frame_offset_start, and frame_offset_end
   
2. **Explicit Strip Tracking:**
   - Used strip.as_pointer() to reliably track strip identity across operations
   - Identified new vs. updated strips after splitting
   
3. **Robust Error Handling:**
   - Added comprehensive error checking at each step
   - Implemented graceful fallback when strip identification failed

## Key Takeaways for Blender VSE Scripting

1. **Strip Identity and Modification:**
   - Always track strips by pointer (strip.as_pointer()) when their identity matters across operations
   - Be aware that operators can modify existing strips rather than just creating new ones

2. **Offset vs. Position:**
   - Understand that frame_offset_start/end control what portion of source content is visible
   - The actual timeline position is a combination of frame_start and these offsets

3. **Design Patterns for VSE Scripting:**
   - Use extensive error handling for operations that might fail
   - Add multiple identification heuristics rather than relying on a single criterion
   - Implement detailed logging to assist in debugging VSE operations
   - Use isolated testing environments for each feature to avoid state contamination