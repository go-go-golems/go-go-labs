declare module 'wrap-ansi' {
  interface Options {
    trim?: boolean;
    hard?: boolean;
    wordWrap?: boolean;
  }
  
  export default function wrapAnsi(
    input: string,
    columns: number,
    options?: Options
  ): string;
}

declare module 'cli-truncate' {
  interface Options {
    position?: 'start' | 'middle' | 'end';
    preferTruncationOnSpace?: boolean;
    truncationCharacter?: string;
    space?: number;
  }
  
  export default function cliTruncate(
    input: string,
    columns: number,
    options?: Options
  ): string;
}

declare module 'widest-line' {
  export default function widestLine(input: string): number;
}

declare module 'yoga-layout' {
  export const EDGE_LEFT: number;
  export const EDGE_TOP: number;
  export const EDGE_RIGHT: number;
  export const EDGE_BOTTOM: number;
  export const DISPLAY_NONE: number;
  
  export interface Node {
    getComputedWidth(): number;
    getComputedHeight(): number;
    getComputedLeft(): number;
    getComputedTop(): number;
    getComputedPadding(edge: number): number;
    getComputedBorder(edge: number): number;
    getDisplay(): number;
    getComputedLayout(): {
      left: number;
      top: number;
      width: number;
      height: number;
      [key: string]: any;
    };
  }
  
  const Yoga: {
    EDGE_LEFT: number;
    EDGE_TOP: number;
    EDGE_RIGHT: number;
    EDGE_BOTTOM: number;
    DISPLAY_NONE: number;
    // Add other yoga constants as needed
  };
  
  export default Yoga;
} 