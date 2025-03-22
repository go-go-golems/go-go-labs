import React, { useState, useRef } from 'react';
import { Box, Text, useInput } from 'ink';
import { useOnMouseClick, useElementPosition } from '@zenobius/ink-mouse';
import { getTheme } from '../utils/theme.js';
import { Button } from './Button.js';

type Props = {
  onSubmit: (input: string) => void;
  isLoading: boolean;
  placeholder?: string;
};

export function PromptInput({ onSubmit, isLoading, placeholder = 'Type a message...' }: Props): JSX.Element {
  const [input, setInput] = useState('');
  const [cursorPosition, setCursorPosition] = useState(0);
  const inputRef = useRef(null);
  const theme = getTheme();

  // Get element position to calculate cursor position from click
  const elementPosition = useElementPosition(inputRef);

  // Handle mouse clicks on the input field
  useOnMouseClick(inputRef, (event) => {
    if (event && !isLoading) {
      // Calculate cursor position based on click position
      // This is a simplified calculation - may need adjustment based on font size
      const clickX = event.x - elementPosition.x - 3; // Adjust for left margin and borders
      
      // Calculate approximately where the click happened in the string
      const textBeforeClick = input.slice(0, input.length);
      const estimatedPosition = Math.min(
        Math.max(0, Math.round(clickX)),
        input.length
      );
      
      setCursorPosition(estimatedPosition);
    }
  });

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

  // Handle clear button click
  const handleClear = () => {
    setInput('');
    setCursorPosition(0);
  };

  // Handle submit button click
  const handleSubmit = () => {
    if (input.trim() && !isLoading) {
      onSubmit(input);
      setInput('');
      setCursorPosition(0);
    }
  };

  return (
    <Box flexDirection="column" marginTop={1}>
      <Box 
        ref={inputRef}
        borderStyle="round" 
        borderColor={theme.border} 
        padding={0}
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
      
      {/* Button controls */}
      <Box marginTop={1} justifyContent="flex-end">
        <Button 
          label="Clear" 
          onClick={handleClear} 
          type="secondary" 
          disabled={isLoading || !input}
        />
        <Button 
          label="Send" 
          onClick={handleSubmit} 
          type="primary" 
          disabled={isLoading || !input.trim()}
        />
      </Box>
    </Box>
  );
} 