import React, { useEffect } from 'react';
import { Box, Text } from 'ink';
import { useOnMouseClick } from '@zenobius/ink-mouse';
import { useAppDispatch, useAppSelector } from '../store/hooks.js';
import { setOffset, setDimensions } from '../store/scrollSlice.js';
import { getTheme } from '../utils/theme.js';
import { Button } from './Button.js';

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

  // Calculate maximum scroll offset
  const maxScroll = Math.max(0, childrenArray.length - height);
  
  // Handler functions for scrolling
  const scrollUp = () => {
    dispatch(setOffset(Math.max(0, offset - 1)));
  };
  
  const scrollDown = () => {
    dispatch(setOffset(Math.min(maxScroll, offset + 1)));
  };

  // Calculate visible children
  const visibleChildren = childrenArray.slice(
    offset,
    offset + height
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
            disabled={offset === 0}
          />
          <Box marginY={1}>
            <Text color={theme.dimText}>
              {offset + 1}-{Math.min(offset + height, childrenArray.length)}/{childrenArray.length}
              {isAutoScrollEnabled ? ' (A)' : ''}
            </Text>
          </Box>
          <Button 
            label="▼" 
            onClick={scrollDown} 
            type="secondary"
            disabled={offset >= maxScroll}
          />
        </Box>
      )}
    </Box>
  );
} 