import React, { useReducer, useRef, useEffect } from 'react';
import { Box, useInput, measureElement } from 'ink';

type ScrollState = {
  height: number;
  innerHeight: number;
  scrollTop: number;
};

type ScrollAction = 
  | { type: 'SET_INNER_HEIGHT'; innerHeight: number }
  | { type: 'SCROLL_DOWN' }
  | { type: 'SCROLL_UP' };

const reducer = (state: ScrollState, action: ScrollAction): ScrollState => {
  switch (action.type) {
    case 'SET_INNER_HEIGHT':
      return {
        ...state,
        innerHeight: action.innerHeight
      };

    case 'SCROLL_DOWN':
      return {
        ...state,
        scrollTop: Math.min(
          state.innerHeight - state.height,
          state.scrollTop + 1
        )
      };

    case 'SCROLL_UP':
      return {
        ...state,
        scrollTop: Math.max(0, state.scrollTop - 1)
      };

    default:
      return state;
  }
};

type Props = {
  height: number;
  children: React.ReactNode;
};

export function ScrollArea({ height, children }: Props): JSX.Element {
  const [state, dispatch] = useReducer(reducer, {
    height,
    innerHeight: 0,
    scrollTop: 0
  });

  const innerRef = useRef<any>();

  useEffect(() => {
    if (innerRef.current) {
      const dimensions = measureElement(innerRef.current);
      dispatch({
        type: 'SET_INNER_HEIGHT',
        innerHeight: dimensions.height
      });
    }
  }, [children]);

  useInput((_input, key) => {
    if (key.downArrow) {
      dispatch({
        type: 'SCROLL_DOWN'
      });
    }

    if (key.upArrow) {
      dispatch({
        type: 'SCROLL_UP'
      });
    }
  });

  return (
    <Box height={height} borderStyle="single" borderColor="gray" flexDirection="column" overflow="hidden">
      <Box
        ref={innerRef}
        flexShrink={0}
        flexDirection="column"
        marginTop={-state.scrollTop}
      >
        {children}
      </Box>
    </Box>
  );
} 