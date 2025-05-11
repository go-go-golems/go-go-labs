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

1. **Stream Information State**
   - title: string
   - description: string
   - startTime: ISO date string
   - language: string
   - githubRepo: string
   - viewerCount: number

2. **Tasks State**
   - completedSteps: string[]
   - activeStep: string
   - upcomingSteps: string[]

3. **UI State**
   - isEditing: boolean
   - editableInfo: copy of stream information
   - newStep: string (for task input)
   - newTopic: string (for quick topic input)

#### 5.2 State Transitions

1. **Editing Workflow**
   - Toggle edit mode
   - Update editable copy of information
   - Save or cancel changes

2. **Task Progression Workflow**
   - Add task to upcoming queue
   - Set task as active
   - Complete active task (moves to completed, pulls next from upcoming)
   - Reactivate completed task

3. **Timer Management**
   - Automatic updates every second
   - Manual reset functionality

### 6. Technical Implementation

#### 6.1 Technology Stack
- **Framework**: React
- **Styling**: TailwindCSS
- **Icons**: Lucide React
- **State Management**: React Hooks (useState, useEffect)

#### 6.2 Component Structure
- StreamInfoDisplay (main component)
  - Header section
  - Stream Information Panel
  - Task Management Panel
    - Active Task Component
    - Completed Tasks List
    - Upcoming Tasks List

#### 6.3 Key Functions

1. **Time Management**
   - `updateDuration()`: Calculates and formats elapsed time
   - `resetTimer()`: Resets start time to current time

2. **Information Management**
   - `handleInputChange()`: Updates editable information
   - `saveChanges()`: Commits editable information to main state
   - `cancelChanges()`: Reverts editable information to main state

3. **Task Management**
   - `addNewStep()`: Adds task to upcoming queue
   - `setNewActiveTopic()`: Sets new active task
   - `completeCurrentStep()`: Marks active task as complete
   - `makeStepActive()`: Promotes task to active status

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

### 10. Deployment Requirements
- Static web hosting capability
- Modern browser support
- Responsive design for various display configurations
- No backend requirements (client-side only)