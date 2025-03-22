import React, { useRef } from 'react';
import { Text, Box } from 'ink';
import { useMousePosition, useElementPosition, useElementDimensions } from '@zenobius/ink-mouse';
import { getTheme } from '../utils/theme.js';
import { useAppSelector } from '../store/hooks.js';
import { ChatMessage } from '../store/chatSlice.js';
import { ScrollArea } from './ScrollArea.js';
type MouseTrackerProps = {
  scrollAreaRef: React.RefObject<any>;
};

export function MouseTracker({ scrollAreaRef }: MouseTrackerProps): JSX.Element {
  const mousePosition = useMousePosition();
  const componentRef = useRef(null);
  const scrollAreaPosition = useElementPosition(scrollAreaRef);
  const scrollAreaDimensions = useElementDimensions(scrollAreaRef);
  const theme = getTheme();
  
  // Get the messages from the store to access text content
  const messages = useAppSelector(state => state.chat.messages) as ChatMessage[];
  const scrollState = useAppSelector(state => state.scroll);
  
  // Calculate relative mouse position within ScrollArea
  const relativeX = mousePosition.x - scrollAreaPosition.left;
  const relativeY = mousePosition.y - scrollAreaPosition.top;

  const foo = scrollAreaRef.current;
  
  // Get text under cursor
  let textUnderCursor = '';
  if (
    relativeX >= 0 && 
    relativeY >= 0 && 
    relativeX < scrollAreaDimensions.width && 
    relativeY < scrollAreaDimensions.height
  ) {
    // Calculate which message is under the cursor
    // Each message might take multiple lines
    // For simplicity, we'll assume each message is one line for now
    
    // Adjust for scroll offset
    const cursorRow = relativeY + scrollState.offset;
    
    // Find the message that contains this row
    let currentRow = 0;
    let targetMessage: ChatMessage | null = null;
    
    for (const message of messages) {
      // For a more accurate implementation, calculate message height based on content
      // Here we assume each message takes 1 row
      const messageHeight = 1;
      
      if (cursorRow >= currentRow && cursorRow < currentRow + messageHeight) {
        targetMessage = message;
        break;
      }
      
      currentRow += messageHeight;
    }
    
    if (targetMessage && typeof targetMessage.content === 'string') {
      // Estimate character position based on X coordinate
      // Assuming monospace font with 1 character = 1 horizontal unit
      const charPos = Math.max(0, Math.floor(relativeX));
      const content = targetMessage.content;
      
      if (charPos < content.length) {
        textUnderCursor = content.substring(charPos, charPos + 10);
      }
    }

  }

  if (foo) {
    foo.childNodes.forEach((child: any) => {
      console.log(JSON.stringify(child, ['childNodes', 'parentNode'], 2));
    });
  }
  
  return (
    <Box ref={componentRef} marginTop={1}>
      <Text color={theme.dimText}>
        Mouse: ({mousePosition.x}, {mousePosition.y}) | 
        MouseRelative: ({relativeX}, {relativeY}) | 
        ScrollArea: ({scrollAreaPosition.left}, {scrollAreaPosition.top}) | 
        Text: {textUnderCursor || '[no text]'} | 
        ScrollRef: {foo ? `${JSON.stringify(foo, null, 2)}` : 'undefined'}
      </Text>
    </Box>
  );
} 