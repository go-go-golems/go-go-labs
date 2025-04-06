# Scenario YAML DSL Specification v1.0

This document specifies the YAML format used to structure the narrative and technical details for the "Debugging After Dark" scenario, intended for visualization or further processing.

## Root Object

The root object must contain a single key: `show`.

```yaml
show:
  # ... show details ...
```

## `show` Object

Contains the overall information about the episode/scenario.

- `title` (String, Required): The main title of the show/episode.
- `logline` (String, Required): A brief, engaging summary of the episode's plot.
- `characters` (List, Required): A list of character objects involved.
- `arc` (List, Required): A list of act objects defining the narrative structure.
- `key_concepts` (List, Required): A list of key technical concept objects explained.

```yaml
show:
  title: "Debugging After Dark: The Case of the Lagging Listener"
  logline: "In the neon-drenched alleys..."
  characters: [ ... ]
  arc: [ ... ]
  key_concepts: [ ... ]
```

## `character` Object

Describes a character in the show.

- `name` (String, Required): The name of the character (can include qualifiers like "(Voice)").
- `description` (String, Required): A description of the character's role and personality. Can be multi-line using `|`.
- `role` (String, Optional): Specific role, especially used for suspects.
- `list` (List, Optional): Used for grouping related characters/entities (like suspects). Contains nested `character` objects (only `name` and `role` needed for nested suspects).

```yaml
characters:
  - name: "Detective Trace"
    description: "Our protagonist..."
  - name: "The Client (Voice)"
    description: "Represents the user/stakeholder..."
  - name: "The Suspects (Personified Code/Libraries)"
    description: |
      List of potential culprits...
    list:
      - name: "Bubble Tea"
        role: "The framework..."
      - name: "Glamour"
        role: "The flashy renderer..."
      # ... more suspects
```

## `act` Object

Represents a major section of the narrative arc.

- `act` (Integer, Required): The act number (e.g., 1, 2, 3).
- `title` (String, Required): The title of the act.
- `scenes` (List, Required): A list of scene objects within this act.

```yaml
arc:
  - act: 1
    title: "The Setup - \"It Just... Hangs.\""
    scenes: [ ... ]
  # ... more acts
```

## `scene` Object

Represents a specific scene within an act.

- `scene_number` (String, Required): A unique identifier for the scene (e.g., "1.1", "2.3").
- `title` (String, Required): The title of the scene.
- `summary` (String, Required): A brief description of the scene's purpose or events.
- `shots` (List, Required): A list of shot objects that make up the scene.

```yaml
scenes:
  - scene_number: "1.1"
    title: "The Office (Night)"
    summary: "Introduction to Detective Trace..."
    shots: [ ... ]
  # ... more scenes
```

## `shot` Object

Represents a single camera shot or moment within a scene, detailing the visual and audio elements.

- `shot_number` (String, Required): A unique identifier for the shot within the scene (e.g., "1.1.1", "2.3.2"). Can include comments linking to original storyboard shots (`# Corresponds to...`).
- `visual` (String, Required): Description of what is seen on screen. Can include references to files (` `), UI elements, code snippets, or graphics. Multi-line allowed using `|`.
- `audio` (String, Required): Description of the audio elements, including dialogue (e.g., `**Trace (VO):**`), sound effects (`(Sound of keyboard typing)`), or music cues. Multi-line allowed using `|`.

```yaml
shots:
  - shot_number: "1.1.1"
    visual: "Rain streaks down a digital window pane..."
    audio: |
      **Client:** "Trace? I've got a problem..."
      **Trace (VO):** "Another one..."
  # ... more shots
```

## `key_concept` Object

Describes a key technical concept introduced or explained during the show.

- `name` (String, Required): The name of the concept (e.g., "Bubble Tea", "Logging").
- `details` (String, Required): A brief explanation of the concept and its relevance.

```yaml
key_concepts:
  - name: "Bubble Tea"
    details: "MVU (Model-View-Update) architecture..."
  - name: "Logging"
    details: "Importance of timestamps..."
  # ... more concepts
```

## Data Types

- **String:** Standard YAML string. Use `|` for multi-line strings where formatting is important.
- **Integer:** Standard YAML integer.
- **List:** Standard YAML list/sequence.
- **Object:** Standard YAML map/dictionary. 