import { useCallback, useEffect, useMemo, type RefObject } from 'react';
import { type DOMElement } from 'ink';

import { useMouse } from '@zenobius/ink-mouse';
import { useElementDimensions, useElementPosition } from '@zenobius/ink-mouse';
type MousePosition = {
  x: number;
  y: number;
};

type MouseClickAction = 'press' | 'release' | null;
type MouseScrollAction = 'scrollup' | 'scrolldown' | null;
type MouseDragAction = 'dragging' | null;

type MouseAction = MouseClickAction | MouseScrollAction;

function isIntersecting({
  mouse: { x, y },
  element,
}: {
  mouse: MousePosition;
  element: { left: number; top: number; width: number; height: number };
}) {
  /**
   * for some reason the position is off by 13 and 2
   */
  const left = element.left;
  const top = element.top;
  const width = element.width;
  const height = element.height;
  const isOutsideHorizontally = x < left || x > left + width;
  const isOutsideVertically = y < top || y > top + height;

  return !isOutsideHorizontally && !isOutsideVertically;
}
export { isIntersecting };

function useOnMouseClick(
  ref: RefObject<DOMElement>,
  onChange: (event: boolean) => void,
) {
  const mouse = useMouse();
  const elementPosition = useElementPosition(ref);
  const elementDimensions = useElementDimensions(ref);
  const element = useMemo(() => {
    return {
      ...elementPosition,
      ...elementDimensions,
    }
  }, [
    elementPosition,
    elementDimensions
  ])

  const handler = useCallback(
    (position: MousePosition, action: MouseClickAction) => {
      onChange(
        isIntersecting({ element, mouse: position }) && action === 'press',
      );
    },
    [ref.current, onChange],
  );

  useEffect(
    function HandleIntersection() {
      const events = mouse.events;

      events.on('click', handler);
      return () => {
        events.off('click', handler);
      };
    },
    [ref.current, handler],
  );
}

export { useOnMouseClick };
