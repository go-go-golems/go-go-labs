import React, { useRef, useState } from 'react';
import { Box, Text } from 'ink';
import { useOnMouseHover } from '@zenobius/ink-mouse';
import { Message } from '../types/index.js';
import { getTheme } from '../utils/theme.js';
import { applyMarkdown } from '../utils/markdown.js';

type Props = {
  message: Message;
  showIndicator?: boolean;
};

export function ChatMessage({ message, showIndicator = true }: Props): JSX.Element {
  const theme = getTheme();
  const isUser = message.role === 'user';
  const [isHovering, setIsHovering] = useState(false);
  const messageRef = useRef(null);
  
  // Add hover effect
  useOnMouseHover(messageRef, setIsHovering);
  
  // Get the timestamp formatted
  const timestamp = message.timestamp.toLocaleTimeString([], { 
    hour: '2-digit', 
    minute: '2-digit'
  });

  return (
    <Box 
      ref={messageRef}
      flexDirection="row" 
      marginY={1}
      // borderStyle={isHovering ? "round" : "single"}
      // borderColor={isHovering ? (isUser ? theme.accent : theme.primary) : "transparent"}
      padding={isHovering ? 1 : 0}
    >
      {showIndicator && (
        <Box marginRight={1} width={2}>
          <Text color={isUser ? theme.accent : theme.primary}>
            {isUser ? '>' : '•'}
          </Text>
        </Box>
      )}
      <Box flexGrow={1} flexDirection="column">
        <Text wrap="wrap" color={isUser ? theme.text : undefined}>
          {applyMarkdown(message.content)}
        </Text>
        
        {isHovering && (
          <Box marginTop={1}>
            <Text color={theme.dimText} italic>
              {timestamp}
            </Text>
          </Box>
        )}
      </Box>
    </Box>
  );
} 