import React, { useRef, useState } from 'react';
import { Box, Text } from 'ink';
import { useOnMouseClick, useOnMouseHover } from '@zenobius/ink-mouse';
import { getTheme } from '../utils/theme.js';

type Props = {
  label: string;
  onClick: () => void;
  type?: 'primary' | 'secondary' | 'danger' | 'success';
  disabled?: boolean;
};

export function Button({ 
  label, 
  onClick, 
  type = 'primary', 
  disabled = false 
}: Props): JSX.Element {
  const ref = useRef(null);
  const [hovering, setHovering] = useState(false);
  const [clicking, setClicking] = useState(false);
  const theme = getTheme();
  
  // Get the color based on the button type
  const getColor = () => {
    if (disabled) return theme.dimText;
    
    switch (type) {
      case 'primary': return theme.primary;
      case 'secondary': return theme.secondary;
      case 'danger': return theme.error;
      case 'success': return theme.accent;
      default: return theme.primary;
    }
  };
  
  // Handle hover events
  useOnMouseHover(ref, (isHovering) => {
    if (!disabled) {
      setHovering(!!isHovering);
    }
  });
  
  // Handle click events
  useOnMouseClick(ref, (event) => {
    if (disabled || !event) return;
    
    setClicking(true);
    onClick();
    
    // Reset clicking state after a brief delay
    setTimeout(() => {
      setClicking(false);
    }, 150);
  });
  
  // Determine border style based on state
  const borderStyle = clicking ? 'double' : (hovering ? 'round' : 'single');
  const color = getColor();
  
  return (
    <Box 
      ref={ref}
      borderStyle={borderStyle} 
      borderColor={color}
      paddingX={2}
      paddingY={0}
      marginX={1}
    >
      <Text color={color}>
        {label}
      </Text>
    </Box>
  );
} 