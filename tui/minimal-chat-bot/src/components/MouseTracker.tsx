import React, { useRef } from 'react';
import { Text, Box } from 'ink';
import { useMousePosition, useElementPosition } from '@zenobius/ink-mouse';
import { getTheme } from '../utils/theme.js';

export function MouseTracker(): JSX.Element {
  const mousePosition = useMousePosition();
  const ref = useRef(null);
  const elementPosition = useElementPosition(ref);
  const theme = getTheme();
  
  return (
    <Box ref={ref} marginTop={1}>
      <Text color={theme.dimText}>
        Mouse: ({mousePosition.x}, {mousePosition.y}) | Component: ({elementPosition.left}, {elementPosition.top})
      </Text>
    </Box>
  );
} 