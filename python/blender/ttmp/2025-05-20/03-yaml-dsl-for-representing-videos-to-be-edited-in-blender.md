# BlenderEdit: A YAML DSL for Defining Blender Video Edits

BlenderEdit is a YAML-based domain-specific language (DSL) that provides a declarative way to represent video edits for Blender's Video Sequence Editor (VSE). This specification defines how to express strips, transitions, cuts, effects, and other video editing operations in a human-readable format that can be parsed and executed by Blender's Python API.

## Core Concepts

BlenderEdit YAML files represent a complete video editing project with the following core concepts:

1. **Project Settings** - Overall settings for the video project (resolution, frame rate, output format)
2. **Strips** - Media elements (video, audio, image, text) placed on the timeline
3. **Effects** - Modifications to media (color grading, transforms, etc.)
4. **Transitions** - Ways to blend between strips (crossfades, wipes)
5. **Meta-strips** - Groups of strips treated as a single unit
6. **Channel Management** - Organization of strips across channels (tracks)
7. **Markers** - Named points in time for reference

## Specification

### Project Settings

The top level of a BlenderEdit YAML document defines project-wide settings:

```yaml
project:
  name: "My Project"
  resolution: [1920, 1080]  # Width, height
  fps: 30
  start_frame: 1
  end_frame: 300  # Optional, can be auto-calculated from strips
  render:
    path: "/path/to/output.mp4"
    format: "mp4"
    codec: "h264"
    audio_codec: "aac"
    quality: "high"  # Options: low, medium, high, lossless
```

### Channels

Channels can be defined and named for better organization:

```yaml
channels:
  - id: 1
    name: "Main Video"
  - id: 2
    name: "Overlay"
  - id: 3
    name: "Music"
  - id: 4
    name: "Voice"
```

### Strips

Strips are the core building blocks. They have common properties and type-specific properties:

```yaml
strips:
  - id: "intro"  # Unique identifier for this strip
    type: "movie"
    path: "/path/to/intro.mp4"
    channel: 1
    start: 1  # Timeline frame where strip starts
    # Additional type-specific properties
    with_audio: true  # Whether to include the video's audio
    
  - id: "voice"
    type: "sound"
    path: "/path/to/narration.wav"
    channel: 4
    start: 30
    volume: 0.8
    
  - id: "logo"
    type: "image"
    path: "/path/to/logo.png"
    channel: 2
    start: 50
    duration: 100  # For images, which don't have inherent length
    
  - id: "title"
    type: "text"
    channel: 2
    start: 10
    duration: 90
    text: "My Awesome Video"
    size: 80
    position: [0.5, 0.5]  # Normalized coordinates (center of screen)
    color: [1, 1, 1, 1]  # RGBA values
    alignment: "center"
```

### Trimming

Trimming can be specified on any strip to use only part of the source media:

```yaml
strips:
  - id: "interview"
    type: "movie"
    path: "/path/to/interview.mp4"
    channel: 1
    start: 100  # Timeline position
    trim:
      start: 30  # Skip the first 30 frames of source
      end: 20    # Skip the last 20 frames of source
```

### Effects

Effects can be applied to strips:

```yaml
effects:
  - id: "color_grade_intro"
    type: "color_balance"
    target: "intro"  # References strip ID
    lift: [1.1, 1.0, 0.9]
    gamma: [0.9, 0.9, 0.9]
    gain: [1.0, 1.0, 1.0]
    
  - id: "transform_logo"
    type: "transform"
    target: "logo"
    scale: [0.5, 0.5]  # 50% size
    position: [100, 50]  # Move 100px right, 50px up
    rotation: 0.2  # Radians (~11.5 degrees)
    
  - id: "slow_motion"
    type: "speed"
    target: "action_shot"
    factor: 0.5  # 50% speed (slow motion)
```

### Transitions

Transitions blend between strips:

```yaml
transitions:
  - id: "intro_to_main"
    type: "cross"  # crossfade
    source: "intro"
    target: "main_video"
    duration: 30  # frames
    
  - id: "main_to_outro"
    type: "wipe"
    source: "main_video"
    target: "outro"
    duration: 45
    wipe_type: "single"  # SINGLE/DOUBLE/etc.
    wipe_direction: "in"  # IN/OUT
    wipe_blur: 0.5  # Blur factor
```

### Meta-strips

Group strips together:

```yaml
meta_strips:
  - id: "intro_sequence"
    strips: ["title", "logo", "intro"]
    start: 1
    end: 150
```

### Markers

Set named points on the timeline:

```yaml
markers:
  - name: "Chapter 1"
    frame: 1
  - name: "Interview Start"
    frame: 300
```

## Complete Example

Here's a full example of a short video project:

```yaml
project:
  name: "Product Introduction"
  resolution: [1920, 1080]
  fps: 30
  render:
    path: "/videos/product_intro.mp4"
    format: "mp4"
    codec: "h264"
    quality: "high"

channels:
  - id: 1
    name: "Main Video"
  - id: 2
    name: "Overlays"
  - id: 3
    name: "Music"
  - id: 4
    name: "Voice"

strips:
  # Main video segments
  - id: "intro"
    type: "movie"
    path: "/footage/company_intro.mp4"
    channel: 1
    start: 1
    with_audio: false
    
  - id: "product_demo"
    type: "movie"
    path: "/footage/product_demo.mp4"
    channel: 1
    start: 90  # Starts after intro
    trim:
      start: 15  # Skip first 15 frames of source
      end: 0
    with_audio: true
    
  - id: "closing"
    type: "movie"
    path: "/footage/closing.mp4"
    channel: 1
    start: 390  # Starts after product demo
    with_audio: false
  
  # Overlays
  - id: "company_logo"
    type: "image"
    path: "/assets/logo.png"
    channel: 2
    start: 10
    duration: 80
  
  - id: "title"
    type: "text"
    channel: 2
    start: 30
    duration: 60
    text: "Revolutionary Product"
    size: 80
    position: [0.5, 0.4]
    color: [1, 1, 1, 1]
    alignment: "center"
    
  - id: "website"
    type: "text"
    channel: 2
    start: 390
    duration: 120
    text: "www.example.com"
    size: 60
    position: [0.5, 0.2]
    color: [1, 1, 1, 1]
    alignment: "center"
  
  # Audio
  - id: "background_music"
    type: "sound"
    path: "/audio/upbeat_track.mp3"
    channel: 3
    start: 1
    volume: 0.4
    
  - id: "voiceover"
    type: "sound"
    path: "/audio/narration.wav"
    channel: 4
    start: 30
    volume: 1.0

effects:
  - id: "warm_color_grade"
    type: "color_balance"
    target: "product_demo"
    lift: [1.05, 1.0, 0.95]
    gamma: [1.0, 0.95, 0.95]
    
  - id: "logo_transform"
    type: "transform"
    target: "company_logo"
    scale: [0.3, 0.3]
    position: [50, 50]  # Top-left corner
    
  - id: "fade_music_end"
    type: "volume_keyframes"
    target: "background_music"
    keyframes:
      - frame: 480
        value: 0.4
      - frame: 510
        value: 0.0

transitions:
  - id: "intro_to_demo"
    type: "cross"
    source: "intro"
    target: "product_demo"
    duration: 30
    
  - id: "demo_to_closing"
    type: "cross"
    source: "product_demo"
    target: "closing"
    duration: 30

markers:
  - name: "Intro"
    frame: 1
  - name: "Product Features"
    frame: 150
  - name: "Call to Action"
    frame: 420
```

## More Examples

### Example 1: Simple Slideshow

```yaml
project:
  name: "Vacation Slideshow"
  resolution: [1920, 1080]
  fps: 24
  render:
    path: "/videos/vacation.mp4"
    format: "mp4"
    codec: "h264"

strips:
  - id: "photo1"
    type: "image"
    path: "/photos/beach1.jpg"
    channel: 1
    start: 1
    duration: 72  # 3 seconds at 24fps
    
  - id: "photo2"
    type: "image"
    path: "/photos/sunset.jpg"
    channel: 1
    start: 73
    duration: 72
    
  - id: "photo3"
    type: "image"
    path: "/photos/family.jpg"
    channel: 1
    start: 145
    duration: 72
    
  - id: "music"
    type: "sound"
    path: "/audio/vacation_music.mp3"
    channel: 2
    start: 1
    volume: 0.7

transitions:
  - id: "photo1_to_photo2"
    type: "cross"
    source: "photo1"
    target: "photo2"
    duration: 24  # 1 second crossfade
    
  - id: "photo2_to_photo3"
    type: "cross"
    source: "photo2"
    target: "photo3"
    duration: 24

effects:
  - id: "photo1_ken_burns"
    type: "transform"
    target: "photo1"
    keyframes:
      - frame: 1
        scale: [1.0, 1.0]
        position: [0, 0]
      - frame: 72
        scale: [1.2, 1.2]
        position: [-50, -20]
```

### Example 2: Multi-camera Interview

```yaml
project:
  name: "Interview Program"
  resolution: [1920, 1080]
  fps: 30
  
channels:
  - id: 1
    name: "Camera 1 (Wide)"
  - id: 2
    name: "Camera 2 (Close-up)"
  - id: 3
    name: "Graphics"
  - id: 4
    name: "Main Audio"

strips:
  - id: "intro_graphic"
    type: "movie"
    path: "/graphics/show_intro.mp4"
    channel: 3
    start: 1
    duration: 90
    with_audio: true
    
  - id: "cam1_full"
    type: "movie"
    path: "/footage/interview_wide.mp4"
    channel: 1
    start: 91
    with_audio: false
    
  - id: "cam2_full"
    type: "movie"
    path: "/footage/interview_closeup.mp4"
    channel: 2
    start: 91
    with_audio: false
    
  - id: "main_audio"
    type: "sound"
    path: "/audio/interview_main_audio.wav"
    channel: 4
    start: 91
    volume: 1.0
    
  - id: "lower_third"
    type: "movie"
    path: "/graphics/lower_third.mp4"
    channel: 3
    start: 150
    duration: 180
    with_audio: false

# Camera cuts (achieved by enabling/disabling strips)
cuts:
  - action: "enable"
    target: "cam1_full"
    start: 91
    end: 200
    
  - action: "enable"
    target: "cam2_full"
    start: 201
    end: 300
    
  - action: "enable"
    target: "cam1_full"
    start: 301
    end: 400
    
  - action: "enable"
    target: "cam2_full"
    start: 401
    end: 500
```

### Example 3: Documentary Style with B-roll

```yaml
project:
  name: "Wildlife Documentary"
  resolution: [3840, 2160]  # 4K UHD
  fps: 25
  
channels:
  - id: 1
    name: "Main Footage"
  - id: 2
    name: "B-roll"
  - id: 3
    name: "Lower Thirds"
  - id: 4
    name: "Music"
  - id: 5
    name: "Narration"

strips:
  # Main interview segments
  - id: "interview_1"
    type: "movie"
    path: "/footage/expert_interview_1.mp4"
    channel: 1
    start: 1
    trim:
      start: 48  # Skip first 2 seconds
      end: 0
    with_audio: false
    
  # B-roll overlays
  - id: "animals_1"
    type: "movie"
    path: "/footage/lions_walking.mp4"
    channel: 2
    start: 75
    duration: 125
    with_audio: false
    
  - id: "animals_2"
    type: "movie"
    path: "/footage/zebras_running.mp4"
    channel: 2
    start: 225
    duration: 100
    with_audio: false
    
  # Text overlays
  - id: "expert_name"
    type: "text"
    channel: 3
    start: 25
    duration: 100
    text: "Dr. Jane Smith\nWildlife Biologist"
    size: 45
    position: [0.1, 0.1]  # Bottom left
    color: [1, 1, 1, 0.8]
    alignment: "left"
    
  # Audio tracks
  - id: "narration"
    type: "sound"
    path: "/audio/narrator_track.wav"
    channel: 5
    start: 1
    
  - id: "ambient_music"
    type: "sound"
    path: "/audio/savanna_ambient.mp3"
    channel: 4
    start: 1
    volume: 0.3

effects:
  # Color grade the main footage
  - id: "main_grade"
    type: "color_balance"
    target: "interview_1"
    lift: [1.0, 1.0, 1.0]
    gamma: [1.0, 1.02, 1.05]  # Slightly cooler look
    
  # Add gaussian blur to the B-roll for stylistic effect
  - id: "animals_1_blur"
    type: "gaussian_blur"
    target: "animals_1"
    size: 20
    
  # Slow down some B-roll footage
  - id: "animals_2_slowmo"
    type: "speed"
    target: "animals_2"
    factor: 0.5

transitions:
  # Fade from interview to B-roll
  - id: "interview_to_broll"
    type: "cross"
    source: "interview_1"
    target: "animals_1"
    duration: 25  # 1 second at 25fps
```

### Example 4: Music Video with Effects

```yaml
project:
  name: "Music Video Project"
  resolution: [1920, 1080]
  fps: 24
  
strips:
  - id: "performer_1"
    type: "movie"
    path: "/footage/singer_take1.mp4"
    channel: 1
    start: 1
    with_audio: true
    
  - id: "performer_2"
    type: "movie"
    path: "/footage/singer_take2.mp4"
    channel: 1
    start: 289  # 12 seconds in
    with_audio: true
    
  - id: "overlay_lights"
    type: "movie"
    path: "/effects/light_leaks.mp4"
    channel: 2
    start: 100
    duration: 200
    blend_type: "screen"
    opacity: 0.7
    
  - id: "studio_logo"
    type: "image"
    path: "/branding/record_label.png"
    channel: 3
    start: 500
    duration: 120
    
  - id: "background_track"
    type: "sound"
    path: "/audio/song_instrumental.wav"
    channel: 4
    start: 1
    volume: 0.8

effects:
  - id: "vhs_effect"
    type: "custom_strip"
    target: "performer_1"
    effect_code: """
    import bpy
    # This would be implemented as a custom strip effect
    # Simulating VHS tape distortion
    strip = bpy.context.scene.sequence_editor.strips.get('performer_1')
    mod = strip.modifiers.new(name='VHS Look', type='CURVES')
    # Set up curves to get VHS look
    """
    
  - id: "strobe_keyframes"
    type: "transform"
    target: "performer_2"
    keyframes:
      - frame: 300
        opacity: 1.0
      - frame: 303
        opacity: 0.0
      - frame: 306
        opacity: 1.0
      - frame: 309
        opacity: 0.0
      - frame: 312
        opacity: 1.0

transitions:
  - id: "performer_transition"
    type: "wipe"
    source: "performer_1"
    target: "performer_2"
    duration: 12
    wipe_type: "clock"
```

### Example 5: Tutorial Video with Split Screen

```yaml
project:
  name: "Software Tutorial"
  resolution: [1920, 1080]
  fps: 30
  
channels:
  - id: 1
    name: "Screen Recording"
  - id: 2
    name: "Face Cam"
  - id: 3
    name: "Text Overlays"
  - id: 4
    name: "Audio"

strips:
  - id: "intro"
    type: "movie"
    path: "/footage/tutorial_intro.mp4"
    channel: 1
    start: 1
    with_audio: true
    
  - id: "screen_recording"
    type: "movie"
    path: "/footage/screen_capture.mp4"
    channel: 1
    start: 91  # After intro
    with_audio: false
    
  - id: "face_cam"
    type: "movie"
    path: "/footage/presenter.mp4"
    channel: 2
    start: 91
    with_audio: true
    
  - id: "step_one"
    type: "text"
    channel: 3
    start: 120
    duration: 150
    text: "Step 1: Open the application"
    size: 50
    position: [0.5, 0.9]  # Top center
    color: [1, 1, 0.5, 1]  # Yellow-ish
    alignment: "center"

effects:
  # Picture-in-picture for face cam
  - id: "pip_face"
    type: "transform"
    target: "face_cam"
    scale: [0.25, 0.25]  # Small size
    position: [1700, 200]  # Bottom right
    
  # Highlight cursor in screen recording
  - id: "cursor_highlight"
    type: "glow"
    target: "screen_recording"
    threshold: 0.8
    radius: 10
    
  # Zoom in on important UI element
  - id: "zoom_ui"
    type: "transform"
    target: "screen_recording" 
    keyframes:
      - frame: 200
        scale: [1.0, 1.0]
        position: [0, 0]
      - frame: 210
        scale: [1.5, 1.5]
        position: [-200, -150]
      - frame: 300
        scale: [1.5, 1.5]
        position: [-200, -150]
      - frame: 310
        scale: [1.0, 1.0]
        position: [0, 0]

transitions:
  - id: "intro_to_tutorial"
    type: "cross"
    source: "intro"
    target: "screen_recording"
    duration: 30
```

## Implementation Notes

To implement a parser for this YAML DSL:

1. Parse the YAML document using a standard YAML library
2. Validate against the schema defined in this specification
3. Convert the parsed data into Blender Python API calls
4. Execute the calls in the correct order to build the sequence in Blender

The implementation would handle:
- Converting paths to absolute paths if needed
- Calculating frames for transitions and effects
- Ensuring strips are created in the correct order
- Managing dependencies between strips, effects, and transitions
- Applying the appropriate Blender API calls for each element

With this specification, video editors can define complex edits in a simple, human-readable format that can be version-controlled, shared, and executed programmatically in Blender. 