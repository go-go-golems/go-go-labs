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

// Handle CLI with meow
const cli = meow(
  `
  Usage
    $ minimal-chat-bot [options] [initial-prompt]

  Options
    --help             Show this help message
    --version          Show version
    --initial-prompt   Set an initial prompt for the chatbot
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
      initialPrompt: {
        type: 'string',
        alias: 'i',
      },
    },
  }
);

// Log application start
logger.info('Application starting', { 
  version: process.env.npm_package_version,
  nodeVersion: process.version,
  initialPrompt: cli.flags.initialPrompt || cli.input[0] || null
});

// // Add initial messages to the chat
// store.dispatch(addMessage({
//   id: uuidv4(),
//   role: 'assistant',
//   content: 'Hello! I\'m your minimal terminal chatbot. How can I help you today? We are going to say something very long: "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum."',
// }));

// store.dispatch(addMessage({
//   id: uuidv4(),
//   role: 'assistant',
//   content: 'Type a message below and press Enter to chat. Use arrow keys to scroll.',
// }));

// store.dispatch(addMessage({
//   id: uuidv4(),
//   role: 'assistant', 
//   content: 'Here is another very long message to test wrapping and scrolling: "Sed ut perspiciatis unde omnis iste natus error sit voluptatem accusantium doloremque laudantium, totam rem aperiam, eaque ipsa quae ab illo inventore veritatis et quasi architecto beatae vitae dicta sunt explicabo. Nemo enim ipsam voluptatem quia voluptas sit aspernatur aut odit aut fugit, sed quia consequuntur magni dolores eos qui ratione voluptatem sequi nesciunt. Neque porro quisquam est, qui dolorem ipsum quia dolor sit amet, consectetur, adipisci velit, sed quia non numquam eius modi tempora incidunt ut labore et dolore magnam aliquam quaerat voluptatem."'
// }));

// store.dispatch(addMessage({
//   id: uuidv4(),
//   role: 'user',
//   content: 'Thanks for all the information!'
// }));

store.dispatch(addMessage({
  id: uuidv4(),
  role: 'user',
  content: `
  <uiInstructions>
  You are responsible for rendering a consistent ASCII UI.  Clicks will be sent with an X at the click position.
  For example, clicking on [UPDATE] will be signaled as [UPDXTE]. You should not render the X in the UI.
  
  UI Controls:
  ( ) / (x) = Checkboxes (click to toggle)
  [ Button ] = Clickable buttons
  [v| Value ] = Dropdown menu (click v to expand, which will cause you to render the dropdown menu in the UI).
  Only expand the dropdown menu if the user clicks on the v. Close the dropdown menu once the user has made a selection.

  Render the full UI after each user interaction. Keep it consistent, don't drop or add elements unless asked to.

  ALWAYS RENDER THE FULL UI AFTER EACH USER INTERACTION.
  </uiInstructions>
  `
}));

logger.info('Added initial messages to the chat');

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
  const [inputValue, setInputValue] = useState('');

  useEffect(() => {
    const initialPrompt = cli.flags.initialPrompt || cli.input[0] || 'Make the UI for a crazy spaceship. Multiple panels, weird widgets. Show action log terminal. Show incoming alien radar.';
    sendMessage(initialPrompt);
  }, []);


  const onlyShowLastAssistantMessage = true
  const messagesToShow = onlyShowLastAssistantMessage ? messages.filter((message) => message.role === 'assistant').slice(-1) : messages;

  // Wrap sendMessage to log message sending
  const handleSendMessage = useCallback((message: string) => {
    logger.info('Sending message', { message });
    sendMessage(message);
  }, [sendMessage]);

  // Handle input changes
  const handleInputChange = useCallback((value: string) => {
    setInputValue(value);
  }, [setInputValue]);

  // Handle message clicks
  const handleMessageClick = useCallback((event: ChatMessageClickEvent) => {
    logger.info('Message clicked in app', {
      clickedLine: event.clickedLine,
      timestamp: new Date().toISOString()
    });

    // Create a string with an X at the clicked position
    const line = event.clickedLine.content;
    const pos = event.textAtPosition.charPosition;
    const markedLine = line.substring(0, pos) + 'X' + line.substring(pos + 1);
    
    // Send the message with the marked line and current input context
    const clickContext = `\n\n<click>\noriginal line: "${line}"\nline index: ${event.clickedLine.index}\nclicked line: "${markedLine}"\n</click>`;
    const messageWithContext = inputValue ? `${inputValue}${clickContext}` : clickContext;
    handleSendMessage(messageWithContext);
  }, [inputValue, handleSendMessage]);

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
      {/* <MouseTracker scrollAreaRef={scrollAreaRef} /> */}

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
          {messagesToShow.map((message) => (
            <ChatMessage 
              key={message.id} 
              message={message} 
              onClick={handleMessageClick}
              isLoading={isLoading}
              showPrefix={!onlyShowLastAssistantMessage}
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
          onChange={handleInputChange}
          value={inputValue}
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