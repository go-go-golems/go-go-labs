# SciFiEdit: A YAML DSL for Sci-Fi Movie Editing with Mood/Tension Organization

SciFiEdit is a YAML-based domain-specific language designed specifically for science fiction film editors who work with complex VFX compositions and want to organize their edits around mood, tension arcs, and narrative beats. This specification builds on the BlenderEdit foundation but introduces specialized concepts for sci-fi production workflows.

## Core Concepts

SciFiEdit extends traditional timeline-based editing with these sci-fi-specific concepts:

1. **Mood/Tension Blocks** - Sections of the timeline defined by emotional intensity and atmosphere
2. **VFX Composition Layers** - Specialized organization for complex effects stacks
3. **Scene Types** - Categorization of shots (e.g., "space exterior", "hologram interface")
4. **Sound Design Integration** - Specialized audio categories for sci-fi elements
5. **Color Grading Profiles** - Pre-defined looks for different worlds, timelines, or realities

## Specification

### Project Settings

The top level of a SciFiEdit YAML document defines project-wide settings:

```yaml
project:
  name: "Quantum Horizon"
  resolution: [3840, 2160]  # 4K UHD
  fps: 24
  start_frame: 1
  aspect_ratio: "2.39:1"  # Cinematic widescreen
  color_science: "ACES"   # Academy Color Encoding System
  render:
    path: "/renders/quantum_horizon_final.mp4"
    format: "mp4"
    codec: "h265"
    quality: "high"
    bit_depth: 10  # 10-bit color
```

### Mood/Tension Blocks

Define narrative sections by emotional intensity, useful for organizing shots by feeling rather than just chronology:

```yaml
mood_blocks:
  - id: "opening_mystery"
    name: "Opening Mystery"
    tension: 3  # Scale 1-10
    atmosphere: "wonder"  # Options: wonder, dread, tension, action, calm, etc.
    color_palette: "cool_blue"
    start_frame: 1
    end_frame: 720  # 30 seconds
    
  - id: "rising_danger"
    name: "Rising Danger"
    tension: 6
    atmosphere: "dread"
    color_palette: "amber_warning"
    start_frame: 721
    end_frame: 2160  # 1 minute mark
    
  - id: "climactic_confrontation"
    name: "Climactic Confrontation"
    tension: 9
    atmosphere: "action"
    color_palette: "high_contrast"
    start_frame: 2161
    end_frame: 4320  # 3 minute mark
```

### VFX Layer Categories

Organize VFX elements into meaningful groups:

```yaml
vfx_categories:
  - id: "practical"
    name: "Practical Effects"
    channel_range: [1, 5]
    
  - id: "cgi_base"
    name: "CGI Base Elements"
    channel_range: [6, 10]
    
  - id: "set_extensions"
    name: "Set Extensions"
    channel_range: [11, 15]
    
  - id: "creature_fx"
    name: "Creature Effects"
    channel_range: [16, 20]
    
  - id: "particle_fx"
    name: "Particle Systems"
    channel_range: [21, 25]
    
  - id: "holograms"
    name: "Holographic Displays"
    channel_range: [26, 30]
    
  - id: "space_bg"
    name: "Space Backgrounds"
    channel_range: [31, 35]
```

### Channel Definitions

Channels are grouped by VFX category and purpose:

```yaml
channels:
  # Base footage
  - id: 1
    name: "Main Footage"
    category: "practical"
    
  - id: 3
    name: "Greenscreen Elements"
    category: "practical"
    
  # VFX Elements
  - id: 7
    name: "Character CGI"
    category: "cgi_base"
    
  - id: 12
    name: "Spacecraft Extensions"
    category: "set_extensions"
    
  - id: 17
    name: "Alien Creature"
    category: "creature_fx"
    
  - id: 22
    name: "Energy Weapons"
    category: "particle_fx"
    
  - id: 27
    name: "Computer Interface"
    category: "holograms"
    
  - id: 32
    name: "Nebula Background"
    category: "space_bg"
    
  # Audio categories
  - id: 50
    name: "Dialogue"
    category: "audio"
    
  - id: 52
    name: "Ambient Space"
    category: "audio"
    
  - id: 54
    name: "Tech Sound Design"
    category: "audio"
    
  - id: 56
    name: "Creature Vocals"
    category: "audio"
    
  - id: 58
    name: "Music Score"
    category: "audio"
```

### Scene Types

Define specialized scene type categories to help organize footage:

```yaml
scene_types:
  - id: "space_ext"
    name: "Space Exterior"
    default_atmosphere: "wonder"
    
  - id: "ship_int"
    name: "Ship Interior"
    default_atmosphere: "tension"
    
  - id: "alien_world"
    name: "Alien Planet Surface"
    default_atmosphere: "discovery"
    
  - id: "cryosleep"
    name: "Cryosleep Sequence"
    default_atmosphere: "ethereal"
    
  - id: "holo_interface"
    name: "Holographic Interface"
    default_atmosphere: "technical"
```

### Color Grading Profiles

Pre-defined color grades for different worlds or realities:

```yaml
color_profiles:
  - id: "earth_reality"
    name: "Earth Reality"
    lift: [1.0, 1.02, 1.04]
    gamma: [1.0, 1.0, 0.98]
    gain: [0.99, 1.0, 1.02]
    
  - id: "alien_world"
    name: "Alien World"
    lift: [0.98, 1.0, 1.1]
    gamma: [0.9, 1.05, 1.1]
    gain: [1.0, 1.1, 1.2]
    
  - id: "virtual_reality"
    name: "Virtual Reality"
    lift: [1.05, 1.05, 1.15]
    gamma: [1.1, 0.95, 1.15] 
    gain: [1.2, 1.0, 1.3]
    
  - id: "flashback"
    name: "Flashback Sequences"
    lift: [1.1, 1.05, 0.95]
    gamma: [1.1, 1.0, 0.9]
    gain: [1.2, 1.1, 0.9]
    saturation: 0.8
```

### Strips

Strips now include mood block association and scene type:

```yaml
strips:
  - id: "establishing_shot"
    type: "movie"
    path: "/footage/space_station_approach.exr"
    channel: 1
    start: 1
    scene_type: "space_ext"
    mood_block: "opening_mystery"
    vfx_ready: true  # Indicates this is a finished VFX shot
    
  - id: "alien_planet_bg"
    type: "movie"
    path: "/vfx/alien_landscape_bg.exr"
    channel: 32  # Space backgrounds channel
    start: 240
    scene_type: "alien_world"
    mood_block: "rising_danger"
    
  - id: "character_over_greenscreen"
    type: "movie"
    path: "/footage/commander_speech_greenscreen.exr"
    channel: 3  # Greenscreen elements
    start: 240
    trim:
      start: 12
      end: 0
    scene_type: "ship_int"
    mood_block: "rising_danger"
    
  - id: "hologram_overlay"
    type: "movie"
    path: "/vfx/tactical_hologram.exr"
    channel: 27  # Hologram channel
    start: 300
    duration: 480
    blend_type: "add"
    opacity: 0.8
    scene_type: "holo_interface"
    mood_block: "rising_danger"
    
  - id: "energy_beam"
    type: "movie"
    path: "/vfx/weapon_beam.exr"
    channel: 22  # Energy weapons
    start: 2400
    scene_type: "space_ext"
    mood_block: "climactic_confrontation"
    
  # Audio elements
  - id: "commander_dialogue"
    type: "sound"
    path: "/audio/commander_speech_final.wav"
    channel: 50  # Dialogue channel
    start: 240
    mood_block: "rising_danger"
    
  - id: "space_ambience"
    type: "sound"
    path: "/audio/space_environment.wav"
    channel: 52  # Ambient space
    start: 1
    mood_block: "opening_mystery"
    
  - id: "alien_vocals"
    type: "sound"
    path: "/audio/creature_vocalizations.wav"
    channel: 56  # Creature vocals
    start: 2160
    mood_block: "climactic_confrontation"
    
  - id: "main_theme"
    type: "sound"
    path: "/audio/main_score.wav"
    channel: 58  # Music score
    start: 1
    volume: 0.7
    volume_keyframes:
      - frame: 720
        value: 0.7
      - frame: 760
        value: 0.9  # Swell music for tension change
```

### Effects with VFX-Specific Properties

Effects now include sci-fi specific parameters:

```yaml
effects:
  - id: "space_grade"
    type: "color_balance"
    target: "establishing_shot"
    profile: "earth_reality"  # Reference predefined color profile
    
  - id: "alien_world_look"
    type: "color_balance"
    target: "alien_planet_bg"
    profile: "alien_world"
    
  - id: "hologram_glow"
    type: "glow"
    target: "hologram_overlay"
    threshold: 0.3
    size: 15
    intensity: 2.5
    color: [0.2, 0.8, 1.0]  # Sci-fi blue glow
    
  - id: "weapon_impact"
    type: "blur_motion"
    target: "energy_beam"
    length: 0.5
    direction: "beam_direction"  # Could be calculated from VFX metadata
    
  - id: "chromatic_aberration"
    type: "lens_distortion"
    target: "character_over_greenscreen"
    distortion: 0.05
    dispersion: 0.02  # RGB separation amount
```

### Transitions Aligned with Mood Changes

Transitions can now be aligned with mood changes:

```yaml
transitions:
  - id: "mystery_to_danger"
    type: "cross"
    source_mood: "opening_mystery"
    target_mood: "rising_danger"
    duration: 48  # 2 seconds
    
  - id: "danger_to_confrontation"
    type: "wipe"
    source_mood: "rising_danger"
    target_mood: "climactic_confrontation"
    duration: 36
    wipe_type: "iris"
    
  # Individual strip transitions
  - id: "holo_transition"
    type: "add_dissolve"  # Special additive dissolve for holograms
    source: "hologram_overlay"
    target: "energy_beam"
    duration: 24
```

### Composite Nodes

For complex VFX shots, define specific composite relationships:

```yaml
composites:
  - id: "character_in_space"
    output_channel: 5
    elements:
      - id: "space_bg"
        strip: "alien_planet_bg"
        blend_mode: "normal"
        order: 1
      
      - id: "character"
        strip: "character_over_greenscreen"
        blend_mode: "alpha_over"
        order: 2
        mask: "character_key"  # References a matte or alpha channel
        
      - id: "hologram"
        strip: "hologram_overlay"
        blend_mode: "add"
        order: 3
```

### Tension Timing Markers

Mark key narrative points for tension and release:

```yaml
tension_markers:
  - name: "Discovery of Anomaly"
    frame: 480
    tension_change: +2  # Increase in tension
    
  - name: "First Contact"
    frame: 1440
    tension_change: +3
    
  - name: "Weapon Systems Failure"
    frame: 2880
    tension_change: +1
    
  - name: "Resolution"
    frame: 3840
    tension_change: -4  # Release of tension
```

## Complete Example: Space Battle Sequence

Here's a full example of a space battle sequence from a sci-fi film:

```yaml
project:
  name: "Quantum Horizon - Battle Sequence"
  resolution: [3840, 2160]
  fps: 24
  color_science: "ACES"

mood_blocks:
  - id: "calm_before_storm"
    name: "Calm Before the Storm"
    tension: 4
    atmosphere: "suspense"
    color_palette: "cool_blue"
    start_frame: 1
    end_frame: 480
    
  - id: "enemy_approach"
    name: "Enemy Ship Approach"
    tension: 6
    atmosphere: "tension"
    color_palette: "red_alert"
    start_frame: 481
    end_frame: 960
    
  - id: "battle_erupts"
    name: "Battle Erupts"
    tension: 8
    atmosphere: "action"
    color_palette: "high_contrast"
    start_frame: 961
    end_frame: 1920
    
  - id: "near_defeat"
    name: "Near Defeat"
    tension: 9
    atmosphere: "dread"
    color_palette: "dark_contrast"
    start_frame: 1921
    end_frame: 2400
    
  - id: "final_victory"
    name: "Final Victory"
    tension: 7
    atmosphere: "triumph"
    color_palette: "golden_victory"
    start_frame: 2401
    end_frame: 2880

vfx_categories:
  - id: "main_footage"
    name: "Main Footage"
    channel_range: [1, 5]
    
  - id: "space_bg"
    name: "Space Backgrounds"
    channel_range: [6, 10]
    
  - id: "ship_models"
    name: "Ship 3D Models"
    channel_range: [11, 15]
    
  - id: "weapons_fx"
    name: "Weapons Effects"
    channel_range: [16, 20]
    
  - id: "damage_fx"
    name: "Damage Effects"
    channel_range: [21, 25]
    
  - id: "ui_elements"
    name: "Interface Elements"
    channel_range: [26, 30]

strips:
  # Main Footage
  - id: "bridge_crew"
    type: "movie"
    path: "/footage/bridge_crew_reaction.exr"
    channel: 1
    start: 1
    scene_type: "ship_int"
    mood_block: "calm_before_storm"
    
  - id: "captain_closeup"
    type: "movie"
    path: "/footage/captain_closeup.exr"
    channel: 1
    start: 240
    scene_type: "ship_int"
    mood_block: "calm_before_storm"
    
  # Space Backgrounds
  - id: "space_nebula"
    type: "movie"
    path: "/vfx/space_nebula_bg.exr"
    channel: 6
    start: 1
    duration: 2880
    scene_type: "space_ext"
    
  - id: "approaching_fleet"
    type: "movie"
    path: "/vfx/enemy_fleet_approach.exr"
    channel: 7
    start: 481
    scene_type: "space_ext"
    mood_block: "enemy_approach"
    
  # Ship Models
  - id: "hero_ship_ext"
    type: "movie"
    path: "/vfx/hero_ship_model.exr"
    channel: 11
    start: 720
    scene_type: "space_ext"
    mood_block: "enemy_approach"
    
  - id: "enemy_flagship"
    type: "movie"
    path: "/vfx/enemy_flagship.exr"
    channel: 12
    start: 961
    scene_type: "space_ext"
    mood_block: "battle_erupts"
    
  # Weapons Effects
  - id: "hero_torpedoes"
    type: "movie"
    path: "/vfx/hero_torpedoes.exr"
    channel: 16
    start: 1200
    duration: 120
    scene_type: "space_ext"
    mood_block: "battle_erupts"
    blend_type: "add"
    
  - id: "enemy_beam"
    type: "movie"
    path: "/vfx/enemy_laser_beam.exr"
    channel: 17
    start: 1440
    duration: 48
    scene_type: "space_ext"
    mood_block: "battle_erupts"
    blend_type: "screen"
    
  # Damage Effects  
  - id: "hull_breach"
    type: "movie"
    path: "/vfx/hull_breach_explosion.exr"
    channel: 21
    start: 1968
    duration: 72
    scene_type: "ship_int"
    mood_block: "near_defeat"
    blend_type: "screen"
    
  # UI Elements
  - id: "tactical_display"
    type: "movie"
    path: "/vfx/bridge_tactical.exr"
    channel: 26
    start: 120
    duration: 360
    scene_type: "holo_interface"
    mood_block: "calm_before_storm"
    blend_type: "add"
    opacity: 0.7
    
  - id: "alert_status"
    type: "movie"
    path: "/vfx/red_alert_ui.exr"
    channel: 27
    start: 500
    duration: 2380
    scene_type: "holo_interface"
    mood_block: "enemy_approach"
    blend_type: "screen"
    opacity: 0.5
    
  # Audio Tracks  
  - id: "captain_orders"
    type: "sound"
    path: "/audio/captain_battle_orders.wav"
    channel: 50
    start: 500
    mood_block: "enemy_approach"
    
  - id: "engine_rumble"
    type: "sound"
    path: "/audio/engine_background.wav"
    channel: 51
    start: 1
    duration: 2880
    volume: 0.3
    
  - id: "torpedo_launch"
    type: "sound"
    path: "/audio/torpedo_launch.wav"
    channel: 53
    start: 1200
    mood_block: "battle_erupts"
    
  - id: "enemy_weapon"
    type: "sound"
    path: "/audio/enemy_beam_weapon.wav"
    channel: 53
    start: 1440
    mood_block: "battle_erupts"
    
  - id: "hull_breach_alarm"
    type: "sound"
    path: "/audio/hull_breach_alarm.wav"
    channel: 54
    start: 1968
    mood_block: "near_defeat"
    
  - id: "battle_score"
    type: "sound"
    path: "/audio/battle_theme.wav"
    channel: 58
    start: 961
    mood_block: "battle_erupts"
    volume_keyframes:
      - frame: 961
        value: 0.5
      - frame: 1100
        value: 0.8
      - frame: 2600
        value: 0.9
      - frame: 2880
        value: 0.4

effects:
  - id: "space_look"
    type: "color_balance"
    target: "space_nebula"
    lift: [0.97, 0.98, 1.05]
    gamma: [0.95, 0.98, 1.1]
    gain: [0.9, 0.95, 1.2]
    
  - id: "red_alert_tint"
    type: "color_balance"
    target: "bridge_crew"
    mood_block: "enemy_approach"
    lift: [1.05, 0.95, 0.95]
    gamma: [1.1, 0.9, 0.9]
    
  - id: "torpedo_glow"
    type: "glow"
    target: "hero_torpedoes"
    threshold: 0.5
    size: 10
    intensity: 2.0
    color: [0.2, 0.8, 1.0]  # Blue photon torpedo glow
    
  - id: "enemy_weapon_distortion"
    type: "lens_distortion"
    target: "enemy_beam"
    distortion: 0.03
    dispersion: 0.01
    
  - id: "battle_camera_shake"
    type: "transform"
    target: "bridge_crew"
    mood_block: "battle_erupts"
    keyframes:
      - frame: 961
        position: [0, 0]
      - frame: 966
        position: [5, -7]
      - frame: 971
        position: [-3, 4]
      - frame: 976
        position: [6, 2]

transitions:
  - id: "suspense_to_tension"
    type: "cross"
    source_mood: "calm_before_storm"
    target_mood: "enemy_approach"
    duration: 24
    
  - id: "tension_to_battle"
    type: "custom_warp"
    source_mood: "enemy_approach"
    target_mood: "battle_erupts"
    duration: 12
    distortion: 0.3  # Special distortion parameter for custom transition
    
  - id: "bridge_to_space"
    type: "cross"
    source: "bridge_crew"
    target: "hero_ship_ext"
    duration: 36

composites:
  - id: "space_battle_composite"
    output_channel: 3
    elements:
      - id: "background"
        strip: "space_nebula"
        blend_mode: "normal"
        order: 1
        
      - id: "enemy_ships"
        strip: "approaching_fleet"
        blend_mode: "normal"
        order: 2
        
      - id: "hero_ship"
        strip: "hero_ship_ext"
        blend_mode: "normal"
        order: 3
        
      - id: "weapon_fx"
        strip: "hero_torpedoes"
        blend_mode: "add"
        order: 4

tension_markers:
  - name: "Enemy Detected"
    frame: 481
    tension_change: +2
    
  - name: "Battle Begins"
    frame: 961
    tension_change: +2
    
  - name: "First Hit Taken"
    frame: 1440
    tension_change: +1
    
  - name: "Critical Damage"
    frame: 1968
    tension_change: +1
    
  - name: "Turning Point"
    frame: 2401
    tension_change: -2
```

## Implementation Notes

When implementing a parser for this sci-fi editing DSL:

1. First parse the mood blocks and VFX categories to establish the narrative structure
2. Create the channels according to VFX categories
3. Create the base strips on each channel
4. Apply effects with appropriate mood-based adjustments
5. Create composite relationships between elements
6. Process transitions between scenes and mood blocks
7. Apply color profiles based on scene types and mood blocks

This workflow allows sci-fi editors to focus on storytelling through tension arcs while maintaining a clear organization of complex VFX elements. The DSL provides:

- A way to visualize the emotional flow of a sequence
- Clear organization of VFX assets by type and purpose
- Connection between technical elements and narrative function
- Color grading that aligns with storytelling intent

By grouping strips by mood and tension rather than just chronology, editors can more easily craft sci-fi narratives where the emotional journey is as important as the visual spectacle. 