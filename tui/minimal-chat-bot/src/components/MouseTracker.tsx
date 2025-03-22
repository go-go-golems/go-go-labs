import React from 'react';
import { Text, Box } from 'ink';
import { useMousePosition } from '@zenobius/ink-mouse';
import { getTheme } from '../utils/theme.js';

export function MouseTracker(): JSX.Element {
  const mousePosition = useMousePosition();
  const theme = getTheme();
  
  return (
    <Box marginTop={1}>
      <Text color={theme.dimText}>
        Mouse: ({mousePosition.x}, {mousePosition.y})
      </Text>
    </Box>
  );
} 