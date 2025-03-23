import React, { FC, useCallback, useRef, useState, useEffect, useMemo } from 'react';
import { Box, Text, measureElement } from 'ink';
import { useOnMouseHover, useMousePosition, useElementPosition, useElementDimensions, useMouse } from '@zenobius/ink-mouse';
import { getElementDimensions, getElementPosition, useOnMouseClick } from './useOnMouseClick.js';
import type { ChatMessage as ChatMessageType } from '../store/chatSlice.ts';
import { getTheme } from '../utils/theme.js';
import { createLogger } from '../utils/logger.js';
import { wrapComponentText } from '../utils/text-wrapper.js';

export interface ChatMessageClickEvent {
  wrappedContent: string;
  position: {
    absolute: { x: number; y: number };
    relative: { x: number; y: number };
  };
  clickedLine: {
    content: string;
    index: number;
    hasPreamble: boolean;
  };
  textAtPosition: {
    raw: string;
    adjusted: string;
    charPosition: number;
  };
}

interface Props {
  message: ChatMessageType;
  onClick?: (event: ChatMessageClickEvent) => void;
  isLoading?: boolean;
  showPrefix?: boolean;
}

// Create a logger for this component
const logger = createLogger('ChatMessage');

export const ChatMessage: FC<Props> = ({ 
  message, 
  onClick, 
  isLoading = false, 
  showPrefix = true 
}) => {
  const { content, role } = message;
  const ref = useRef(null);
  const [hovering, setHovering] = useState(false);
  const [clicking, setClicking] = useState(false);
  const [wrappedContent, setWrappedContent] = useState('');
  const [colorIndex, setColorIndex] = useState(0);
  const theme = getTheme();
  
  // Array of colors for glitter effect
  const glitterColors = ['blue', 'cyan', 'magenta', 'green', 'yellow'];
  
  // Update color index for glitter effect when loading
  useEffect(() => {
    let intervalId: NodeJS.Timeout;
    
    if (isLoading) {
      intervalId = setInterval(() => {
        setColorIndex((prevIndex) => (prevIndex + 1) % glitterColors.length);
      }, 150); // Change color every 150ms
    }
    
    return () => {
      if (intervalId) clearInterval(intervalId);
    };
  }, [isLoading]);
  
  // Get mouse position and element dimensions for detailed logging
  const mouse = useMouse();
  const mousePosition = useMousePosition();
  
  // Create the complete text with emoji preamble
  const fullText = useMemo(() => {
    if (!showPrefix) return content;
    const preamble = role === 'user' ? '🧑 You: ' : '🤖 Bot: ';
    return preamble + content;
  }, [role, content, showPrefix]);
  
  // Update wrapped content when element dimensions or content changes
  useEffect(() => {
    // Delay the wrapping calculation until after the initial render
    // so that the element measurements are available
    const timeoutId = setTimeout(() => {
      const wrapped = wrapComponentText(fullText, ref);
      setWrappedContent(wrapped);
      
      if (wrapped !== fullText) {
        logger.debug('Text wrapped', {
          originalText: fullText,
          originalLength: fullText.length,
          wrappedLength: wrapped.length,
          originalFirstLine: fullText.split('\n')[0],
          wrappedFirstLine: wrapped.split('\n')[0],
          lineCount: wrapped.split('\n').length
        });
      }
    }, 600);
    
    return () => clearTimeout(timeoutId);
  }, [fullText, ref.current]);

  const onSetHovering = useCallback((isHovering: boolean) => {
    // logger.info('Hovering', { isHovering,
    //   beginningText: wrappedContent.split('\n')[0],
    //   wrappedContent,
    //  });
    setHovering(isHovering);
  }, [wrappedContent]);
  
  // Handle hover events
  useOnMouseHover(ref, onSetHovering);

  const handler = useCallback((isClicked: boolean) => {
    setClicking(isClicked);

    const elementPosition = getElementPosition(ref.current);
    const elementDimensions = getElementDimensions(ref.current);
    if (!elementPosition || !elementDimensions) {
      return;
    }

    if (isClicked) {
      // Calculate relative mouse position within the component
      const relativeX = mousePosition.x - elementPosition.left - 3;
      const relativeY = mousePosition.y - elementPosition.top - 2;
      
      // Determine which line was clicked
      const contentLines = wrappedContent.split('\n');
      const clickedLineIndex = Math.min(Math.max(0, relativeY), contentLines.length - 1);
      const clickedLine = contentLines[clickedLineIndex] || '';
      
      // Estimate character position based on X coordinate
      const charPos = Math.max(0, Math.floor(relativeX));
      
      // Get character under cursor if possible
      let textUnderCursor = '';
      if (charPos < clickedLine.length) {
        textUnderCursor = clickedLine.substring(charPos, charPos + 10) + 
          (clickedLine.length > charPos + 10 ? '...' : '');
      }

      // Extract the content part of the wrapped text (removing preamble)
      const preambleLength = showPrefix ? (role === 'user' ? '🧑 You: ' : '🤖 Bot: ').length : 0;
      const hasPreamble = showPrefix && (clickedLine.startsWith('🧑') || clickedLine.startsWith('🤖'));
      const contentStart = hasPreamble ? preambleLength : 0;
        

      // Create click event object
      const clickEvent: ChatMessageClickEvent = {
        wrappedContent,
        position: {
          absolute: { x: mousePosition.x, y: mousePosition.y },
          relative: { x: relativeX, y: relativeY }
        },
        clickedLine: {
          content: clickedLine,
          index: clickedLineIndex,
          hasPreamble
        },
        textAtPosition: {
          raw: textUnderCursor || '',
          adjusted: textUnderCursor.slice(contentStart) || '',
          charPosition: charPos,
        }
      };

      // Log detailed information about the click
      logger.info(`Message clicked: ${role}`, {
        // ...clickEvent,
        position: {
          absolute: { x: mousePosition.x, y: mousePosition.y },
          relative: { x: relativeX, y: relativeY }
        },
        role,
        elementPos: { 
          left: elementPosition.left, 
          top: elementPosition.top,
          right: elementPosition.left + elementDimensions.width,
          bottom: elementPosition.top + elementDimensions.height
        },
        elementDim: { 
          width: elementDimensions.width, 
          height: elementDimensions.height 
        },
        inkMeasure: ref.current ? measureElement(ref.current) : null,
        textUnderCursor: textUnderCursor,

      });
      
      // Call the onClick handler if provided
      if (onClick) {
        onClick(clickEvent);
      }
    }
  }, [mousePosition, wrappedContent, fullText, role, onClick, showPrefix]);
  
  // Handle click events with detailed logging
  useOnMouseClick(ref, handler);
  
  // Determine border style based on state
  const borderStyle = onClick ? (clicking ? 'double' : (hovering ? 'round' : 'single')) : undefined;
  const roleColor = role === 'user' ? theme.accent : theme.primary;
  
  // Determine text color based on role and loading state
  const getTextColor = () => {
    if (!isLoading) {
      return role === 'user' ? 'green' : 'blue';
    }
    return glitterColors[colorIndex];
  };
  
  return (
    <Box 
      ref={ref}
      borderStyle={borderStyle}
      borderColor={hovering ? roleColor : undefined}
      paddingX={1}
      paddingY={0}
      marginY={0}
    >
      <Text color={getTextColor()}>
        {showPrefix ? (role === 'user' ? '🧑 You: ' : '🤖 Bot: ') : ''}
        {content}
      </Text>
    </Box>
  );
}; 