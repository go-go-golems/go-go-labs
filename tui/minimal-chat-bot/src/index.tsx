#!/usr/bin/env node
import React, { FC, useState, useEffect } from 'react';
import { Box, render, Text, useStdout } from 'ink';
import { MouseProvider } from '@zenobius/ink-mouse';
import meow from 'meow';
import { ChatMessage } from './components/ChatMessage.js';
import { PromptInput } from './components/PromptInput.js';
import { Spinner } from './components/Spinner.js';
import { MouseTracker } from './components/MouseTracker.js';
import { ScrollArea } from './components/ScrollArea.js';
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
  // Get chat functionality from our hook
  const { messages, isLoading, error, sendMessage } = useChat();
  const theme = getTheme();
  const size = useTerminalSize();

  // Calculate content height (total height minus padding and fixed elements)
  const contentHeight = size.height - 6; // Adjust for header, input, padding, etc.

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
        <ScrollArea height={contentHeight - 8}>
          {messages.map((message) => (
            <ChatMessage 
              key={message.id} 
              message={message} 
            />
          ))}
        </ScrollArea>
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

// Wrap the App with MouseProvider and render with stdin in raw mode to properly handle mouse events
render(
  <MouseProvider>
    <App />
  </MouseProvider>,
  {
    // Set stdin to raw mode to better handle mouse events
    stdin: process.stdin,
    stdout: process.stdout
  }
); 