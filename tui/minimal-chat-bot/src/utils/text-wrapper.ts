import wrapAnsi from 'wrap-ansi';
import cliTruncate from 'cli-truncate';
import widestLine from 'widest-line';
import Yoga from 'yoga-layout';
import { createLogger } from './logger.js';

// Cache for wrapped text to avoid redundant calculations
const cache: Record<string, string> = {};

type WrapType = 'wrap' | 'truncate' | 'truncate-start' | 'truncate-middle' | 'truncate-end';

const logger = createLogger('TextWrapper');

/**
 * Wrap or truncate text to fit within a maximum width
 */
export function wrapText(
  text: string,
  maxWidth: number,
  wrapType: WrapType = 'wrap',
): string {
  const cacheKey = text + String(maxWidth) + String(wrapType);
  const cachedText = cache[cacheKey];

  if (cachedText) {
    return cachedText;
  }

  let wrappedText = text;

  if (wrapType === 'wrap') {
    wrappedText = wrapAnsi(text, maxWidth, {
      trim: false,
      hard: true,
    });
  }

  if (wrapType.startsWith('truncate')) {
    let position: 'end' | 'middle' | 'start' = 'end';

    if (wrapType === 'truncate-middle') {
      position = 'middle';
    }

    if (wrapType === 'truncate-start') {
      position = 'start';
    }

    wrappedText = cliTruncate(text, maxWidth, { position });
  }

  cache[cacheKey] = wrappedText;

  return wrappedText;
}

/**
 * Calculate the max width available for text content within a yoga node
 */
export function getMaxWidth(yogaNode: any): number {
  if (!yogaNode) return 0;
  
  try {
    return (
      yogaNode.getComputedWidth() -
      yogaNode.getComputedPadding(Yoga.EDGE_LEFT) -
      yogaNode.getComputedPadding(Yoga.EDGE_RIGHT) -
      yogaNode.getComputedBorder(Yoga.EDGE_LEFT) -
      yogaNode.getComputedBorder(Yoga.EDGE_RIGHT)
    );
  } catch (error) {
    // If yoga node methods are not available, return a sensible default
    return 80;
  }
}

/**
 * Wrapper for component content that performs text wrapping based on element dimensions
 */
export function wrapComponentText(
  text: string,
  elementRef: React.RefObject<any>,
  fallbackWidth: number = 80,
  wrapType: WrapType = 'wrap'
): string {
  if (!text) return '';
  
  // Use Ink's element ref to determine available width if possible
  let maxWidth = fallbackWidth;
  
  if (elementRef?.current) {
    try {
      // Attempt to get yoga node from the element
      const yogaNode = elementRef.current.yogaNode;
      
      if (yogaNode) {
        maxWidth = getMaxWidth(yogaNode);
      } else if (typeof elementRef.current.getComputedWidth === 'function') {
        // If element has direct computed width methods
        maxWidth = elementRef.current.getComputedWidth();
      }
    } catch (error) {
      // Fallback to a reasonable default if we can't access the yoga node
      maxWidth = fallbackWidth;
    }
  }
  
  // Get the width of the widest line in the text
  const currentWidth = widestLine(text);

  logger.info('Current width', { currentWidth })
      logger.info('Max width', { maxWidth })
  
  // Only wrap if text exceeds available width
  if (currentWidth > maxWidth) {
    const wrapped = wrapText(text, maxWidth, wrapType);
    logger.info('Wrapped', { wrapped })
    return wrapped;
  }
  
  return text;
} 