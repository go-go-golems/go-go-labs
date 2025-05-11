# Live Coding Stream Information Display
## System Specification Document

### 1. Overview

The Live Coding Stream Information Display is a web application designed for content creators who stream live coding sessions. It provides viewers with essential context about the current stream, progress tracking, and upcoming tasks in a visually distinctive interface inspired by the aesthetic of the TV show "Severance."

### 2. Purpose

This application solves several problems for live coding streamers:
- Provides context for viewers who join mid-stream
- Tracks progress through planned tasks systematically
- Maintains a visual history of completed work
- Shows upcoming tasks to set expectations
- Offers easy updating of information during active streams
- Creates a distinctive, branded look for the stream overlay

### 3. User Interface

#### 3.1 Visual Design
- **Primary Aesthetic**: Minimalist corporate design inspired by "Severance"
- **Color Palette**: Primarily black and white with strategic green accents
- **Typography**: Monospace font family throughout
- **Layout**: Grid-based, compartmentalized sections with rigid borders
- **Branding**: "Lumon Industries" header (customizable)

#### 3.2 Primary Sections

1. **Header Section**
   - Stream title (customizable)
   - Duration timer (automatically updates)
   - Edit controls

2. **Stream Information Panel**
   - Project designation (title)
   - Project description
   - Programming language/framework
   - GitHub repository link
   - Current viewer count
   - Quick topic entry field

3. **Task Management Panel**
   - Active task display (with completion button)
   - Completed tasks list (with reactivation options)
   - Upcoming tasks list (with promotion options)
   - New task entry field

### 4. Features & Functionality

#### 4.1 Core Features

1. **Stream Information Management**
   - Editable stream title and description
   - Editable programming language/framework
   - Editable GitHub repository link
   - Editable viewer count (manual update)
   - Automatic session duration timer

2. **Task Tracking System**
   - Current active task display
   - Task completion tracking
   - Task history with timestamps
   - Upcoming task queue
   - Task reordering capabilities

3. **Quick Actions**
   - Set new topic instantly
   - Add new tasks to queue
   - Complete current task
   - Reactivate completed tasks
   - Promote upcoming tasks to active

#### 4.2 Interactive Elements

1. **Edit Mode Toggle**
   - Button to enter/exit edit mode
   - Save/cancel buttons in edit mode
   - Form fields for all editable content

2. **Task Management Controls**
   - "Complete" button for active task
   - "Reactivate" button for completed tasks
   - "Make Active" button for upcoming tasks
   - "Add" button for new task entries
   - "Set Topic" button for quick topic changes

3. **Timer Controls**
   - Reset timer button (in edit mode)

### 5. State Management

#### 5.1 Primary State Objects

1. **Redux State** (defined in `streamSlice.ts`)
   ```typescript
   interface StreamState {
     info: StreamInfo;            // Stream metadata
     completedSteps: string[];    // Completed tasks
     activeStep: string;          // Current active task
     upcomingSteps: string[];     // Upcoming tasks
     isEditing: boolean;          // Edit mode toggle
     isLoggedIn: boolean;         // Authentication state (for view-only mode)
   }
   ```

2. **Stream Information State** (defined in `StreamInfo` interface)
   ```typescript
   interface StreamInfo {
     title: string;
     description: string;
     startTime: string;           // ISO date string
     language: string;
     githubRepo: string;
     currentTask: string;
     viewerCount: number;
   }
   ```

3. **Local Component State** (in `StreamInfoDisplay.tsx`)
   - `editableInfo`: Copy of stream information for editing
   - `newStep`: string (for task input)
   - `newTopic`: string (for quick topic input)
   - `duration`: string (formatted elapsed time)

#### 5.2 State Transitions and Redux Actions

1. **Editing Workflow**
   - `toggleEditMode()`: Toggle between view and edit modes
   - `setStreamInfo(info)`: Update stream information after editing
   - Local state manages temporary edits before committing

2. **Authentication State**
   - `toggleLoggedIn()`: Toggle between authenticated and guest views
   - Automatically disables editing when logged out

3. **Task Progression Workflow**
   - `addUpcomingStep(step)`: Add task to upcoming queue
   - `setNewActiveTopic(topic)`: Set a new active task
   - `completeCurrentStep()`: Mark current task as complete (moves to completed, pulls next from upcoming)
   - `makeStepActive({step, source})`: Promote a task to active status

4. **Timer Management**
   - Automatic updates every second via useEffect hook
   - `resetTimer()`: Reset the stream duration

### 6. Technical Implementation

#### 6.1 Technology Stack
- **Framework**: React with TypeScript
- **Styling**: TailwindCSS
- **Icons**: Lucide React
- **State Management**: Redux Toolkit (RTK)
- **Build Tool**: Vite with Bun

#### 6.2 Component Structure
- **Project Structure**:
  ```
  ui/
  ├── src/
  │   ├── components/
  │   │   └── StreamInfoDisplay.tsx     # Main component
  │   ├── store/
  │   │   ├── hooks.ts                  # Redux hooks
  │   │   ├── store.ts                  # Redux store configuration
  │   │   └── slices/
  │   │       └── streamSlice.ts        # Stream state and actions
  │   ├── App.tsx                       # Root application component
  │   ├── main.tsx                      # Entry point
  │   └── index.css                     # Global styles (Tailwind)
  ├── index.html
  ├── package.json
  └── tailwind.config.js
  ```

- **Components**:
  - StreamInfoDisplay (main component)
    - Header section
    - Stream Information Panel
    - Task Management Panel
      - Active Task Component
      - Completed Tasks List
      - Upcoming Tasks List

#### 6.3 Key Functions and Handlers

1. **Redux Action Creators** (in `streamSlice.ts`)
   - `setStreamInfo`: Updates stream metadata
   - `toggleEditMode`: Toggles editing state
   - `toggleLoggedIn`: Toggles between authenticated and guest views
   - `resetTimer`: Resets stream timer
   - `addUpcomingStep`: Adds a task to upcoming queue
   - `setNewActiveTopic`: Sets a new active task
   - `completeCurrentStep`: Completes current task
   - `makeStepActive`: Activates a task from completed or upcoming lists

2. **Component Event Handlers** (in `StreamInfoDisplay.tsx`)
   - `handleInputChange()`: Updates editable form fields
   - `saveChanges()`: Saves edited information to Redux
   - `cancelChanges()`: Discards edits
   - `handleResetTimer()`: Resets stream timer
   - `handleAddNewStep()`: Adds task to upcoming queue
   - `handleSetNewActiveTopic()`: Sets new active task
   - `handleCompleteCurrentStep()`: Completes current task
   - `handleMakeStepActive()`: Activates a task from a list

3. **Time Management**
   - `updateDuration()`: Calculates and formats elapsed time (in useEffect hook)

### 7. Customization Options

#### 7.1 Branding
- Replace "Lumon Industries" with custom stream branding
- Modify color scheme while maintaining aesthetic

#### 7.2 Feature Expansions
- Integration with streaming platforms (Twitch, YouTube) for automatic viewer count
- GitHub integration for automatic repository information
- Custom Severance-inspired animations or transitions
- Light/dark mode toggle with appropriate Severance aesthetics
- Additional information fields for stream-specific data
- Backend API integration for persistent storage

### 8. Future Enhancements

#### 8.1 Potential Additions
- Local storage persistence between sessions
- Export/import configuration
- Multiple project profiles
- Keyboard shortcuts for quick updates
- Automatic task timing statistics
- Stream milestones/achievements tracking
- Chat command integration
- Viewer question queue
- Real authentication with backend (currently simulated with toggleLoggedIn)
- Task categories and filtering
- Time estimates for tasks

#### 8.2 Integration Possibilities
- OBS Browser Source compatibility testing
- Streamlabs overlay integration
- Discord webhook notifications for task completion
- Twitter/social media integration for milestone sharing

### 9. Accessibility Considerations
- Maintain minimum contrast ratios despite stylistic choices
- Ensure all interactive elements are keyboard accessible
- Provide appropriate ARIA labels for custom UI elements
- Test with screen readers for information hierarchy
- Ensure font sizes remain readable at various display sizes

### 10. Implementation Status and Deployment

#### 10.1 Current Implementation Status
- **Completed Features**:
  - Stream information display and editing
  - Task management (add, complete, reactivate)
  - Duration timer
  - Severance-inspired UI
  - Authentication toggle (simulated with isLoggedIn state)
  - View-only mode for non-authenticated users
  - Responsive design

#### 10.2 Deployment Requirements
- Static web hosting capability (GitHub Pages, Vercel, Netlify, etc.)
- Modern browser support
- No backend requirements (client-side only)