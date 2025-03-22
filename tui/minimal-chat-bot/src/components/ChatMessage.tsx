import React, { FC } from 'react';
import { Box, Text } from 'ink';
import type { ChatMessage as ChatMessageType } from '../store/chatSlice.ts';

interface Props {
  message: ChatMessageType;
}

export const ChatMessage: FC<Props> = ({ message }) => {
  const { content, role } = message;
  
  return (
    <Box marginY={1}>
      <Text color={role === 'user' ? 'green' : 'blue'}>
        {role === 'user' ? '🧑 You: ' : '🤖 Bot: '}
        {content}
      </Text>
    </Box>
  );
}; 