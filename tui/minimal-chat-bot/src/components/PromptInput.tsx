import React, { useState } from 'react';
import { Box, Text, useInput } from 'ink';
import { getTheme } from '../utils/theme.js';

type Props = {
  onSubmit: (input: string) => void;
  isLoading: boolean;
  placeholder?: string;
};

export function PromptInput({ onSubmit, isLoading, placeholder = 'Type a message...' }: Props): JSX.Element {
  const [input, setInput] = useState('');
  const [cursorPosition, setCursorPosition] = useState(0);
  const theme = getTheme();

  // Handle keyboard input
  useInput((inputChar, key) => {
    // Disable input during loading
    if (isLoading) return;

    // Submit on Enter
    if (key.return) {
      if (input.trim()) {
        onSubmit(input);
        setInput('');
        setCursorPosition(0);
      }
      return;
    }

    // Clear input on Escape
    if (key.escape) {
      setInput('');
      setCursorPosition(0);
      return;
    }

    // Handle backspace/delete
    if (key.backspace || key.delete) {
      if (cursorPosition > 0) {
        setInput(
          input.slice(0, cursorPosition - 1) + input.slice(cursorPosition)
        );
        setCursorPosition(cursorPosition - 1);
      }
      return;
    }

    // Cursor movement with arrow keys
    if (key.leftArrow && cursorPosition > 0) {
      setCursorPosition(cursorPosition - 1);
      return;
    }

    if (key.rightArrow && cursorPosition < input.length) {
      setCursorPosition(cursorPosition + 1);
      return;
    }

    // Add typed characters at cursor position
    if (inputChar && !key.ctrl && !key.meta) {
      setInput(
        input.slice(0, cursorPosition) + inputChar + input.slice(cursorPosition)
      );
      setCursorPosition(cursorPosition + 1);
    }
  });

  const placeholderActive = !input;

  // Split input into sections for cursor rendering
  const beforeCursor = input.slice(0, cursorPosition);
  const atCursor = input[cursorPosition] || ' ';
  const afterCursor = input.slice(cursorPosition + 1);

  return (
    <Box 
      borderStyle="round" 
      borderColor={theme.border} 
      padding={0}
      marginTop={1}
    >
      <Box marginLeft={1}>
        <Text color={theme.primary}>{'>'}</Text>
      </Box>
      <Box marginLeft={1} marginRight={1} flexGrow={1}>
        {placeholderActive ? (
          <Text dimColor>{placeholder}</Text>
        ) : (
          <Text>
            <Text>{beforeCursor}</Text>
            <Text inverse>{atCursor}</Text>
            <Text>{afterCursor}</Text>
          </Text>
        )}
      </Box>
    </Box>
  );
} 