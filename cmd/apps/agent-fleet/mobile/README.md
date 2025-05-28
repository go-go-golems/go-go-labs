# Agent Fleet Mobile App

A mobile-first application for monitoring and controlling a fleet of coding agents.

## Features

- **Fleet View**: Monitor all agents with real-time status indicators
- **Updates Feed**: Recent activity and events from your agent fleet  
- **Task Management**: Create and assign tasks to agents
- **Agent Detail**: Detailed view with command interface, todos, and event history
- **Real-time Updates**: Live status updates via Server-Sent Events

## Tech Stack

- **Framework**: Expo React Native
- **State Management**: Redux Toolkit + RTK Query
- **Navigation**: React Navigation
- **Language**: TypeScript
- **UI**: Custom components with emoji icons

## Getting Started

1. Install dependencies:
   ```bash
   npm install
   ```

2. Start the development server:
   ```bash
   npm start
   ```

3. Run on your device:
   - iOS: `npm run ios`
   - Android: `npm run android`  
   - Web: `npm run web`

## Project Structure

```
src/
├── components/          # Reusable UI components
│   ├── AgentCard.tsx   # Agent status card
│   ├── UpdateItem.tsx  # Event update item
│   └── TaskItem.tsx    # Task queue item
├── navigation/         # Navigation configuration
├── screens/            # Main app screens
│   ├── FleetScreen.tsx
│   ├── UpdatesScreen.tsx
│   ├── TasksScreen.tsx
│   └── AgentDetailScreen.tsx
├── services/           # API layer
│   └── api.ts         # RTK Query API definitions
├── store/             # Redux store
│   ├── index.ts
│   └── slices/
└── types/             # TypeScript type definitions
```

## API Configuration

The app expects the Agent Fleet API to be running at `https://api.agentfleet.dev/v1/`.

To use a different API endpoint, update the `baseUrl` in `src/services/api.ts`.

For development, you may need to update the authentication token in the API service.

## Key Components

### AgentCard
Displays agent status with visual indicators for feedback needs, progress bars, and metadata.

### Agent Detail Screen  
Full-featured agent interface with:
- Command/feedback input
- Todo list management
- Event history
- Debug logs (collapsible)

### Real-time Updates
Uses Server-Sent Events for live updates of agent status, progress, and new events.

## Design System

- **Colors**: Dark theme with status-based color coding
- **Typography**: System fonts with emoji icons
- **Layout**: Mobile-first responsive design
- **Interactions**: Touch-optimized with haptic feedback

## Development Notes

- Uses React Navigation 6 for navigation
- RTK Query handles API state management and caching
- Dark theme optimized for mobile viewing
- Portrait-first orientation
- Supports pull-to-refresh patterns
