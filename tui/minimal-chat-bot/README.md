# Minimal TUI Chatbot

A minimal terminal-based LLM chatbot built with Ink and TypeScript. This chatbot provides a beautiful, interactive terminal user interface for conversing with an LLM (Large Language Model).

## Features

- Beautiful terminal UI built with React Ink
- Interactive chat interface with user and assistant messages
- Full mouse support for clicking, hovering, and scrolling
- Scrollable message history with navigation controls
- Real-time loading indication with animated spinner
- Markdown rendering for rich text formatting
- Error handling and display
- Full keyboard navigation and input handling

## Prerequisites

- Node.js (v16+)
- An Anthropic API key (environment variable: `ANTHROPIC_API_KEY`)
- Terminal emulator with mouse support (most modern terminals like iTerm2, Windows Terminal, etc.)

## Installation

```bash
# Clone the repository
git clone <repository-url>
cd tui/minimal-chat-bot

# Install dependencies
npm install
```

## Environment Setup

Create a `.env` file in the project root and add your Anthropic API key:

```
ANTHROPIC_API_KEY=your_api_key_here
```

## Usage

There are multiple ways to run the chatbot:

### Development Mode

For development with automatic recompilation:

```bash
# Run directly with ts-node (development)
npm start
```

### Build and Run

Build the TypeScript code first, then run the compiled JavaScript:

```bash
# Build the project
npm run build

# Run the compiled JavaScript
npm run start:js
```

### Watch Mode

For continuous development with automatic rebuilding:

```bash
# Start watch mode
npm run dev

# In another terminal
npm run start:js
```

### As a Global Command

You can install the chatbot globally:

```bash
# Install globally
npm install -g .

# Run from anywhere
minimal-chat-bot
```

### Controls

#### Mouse Controls
- Click in the input field to position the cursor
- Click the "Send" button to send a message
- Click the "Clear" button to clear input
- Use scroll buttons to navigate through message history
- Hover over messages to see timestamps and highlight messages

#### Keyboard Controls
- Type your message and press `Enter` to send
- Press `Escape` to clear the current input
- Use `Left` and `Right` arrow keys to navigate the cursor within the input field
- Use `Backspace` to delete characters

## Architecture

The chatbot is built with a component-based architecture:

- **ChatMessage**: Renders individual chat messages with styling and hover effects
- **PromptInput**: Handles user input with cursor positioning via keyboard and mouse
- **Button**: Clickable buttons with hover effects
- **ScrollableBox**: Provides scrollable content areas with controls
- **MouseTracker**: Displays current mouse coordinates
- **Spinner**: Shows an animated loading indicator
- **useChat hook**: Manages chat state and API communication

## Manual JavaScript Setup

If you want to modify or extend the chatbot, you might need to understand how Ink works with JavaScript. Here's how to set up Ink with JavaScript manually:

### Setting up Babel

Ink requires the same Babel setup as regular React-based apps in the browser.

1. Set up Babel with a React preset:

```bash
npm install --save-dev @babel/preset-react
```

2. Create a `babel.config.json` file:

```json
{
  "presets": [
    "@babel/preset-env",
    "@babel/preset-react",
    "@babel/preset-typescript"
  ]
}
```

### Creating and Transpiling a Source File

1. Create a file `source.js` with your Ink code:

```jsx
import React from 'react';
import {render, Text} from 'ink';

const Demo = () => <Text>Hello World</Text>;

render(<Demo />);
```

2. Transpile this file with Babel:

```bash
npx babel source.js -o cli.js
```

3. Run the transpiled file:

```bash
node cli
```

### Development Workflow

If you don't like transpiling files during development, you can use `import-jsx` or `@esbuild-kit/esm-loader` to import a JSX file and transpile it on the fly.

## Understanding Ink's Layout System

Ink uses Yoga - a Flexbox layout engine to build user interfaces for CLIs using familiar CSS-like props. Each element in Ink is a Flexbox container (think of it as if each `<div>` in the browser had `display: flex`).

Key points to remember:

- All text must be wrapped in a `<Text>` component
- Use `<Box>` components for layout with flexbox properties
- Ink supports most CSS flexbox properties like:
  - `flexDirection`
  - `alignItems`
  - `justifyContent`
  - `flexGrow`
  - `flexShrink`
  - `flexBasis`
  - `width`/`height`
  - `padding`/`margin`

## Mouse Support

Mouse support is implemented using the `@zenobius/ink-mouse` library, which provides several hooks:

- `useMousePosition`: Tracks the current mouse coordinates
- `useOnMouseClick`: Detects click events on a component
- `useOnMouseHover`: Detects hover events on a component
- `useElementPosition`: Gets the position of a component

Note that mouse support requires a terminal emulator that supports mouse events. Most modern terminal emulators (iTerm2, Windows Terminal, modern versions of xterm) support mouse events, but some configuration may be required.

## Customization

You can easily customize the chatbot by modifying the theme in `src/utils/theme.ts`.

## Dependencies

- React and Ink for terminal UI rendering
- @zenobius/ink-mouse for mouse interaction support
- Anthropic SDK for API communication
- UUID for generating message IDs
- Chalk for terminal styling
- Meow for CLI argument parsing

## License

ISC

## Credits

This project was inspired by the implementation in Claude Code, following the tutorial in [claude-code/ttmp/2025-03-21/01-chatbot-repl-tutorial.md](https://github.com/your-repo/claude-code/ttmp/2025-03-21/01-chatbot-repl-tutorial.md). 