# Film Development Timer - Mobile App

A React Native mobile app built with Expo for timing film development processes, specifically designed for Cinestill Cs41 chemistry.

## Features

- **Offline Timer**: Works completely offline with SQLite database
- **Background Notifications**: Get notified when steps complete, even if app is backgrounded
- **17 Pre-configured Rolls**: Each with specific development times based on chemistry fatigue
- **5 Development Steps**: Pre-soak, Developer, Bleach-fix, Wash, Stabilizer
- **iOS Native Feel**: Dark theme optimized for darkroom use
- **Haptic Feedback**: Feel the timer alerts on iPhone
- **Keep Awake**: Screen stays on during development

## Development Setup

### Prerequisites

- Node.js 18+
- iOS device for testing (notifications don't work in simulator)
- Expo Go app installed on your iPhone

### Installation

```bash
# Install dependencies
npm install

# Start development server
npx expo start

# Scan QR code with iPhone Camera app to load in Expo Go
```

### Testing on Device

1. Start the development server: `npx expo start`
2. Scan the QR code with your iPhone's Camera app
3. The app will open in Expo Go
4. Test timer functionality and notifications

### Building for Distribution

```bash
# Configure EAS Build
npx eas build:configure

# Build for iOS
npx eas build --platform ios --profile production

# Submit to TestFlight
npx eas submit --platform ios --latest
```

## Project Structure

```
src/
├── components/
│   ├── TimerDisplay.tsx      # Timer countdown component
│   └── ProcessStepCard.tsx   # Individual step cards
├── screens/
│   ├── HomeScreen.tsx        # Welcome screen with process overview
│   └── TimerScreen.tsx       # Main timer interface
├── services/
│   ├── database.ts           # SQLite operations
│   └── notifications.ts     # Local notification scheduling
├── data/
│   ├── rolls.json           # Pre-configured roll data
│   └── steps.json           # Development step definitions
└── types/
    └── index.ts             # TypeScript type definitions
```

## Usage

1. **Start the App**: Launch from home screen
2. **Begin Development**: Tap "Start Timer" to begin
3. **Navigate Rolls**: Use prev/next buttons to select your roll number
4. **Run Steps**: Tap any step to select it, then use timer controls
5. **Background Alerts**: Receive notifications even when app is closed
6. **Complete Process**: Follow replenishment instructions after final step

## Key Technologies

- **Expo SDK 53**: React Native framework with managed workflow
- **SQLite**: Local persistent storage for offline functionality
- **Expo Notifications**: Background local notifications
- **React Navigation**: Native iOS navigation patterns
- **Expo Haptics**: iPhone haptic feedback
- **Expo Keep Awake**: Prevent screen lock during timing

## Development Notes

- Uses TypeScript for type safety
- Dark theme optimized for darkroom environment
- Responsive design for different iPhone screen sizes
- Graceful error handling for missing audio assets
- Persistent timer state across app restarts

## Testing Checklist

- [ ] Timer countdown works correctly
- [ ] Notifications fire when app is backgrounded
- [ ] Database persists data across app restarts
- [ ] Navigation works smoothly
- [ ] Haptic feedback triggers appropriately
- [ ] Screen stays awake during timing
- [ ] All 17 rolls have correct development times
- [ ] Step completion tracking works

## Deployment

The app can be deployed to TestFlight and App Store using EAS Build and Submit. Make sure to:

1. Configure proper bundle identifier in app.json
2. Set up Apple Developer account
3. Test thoroughly in TestFlight before App Store submission
4. Include proper screenshots and metadata for App Store listing