# Stream Task Overview

A React-based live coding stream information display with a distinctive Severance-inspired UI. This application helps content creators who stream live coding sessions provide context, track progress, and manage tasks in real-time.

![Severance-inspired UI](https://i.imgur.com/placeholder.png)

## Features

- **Stream Information Management**: Edit title, description, language, and repository details
- **Task Tracking System**: Manage active, completed, and upcoming tasks
- **Viewer-only Mode**: Toggle between editable mode (for the streamer) and view-only mode (for viewers)
- **Duration Timer**: Automatically tracks and displays stream duration
- **Severance-inspired UI**: Minimalist corporate design with monospace typography

## Tech Stack

- **Frontend**: React with TypeScript
- **State Management**: Redux Toolkit (RTK)
- **Styling**: TailwindCSS
- **Icons**: Lucide React
- **Build Tool**: Vite with Bun

## Getting Started

### Prerequisites

- [Bun](https://bun.sh/) (JavaScript runtime & package manager)

### Installation

```bash
# Install dependencies
cd ui
bun install

# Start development server
bun run dev
```

## Project Structure

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

## Redux State Management

The application uses Redux Toolkit for state management with the following state structure:

```typescript
interface StreamState {
  info: StreamInfo;            // Stream metadata
  completedSteps: string[];    // Completed tasks
  activeStep: string;          // Current active task
  upcomingSteps: string[];     // Upcoming tasks
  isEditing: boolean;          // Edit mode toggle
  isLoggedIn: boolean;         // Authentication state
}
```

### Redux Actions

- `setStreamInfo`: Update stream metadata
- `toggleEditMode`: Toggle between view and edit modes
- `toggleLoggedIn`: Toggle between authenticated and guest views
- `resetTimer`: Reset the stream duration timer
- `addUpcomingStep`: Add a new task to the upcoming queue
- `setNewActiveTopic`: Set a new active task
- `completeCurrentStep`: Mark the current task as complete
- `makeStepActive`: Promote a task to active status

## Usage

### Authentication Toggle

The app includes a simulated authentication system to demonstrate the different views:

- **Logged In**: Full editing capabilities for the streamer
- **Logged Out**: Read-only view for viewers

Use the Login/Logout button in the top-right corner to toggle between these modes.

### Managing Tasks

1. Add new tasks using the input field at the bottom of the task panel
2. Complete the current task using the "Complete" button
3. Reactivate completed tasks or promote upcoming tasks as needed
4. Set a new topic directly using the input in the information panel

## Customization

To customize the appearance:

1. Modify the TailwindCSS classes in `StreamInfoDisplay.tsx`
2. Update the initial state in `streamSlice.ts` to set default values
3. Replace "LUMON INDUSTRIES" with your own branding

## License

MIT