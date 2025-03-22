import React, { useRef, useEffect } from 'react';
import { Text, Box, render } from 'ink';
import { useMousePosition, useElementPosition, useElementDimensions } from '@zenobius/ink-mouse';
import { getTheme, createLogger } from '../utils/index.js';
import { useAppSelector } from '../store/hooks.js';
import { ChatMessage } from '../store/chatSlice.js';
import { ScrollArea } from './ScrollArea.js';

// Create a logger for this component
const logger = createLogger('MouseTracker');

type MouseTrackerProps = {
  scrollAreaRef: React.RefObject<any>;
};

// Helper function to log yoga node information
const logYogaNodeInfo = (element: any, elementName: string) => {
  if (!element) return;
  
  logger.debug(`${elementName} basic properties`, {
    toString: element.toString(),
    type: element.type,
    props: element.props,
    key: element.key,
    ref: element.ref,
    children: element.children,
    parent: element.parent,
    nodeName: element.nodeName,
    nodeType: element.nodeType,
    attributes: Object.keys(element),
  });

  // Log yoga node if it exists
  if (element.yogaNode) {
    try {
      const layout = element.yogaNode.getComputedLayout();
      logger.debug(`${elementName} yoga layout`, {
        layout,
        width: element.yogaNode.getComputedWidth(),
        height: element.yogaNode.getComputedHeight(),
        left: element.yogaNode.getComputedLeft(),
        top: element.yogaNode.getComputedTop()
      });
      
      // Log yoga node methods and attributes
      logger.debug(`${elementName} yoga node details`, {
        yogaNodeAttrs: Object.keys(element.yogaNode),
        yogaNodeMethods: Object.getOwnPropertyNames(Object.getPrototypeOf(element.yogaNode)),
      });
    } catch (err) {
      logger.error(`Error getting ${elementName} yoga layout`, { error: err });
    }
  }
  
  // Log text content if available
  if (element.textContent) {
    logger.debug(`${elementName} text content`, { text: element.textContent });
  }
  
  // Log child nodes information
  if (element.childNodes && element.childNodes.length > 0) {
    logger.debug(`${elementName} child nodes`, {
      childrenCount: element.childNodes.length,
      childrenTexts: element.childNodes.map((n: any) => {
        const nodeInfo = {
          text: n.textContent || n.toString(),
          attributes: Object.keys(n),
          methods: Object.getOwnPropertyNames(Object.getPrototypeOf(n)),
          type: typeof n,
          constructor: n.constructor?.name,
          children: []
        };

        // Recursively get info for child yoga nodes
        if (n.childNodes && n.childNodes.length > 0) {
          nodeInfo.children = n.childNodes.map((child: any) => ({
            text: child.textContent || child.toString(),
            attributes: Object.keys(child),
            methods: Object.getOwnPropertyNames(Object.getPrototypeOf(child)), 
            type: typeof child,
            constructor: child.constructor?.name,
            yogaLayout: child.yogaNode ? {
              width: child.yogaNode.getComputedWidth(),
              height: child.yogaNode.getComputedHeight(),
              left: child.yogaNode.getComputedLeft(),
              top: child.yogaNode.getComputedTop()
            } : null
          }));
        }

        return nodeInfo;
      })
    });
  }
};

// Helper function to recursively extract text content from nodes
const extractTextContent = (node: any): string => {
  if (!node) return '';
  
  // If the node has a direct text value
  if (node.nodeValue !== undefined && node.nodeValue !== null) {
    return node.nodeValue;
  }
  
  // If the node has textContent property
  if (node.textContent) {
    return node.textContent;
  }
  
  // If the node has children or childNodes, recursively extract text
  let text = '';
  
  // Try childNodes first (DOM-like structure)
  if (node.childNodes && node.childNodes.length > 0) {
    for (const child of node.childNodes) {
      text += extractTextContent(child);
    }
  }
  
  // Try React children if available
  if (node.children && Array.isArray(node.children)) {
    for (const child of node.children) {
      text += extractTextContent(child);
    }
  }
  
  return text;
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

  // Component string representation
  let scrollAreaText = '';
  
  // if (foo) {
  //   // Extract text content from the scroll area
  //   scrollAreaText = extractTextContent(foo);
    
  //   // Log information about the main scroll area element
  //   logYogaNodeInfo(foo, 'ScrollArea');
    
  //   // Try to access render instances to examine them
  //   logger.debug('Ink render instances', {
  //     hasRender: typeof render === 'function',
  //     renderType: typeof render
  //   });
    
  //   // Log information about each child node
  //   foo.childNodes.forEach((child: any, index: number) => {
  //     logYogaNodeInfo(child, `ScrollArea.child[${index}]`);
  //     // Log individual node text content for debugging
  //     logger.debug(`ScrollArea.child[${index}] text content`, {
  //       extractedText: extractTextContent(child)
  //     });
  //   });
  // }
  
  return (
    <Box ref={componentRef} marginTop={1}>
      <Text color={theme.dimText}>
        Mouse: ({mousePosition.x}, {mousePosition.y}) | 
        MouseRelative: ({relativeX}, {relativeY}) | 
        ScrollArea: ({scrollAreaPosition.left}, {scrollAreaPosition.top}) | 
        Text: {textUnderCursor || '[no text]'} | 
        ComponentText: {scrollAreaText.substring(0, 20) + (scrollAreaText.length > 20 ? '...' : '')}
      </Text>
    </Box>
  );
} 