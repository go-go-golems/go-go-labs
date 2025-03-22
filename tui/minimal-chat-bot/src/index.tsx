#!/usr/bin/env node
import React, { FC, useState, useEffect } from 'react';
import { Box, render, Text, useStdout, useInput } from 'ink';
import { MouseProvider } from '@zenobius/ink-mouse';
import { Provider } from 'react-redux/alternate-renderers';
import meow from 'meow';
import { ChatMessage } from './components/ChatMessage.js';
import { PromptInput } from './components/PromptInput.js';
import { Spinner } from './components/Spinner.js';
import { MouseTracker } from './components/MouseTracker.js';
import { ScrollableBox } from './components/ScrollableBox.js';
import { useChat } from './hooks/useChat.js';
import { getTheme } from './utils/theme.js';
import { store } from './store/store.js';
import { useAppDispatch, useAppSelector } from './store/hooks.js';
import { ScrollState, setOffset } from './store/scrollSlice.js';

// Handle CLI with meow
const cli = meow(
  `
  Usage
    $ minimal-chat-bot

  Options
    --help    Show this help message
    --version Show version
`,
  {
    importMeta: import.meta,
    flags: {
      help: {
        type: 'boolean',
        alias: 'h',
      },
      version: {
        type: 'boolean',
        alias: 'v',
      },
    },
  }
);

// Hook to get and track terminal size
function useTerminalSize() {
  const out = useStdout();
  const [size, setSize] = useState(() => ({
    width: out.stdout.columns,
    height: out.stdout.rows,
  }));

  useEffect(() => {
    const handleTerminalResize = () => {
      setSize({
        width: out.stdout.columns,
        height: out.stdout.rows,
      });
    };

    process.stdout.on('resize', handleTerminalResize);
    process.stdout.on('SIGWINCH', handleTerminalResize);
    return () => {
      process.stdout.off('SIGWINCH', handleTerminalResize);
      process.stdout.off('resize', handleTerminalResize);
    };
  }, []);

  return size;
}

const App: FC = () => {
  const { messages, isLoading, error, sendMessage } = useChat();
  const size = useTerminalSize();
  const theme = getTheme();
  const dispatch = useAppDispatch();
  const scrollState = useAppSelector((state: {scroll: ScrollState}) => state.scroll);

  // Calculate available height for messages
  const messageAreaHeight = Math.max(3, size.height - 10); // Adjust for header, input, status, margins

  // Handle keyboard input for scrolling
  useInput((input, key) => {
    if (key.upArrow) {
      dispatch(setOffset(Math.max(0, scrollState.offset - 1)));
    }
    if (key.downArrow) {
      dispatch(setOffset(Math.min(
        messages.length - messageAreaHeight,
        scrollState.offset + 1
      )));
    }
    if (key.pageUp) {
      dispatch(setOffset(Math.max(0, scrollState.offset - messageAreaHeight)));
    }
    if (key.pageDown) {
      dispatch(setOffset(Math.min(
        messages.length - messageAreaHeight,
        scrollState.offset + messageAreaHeight
      )));
    }
  });

  return (
    <Box 
      flexDirection="column" 
      padding={1}
      width={size.width - 2}  // Account for padding
      height={size.height}
      marginX={1}
    >
      {/* Header */}
      <Box marginBottom={1}>
        <Text bold color={theme.primary}>
          Minimal TUI Chatbot
        </Text>
      </Box>

      {/* Mouse position display */}
      <MouseTracker />

      {/* Error display */}
      {error && (
        <Box marginBottom={1}>
          <Text color={theme.error}>Error: {error}</Text>
        </Box>
      )}

      {/* Message history */}
      <Box 
        borderStyle="single"
        borderColor={theme.primary}
        padding={1}
      >
        <ScrollableBox height={messageAreaHeight}>
          {messages.map((message) => (
            <ChatMessage key={message.id} message={message} />
          ))}
        </ScrollableBox>
      </Box>

      {/* Loading indicator */}
      {isLoading && (
        <Box marginTop={1}>
          <Spinner label="Thinking..." />
        </Box>
      )}

      {/* Input area */}
      <Box marginTop={1}>
        <PromptInput
          onSubmit={sendMessage}
          isLoading={isLoading}
          placeholder="Type a message..."
        />
      </Box>
    </Box>
  );
};

render(
  <Provider store={store}>
    <MouseProvider>
      <App />
    </MouseProvider>
  </Provider>,
  {
    stdin: process.stdin,
    stdout: process.stdout
  }
); 