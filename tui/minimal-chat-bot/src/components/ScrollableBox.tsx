import React, { useRef, useState } from 'react';
import { Box, Text } from 'ink';
import { useOnMouseClick } from '@zenobius/ink-mouse';
import { getTheme } from '../utils/theme.js';
import { Button } from './Button.js';

type Props = {
  height: number;
  children: React.ReactNode;
};

export function ScrollableBox({ height, children }: Props): JSX.Element {
  const [scrollOffset, setScrollOffset] = useState(0);
  const theme = getTheme();
  
  // Convert children to array for easier handling
  const childrenArray = React.Children.toArray(children);
  
  // Calculate maximum scroll offset
  const maxScroll = Math.max(0, childrenArray.length - height);
  
  // Handler functions for scrolling
  const scrollUp = () => {
    setScrollOffset(Math.max(0, scrollOffset - 1));
  };
  
  const scrollDown = () => {
    setScrollOffset(Math.min(maxScroll, scrollOffset + 1));
  };
  
  // Get the slice of children to display
  const visibleChildren = childrenArray.slice(
    scrollOffset,
    scrollOffset + height
  );
  
  // Check if scroll controls should be shown
  const showControls = childrenArray.length > height;
  
  return (
    <Box flexDirection="row">
      {/* Main content area */}
      <Box flexDirection="column" flexGrow={1}>
        {visibleChildren}
      </Box>
      
      {/* Scroll controls */}
      {showControls && (
        <Box flexDirection="column" marginLeft={1}>
          <Button 
            label="▲" 
            onClick={scrollUp} 
            type="secondary"
            disabled={scrollOffset === 0}
          />
          <Box marginY={1}>
            <Text color={theme.dimText}>
              {scrollOffset + 1}-{Math.min(scrollOffset + height, childrenArray.length)}/{childrenArray.length}
            </Text>
          </Box>
          <Button 
            label="▼" 
            onClick={scrollDown} 
            type="secondary"
            disabled={scrollOffset >= maxScroll}
          />
        </Box>
      )}
    </Box>
  );
} 