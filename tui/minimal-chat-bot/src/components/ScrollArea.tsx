import React, { useRef, useEffect } from 'react';
import { Box, useInput, measureElement } from 'ink';
import { useAppDispatch, useAppSelector } from '../store/hooks.js';
import { setOffset, setDimensions } from '../store/scrollSlice.js';

type Props = {
  height: number;
  children: React.ReactNode;
};

export function ScrollArea({ height, children }: Props): JSX.Element {
  const dispatch = useAppDispatch();
  const { offset, contentHeight } = useAppSelector(state => state.scroll);
  const innerRef = useRef<any>();

  // Update dimensions when content changes
  useEffect(() => {
    if (innerRef.current) {
      const dimensions = measureElement(innerRef.current);
      dispatch(setDimensions({
        height,
        contentHeight: dimensions.height
      }));
    }
  }, [dispatch, height, children]);

  // Handle keyboard input
  useInput((_input, key) => {
    if (key.downArrow) {
      dispatch(setOffset(Math.min(
        contentHeight - height,
        offset + 1
      )));
    }

    if (key.upArrow) {
      dispatch(setOffset(Math.max(0, offset - 1)));
    }

    if (key.pageDown) {
      dispatch(setOffset(Math.min(
        contentHeight - height,
        offset + height
      )));
    }

    if (key.pageUp) {
      dispatch(setOffset(Math.max(0, offset - height)));
    }
  });

  return (
    <Box
     height={height}
    //  borderStyle="single"
    //   borderColor="gray"
       flexDirection="column"
       margin={0}
       padding={0}
        overflow="hidden">
      <Box
        ref={innerRef}
        flexShrink={0}
        flexDirection="column"
        marginTop={-offset}
      >
        {children}
      </Box>
    </Box>
  );
} 