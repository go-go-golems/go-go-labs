import React, { useState, useEffect } from 'react';
import { Text, Box } from 'ink';
import { getTheme } from '../utils/theme.js';

// Unicode braille patterns that create a spinning animation
const frames = ['⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'];

type Props = {
  label?: string;
};

export function Spinner({ label = 'Thinking' }: Props): JSX.Element {
  const [frame, setFrame] = useState(0);
  const theme = getTheme();

  useEffect(() => {
    // Set up an interval to cycle through animation frames
    const timer = setInterval(() => {
      setFrame(previousFrame => (previousFrame + 1) % frames.length);
    }, 80);

    // Clean up the interval when component unmounts
    return () => {
      clearInterval(timer);
    };
  }, []);

  return (
    <Box>
      <Text color={theme.primary}>{frames[frame]} </Text>
      <Text dimColor>{label}...</Text>
    </Box>
  );
} 