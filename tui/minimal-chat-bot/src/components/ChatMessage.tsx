import React, { FC } from 'react';
import { Box, Text } from 'ink';
import { getTheme } from '../utils/theme.js';

interface Message {
  id: string;
  role: 'user' | 'assistant';
  content: string;
}

interface Props {
  message: Message;
}

export const ChatMessage: FC<Props> = ({ message }) => {
  const theme = getTheme();
  const isUser = message.role === 'user';

  return (
    <Box 
      flexDirection="column"
      marginY={1}
    >
      <Text 
        color={isUser ? theme.accent : theme.primary}
        bold
      >
        {isUser ? 'You' : 'Assistant'}:
      </Text>
      <Box marginLeft={1}>
        <Text wrap="wrap">
          {message.content}
        </Text>
      </Box>
    </Box>
  );
}; 