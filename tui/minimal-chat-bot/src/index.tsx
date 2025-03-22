#!/usr/bin/env node
import React, { FC, useState, useEffect, useRef, useCallback } from 'react';
import { Box, Text, useStdout, useInput, render } from 'ink';
import { render as testRender } from 'ink-testing-library';
import { MouseProvider } from '@zenobius/ink-mouse';
import { Provider } from 'react-redux/alternate-renderers';
import meow from 'meow';
import { ChatMessage } from './components/ChatMessage.js';
import type { ChatMessageClickEvent } from './components/ChatMessage.js';
import { PromptInput } from './components/PromptInput.js';
import { Spinner } from './components/Spinner.js';
import { MouseTracker } from './components/MouseTracker.js';
import { ScrollArea } from './components/ScrollArea.js';
import { useChat } from './hooks/useChat.js';
import { getTheme, createLogger } from './utils/index.js';
import { store } from './store/store.js';
import { useAppDispatch, useAppSelector } from './store/hooks.js';
import { ScrollState, setOffset } from './store/scrollSlice.js';
import { addMessage } from './store/chatSlice.js';
import { v4 as uuidv4 } from 'uuid';

// Create a logger for the main app
const logger = createLogger('App');

// Log application start
logger.info('Application starting', { 
  version: process.env.npm_package_version,
  nodeVersion: process.version 
});

// Add initial messages to the chat
store.dispatch(addMessage({
  id: uuidv4(),
  role: 'assistant',
  content: 'Hello! I\'m your minimal terminal chatbot. How can I help you today? We are going to say something very long: "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum."',
}));

store.dispatch(addMessage({
  id: uuidv4(),
  role: 'assistant',
  content: 'Type a message below and press Enter to chat. Use arrow keys to scroll.',
}));

store.dispatch(addMessage({
  id: uuidv4(),
  role: 'assistant', 
  content: 'Here is another very long message to test wrapping and scrolling: "Sed ut perspiciatis unde omnis iste natus error sit voluptatem accusantium doloremque laudantium, totam rem aperiam, eaque ipsa quae ab illo inventore veritatis et quasi architecto beatae vitae dicta sunt explicabo. Nemo enim ipsam voluptatem quia voluptas sit aspernatur aut odit aut fugit, sed quia consequuntur magni dolores eos qui ratione voluptatem sequi nesciunt. Neque porro quisquam est, qui dolorem ipsum quia dolor sit amet, consectetur, adipisci velit, sed quia non numquam eius modi tempora incidunt ut labore et dolore magnam aliquam quaerat voluptatem."'
}));

store.dispatch(addMessage({
  id: uuidv4(),
  role: 'user',
  content: 'Thanks for all the information!'
}));

store.dispatch(addMessage({
  id: uuidv4(),
  role: 'assistant',
  content: 'And one final long message to really test things out: "At vero eos et accusamus et iusto odio dignissimos ducimus qui blanditiis praesentium voluptatum deleniti atque corrupti quos dolores et quas molestias excepturi sint occaecati cupiditate non provident, similique sunt in culpa qui officia deserunt mollitia animi, id est laborum et dolorum fuga. Et harum quidem rerum facilis est et expedita distinctio. Nam libero tempore, cum soluta nobis est eligendi optio cumque nihil impedit quo minus id quod maxime placeat facere possimus, omnis voluptas assumenda est, omnis dolor repellendus."'
}));


logger.info('Added initial messages to the chat');

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
  const scrollAreaRef = useRef(null);

  // Handle message clicks
  const handleMessageClick = useCallback((event: ChatMessageClickEvent) => {
    // logger.info('Message clicked in app', {
    //   // clickEvent: event,
    //   timestamp: new Date().toISOString()
    // });
  }, []);

  // Log terminal size changes
  useEffect(() => {
    logger.debug('Terminal size updated', { width: size.width, height: size.height });
  }, [size.width, size.height]);

  // Log errors
  useEffect(() => {
    if (error) {
      logger.error('Chat error occurred', { error });
    }
  }, [error]);

  // Calculate available height for messages
  const messageAreaHeight = Math.max(3, size.height - 16); // Adjust for header, input, status, margins

  // Handle keyboard input for scrolling
  useInput((_, key) => {
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

  // Wrap sendMessage to log message sending
  const handleSendMessage = (message: string) => {
    logger.info('Sending message', { message });
    sendMessage(message);
  };

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
      <MouseTracker scrollAreaRef={scrollAreaRef} />

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
        paddingX={1}
      >
        <ScrollArea height={messageAreaHeight} ref={scrollAreaRef}>
          {messages.map((message) => (
            <ChatMessage 
              key={message.id} 
              message={message} 
              onClick={handleMessageClick}
            />
          ))}
        </ScrollArea>
      </Box>

        <Box marginTop={1}>
      {/* Loading indicator */}
      {isLoading ? (
          <Spinner label="Thinking..." />
      ) : (
          <Text>...</Text>
      )}
        </Box>

      {/* Input area */}
      <Box marginTop={1}>
        <PromptInput
          onSubmit={handleSendMessage}
          isLoading={isLoading}
          placeholder="Type a message..."
        />
      </Box>
    </Box>
  );
};

// Log when application is about to render
logger.info('Rendering application');

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

// Handle process termination
process.on('SIGINT', () => {
  logger.info('Application shutting down');
  process.exit(0);
});

process.on('uncaughtException', (error) => {
  logger.error('Uncaught exception', { error: error.toString(), stack: error.stack });
  process.exit(1);
}); 