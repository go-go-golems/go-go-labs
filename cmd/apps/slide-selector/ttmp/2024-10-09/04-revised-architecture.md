> https://chatgpt.com/c/6706b500-6910-8012-98e3-7f67373e570a

## Slide Editor / Selector Software Specification

### Overview

The Slide Editor / Selector is a Terminal User Interface (TUI) application built using the [Charmbracelet](https://github.com/charmbracelet) suite of libraries, specifically leveraging [Bubbletea](https://github.com/charmbracelet/bubbletea) for managing the application state and [Glamour](https://github.com/charmbracelet/glamour) for rendering Markdown. The application facilitates the selection, filtering, and management of AI-generated slides written in Markdown format. Users can efficiently curate their slide decks by flagging important slides, deleting unwanted ones, and saving their selections into a new Markdown file.

### Key Features

- **Load Slides from Markdown:** Import slides separated by `---` dividers in a Markdown file.
- **View Modes:**
  - **Multi-Slide View:** Display up to six slides simultaneously in a grid layout.
  - **Single-Slide View:** Focus on a single slide for detailed inspection.
- **Slide Management:**
  - **Flagging:** Mark slides as interesting for easy identification.
  - **Deletion:** Remove slides that are not needed.
  - **Filtering:** Toggle between viewing all slides or only flagged slides.
- **Navigation:**
  - Use arrow keys (left, right, up, down) to navigate between slides.
  - Page Up/Page Down for rapid navigation.
- **Actions:**
  - **Delete:** Remove selected slides.
  - **Flag:** Mark/unmark slides as interesting.
  - **Toggle Views:** Switch between multi-slide and single-slide views.
  - **Toggle Filtering:** Show only flagged slides or all slides.
  - **Undo/Redo:** Revert or reapply recent actions.
- **File Saving:**
  - Save selected slides into a new Markdown file using a file picker interface.
- **Responsive UI:**
  - Components can be resized to fit different terminal window sizes.
  - Scrollable views for content that exceeds the display area.

### User Interface Design

#### Layout

1. **Header:**
   - **Title:** Displays the application name, e.g., "AI Slide Editor".
   - **Mode Indicator:** Shows the current view mode ("Multi" or "Single").

2. **Main Content Area:**
   - **Multi-Slide View:**
     - Displays up to six slides in a 3x2 grid.
     - Each slide preview includes the title and a truncated version of the content.
     - The focused slide is highlighted.
   - **Single-Slide View:**
     - Displays the full content of the selected slide.
     - Includes visual indicators for flagged slides.

3. **Footer:**
   - **Instructions:** Key bindings and navigation hints.
   - **Message Area:** Displays feedback messages (e.g., "Slide flagged", "Slide deleted").

#### Interaction

- **Navigation:**
  - **Arrow Keys:** Move focus between slides.
  - **Page Up/Page Down:** Jump between sections of slides.
- **Actions:**
  - **f:** Flag/unflag the focused slide.
  - **d:** Delete the focused slide.
  - **Tab:** Toggle between Multi-Slide and Single-Slide views.
  - **s:** Toggle between showing all slides or only flagged slides.
  - **Ctrl-s:** Open the file picker to save selected slides.
  - **Undo/Redo:** Revert or redo recent actions.

### User Workflow

1. **Loading Slides:**
   - Launch the application with a command like `./slide-editor slides.md`.
   - Slides are loaded and displayed in Multi-Slide View.

2. **Navigating Slides:**
   - Use arrow keys to move focus between slides.
   - Switch to Single-Slide View for detailed content inspection.

3. **Managing Slides:**
   - Flag slides that are interesting by pressing `f`.
   - Delete unwanted slides by pressing `d`.
   - Toggle to view only flagged slides by pressing `s`.

4. **Saving Selections:**
   - Press `Ctrl-s` to open the file picker.
   - Choose the destination and filename to save the selected slides.

5. **Undo/Redo Actions:**
   - Use designated key bindings to undo or redo recent changes.

---

## Application Architecture Using Charmbracelet

### Overview

The Slide Editor application is structured using the Model-View-Update (MVU) pattern provided by Bubbletea. The architecture is modular, with distinct components handling different parts of the UI and functionality. The key components include the main application model, the Markdown slide view, the file picker for saving files, and auxiliary components for message display.

### Components

1. **Main Application (`App`):**
   - **Responsibilities:**
     - Manage the overall application state.
     - Coordinate interactions between different components.
     - Handle global key bindings and actions.
   - **Structure:**
     ```go
     type App struct {
         State        AppState
         SlideView    SlideViewComponent
         FilePicker   filepicker.Model
         Message      string
         Mode         ViewMode // Enum: Multi, Single, FileSelector
         PreviousMode ViewMode // To store the mode before opening FilePicker
     }
     ```

2. **Slide View Component (`SlideViewComponent`):**
   - **Responsibilities:**
     - Render slides in either Multi-Slide or Single-Slide view.
     - Handle resizing and scrolling of slide content.
   - **Structure:**
     ```go
     type SlideViewComponent struct {
         Slides        []Slide
         CurrentIndex  int
         MultiView     bool
         ScrollOffset  int
         Dimensions    Dimensions // Width and Height for resizing
         GlamourRenderer glamour.TermRenderer
     }
     ```

3. **File Picker Component (`filepicker.Model`):**
   - **Responsibilities:**
     - Provide an interface for users to select the destination filename and path when saving slides.
     - Handle user interactions within the file picker.
   - **Structure:**
     - Utilizes the provided `filepicker.go` API from Charmbracelet Bubbles.

4. **Message Display:**
   - **Responsibilities:**
     - Show feedback messages to the user (e.g., "Slide flagged").
   - **Structure:**
     ```go
     type MessageDisplay struct {
         Message string
     }
     ```

### Data Structures

1. **Slide:**
   ```go
   type Slide struct {
       Content   string // Full Markdown content
       Title     string // Extracted from the first header
       Flagged   bool
       Deleted   bool
   }
   ```

2. **AppState:**
   ```go
   type AppState struct {
       Slides          []Slide
       CurrentIndex    int
       ShowFlaggedOnly bool
       MultiView       bool
       History         []AppState // For undo/redo functionality
   }
   ```

3. **ViewMode:**
   ```go
   type ViewMode int

   const (
       MultiView ViewMode = iota
       SingleView
       FileSelector
   )
   ```

4. **Dimensions:**
   ```go
   type Dimensions struct {
       Width  int
       Height int
   }
   ```

### Component Interactions

1. **Main Application (`App`):**
   - Initializes all components.
   - Routes messages and commands between components.
   - Maintains the central `AppState` which is shared across components.
   - Handles global actions like saving files, toggling views, and filtering slides.

2. **Slide View Component:**
   - Receives the current list of slides based on filtering.
   - Updates its display based on the current view mode (Multi or Single).
   - Handles user inputs for navigating and interacting with slides within its scope.
   - Communicates slide actions (flagging, deletion) back to the main `App`.

3. **File Picker Component:**
   - Activated when the user initiates a save action (`Ctrl-s`).
   - Uses the provided `filepicker.go` API to render the file selection UI.
   - Once a file is selected, it sends a message to the main `App` to perform the save operation.
   - Can be canceled to revert to the previous view without saving.

4. **Message Display:**
   - Receives messages from various components to display feedback.
   - Ensures messages are visible without obstructing the main content.

### Markdown Slide View Component Design

The Markdown Slide View is encapsulated as its own component within the application, providing flexibility in rendering and interaction.

#### Responsibilities

- **Rendering Slides:**
  - Uses Glamour to render Markdown content into styled terminal output.
  - Supports both Multi-Slide and Single-Slide views.
  
- **Resizing:**
  - Adjusts the layout based on terminal window size changes.
  - Dynamically reallocates space between slides in Multi-Slide View.
  
- **Scrolling:**
  - Enables vertical scrolling within a slide if the content exceeds the available display area.
  - Maintains scroll position when navigating between slides.

#### Structure

```go
type SlideViewComponent struct {
    Slides          []Slide
    CurrentIndex    int
    MultiView       bool
    ScrollOffset    int
    Dimensions      Dimensions
    GlamourRenderer glamour.TermRenderer
}
```

#### Functionality

1. **Rendering:**
   - In **Multi-Slide View**, renders up to six slides in a grid (e.g., 3x2).
   - In **Single-Slide View**, renders the full content of the focused slide.
   - Applies styles to indicate flagged or deleted slides.

2. **Resizing:**
   - Listens for terminal resize events.
   - Adjusts the `Dimensions` accordingly.
   - Recalculates the grid layout for Multi-Slide View.

3. **Scrolling:**
   - Tracks `ScrollOffset` to determine the visible portion of the slide content.
   - Provides key bindings (e.g., Page Up/Page Down) to scroll through content.

4. **Interactivity:**
   - Highlights the currently focused slide.
   - Updates focus based on user navigation inputs.
   - Communicates slide actions (flagging, deletion) back to the main `App`.

### File Picker Integration

Utilizing the provided `filepicker.go` API from Charmbracelet Bubbles, the File Picker component is integrated to handle file saving operations.

#### Integration Steps

1. **Initialization:**
   - Instantiate the `filepicker.Model` using `filepicker.New()`.
   - Configure styles and key bindings as needed using `filepicker.DefaultStyles()` and `filepicker.DefaultKeyMap()`.

2. **Activation:**
   - Triggered when the user presses `Ctrl-s`.
   - Switch the application's `Mode` to `FileSelector`.
   - Render the File Picker overlay on top of the current UI.

3. **Interaction:**
   - Users navigate the file system using the key bindings defined in `KeyMap`.
   - Users can select a destination directory and specify a filename.

4. **Completion:**
   - Upon selection, the File Picker sends the chosen path back to the main `App`.
   - The application then saves the filtered slides to the specified file.
   - Switch the `Mode` back to the previous view (Multi or Single).

5. **Cancellation:**
   - Users can cancel the save operation, reverting to the previous view without saving.

#### Example Integration Code Snippet

```go
func (m *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch m.Mode {
    case FileSelector:
        var cmd tea.Cmd
        m.FilePicker, cmd = m.FilePicker.Update(msg)
        if path, ok := m.FilePicker.DidSelectFile(msg); ok {
            // Perform save operation
            saveCmd := m.saveSlides(path)
            // Reset mode
            m.Mode = m.PreviousMode
            return m, saveCmd
        }
        return m, cmd
    default:
        // Handle other modes
    }
}
```

### State Management

The application's state is centralized in the `AppState` structure, ensuring consistency across components.

```go
type AppState struct {
    Slides          []Slide
    CurrentIndex    int
    ShowFlaggedOnly bool
    MultiView       bool
    History         []AppState // For undo/redo functionality
}
```

- **Slides:** Stores all loaded slides with their current status (flagged, deleted).
- **CurrentIndex:** Tracks the currently focused slide.
- **ShowFlaggedOnly:** Determines whether to display all slides or only flagged ones.
- **MultiView:** Indicates the current view mode.
- **History:** Maintains a stack of previous states for undo/redo functionality.

### Message Passing

Components communicate through messages, adhering to the Bubbletea architecture.

- **User Actions:** Generate messages (e.g., `FlagSlideMsg`, `DeleteSlideMsg`) that are handled by the `Update` function.
- **Component Updates:** Components emit messages when their internal state changes (e.g., File Picker completion).

### Styling and Theming

Utilize [Lip Gloss](https://github.com/charmbracelet/lipgloss) for consistent styling across components.

- **Slide Highlighting:** Use distinct colors or borders to indicate focus and flagged status.
- **Deleted Slides:** Greyed out or strikethrough text to signify deletion.
- **File Picker:** Styled according to `filepicker.Styles` to ensure a cohesive look.

### Error Handling

- **File Operations:** Handle errors during file saving (e.g., permission issues) and provide user feedback.
- **Input Validation:** Ensure valid filenames and paths are provided in the File Picker.
- **State Consistency:** Prevent actions that could lead to inconsistent states (e.g., deleting a slide that's already deleted).

### Extensibility

The architecture allows for easy addition of new features, such as:

- **Export Formats:** Support exporting slides in formats other than Markdown.
- **Collaborative Editing:** Integrate with version control systems for team collaboration.
- **Advanced Filtering:** Provide more granular filtering options based on slide metadata.

### Summary

The Slide Editor / Selector application leverages the modularity and flexibility of the Charmbracelet ecosystem to deliver a robust TUI for managing AI-generated slides. By compartmentalizing functionality into distinct components and maintaining a centralized state, the application ensures a seamless and responsive user experience. The integration of specialized components like the Markdown Slide View and the File Picker enhances usability, making the application both powerful and user-friendly.

---

## Revised Application Architecture Without Table of Contents and Title List

### Overview

Based on the user's feedback, the Slide Editor / Selector application will no longer include the Table of Contents or Title List features. This simplifies the architecture and UI, focusing solely on slide management and selection.

### Updated Components

1. **Main Application (`App`):**
   - **Responsibilities:**
     - Manage the overall application state.
     - Coordinate interactions between different components.
     - Handle global key bindings and actions.
   - **Structure:**
     ```go
     type App struct {
         State        AppState
         SlideView    SlideViewComponent
         FilePicker   filepicker.Model
         Message      string
         Mode         ViewMode // Enum: Multi, Single, FileSelector
         PreviousMode ViewMode // To store the mode before opening FilePicker
     }
     ```

2. **Slide View Component (`SlideViewComponent`):**
   - **Responsibilities:**
     - Render slides in either Multi-Slide or Single-Slide view.
     - Handle resizing and scrolling of slide content.
   - **Structure:**
     ```go
     type SlideViewComponent struct {
         Slides          []Slide
         CurrentIndex    int
         MultiView       bool
         ScrollOffset    int
         Dimensions      Dimensions // Width and Height for resizing
         GlamourRenderer glamour.TermRenderer
     }
     ```

3. **File Picker Component (`filepicker.Model`):**
   - **Responsibilities:**
     - Provide an interface for users to select the destination filename and path when saving slides.
     - Handle user interactions within the file picker.
   - **Structure:**
     - Utilizes the provided `filepicker.go` API from Charmbracelet Bubbles.

4. **Message Display:**
   - **Responsibilities:**
     - Show feedback messages to the user (e.g., "Slide flagged").
   - **Structure:**
     ```go
     type MessageDisplay struct {
         Message string
     }
     ```

### Updated Data Structures

1. **Slide:**
   ```go
   type Slide struct {
       Content   string // Full Markdown content
       Title     string // Extracted from the first header
       Flagged   bool
       Deleted   bool
   }
   ```

2. **AppState:**
   ```go
   type AppState struct {
       Slides          []Slide
       CurrentIndex    int
       ShowFlaggedOnly bool
       MultiView       bool
       History         []AppState // For undo/redo functionality
   }
   ```

3. **ViewMode:**
   ```go
   type ViewMode int

   const (
       MultiView ViewMode = iota
       SingleView
       FileSelector
   )
   ```

4. **Dimensions:**
   ```go
   type Dimensions struct {
       Width  int
       Height int
   }
   ```

### Updated Component Interactions

1. **Main Application (`App`):**
   - Initializes all components.
   - Routes messages and commands between components.
   - Maintains the central `AppState` which is shared across components.
   - Handles global actions like saving files, toggling views, and filtering slides.

2. **Slide View Component:**
   - Receives the current list of slides based on filtering.
   - Updates its display based on the current view mode (Multi or Single).
   - Handles user inputs for navigating and interacting with slides within its scope.
   - Communicates slide actions (flagging, deletion) back to the main `App`.

3. **File Picker Component:**
   - Activated when the user presses `Ctrl-s`.
   - Uses the provided `filepicker.go` API to render the file selection UI.
   - Once a file is selected, it sends a message to the main `App` to perform the save operation.
   - Can be canceled to revert to the previous view without saving.

4. **Message Display:**
   - Receives messages from various components to display feedback.
   - Ensures messages are visible without obstructing the main content.

### Updated Markdown Slide View Component Design

The Markdown Slide View remains largely the same but without the need to manage a table of contents or title list.

#### Responsibilities

- **Rendering Slides:**
  - Uses Glamour to render Markdown content into styled terminal output.
  - Supports both Multi-Slide and Single-Slide views.
  
- **Resizing:**
  - Adjusts the layout based on terminal window size changes.
  - Dynamically reallocates space between slides in Multi-Slide View.
  
- **Scrolling:**
  - Enables vertical scrolling within a slide if the content exceeds the available display area.
  - Maintains scroll position when navigating between slides.

#### Structure

```go
type SlideViewComponent struct {
    Slides          []Slide
    CurrentIndex    int
    MultiView       bool
    ScrollOffset    int
    Dimensions      Dimensions
    GlamourRenderer glamour.TermRenderer
}
```

#### Functionality

1. **Rendering:**
   - In **Multi-Slide View**, renders up to six slides in a grid (e.g., 3x2).
   - In **Single-Slide View**, renders the full content of the focused slide.
   - Applies styles to indicate flagged or deleted slides.

2. **Resizing:**
   - Listens for terminal resize events.
   - Adjusts the `Dimensions` accordingly.
   - Recalculates the grid layout for Multi-Slide View.

3. **Scrolling:**
   - Tracks `ScrollOffset` to determine the visible portion of the slide content.
   - Provides key bindings (e.g., Page Up/Page Down) to scroll through content.

4. **Interactivity:**
   - Highlights the currently focused slide.
   - Updates focus based on user navigation inputs.
   - Communicates slide actions (flagging, deletion) back to the main `App`.

### Updated File Picker Integration

The integration of the File Picker remains the same, facilitating the saving of selected slides without the need for a table of contents or title list.

#### Integration Steps

1. **Initialization:**
   - Instantiate the `filepicker.Model` using `filepicker.New()`.
   - Configure styles and key bindings as needed using `filepicker.DefaultStyles()` and `filepicker.DefaultKeyMap()`.

2. **Activation:**
   - Triggered when the user presses `Ctrl-s`.
   - Switch the application's `Mode` to `FileSelector`.
   - Render the File Picker overlay on top of the current UI.

3. **Interaction:**
   - Users navigate the file system using the key bindings defined in `KeyMap`.
   - Users can select a destination directory and specify a filename.

4. **Completion:**
   - Upon selection, the File Picker sends the chosen path back to the main `App`.
   - The application then saves the filtered slides to the specified file.
   - Switch the `Mode` back to the previous view (Multi or Single).

5. **Cancellation:**
   - Users can cancel the save operation, reverting to the previous view without saving.

#### Example Integration Code Snippet

```go
func (m *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch m.Mode {
    case FileSelector:
        var cmd tea.Cmd
        m.FilePicker, cmd = m.FilePicker.Update(msg)
        if path, ok := m.FilePicker.DidSelectFile(msg); ok {
            // Perform save operation
            saveCmd := m.saveSlides(path)
            // Reset mode
            m.Mode = m.PreviousMode
            return m, saveCmd
        }
        return m, cmd
    default:
        // Handle other modes
    }
}
```

### Updated State Management

The application's state remains centralized in the `AppState` structure, without the table of contents or title list.

```go
type AppState struct {
    Slides          []Slide
    CurrentIndex    int
    ShowFlaggedOnly bool
    MultiView       bool
    History         []AppState // For undo/redo functionality
}
```

- **Slides:** Stores all loaded slides with their current status (flagged, deleted).
- **CurrentIndex:** Tracks the currently focused slide.
- **ShowFlaggedOnly:** Determines whether to display all slides or only flagged ones.
- **MultiView:** Indicates the current view mode.
- **History:** Maintains a stack of previous states for undo/redo functionality.

### Updated Message Passing

Components continue to communicate through messages, adhering to the Bubbletea architecture.

- **User Actions:** Generate messages (e.g., `FlagSlideMsg`, `DeleteSlideMsg`) that are handled by the `Update` function.
- **Component Updates:** Components emit messages when their internal state changes (e.g., File Picker completion).

### Updated Styling and Theming

Utilize [Lip Gloss](https://github.com/charmbracelet/lipgloss) for consistent styling across components.

- **Slide Highlighting:** Use distinct colors or borders to indicate focus and flagged status.
- **Deleted Slides:** Greyed out or strikethrough text to signify deletion.
- **File Picker:** Styled according to `filepicker.Styles` to ensure a cohesive look.

### Updated Error Handling

- **File Operations:** Handle errors during file saving (e.g., permission issues) and provide user feedback.
- **Input Validation:** Ensure valid filenames and paths are provided in the File Picker.
- **State Consistency:** Prevent actions that could lead to inconsistent states (e.g., deleting a slide that's already deleted).

### Updated Extensibility

The simplified architecture allows for easy addition of new features, such as:

- **Export Formats:** Support exporting slides in formats other than Markdown.
- **Collaborative Editing:** Integrate with version control systems for team collaboration.
- **Advanced Filtering:** Provide more granular filtering options based on slide metadata.

### Summary

By removing the Table of Contents and Title List features, the Slide Editor / Selector application becomes more streamlined, focusing solely on the core functionalities of slide management and selection. The modular architecture using Charmbracelet ensures that the application remains maintainable and extensible, providing a responsive and user-friendly experience for managing AI-generated slides.