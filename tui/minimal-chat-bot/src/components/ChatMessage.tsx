import React from 'react';
import { Box, Text } from 'ink';
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

  return (
    <Box 
      flexDirection="row" 
      marginY={1}
    >
      {showIndicator && (
        <Box marginRight={1} width={2}>
          <Text color={isUser ? theme.accent : theme.primary}>
            {isUser ? '>' : '•'}
          </Text>
        </Box>
      )}
      <Box flexGrow={1}>
        <Text wrap="wrap" color={isUser ? theme.text : undefined}>
          {applyMarkdown(message.content)}
        </Text>
      </Box>
    </Box>
  );
} 