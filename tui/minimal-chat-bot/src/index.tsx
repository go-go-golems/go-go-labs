#!/usr/bin/env node
import React, { FC } from 'react';
import { Box, render, Text, Static } from 'ink';
import { MouseProvider } from '@zenobius/ink-mouse';
import meow from 'meow';
import { ChatMessage } from './components/ChatMessage.js';
import { PromptInput } from './components/PromptInput.js';
import { Spinner } from './components/Spinner.js';
import { MouseTracker } from './components/MouseTracker.js';
import { ScrollableBox } from './components/ScrollableBox.js';
import { useChat } from './hooks/useChat.js';
import { getTheme } from './utils/theme.js';

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

const App: FC = () => {
  // Get chat functionality from our hook
  const { messages, isLoading, error, sendMessage } = useChat();
  const theme = getTheme();

  return (
    <Box flexDirection="column" padding={1}>
      {/* Header */}
      <Box marginBottom={1}>
        <Text bold color={theme.primary}>
          Minimal TUI Chatbot (with Mouse Support)
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
      
      {/* Message history with scrolling */}
      <Box marginY={1} height={10}>
        <ScrollableBox height={10}>
          {messages.map((message) => (
            <ChatMessage 
              key={message.id} 
              message={message} 
            />
          ))}
        </ScrollableBox>
      </Box>
      
      {/* Loading indicator */}
      {isLoading && <Spinner />}
      
      {/* Input area */}
      <PromptInput 
        onSubmit={sendMessage} 
        isLoading={isLoading}
      />
    </Box>
  );
};

// Wrap the App with MouseProvider and render
render(
  <MouseProvider>
    <App />
  </MouseProvider>
); 