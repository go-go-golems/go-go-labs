import React, { useEffect } from 'react';
import { Box, Text } from 'ink';
import { useAppDispatch, useAppSelector } from '../store/hooks.js';
import { setOffset, setDimensions } from '../store/scrollSlice.js';
import { getTheme } from '../utils/theme.js';

type Props = {
  height: number;
  children: React.ReactNode;
};

export function ScrollableBox({ height, children }: Props): JSX.Element {
  const dispatch = useAppDispatch();
  const { offset, contentHeight, isAutoScrollEnabled } = useAppSelector(state => state.scroll);
  const theme = getTheme();
  
  // Convert children to array for easier handling
  const childrenArray = React.Children.toArray(children);
  
  // Update dimensions whenever children or height changes
  useEffect(() => {
    dispatch(setDimensions({ 
      height, 
      contentHeight: childrenArray.length 
    }));
  }, [dispatch, height, childrenArray.length]);

  // Calculate visible children
  const visibleChildren = childrenArray.slice(
    offset,
    offset + height
  );

  // Calculate scroll indicators
  const canScrollUp = offset > 0;
  const canScrollDown = offset + height < childrenArray.length;
  
  return (
    <Box flexDirection="column" height={height}>
      {/* Scroll position indicator */}
      <Box>
        <Text color={theme.dimText}>
          {offset + 1}-{Math.min(offset + height, childrenArray.length)}/{childrenArray.length}
          {isAutoScrollEnabled ? ' (Auto-scroll)' : ''}
        </Text>
      </Box>

      {/* Main content area */}
      <Box 
        flexDirection="column" 
        height={height - 1} // Account for the indicator line
        overflow="hidden"
      >
        {visibleChildren}
      </Box>

      {/* Scroll indicators */}
      <Box position="absolute" flexDirection="column" marginLeft={1}>
        {canScrollUp && (
          <Text color={theme.dimText}>↑</Text>
        )}
        {canScrollDown && (
          <Text color={theme.dimText}>↓</Text>
        )}
      </Box>
    </Box>
  );
} 